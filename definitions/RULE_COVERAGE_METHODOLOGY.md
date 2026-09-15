# Rule Coverage & Verification Methodology

How we validate that a platform's detection rules (`definitions/rules/<platform>/`)
and filters (`definitions/filters/<platform>/`) are actually sufficient, instead of
just trusting whatever already exists. Written after doing this for macOS (see
`macos-logs-client` in the `UTMStackEnterprise` repo, commit `c4067a3` and
follow-ups) and applying the same process to Linux next.

## Why this exists

Two things kept being true every time we looked closely at a platform's pipeline:

1. **Our own agent was shipping mostly noise.** Not "too many logs" in the
   abstract -- specific, identifiable internal housekeeping (OS/daemon plumbing,
   connection teardown, routine self-checks) that no rule could ever use, mixed
   in with the handful of events that actually mattered. On macOS this was a
   ~99.84% reduction (287K events/hour -> ~465/hour on one endpoint) without
   losing real signal, verified against production ClickHouse data at every
   step, not estimated.
2. **We had no independent way to know if our rule set was *complete*.** Our own
   judgment about "this covers the important techniques" isn't evidence of
   anything -- it's just our judgment. Comparing against what mature, widely
   deployed open-source projects already detect is a much stronger bar than "we
   thought of everything."

Both problems compound: a rule that's technically correct is worthless if the
agent never ships the field it needs, or if the filter drops the log before it
reaches the field a rule reads.

## The three parallel workstreams

For a given platform, work these together, not sequentially -- a fix in one
often depends on what you find in another:

1. **Agent-side filtering** (`agent/collector/platform/<platform>_*.go`) --
   ship logs that describe something a security analyst would act on, not
   general system telemetry. Positive-gate by content where possible (require
   a specific keyword/verb) rather than blocklisting individual noisy messages
   one at a time -- blocklisting is a losing game against free-text logs; new
   phrasings keep surfacing. Validate every exclusion against real, current
   production log volume before applying it (see "Verifying a filter change"
   below) -- never exclude based on what a message *sounds like* it might be.
2. **Server-side parsing** (`definitions/filters/<platform>/*.yaml`) -- the
   raw event has to actually land in the field names a rule's `where` clause
   reads. A rule can be perfectly reasonable and still never fire because the
   filter renamed or never mapped the field it depends on.
3. **Attack simulation + verification** -- for a given rule (existing or new),
   actually perform the technique it claims to detect, on a real test host,
   and confirm: the raw log is generated -> the agent ships it -> the filter
   parses it into the expected fields -> the rule's `where` clause matches ->
   an alert fires. Any one of those four links can silently break independent
   of the other three. "The rule logic looks right" is not verification.

## Reference materials for coverage benchmarking

Cloned locally at `/Users/yorjanderhernandez/code/UTMStack/Rules-Resources/`:

