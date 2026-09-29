# Cisco switch v11 filter and rule review

The Cisco switch filter (`filters/cisco/cs_switch.yml`) and its three rules had defects that real
switch records, the filter's own patterns, the event engine and the SDK make visible. The MAC
address spoofing rule matched every MAC flap notification, but no step wrote the `origin.mac` its
history search needs, so the rule failed on every flap, disabled itself after five failures and
raised `Circuit Breaker` alerts in production; it has never produced a detection. The ARP
poisoning rule had the same unresolved history value on every branch. The one raw `log.*`
comparison in the filter stored an error on every line without a Cisco header. Addresses that six
real message types carry in their text were never mapped. This revision fixes those points and
nothing else. The schema is ThreatWinds go-sdk **v1.1.36**, as pinned by `plugins/alerts/go.mod`
since official `v11` (`d2479c1a`) was merged into this branch. The review itself used v1.1.33,
whose `plugins.proto`, `plugins/cel.go` and `plugins/rules.go` are identical. The draft was
checked again on the latest versions; see [Re-validation on the latest versions](#re-validation-on-the-latest-versions).

## Evidence basis

- **Real records, read only.** 29 of the 31 v11 instances could be searched; on one the search
  service did not answer, and one had no SSH access. Two instances hold Cisco switch records under
  this data type, and a third receives only other devices' logs on this input (see "For the
  owner"). Nothing was written to any instance. Record contents, addresses, instance names and
  document identifiers stay in the private review; this file gives counts only.

  | Message class (all time, 2026-09-24) | Instance A | Instance B |
  |---|---|---|
  | Records (senders), first record | 225,974 (5), 2026-08-24 | 2,618 (1), 2026-08-26 |
  | `SW_MATM-4-MACFLAP_NOTIF` | 179,487 | 2,519 |
  | `LINK-3-UPDOWN` / `LINEPROTO-5-UPDOWN` | 36,529 / 0 | 41 / 49 |
  | `SISF-4-EXCESS_ARP_ACTIVITY` | 4,048 | 0 |
  | Wireless access point traces (`CAPWAPAC_SMGR_TRACE_MESSAGE`, `APMGR_TRACE_MESSAGE`) | 3,143 | 0 |
  | `PKI` (7 mnemonics) | 2,512 | 0 |
  | `IP-3-LOOPPAK` | 112 | 0 |
  | `SYS-3-LOGGINGHOST_FAIL` / `SYS-6-LOGGINGHOST_STARTSTOP` | 97 / 0 | 0 / 1 |
  | `SSH-4-SSH2_UNEXPECTED_MSG` / `SSH-5-SSH_CLOSE` | 5 / 0 | 0 / 2 |
  | `DHCPD-4-PING_CONFLICT` | 0 | 5 |
  | Optics, licensing, CPU and platform messages | 41 | 1 |
  | Records stored with an error | 0 | 0 |

  Every record on both instances has one of two header shapes, both handled by the filter's first
  time step: `<PRI>seq: Mon DD HH:MM:SS.mmm: %FACILITY-SEVERITY-MNEMONIC: text` (instance A; a
  small share has a leading `.` before the time) and `<PRI>seq: *Mon DD YYYY HH:MM:SS: %...`
  (instance B). No access-list, 802.1X, `SW_DAI`, `DTP`, `SW_VLAN`, `IP-4-DUPADDR` or
  `IP SOURCEGUARD` record exists on any searched instance. The deployed filter and rules (engine
  v11.2.13) are identical to this repository's on the three instances.
- **Production circuit breakers.** The alert indices hold 12 `Circuit Breaker: MAC Address
  Spoofing Detection` alerts: 2 on instance A (2026-09-04 to 2026-09-21) and 10 on instance B
  (2026-08-19 to 2026-09-21). Each reads "The rule MAC Address Spoofing Detection has been
  temporarily disabled after failing 5 times during processing.", carries the error
  `expression value cannot be nil after placeholder resolution`, was triggered by a
  `SW_MATM`/`MACFLAP_NOTIF` event and is stored with severity High; the latest ones are Open.
  No `MAC Address Spoofing Detection`, `ARP Poisoning Attack Detection` or `VLAN Hopping Attack
  Detection` alert exists on any of the three instances.
- **Cisco documentation was not available.** The allowed fetch tool received HTTP 403 for the
  system message guide the filter cites and for the two Cisco pages the rules cite (one request
  each, 2026-09-24 15:31 UTC; not retried or worked around). No fixture is taken from Cisco's
  documentation, and nothing that depends on what a message means, which port a flap moved
  from, or how an unobserved message is laid out is changed. MITRE ATT&CK v19.2 was read; no
  label changes (see D-12).
- **What the corrections rest on:** the real text shapes above; the public EventProcessor at
  commit `497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1`, whose
  [parser](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/pkg/parsing/parsing.go)
  stores an error and skips the step when a `where` clause fails, whose
  [grok plugin](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/plugins/grok/main.go)
  writes nothing unless every pattern matched at the start of the remaining text, and whose
  [CEL plugin](https://github.com/utmstack/EventProcessor/blob/497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1/plugins/cel/main.go)
  disables a rule at its fifth error with a `Circuit Breaker: <rule name>` alert (the latest
  EventProcessor, `main` at `8a3ade72bd9d12db21f6b273200588fb49540f14`, changes only these
  plugins' go-sdk version and how the CEL plugin reads its OpenSearch address, so this is the
  same there); and go-sdk v1.1.33 (the three files below are identical in v1.1.36), whose
  [`plugins/cel.go`](https://github.com/threatwinds/go-sdk/blob/v1.1.33/plugins/cel.go)
  declares a CEL variable only for the top-level keys an event has (so `log.severity=="4"` fails
  without a `log` object, while `equals` returns false), whose
  [`plugins/rules.go`](https://github.com/threatwinds/go-sdk/blob/v1.1.33/plugins/rules.go)
  returns an error when a `{{.field}}` history value cannot be resolved, and whose
  [`plugins/plugins.proto`](https://github.com/threatwinds/go-sdk/blob/v1.1.33/plugins/plugins.proto)
  makes `origin.mac` and `origin.ip` strings and `origin.port`/`target.port` whole numbers.

## Filter changes (version 3.1.0)

| Id | Change | Why | Proof |
|---|---|---|---|
| F-1 | Line 206: `where: log.severity=="4"` becomes `equals("log.severity", "4")`. | A line without a `%FACILITY-SEVERITY-MNEMONIC` header has no `log` object or no `log.severity`, so the raw clause failed and an error was stored on the event. It was the only raw clause among the filter's 16. | Playground, 398 inputs (327 real, 71 fabricated): events with errors 25 to 0; severity identical on 398 of 398. Go test: every clause runs without an error on drafts without `log`, without a severity and parsed; severity unchanged on levels 0 to 7. |
| F-2 | `SW_MATM-4-MACFLAP_NOTIF`: `Host <mac> in vlan <n> is flapping between port <if> and port <if>` gives `origin.mac`, `log.vlan`, `log.firstPort` and `log.secondPort`. | Every stored flap record has this text, but the address stayed inside `log.ciscoMsg`. Interface names are text, so they stay under `log.*`; the ports keep the order the message gives, because which one is the previous port is not established (D-2). | Playground: 30 of 30 real flap records get all four fields, each equal to an independent reading of the text; fabricated upper-case, port-channel, card-slot and trailing-text lines extracted; colon MAC, 13 hex digits, missing VLAN, truncated port and `SW_VLAN` lines untouched. |
| F-3 | `SISF-4-EXCESS_ARP_ACTIVITY`: the client address to `origin.mac`. | All stored records have `... Excessive ARP activity detected for the client <mac>. client is brought down ...`. | Playground: 20 of 20 real records; colon-form refused. |
| F-4 | `SSH-4-SSH2_UNEXPECTED_MSG` and `SSH-5-SSH_CLOSE`: the client address to `origin.ip`. | `Terminating the connection from <ip>` and `SSH Session from <ip> (tty ...` name the client. | Playground: 5 of 5 and 1 of 1 real records; IPv6, invalid octet and trailing-text lines refused. |
| F-5 | `DHCPD-4-PING_CONFLICT`: the pinged address to `target.ip`. | `server pinged <ip>.` names the address the server tested. | Playground: 5 of 5 real records; no final period and invalid octet refused. |
| F-6 | `SYS-3-LOGGINGHOST_FAIL` and `SYS-6-LOGGINGHOST_STARTSTOP`: the logging host to `target.ip` and its port to `target.port`, cast to a number. | `Logging to host <ip> port <n> failed/started` names where the switch sends its logs. | Playground: 10 of 10 and 1 of 1 real records, `target.port` a JSON number; `stopped`, host name, text, 11-digit and negative ports refused. |

Every new step runs only for its facility and mnemonic. The `actionResult` steps (lines 159-192)
and the severity words are unchanged (D-9, D-10).

## Rule changes

Names, impact, category, technique, adversary side, references, thresholds and windows are
unchanged. `vlan_hopping_attempts.yml` is unchanged.

| Rule | Change | Why |
|---|---|---|
| `mac_address_spoofing` (v1.0.1) | Require `origin.mac`; leave out flap notifications (`!regexMatch("log.msg", "(?i)(mac.*flap\|is flapping between port)")`); read `log.msg` instead of `log.message`; `deduplicateBy: adversary.mac` instead of `groupBy`; the description says flaps are not used. | The flap branch matched every flap and failed its `{{.origin.mac}}` history value, which caused the production circuit breakers. With F-2 the value resolves, so the unchanged rule would run a history search on every flap. Nothing shows that a flap means an address was copied (see the decision below). Nothing writes `log.message`. |
| `arp_poisoning_detection` (v1.0.1) | Wrap the condition in `exists("origin.ip") && (...)`; read `log.msg` instead of `log.message`. | No step writes `origin.ip` for `SW_DAI`, `IP DUPADDR/SOURCEGUARD` or the text branches, so any match would fail its `{{.origin.ip}}` history value in the same way. No such message has arrived yet. |

F-2 and the MAC rule change must ship together. The rule commit comes first on this branch, so
every commit is safe on its own.

## The MAC rule decision

The flap trigger is left out of the spoofing rule. This is an owner decision that the reviewer
can revisit. It rests on the real flap patterns, because Cisco's explanation could not be read
and ATT&CK v19.2 has no MAC spoofing technique:

- Instance A: one address alternates between the same two ports about every 15 seconds, on 31
  of 31 days, with no other address involved.
- Instance B: 438 addresses in 30 days. Five port pairs carry 93% of the flaps, and each pair
  saw 73 to 116 different addresses. 64% of the unchanged rule's would-be firings happen while
  other addresses flap on the same ports.

A standing path problem or a moving host fits both patterns; a copied address would also cause
flaps, but the message carries nothing that tells the cases apart, and a history block can only
require at least N hits. Estimated volumes over the same 30 days, from the stored flap
timestamps (a model of the SDK history search, not observed alerts):

| Option (30 days) | Instance A | Instance B |
|---|---|---|
| O0: map `origin.mac`, rule unchanged | 172,614-175,323 indexed alerts under one parent | 488-500 indexed alerts |
| O1: deduplicate, 3 in 10 minutes | 5 alerts from 172,614-175,323 firings | 115-119 alerts from 488-500 firings |
| O3: deduplicate, 10 in 10 minutes | 5 alerts from 172,576-175,253 firings | 9 alerts from 54 firings |
| O5a / O5b: storms only, 50 / 100 in 10 minutes | 4-5 / 2 alerts from 1,212-12,445 / 620-3,058 firings | 0 / 0 |
| **O4: no flap trigger (this revision)** | **0** | **0** |

Every firing also reaches the correlation plugins, including automatic AI analysis when it is
on, even when deduplication hides the alert. If the owner wants flaps to alert, a separately
named network-health rule on flap storms (O5) fits the data better than the spoofing rule (D-1).
With O4 the rule keeps its duplicate-MAC, MAC-conflict, MAC-move and `SW_DAI` branches, which
need an `origin.mac` producer for those messages (D-3); on today's data it has no live path, as
before, but it no longer disables itself.

## Re-validation on the latest versions

On 2026-09-24 official `v11` moved to `d2479c1a3705eec6a00016689c2bf5fbcc1814f2`, whose
`plugins/alerts` pins go-sdk v1.1.36, and EventProcessor `main` moved to
`8a3ade72bd9d12db21f6b273200588fb49540f14`, whose playground and parser, writer and CEL
plugins all link go-sdk v1.1.36. `v11` was merged into this branch. No file overlaps this
draft, so nothing conflicted.

The newest published engine image, `ghcr.io/utmstack/utmstack/eventprocessor:v11.2.14`
(built 2026-09-24 19:13 UTC on base image `eventprocessor/base:1.1.7`), embeds Go build
information showing that its playground and plugin binaries come from the same
EventProcessor revision `8a3ade7` with go-sdk v1.1.36, built with go1.26.8 for linux/amd64.
The local build used here is that source revision compiled natively for darwin/arm64 with
go1.25.7; only the Go toolchain and platform differ.

What changed in the SDK, and what it means here:

- Since v1.1.35, `utils.SanitizeField` keeps `_` in the field names that the `json`
  (top-level keys), `kv`, `grok`, `csv`, `xml`, `add` and `rename` plugins write. Other
  characters are still removed. This filter has no `json`, `kv` or `csv` step, and none of
  the 25 names it writes contains `_` or another removed character. The underscores in
  `SW_MATM`, `MACFLAP_NOTIF` and similar are values, not names. So every stored name stays the
  same; the comparisons below confirm it on real records.
- v1.1.36 makes `regexMatch` match string values only again. Since v1.1.34, `contains`,
  `containsAll`, `startsWith` and `endsWith` also search the JSON text of objects and lists.
  Every such call in this filter and its rules reads a text field (`log.msg`, `log.message`),
  so no result changes. `plugins.proto`, `plugins/cel.go` and `plugins/rules.go` are identical
  in v1.1.33 and v1.1.36.
- No filter, rule or fixture needed a change.

| Check on the latest versions | Result |
|---|---|
| Full `plugins/alerts` suite, go-sdk v1.1.36 | 51 tests pass, 11 skip, none fail (2,281 passing results with subtests). The eight Cisco switch tests pass; the Go model of the step plugins uses the SDK's own `SanitizeField`, so it follows v1.1.36. In a throwaway copy with a Cisco switch manifest added, `TestFilterAndRuleContracts` passes for the filter and the three rules (312 subtests, none fail). The skipped tests need other technologies' private evidence and skip on the base commit too. |
| `replay.py` on EventProcessor 8a3ade7 | 44 events without errors, every stored field as in `expected.json`; six alerts, all from the VLAN hopping rule; no `Circuit Breaker` and no history search attempted. |
| Original and corrected filter, the same 398 inputs (327 real records, 71 fabricated) | Both runs are identical, event by event, to the original review's runs (with the new port names). Errors 25 to 0; severity identical on 398; 30 of 30 real flap records and 42 of 42 real SISF, SSH, DHCPD and logging-host records carry their fields; 18 of 18 fabricated near-misses carry none; `target.port` is a number on all 13; no event has `origin.port`. |
| Corrected filter, the 2,795 distinct real texts and 28 fabricated lines | 2,823 events, identical to the original review's output for every line. |
| Corrected filter and the three rules, the same 20 fabricated lines | Six VLAN hopping alerts, exactly on the six `SW_VLAN`/`DTP` lines; no MAC, ARP or `Circuit Breaker` alert; no rule or history search error. 14 of 14 checks pass. |
| Positive control, the same 2 contrived lines | Both the MAC and the ARP rule reached their history search with the value resolved; both searches failed because nothing listened. No alert. |
| Corrected filter and the original rules, the same 20 lines | The original MAC rule reached its history search on all 8 flaps; both `SW_DAI` lines failed the MAC and ARP rules with `expression value cannot be nil after placeholder resolution`; one `Circuit Breaker: MAC Address Spoofing Detection` alert. |
| go-sdk v1.1.36 rule replay | Over the 2,823 latest outputs above plus the 8 synthetic events: the committed MAC and ARP rules match none of the 229,017 real records they stand for; the original MAC rule matches all 182,326 flap records with the value resolved; 15 of 15 synthetic checks pass. Over the 398 latest events: committed MAC 0, ARP 0, VLAN 1 fabricated line. |

At 8a3ade7 the CEL plugin reads its OpenSearch address from separate `host`, `port`, `user`
and `password` settings. `replay.py` still gives one URL, so the client gets an empty host
and connects to port 443 on the test computer, where nothing listened. Any history search
therefore still fails and is reported; none was attempted with the committed rules and lines.

## Validation

These are the original review's results, on EventProcessor `497bf53` and go-sdk v1.1.33.
The section above repeats them on the latest versions.

**Fabricated regression, committed.** `plugins/alerts/testdata/cisco-switch/` holds 44 invented
raw lines (`raw.json`), their expected fields and alerts (`expected.json`), the 8 shared grok
definitions the filter uses (`patterns.yaml`, copied from
`20250616001_insert_utm_regex_pattern.xml`; identical to the deployed definitions) and
`replay.py`. MAC addresses are in the locally administered range 02:00:00:xx:xx:xx in Cisco's
dotted form, addresses are RFC 5737 and RFC 3849 documentation addresses, and host, user and
interface names are examples. The flap, SISF, SSH, DHCPD and logging-host lines copy the observed
text shapes; the `SW_DAI`, `IP`, `SW_VLAN` and `DTP` lines follow the filter's and rules' own
patterns, because no such record or documentation was available.

**Playground.** A clean build of the EventProcessor commit above, whose parser and writer plugins
link go-sdk v1.1.26 and whose CEL plugin links v1.1.34; every binary reproduced its recorded
checksum. File input, one fresh private working directory per run.

| Run | Result |
|---|---|
| Original filter, 398 inputs (327 real records from three instances, 71 fabricated) | 398 events, identical to an earlier run of the same inputs. 25 events carry the line-206 error (17 without a `log` object, 8 without a severity); on the 327 real records the playground reproduced every stored document, error texts included. |
| Corrected filter, same 398 inputs | 398 events, no errors, severity identical on 398. Every difference from the original run is an intended new field: `origin.mac` 59, `log.vlan`/`log.firstPort`/`log.secondPort` 37 each, `target.ip` 21, `target.port` 13 (all numbers), `origin.ip` 9; 30 of 30 real flap records and 42 of 42 real SISF, SSH, DHCPD and logging-host records carry their fields; 18 of 18 fabricated near-misses carry none; no event has `origin.port`. |
| Corrected filter and the three rules, 20 fabricated lines: 8 flaps of one address within two minutes, 2 `SW_DAI`, 1 each of SISF, SSH, DHCPD and logging host, 6 `SW_VLAN`/`DTP` | 20 events without errors. Six VLAN hopping alerts, exactly on the six `SW_VLAN`/`DTP` lines; no MAC, ARP or `Circuit Breaker` alert; no compile, rule or history search error. The rules' OpenSearch address was a closed local port and no history search was attempted. |
| Corrected filter and the three rules, 2 contrived lines: a SISF line whose text also says `duplicate mac`, and an SSH session line that ends in `gratuitous arp` | Positive control for the MAC and ARP rules: both conditions were true on the addresses the new steps wrote (`origin.mac`, `origin.ip`), and each rule reached its history search with the value resolved; both searches failed because nothing listened. No alert. These lines are not real message shapes and are not committed. |
| Corrected filter and the ORIGINAL rules, same 20 lines | The original MAC rule reached its history search on all 8 flaps (their `origin.mac` now resolves); each search failed because nothing listened. Both `SW_DAI` lines failed the MAC and the ARP rule with `expression value cannot be nil after placeholder resolution`. One `Circuit Breaker: MAC Address Spoofing Detection` alert; the same six VLAN alerts. |
| Committed `replay.py`, 44 lines | 44 events without errors, every stored field as recorded in `expected.json`; six alerts, all from the VLAN hopping rule; no `Circuit Breaker` and no history search attempted. |

**SDK predicate checks.** The go-sdk v1.1.33 rule replay evaluated the committed rules over the
corrected filter's playground output of all 2,795 distinct texts behind the 229,017 switch
records stored on the two instances when they were collected (182,326 of them flaps), and over
the 398 events above. The committed MAC and ARP rules match none of them; the original MAC rule
matches all 182,326 flap records, and with the corrected filter their history value now resolves.
15 of 15 synthetic checks pass: duplicate-MAC, MAC-conflict and `SW_DAI` events with `origin.mac`
match the MAC rule with the value resolved, `SW_DAI` with `origin.ip` matches the ARP rule, and
flap, SISF, SSH and address-less `SW_DAI` events match neither. `TestFilterAndRuleContracts`
passes for the filter and the three rules.

**Go tests.** `cisco_switch_filter_test.go` has eight tests. They check that no `where` clause
compares `log.*` directly and every clause runs without an error without a `log` object or a
severity; that the severity steps keep their results; that a model of the engine's step plugins,
with every `where` clause evaluated by go-sdk v1.1.33, gives the playground's result for every
stored field of the 44 lines, including a positive and a near-miss line for each new mapping;
that no interface name reaches `origin.port` or `target.port`; the rules' names, metadata,
impact, grouping and history searches and the unchanged VLAN condition; 28 synthetic rule cases;
that the MAC and ARP rules match none of the 44 lines and the VLAN rule exactly the six
`SW_VLAN`/`DTP` lines; and that whenever a rule with a history search matches, its history values
resolve. All eight fail against the original filter and rules and pass against this revision.
The full `plugins/alerts` suite passes: 51 tests pass, and the same 11 tests that need other
technologies' private evidence skip, as they do on the base commit (43 pass, 11 skip).

## Deferred

Each of these needs Cisco's documentation, real records or an owner decision, and is unchanged
here.

| Id | What | What would unblock it |
|---|---|---|
| D-1 | Whether flap notifications should raise any alert, including the MAC-move text branch. | Cisco's explanation of `SW_MATM-4-MACFLAP_NOTIF`, and an owner decision; if wanted, a separately named network-health rule, not the spoofing rule. |
| D-2 | Which flap port is the previous one (rename `log.firstPort`/`log.secondPort` to a direction). | Cisco's message layout, or a lab test moving one host between known ports. |
| D-3 | Address mapping for `SW_DAI`, `IP DUPADDR` and `SOURCEGUARD`; the wording and case of the ARP and MAC text branches (the ARP phrases are case-sensitive); one event firing both rules. | Cisco layouts and real records (none exist). |
| D-4 | VLAN rule: its `log.message` text branches, `MACFLAP_NOTIF` listed under `SW_VLAN`, grouping keys no step writes. Reading `log.msg` would wake text branches that have no threshold and no guard. | Cisco wording and real `SW_VLAN`/`DTP` records. |
| D-5 | New rules for SISF excess ARP, logging-host failures, DHCP conflicts and PKI failures. | Cisco explanations and an owner decision. |
| D-6 | Access-list field extraction. | Cisco formats and real records (none). |
| D-7 | `deviceTime` from the header time: no sender includes a time zone, one writes local time. | Cisco timestamp options; a sender that sends year and zone. |
| D-8 | Header variants and subfacilities that contain digits. | Cisco documentation or real records. |
| D-9 | Severity words `high`/`medium`/`low` instead of the SDK wiki's values. | Owner decision after a dashboard and saved-search review. |
| D-10 | `actionResult` values `failed` and `blocked`. | The separate action-result correction. |
| D-11 | Interface and state of link messages. | A consumer that needs them. |
| D-12 | ATT&CK v19 tactic names (Defense Evasion is now Stealth; T1599 moved to Defense Impairment). | A repository-wide owner decision. |

## For the owner

On one instance, Firepower Threat Defense events and the management center's Linux system lines
arrive on the Cisco switch input, because those devices send to the port assigned to this data
type. About 1.12 million of those records carried the line-206 error. F-1 only removes that
error. The Firepower events are still parsed as switch messages, stored with severity `high`
and their addresses left in text, and neither the Firepower filter nor its rules see them. After
this change the misrouting no longer shows up as errors; track it by facility `FTD` and by
records without `log.facility`. The fix is on the customer side: point those devices at the
Firepower input, decide whether the management center's system log should be collected, and
check the TCP framing between them and the collector (some records hold two messages glued
together).

## Known limits

- The latest check used EventProcessor `8a3ade7`, whose playground and plugins link go-sdk
  v1.1.36, the version the alerts module now pins; the predicates were also checked with
  v1.1.36. The original review used `497bf53` (parser and writer plugins v1.1.26, CEL plugin
  v1.1.34) and v1.1.33 predicates. Neither build is asserted to match a customer deployment.
- No history search completed: there was no OpenSearch, so the searches that the original rules
  and the two contrived lines started failed to connect. History, indexing, grouping, the MAC
  rule's new deduplication, notifications and production alerts were not tested. The volume table
  above is a model of the SDK search over stored timestamps.
- The MAC and ARP rules have no positive case in a real message shape, because this filter cannot
  give their remaining messages an address yet (D-3). Their positive cases are the synthetic
  normalized events and the two contrived playground lines above.
- The Go extraction test is a model of the engine's step plugins. It agreed with the playground on
  every field of the 44 lines, but `replay.py` is the check that runs the engine.
- `equals("log.severity", "4")` compares numbers, like the neighbouring `oneOf` severity clauses,
  so a level written `04` or `+4` now counts as 4, as `03` already counted as 3; the old clause
  gave such a level no severity. No stored record has such a level.
- The separate action-result correction edits the same filter; combining the two needs a rebase
  at the version line.

## Reproduce

Build EventProcessor `8a3ade72bd9d12db21f6b273200588fb49540f14` (the latest check) without
changing its dependencies. With `EP` set to that checkout's absolute path:

```sh
mkdir -p "$EP/test-bin" "$EP/test-plugins"
(cd "$EP" && go build -mod=readonly -o "$EP/test-bin/playground" ./cmd/playground)
for plugin in add cast cel delete grok saw sew trim; do
  (cd "$EP/plugins/$plugin" && go build -mod=readonly -o "$EP/test-plugins/$plugin.plugin" .)
done
```

From this UTMStack checkout, with PyYAML installed:

```sh
python3 plugins/alerts/testdata/cisco-switch/replay.py --playground "$EP/test-bin/playground" \
  --plugins "$EP/test-plugins"
(cd plugins/alerts && go test ./... -count=1)
```

The playground run takes about two and a half minutes. The Go suite alone does not run the
engine.