- **`sigma/`** ([SigmaHQ/sigma](https://github.com/SigmaHQ/sigma)) -- the
  portable, vendor-agnostic detection rule standard (3,000+ community rules,
  MITRE ATT&CK-mapped). Relevant subtree: `rules/linux/` (auditd, syslog,
  builtin categories). Best single source for "what do other systems already
  detect" since it's explicitly designed to be converted into other SIEMs'
  query languages.
- **`wazuh-ruleset/`** ([wazuh/wazuh-ruleset](https://github.com/wazuh/wazuh-ruleset))
  -- Wazuh's own built-in rules. Most directly comparable to our own
  architecture (declarative rules over structured log fields, not a query
  language), and has the deepest Linux/auditd-specific coverage of the major
  open-source SIEMs.
- **`atomic-red-team/`** ([redcanaryco/atomic-red-team](https://github.com/redcanaryco/atomic-red-team))
  -- not rules, **attack simulation scripts** organized by MITRE ATT&CK
  technique ID (e.g. `atomics/T1548.003/`). This is the actual attack-execution
  tool for workstream 3 -- run the atomic test for a technique instead of
  improvising an attack by hand, so the test is reproducible and its ATT&CK
  mapping is already correct.

**How to use them**: don't port rule logic literally -- field names never
match (Sigma's `logsource`/`detection` fields, Wazuh's `<rule>` XML, and our
own `log.process`/`log.subsystem`/`log.eventMessage` schema are all shaped
differently). Extract the *techniques covered* (their MITRE ATT&CK tags) as
the comparison unit, diff that against the `technique:` field already present
in every rule under `definitions/rules/<platform>/`, and treat the resulting
gap list as the backlog. Then write the detection logic fresh, against
whatever our own agent/filter pipeline actually produces for that platform.

## Test infrastructure

- **`V11-DEV-18`** (`10.11.11.18`) -- has the UTMStack Linux agent installed
  and running (`/opt/utmstack-linux-agent/utmstack_agent_service_linux_amd64`).
  Also running the full v11 stack itself (event-processor plugins), so it
  doubles as both the SIEM under test and a monitored endpoint.
- The agent's `auditd` setup (`agent/dependency/auditd_linux.go`) is
  automatic on install: detects the distro, installs `auditd`, deploys
  `/etc/audit/rules.d/50-utmstack.rules`, starts and reloads the service. No
  manual audit configuration needed before running atomic tests -- confirmed
  live on V11-DEV-18 (`auditctl -l` matches the rule source exactly).

### What the Linux agent's current audit rules actually cover

Audit keys currently deployed (`agent/dependency/auditd_linux.go`,
`AuditdVersion` tracks the version -- bump it when this list changes):

| Key | Covers |
|---|---|
| `utmstack_exec` | every `execve` (no `auid` filter -- deliberately includes root and no-login-session processes: systemd services, cron, containers, web workers) |
| `utmstack_priv` | `setuid`/`setgid`/`setreuid`/`setregid`/`setresuid`/`setresgid` |
| `utmstack_sensitive` | `/etc/shadow`, `/etc/passwd`, `/etc/group`, `/etc/gshadow`, `/etc/sudoers(.d)`, `/etc/ssh/sshd_config`, `/root/.ssh` |
| `utmstack_persistence` | cron dirs/files, `/var/spool/cron`, `/etc/systemd/system`, `/etc/ld.so.preload`, `/etc/rc.local`, `/etc/environment`, `/etc/pam.d`, `/etc/hosts` |
| `utmstack_log_tampering` | `/var/log/wtmp`, `/var/log/btmp`, `/var/log/lastlog` |
| `utmstack_modules` | `init_module`/`finit_module`/`delete_module` |
| `utmstack_time` | `adjtimex`/`settimeofday`/`clock_settime`, `/etc/localtime` |
| `utmstack_audit_config` | `/etc/audit`, `/etc/audisp` |

**Known, deliberate gap**: no network telemetry from auditd. `connect()`
can't be filtered by destination in an audit rule (the address is behind a
pointer arg), so enabling it means one full audit record per socket connect
on the host -- DNS, health checks, everything. Real network visibility here
needs an eBPF-based flow monitor, not an audit rule; treat any ATT&CK
technique that's purely network-signature (not exec/file/priv-adjacent) as
out of scope for this pipeline until that separate work exists, not as a
rule-writing gap.

## Verifying a filter change (agent-side)

Same process used for macOS, repeat for Linux/Windows:

1. Pull real production log volume for the target host/dataSource from
   ClickHouse (`utmstack.logs`) over a representative window (an hour is
   usually enough to separate routine chatter from rare real signal).
2. Group by whatever the platform's equivalent of process/subsystem is, find
   the highest-volume groups, and sample actual message content for each --
   never assume from the field names alone what a group contains.
3. For each high-volume group: is this something a security analyst would
   ever act on, regardless of whether a rule reads it *today*? If yes, does a
   rule read it? If not yet, that's a gap for workstream 3, not a reason to
   exclude it.
4. Simulate the proposed exclusion as a SQL filter against the same
   ClickHouse data before touching any code, to get a real before/after
   number -- not an estimate.
5. Implement, then validate with a deterministic test suite covering every
   rule and fallback path (see `macos-logs-client/main.swift`'s
   `logictest.swift`-style harness for the pattern -- copy the pure
   matching/decision functions into a standalone test file, feed it synthetic
   cases per rule, assert expected outcomes). Rule *ordering* matters when
   rules can shadow each other (a broad terminal rule placed before a more
   specific one makes the specific one unreachable) -- trace this by hand for
   every rule addition, the test suite alone won't catch it if the synthetic
   cases don't happen to exercise that specific ordering.
6. Where possible, also validate live: run the collector against real system
   activity for a short window and inspect what actually gets shipped, not
   just the ClickHouse-based simulation. Live testing on macOS caught real
   gaps (bursty activity, e.g. iCloud sync chatter) that the hourly
   ClickHouse aggregate underrepresented.

## Status log

Keep this updated as work progresses -- newest entry first, one line per
platform per work session, with the concrete number (not "reduced noise").

- 2026-09-15 (Linux, session 3): Live-fire verified every remaining testable
  rule on V11-DEV-18 (Debian/Ubuntu). Final tally out of 70 total Linux rule
  files: 58 confirmed firing via real triggered commands + real alerts this
  session (all of top-level minus the 10 rhel_family/ ones, plus 15 of 17
  debian_family/ ones); 10 rhel_family/ rules cannot be tested here at all
  (wrong OS family -- no yum/dnf/rpm/SELinux/OpenShift on this host, user is
  installing a RHEL VM separately for that); 2 debian_family/ rules
  (ebpf_rootkit_detection.yaml, process_masquerading.yaml) remain documented
  dead-on-arrival pending real redesign/new telemetry, deliberately not
  live-fire tested since they're already known-broken by construction.
  Found and fixed 3 more real bugs via this pass (none were test-methodology
  artifacts):
  1. linux_nping_activity.yaml and linux_hping_activity.yaml used
     `deduplicateBy` (the only 2 of 70 files doing so) instead of `groupBy`.
     Confirmed on live data this silently drops repeat matches for the same
     adversary.ip/user with no new alert and no lastEvent/events-array
     update -- both fired correctly on a first-ever trigger, then a second
     real trigger ~5 min later against the same ip/user produced nothing at
     all (checked events array length: stuck at 1). Switched both to
     `groupBy` for consistency with the other 68 rules and reconfirmed a
     repeat trigger now creates a fresh alert.
  2. systemd_timer_persistence.yaml's entire premise was wrong: verified via
     direct `journalctl` inspection during a real `systemctl enable
     foo.timer` that systemd never journals the "Created symlink ...timer"
     confirmation text (that's CLI stdout only) or any ".timer"-scoped
     message at all -- the only journald activity is a generic
     "Reloading requested/Reloading/Reloading finished" trio with no unit
     name in it, so the rule's `contains(log.message,".timer") AND
     (contains("Created symlink") OR ...)` could never be satisfied by any
     real timer operation. Rewritten to detect the same intent directly from
     the execve (`systemctl enable/start *.timer`) instead of assumed-but-
     nonexistent journald text; redeployed and reconfirmed firing on a real
     enable/disable cycle.
  Two rules that looked "missing" on first pass turned out to be self-
  inflicted test artifacts once traced through lastEvent, not bugs: (a) the
  first nping/hping alerts of the day actually came from my own `which
  nping`/`which ... hping3 ...` tool-availability check (origin.command
  contains "nping"/"hping3" as a substring of the `which` invocation itself)
  -- same class of self-matching-filename artifact documented earlier this
  week; (b) a query-window off-by-a-few-seconds boundary, not a detection
  gap. Also applied the previously-flagged `auid`-over-`subj_user` filter
  improvement (definitions/filters/linux/linux.yaml) for a more useful
  origin.user; confirmed working for the large majority of events (real
  numeric UID instead of "unconfined") but left one open, uninvestigated
  inconsistency: ~45 of ~166 sampled events with a real non-"unset" auid
  still showed "unconfined" instead of the UID -- root cause not identified
  (suspect rule/filter evaluation-order nuance in the engine, not the YAML
  itself), flagged for a future session.
  All 70 rule files confirmed byte-identical between the local repo and the
  deployed copy on V11-DEV-18 (sha256sum comparison) after every edit this
  session.
- 2026-09-15 (Linux, session 2): Root-caused why almost nothing fired despite
  correct-looking `where` clauses: EventProcessor's `json` pipeline step runs
  every parsed field name through go-sdk's `SanitizeField()`
  (threatwinds/go-sdk/utils/fields.go), whose allow-list regex
  (`[^a-zA-Z0-9.]`) strips every underscore from TOP-LEVEL keys before they
  reach the `log` JSON column -- `_HOSTNAME` -> `log.HOSTNAME`,
  `_SYSTEMD_CGROUP` -> `log.SYSTEMDCGROUP`, auditd's `subj_user` ->
  `log.subjuser`. This is why `origin.host` was NEVER populated for any
  journald-sourced Linux event and `origin.user` was never populated from
  `subj_user` for auditd events, confirmed via live `log`-column inspection on
  V11-DEV-18. Nested object fields (execve.a0, userauth.acct) are NOT affected
  -- only top-level keys go through SanitizeField. Fixed the allow-list at the
  source (`[^a-zA-Z0-9_.]`, uncommitted local patch in the go-sdk repo -- not
  yet released in any EventProcessor image) AND reworked
  `definitions/filters/linux/linux.yaml`'s Phase 2/3/4 renames to check the
  already-stripped names so detection works against the currently-deployed
  image; comment block documents both bugs and what to revert once a fixed
  image ships. Verified live: `origin.host` now populates
  (`"v11-DEV-18"` for a journald `logger` event) and `origin.user` now
  populates (`"unconfined"`, the real subj_user value, for auditd execve
  events) -- both were unconditionally null before. Known residual gap:
  auditd-native records carry no hostname field at all (confirmed in raw
  parser.go output), so `origin.host`/`adversary.host` stay empty for the
  (more voluminous) auditd-sourced rules -- doesn't block alert firing
  (confirmed empty groupBy dimensions still alert) but does weaken
  host-based dedup; no fix shipped for this yet (would need a new
  copy-field-into-another-field pipeline capability -- `add` only writes
  literal constants, `rename` moves rather than copies, confirmed by reading
  EventProcessor's `add`/`rename` plugin source).
  Live-fire results, 20 new + 32 bulk-fixed existing rules: 7/8 of the
  first "safe discovery/collection" batch confirmed firing via real alerts
  (discovery_account_and_groups, discovery_network_config_and_connections,
  discovery_file_and_directory, discovery_installed_software,
  ingress_tool_transfer, archive_sensitive_data_before_exfil,
  hide_artifacts_hidden_file); the 8th (deobfuscate_decode_command,
  base64|bash) has a real logic bug -- auditd emits one execve per pipeline
  stage, so a piped `base64 -d | bash` never has "bash" inside base64's own
  origin.command, only a single-process `sh -c "...|bash"` compound
  invocation does. Split into two rules matching SigmaHQ's own split
  (proc_creation_lnx_base64_decode.yml level:low + ...execution.yml
  level:medium): kept deobfuscate_decode_command.yaml for the compound form,
  added base64_openssl_decode_executed.yaml for the standalone case. 3 of the
  20 new rules were already confirmed in the prior session
  (discovery_user_whoami_id, discovery_process_listing,
  discovery_system_service). Remaining 9 higher-risk new rules
  (discovery_remote_system_scan, network_sniffing_tool_executed,
  remote_access_software_executed, resource_hijacking_cryptominer,
  account_creation, account_access_removal, security_service_stopped,
  system_shutdown_or_reboot, disk_content_wipe) triggered via safe surrogates
  (localhost-only nmap/tcpdump, a placeholder ELF binary at the anydesk path
  -- a shebang-script placeholder does NOT work, see below --, a literal
  stratum+tcp string via /bin/echo, a real create+lock+delete cycle on a
  throwaway test account, a systemctl stop against a nonexistent decoy unit
  name containing "auditd", `shutdown --help`, `dd` to a nonexistent /dev
  path). All 9 confirmed firing via real alerts once the test methodology
  itself was debugged. Two were genuine test-artifact false negatives, not
  rule bugs:
  (1) remote_access_software_executed.yaml matches on origin.process exact
  path -- a `#!/bin/sh` placeholder at /usr/bin/anydesk gets shebang-resolved
  by the kernel into an execve of /bin/sh with the script path as argv[1], so
  origin.process becomes /usr/bin/dash, never /usr/bin/anydesk; a real ELF
  placeholder (`cp /bin/true /usr/bin/anydesk`) confirmed the rule fires
  correctly.
  (2) A same-minute alerts query used `> '...:41:00'` while the trigger event
  landed at `:40:59.3` -- excluded by one's own off-by-a-few-seconds query
  boundary, not a detection gap.
  One real, non-artifact, high-value bug found and fixed:
  system_shutdown_or_reboot.yaml's origin.process oneOf() against
  /usr/sbin/shutdown|reboot|halt|poweroff can never match on ANY modern
  systemd host (not V11-DEV-18-specific) -- those binaries are the same
  multi-call binary as systemctl, so the kernel-resolved `exe` (origin.process)
  is always /usr/bin/systemctl regardless of which name invoked it, confirmed
  via real data (`shutdown --help` landed as origin.process=/usr/bin/systemctl,
  execve.a0 still literally "/usr/sbin/shutdown"). Added an origin.command-text
  fallback (execve argv[0] still preserves the typed name) and a direct
  `systemctl reboot|poweroff|halt` verb branch; redeployed and reconfirmed
  firing (3 alert rows for one command -- the sudo wrapper's own execve and
  the resolved systemctl execve both carry the same inherited command text,
  so a text-based match legitimately re-fires at each hop of a 2-hop sudo
  chain; this is a small, bounded, already-common duplicate across many rules
  in this set, unlike the 10-25x sudo-privilege-transition-syscall flood
  fixed separately above, and not worth chasing further this session).
  Confirmed (again) that restarting utmstack_event-processor-worker breaks
  the agent's log-shipping stream and needs a separate `systemctl restart
  UTMStackAgent`; this time the worker also took ~90s to fully reinitialize
  all its plugins after a force-update (rule-flood-guard, soc-ai, sophos,
  etc. logged their startup sequentially) before log ingestion resumed --
  a same-minute verification query after redeploy showed 0 new events and
  was a false alarm, not a bug; the fix was just waiting longer.
  Also found and fixed a second systemic bug, independent of the field-name
  ones above: any rule matching on origin.command/origin.process WITHOUT
  gating on `equals("action","execve")` fires once per underlying command,
  not once per matching audit record -- when a command runs via `sudo`, the
  native auditd collector emits the real execve PLUS a cascade of sudo's own
  internal privilege-transition syscalls (setresuid/setresgid/setgid/etc,
  confirmed as high as ~25 records for a single `sudo systemctl stop <x>`),
  and every one of them inherits the SAME PROCTITLE/origin.command as the
  triggering command -- so an ungated rule alerts 10-25x per real event.
  Confirmed empirically: one decoy `sudo systemctl stop <name containing
  "auditd">` produced 24 duplicate alerts from
  debian_family/auditd_syslog_disabling.yaml. Audited all 70 Linux rule files
  (up from 49 at the start of this work): rewrote the 10 rhel_family/ rules
  (previously built on an entirely fictional schema -- log.file_path,
  log.event_type, log.tclass, log.scontext, log.process_name, etc. that this
  collector never produces under any name, stripped or not -- reduced to what
  real execve/journald telemetry can actually support, with genuine
  telemetry-gap sections documented rather than faked, and one branch
  (SELinux AVC via log.avc.*) left explicitly flagged unverified since
  V11-DEV-18 is AppArmor/Ubuntu, not SELinux); added the execve gate to 31
  more files across debian_family/ and top-level (29 straightforward, 1
  needed a real origin.command vs origin.process redesign --
  ssh_tunneling_port_forwarding.yaml, whose sshd branch could never fire
  since sshd's own origin.command is a static "/usr/sbin/sshd -D -R" --  and
  1 left as message-based by design -- debian_kernel_exploits.yaml); fixed 3
  more standalone fictional-field files directly (crontab_persistence.yaml,
  systemd_timer_persistence.yaml: log.process -> origin.process;
  debian_specific_rootkits.yaml: full rewrite, dropped unfixable branches
  needing telemetry we don't collect at all -- EFI/network-listen/proc-maps
  -- rather than fake them); documented (not fixed, needs a new audit rule)
  that process_masquerading.yaml is inert pending a `prctl` audit watch.
  Deployed the entire updated 70-file rules/linux/ tree to V11-DEV-18 (rsync
  --delete + force-update) so all of the above is live, not just locally
  edited. Running tally: 70 rule files total, all now free of the two
  systemic bugs (stripped-field-name mismatches, ungated sudo-cascade
  flooding) and the fictional-schema problem; ~15 confirmed firing via real
  triggered alerts this session, the rest statically corrected but not yet
  individually live-fire tested (most of rhel_family/ can't be -- no
  RHEL/SELinux/OpenShift/yum host available; V11-DEV-18 is Debian/Ubuntu).
- 2026-09-15 (Linux, session 1): Methodology written. Reference repos cloned
  to `Rules-Resources/`. Linux test agent confirmed live on V11-DEV-18. Gap
  analysis against Sigma/Wazuh not started yet.
