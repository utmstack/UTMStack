# UTMStack EDR — Completion Plan and Team Assignments

**Date:** 2026-09-21
**Product owner:** Rick
**Engineers:** Yadian · Alex · Andres · Jose

"EDR" here means the UTMStack endpoint detection and response module: the piece that
watches an endpoint for malware and hostile behaviour, stops it, and reports what it did.

This document says three things: what already works, what is still missing before we can
call the feature finished, and exactly who builds each missing piece.

---

## 1. Where the work stands today

### 1.1 The endpoint module (Windows) — built and proven on a test machine

Roughly 11,700 lines of Go (3,650 of them tests) plus 178 lines of native C++ for the one
piece that cannot be written in Go. Every package's tests pass, the module cross-builds and
passes static checks for both Windows processor families (Intel/AMD 64-bit and ARM 64-bit),
and the branding audit is clean.

| Capability | What it does | State |
|---|---|---|
| Own service | `utmstack_edr.exe` runs as its own Windows system service; the agent only starts, stops and configures it | Working, proven on the test machine |
| Scanning engine | The open-source scanning engine runs as a supervised child process; the module writes its configuration and tunes it to the host's memory and core count | Working, proven |
| File arrival watch | Reads the file system change journal for the whole volume, scans new and changed files, moves malicious ones to quarantine (never deletes), supports restore | Working, proven |
| Process launch watch | Sees every process start, scans its image, and kills the whole process tree when the verdict is malicious. Optional freeze-before-scan exists but ships off because it once froze the test machine | Working, proven |
| Script blocking | A small native library plugs into the Windows script-scanning interface and forwards script text to the module before it runs; a real malicious script was blocked end to end | Working, proven |
| Behavioural telemetry | Forwards process creation and PowerShell script-block records to the platform | Working, proven |
| Ransomware guard | Decoy files plus recovery-tampering command rules plus a file-activity feed drive a per-process score; crossing a threshold freezes, kills and quarantines the culprit | Working, proven. On by default in freeze mode |
| Network blocking | Pulls indicator lists from the server mirror and blocks matching addresses, address ranges and domains using the built-in Windows filtering layer, with self-lockout protection | Working, proven. On and enforcing by default |
| Configuration | One file, `edr.json`. A local command-line tool edits it. Changes to allowlists and response mode apply live; structural changes need a restart | Working |
| False-positive tuning | One combined allowlist for paths, processes, commands and networks, plus a switch per sensor so one noisy detector can be turned off without disabling the module | Working |
| Event path | The module appends events to a durable spool file; the agent drains it into the existing log channel, so there is one outbound connection and at-least-once delivery | Working, proven |

### 1.2 The master-server side — built, not deployed

A new container called `edr` (about 1,600 lines of Go) mirrors two content channels for the
whole fleet from one upstream fetch: the scanning engine's signature databases and the
threat-intelligence indicator lists. Agent-manager serves the shared volume read-only on the
existing dependencies port for deployments with a trusted certificate; the container serves
the same tree over plain HTTP on port 9002 for self-signed deployments, where the signature
updater cannot validate certificates. The installer and the build pipeline are wired.

It has never run on a real customer deployment.

### 1.3 What is missing

1. **Nothing in the UTMStack console shows or controls the EDR.** Today an administrator
   would have to open a remote shell to each endpoint and type commands. There is no page,
   no policy, no fleet view, no remote quarantine.
2. **The platform does not understand the events.** The endpoints already send them, but
   there is no parser, no correlation rule and no playbook, so they land in the log explorer
   as unstructured records and never become alerts.
3. **Linux is stubs only.** Every sensor compiles on Linux and does nothing.
4. **The binaries are not delivered.** The build pipeline does not publish
   `utmstack_edr` and there is no code anywhere that downloads the scanning engine — it was
   installed by hand on the test machine.
5. **Nothing is signed.** Production Windows deployment needs an Authenticode signature on
   both the module and the native script-scanning library.
6. **Two real endpoint gaps for a complete product:** there is no scheduled full-disk scan,
   and there is no host isolation (cutting a compromised endpoint off the network while
   leaving the path back to UTMStack open).

### 1.4 Code location

| What | Where |
|---|---|
| Endpoint module | `utmstack/UTMStack`, branch `edr-phase1`, path `agent/edr/` — two commits on top of `release/v12.0.0` |
| Master-server container | `utmstack/UTMStack`, branch `feature/edr-server-container`, path `edr/` — one commit |
| Designs and plans | `docs/superpowers/specs/` and `docs/superpowers/plans/` |
| Private safety copy | `utmstack/OpenEDR` (private), branch `main` — snapshot plus restorable patches |

Neither branch has been pushed to the shared repository or opened as a pull request yet.

---

## 2. What "done" means

The feature is finished when all of the following are true.

1. An administrator can see, from the UTMStack console, which endpoints have the EDR, whether
   it is healthy, how fresh its signatures are, and what it has caught.
2. An administrator can change behaviour centrally through a policy — allowlists, which
   sensors run, how aggressively ransomware is handled, whether connections are blocked —
   and see the policy actually applied on the endpoint.
3. An administrator can act on an endpoint from the console without a shell: run a scan,
   restore or permanently remove a quarantined file, kill a process, isolate the host.
4. Detections become alerts, with correlation rules and automated response playbooks.
5. Everything above works the same on Windows and on Linux.
6. The signature and threat-intelligence mirrors run on the customer's own server, and the
   endpoint binaries are published, signed and installed by the normal agent update path.

---

## 3. How the work is split

| Engineer | Owns | Why |
|---|---|---|
| **Yadian** | The control path between the server and the agent, and the Linux port of the endpoint module | He knows the agent manager and the agent-to-server conversation better than anyone, and the Linux sensors are agent-internal work |
| **Alex** | The backend module and its interface, the master-server container in production, the build and delivery pipeline, and the assistant integration | Large volume of self-contained backend work that does not require deep agent knowledge |
| **Andres** | Everything the administrator sees | Front-end owner; picks up light interface glue on the backend where the screens need shaping |
| **Jose** | Detection content and proving every capability works on both operating systems | He is the person for parsers, correlation and cross-platform behaviour verification |

Nobody works alone on a boundary: Yadian and Alex agree the message shapes before either
starts; Andres and Alex agree the interface shape before the screens start; Jose reviews
every capability before it is called complete.

---

## 4. Phases

| Phase | Goal | Rough length |
|---|---|---|
| **Phase 0 — Land what exists** | The two branches are reviewed, merged and building; the server container runs on a real deployment; events become alerts | 2 weeks |
| **Phase 1 — Console control, Windows** | An administrator can see and drive the EDR from UTMStack on Windows endpoints | 6–8 weeks |
| **Phase 2 — Linux parity** | Every capability works on Linux and is visible in the same screens | 8–10 weeks |
| **Phase 3 — Production hardening** | Signed binaries, scale testing, false-positive tuning, the two remaining endpoint gaps | 4 weeks, overlapping Phase 2 |

Phases 1 and 2 overlap: Yadian moves to Linux as soon as the control path is handed to Alex
and Andres for wiring.

---

## 5. Phase 0 — Land what exists (2 weeks)

### Yadian

**Y0.1 — Review and merge the endpoint branch.**
Read `edr-phase1` (two commits, about 135 files) against `release/v12.0.0`, focusing on the
four files that touch agent packages: `agent/cmd/{enable_edr,disable_edr,edr_status}.go`,
`agent/agent/edr_relay.go`, `agent/dependency/edr.go` and the one line added to
`agent/serv/service.go`. Confirm the module stays self-contained and that a failure inside it
cannot take the agent down.
*Done when:* the branch is pushed, a pull request into `release/v12.0.0` is approved, and the
agent still starts cleanly with the EDR absent, present-but-disabled, and enabled.

**Y0.2 — Review and merge the server container branch.**
`feature/edr-server-container` touches agent-manager, the installer and the build pipeline.
Check the new static route on the dependencies port cannot expose anything beyond the mirror
volume and that directory listings stay refused.
*Done when:* merged, and a fresh install produces a stack that includes the `edr` service.

### Alex

**A0.1 — Publish the endpoint binaries through the build pipeline.**
The build currently produces nothing for the EDR. Add build steps that produce
`utmstack_edr_windows_amd64.exe` and `utmstack_edr_windows_arm64.exe` and stage them into the
agent-manager dependencies image the same way the collector is staged
(`agent-manager/Dockerfile` copies `./dependencies/agent/`).
*Done when:* a built agent-manager image serves both files at
`https://<server>:9001/private/dependencies/agent/`, and an agent with `edr` in its
dependency list downloads and installs the module by itself.

**A0.2 — Package and host the scanning engine.**
No code anywhere downloads the engine; it was unpacked by hand on the test machine. Produce a
per-platform engine package (the daemon, its support libraries and a seed database set), host
it at the same dependencies endpoint, and write the endpoint-side code that downloads and
unpacks it into the module's `engine/` directory when the module is enabled and the directory
is missing or incomplete. Reuse `shared/http.DownloadFile` and `shared/archive.Unzip`.
*Done when:* a clean Windows machine with only the agent installed ends up with a working
engine after `enable-edr`, with no manual steps.

**A0.3 — Deploy the server container on a real deployment and watch it for a week.**
Bring up the `edr` service on a staging server, confirm the signature mirror fills, confirm
the threat-intelligence lists download with the platform's stored credentials, and confirm
agents pull from it on both transports (trusted certificate on port 9001, self-signed on port
9002). Settle the one open question in the design: whether the upstream feed path expects
`level1` or `1`.
*Done when:* an endpoint's signature database updates from the customer's own server with no
internet access from the endpoint, and the mirror's freshness document is being written.

### Jose

**J0.1 — Write the parser for EDR events.**
Create `definitions/filters/utmstack/utmstack_edr.yaml`. The events arrive as one JSON object
per record with data type `utmstack_edr`. Fields: `timestamp`, `host_id`, `os`, `product`,
`source` (one of `file_watcher`, `process_watcher`, `amsi`, `edr_engine`, `behavioral`,
`network_watcher`), `action` (`detected`, `quarantined`, `killed`, `blocked`, `allowed`,
`scan_complete`, `health`, `ransomware_suspected`, `ransomware_contained`), `object_path`,
`file_hash`, `verdict`, `signature`, `remote_ip`, `remote_port`, `domain`, `direction`,
`indicator`, `engine`, `severity`, and a nested `process` object with `pid`, `ppid`, `image`,
`command_line`, `session`, `user`.
*Done when:* every one of the six sources produces correctly typed, searchable fields in the
log explorer, with no unparsed remainder.

**J0.2 — Write the first correlation rules.**
Create `definitions/rules/edr/`. At minimum:
malware detected and quarantined; malware detected but not contained; malicious process tree
terminated; script blocked before execution; ransomware behaviour suspected; ransomware
contained; connection blocked to a known-bad address; repeated blocked connections from one
host; scanning engine unhealthy; signature database stale; a sensor switched off on an
endpoint that previously had it on (tamper signal); a quarantined file restored (audit trail).
*Done when:* each rule fires on a real event produced by the test machine, and none of them
fire during an hour of ordinary desktop use.

**J0.3 — Alert severity and scoring.**
Map each rule to a severity and make sure the scoring module treats a contained ransomware
incident as higher-impact than a single quarantined file.
*Done when:* Rick agrees the severities read correctly on the alerts page.

### Andres

**N0.1 — Detections view.**
Before any control surface exists, administrators need to see what the EDR found. Add an EDR
section to the alerts and log explorer experience that presents the parsed fields well: file
path, verdict, signature name, the process tree, and for network blocks the address, domain
and matched indicator.
*Done when:* an EDR alert opens into a readable detail panel without raw JSON.

---

## 6. Phase 1 — Console control on Windows (6–8 weeks)

This is the heart of the request: make the EDR a managed feature of UTMStack rather than a
thing that runs on its own.

### 6.1 The control path — Yadian

Today the only way the server can make an endpoint do something is to send a shell command
string over the agent stream and read back its text output
(`agent-manager/protos/agent.proto`, `UtmCommand` and `CommandResult`; handled in
`agent/agent/incident_response.go`). That is fine for an interactive console and wrong for a
product feature: no typed result, no way to tell a failure from empty output, and the audit
trail is a shell line.

**Recommended approach:** add typed messages to the same stream. One connection, one
authentication path, no shell.

**Y1.1 — Define the message contract.**
In `agent-manager/protos/agent.proto` add to the `BidirectionalStream` oneof:

- `EdrCommand { agent_id, cmd_id, action, payload, executed_by, reason }` where `action` is
  one of `enable`, `disable`, `status`, `policy_set`, `quarantine_list`,
  `quarantine_restore`, `quarantine_purge`, `scan_path`, `full_scan`, `kill_process`,
  `allow_add`, `allow_remove`, `blocklist_refresh`, `isolate`, `release_isolation`, and
  `payload` is a JSON argument object.
- `EdrResult { agent_id, cmd_id, ok, error, payload, executed_at }` where `payload` is a JSON
  result object.
- `EdrStatusReport { agent_id, status_json, policy_version, reported_at }` sent by the agent
  without being asked.

Agree the shape with Alex before generating code, because his interface mirrors it.
*Done when:* generated code compiles in agent-manager, the agent and the backend client, and
an old agent that does not understand the new messages still works.

**Y1.2 — Machine-readable output on the endpoint command-line tool.**
`agent/edr/manage.go` already implements `config show|get|set`, `allow …`,
`quarantine list|restore|purge` and `main.go` implements `scan`, `status`, `restore`. Add a
`--json` flag so every verb can emit a structured result instead of human text. This reuses
a surface that is already tested on the test machine rather than inventing a new channel
between the agent and the module.
*Done when:* every verb returns valid JSON with a stable field set, and the existing human
output is unchanged when the flag is absent.

**Y1.3 — Agent-side handler.**
New file `agent/agent/edr_control.go`. It receives `EdrCommand`, maps the action onto either
service control (`shared/svc`), a configuration write (`agent/edr/config`) or an invocation of
the module's command-line tool with `--json`, and returns `EdrResult`. It must never block the
stream: long actions such as a full scan return "accepted" immediately and report completion
through an event.
*Done when:* every action in Y1.1 round-trips from a test client, including the failure cases
(module not installed, service stopped, file already restored).

**Y1.4 — Status reporting upstream.**
The module already writes `status.json` with engine health, signature version and source,
sensor states, ransomware state, network blocking state and engine tuning. Have the agent send
`EdrStatusReport` every five minutes and immediately after any change, and have agent-manager
store the latest report per agent.
*Done when:* stopping the scanning engine on the test machine turns that endpoint unhealthy in
the server's stored state within five minutes.

**Y1.5 — Policy distribution.**
A policy is the centrally managed subset of `edr.json`: enabled, sensor switches, allowlists,
ransomware mode and thresholds, network blocking settings, quarantine retention, scan
concurrency, signature source, engine tier override. The server sends `policy_set` with the
policy document and a version string; the agent merges it into `edr.json`, saves, and lets the
module's live reload pick up what it can. Settings that need a restart must cause the agent to
restart the module. The agent reports the applied policy version in every status report so the
server can show drift.

Two rules that must not be broken: the endpoint must keep working with its last policy when the
server is unreachable, and a local administrator's emergency change must not be silently
overwritten without it being visible as drift in the console.
*Done when:* changing a policy centrally changes behaviour on the test machine, with allowlist
changes taking effect within seconds and structural changes surviving a restart.

**Y1.6 — Host isolation.**
A real EDR can cut an endpoint off. Build it on the same Windows filtering layer the network
blocking already uses: block all traffic except the UTMStack server, the agent's ports, and the
loopback interface. It must fail open — if the module dies, isolation lifts — and it must be
reversible from the console and from the local command-line tool.
*Done when:* an isolated test machine cannot reach the internet or the local network, still
reports to UTMStack, and comes back fully when released.

**Y1.7 — Scheduled full scan.**
There is no scheduled scan today. Add one driven by policy: a schedule, a target set of paths,
and a throttle so it does not swamp the machine. Emit start and completion events.
*Done when:* a scheduled scan runs, reports progress, and stays inside the throttle.

### 6.2 The backend — Alex

**A1.1 — New backend module `backend/modules/edr/`.**
Follow the existing module shape exactly (`connectors`, `domain`, `dto`, `handler`,
`repository`, `usecase`, `module.go`, `routes.go` — copy the structure of
`backend/modules/datasources/`).

Endpoints, all under `/api/v1/edr`:

| Method and path | Purpose |
|---|---|
| `GET /overview` | Fleet summary: endpoints with the module, healthy, unhealthy, stale signatures, detections in the last day |
| `GET /endpoints` | Per-endpoint list with status, policy, last report time |
| `GET /endpoints/:agentId` | Full stored status for one endpoint |
| `POST /endpoints/:agentId/actions` | Dispatch one action (scan, kill, isolate, release, refresh) |
| `GET /policies`, `POST`, `GET /:id`, `PUT /:id`, `DELETE /:id` | Policy documents |
| `PUT /policies/:id/assignments` | Assign a policy to endpoints or groups |
| `GET /quarantine` | Quarantine items across the fleet, filterable by endpoint |
| `POST /quarantine/:itemId/restore`, `POST /quarantine/:itemId/purge` | Act on one item |
| `GET /blocklist/status` | Indicator counts, feed freshness, recent blocks |
| `GET /blocklist/allowlist`, `POST`, `DELETE` | Network allowlist entries |

**A1.2 — Permissions and migration.**
Add `edr.read` and `edr.write` to the permissions table following the pattern in
`backend/migrations/000001_init.up.sql` (a new numbered migration, with the matching removal
in the down migration), and gate every route with `middleware.RequirePermission`.
*Done when:* a user without the permission gets a refusal on every route, and a fresh install
has the permissions seeded.

**A1.3 — Extend the agent-manager client.**
`backend/pkg/agentmanager/client.go` already wraps the panel service. Add the calls the new
messages need, with sensible timeouts and a clear error when the agent is offline.
*Done when:* every backend route that touches an endpoint returns a useful error within ten
seconds when the endpoint is unreachable.

**A1.4 — Storage.**
Backend owns policies and assignments. Agent-manager owns the latest status per agent and the
command log (it already has an `AgentCommand` table — extend rather than duplicate). Define
which side is the source of truth for each table and write it down in the module's README.
*Done when:* a server restart loses no policy and no status older than one reporting interval.

**A1.5 — Assistant tools.**
Add EDR entries to `backend/modules/mcp/catalog.json` and the matching tool file
(`backend/modules/mcp/tools_edr.go`, following `tools_datasources.go`): read fleet status, read
one endpoint, list quarantine, list recent detections. Read-only first; any action tool needs
Rick's explicit approval before it ships.
*Done when:* the assistant can answer "which endpoints are unprotected right now" from real
data.

### 6.3 The console — Andres

New feature folder `frontend/src/features/edr/`, registered in `frontend/src/app/routes/index.tsx`
and in the navigation, gated on `edr.read` and `edr.write`, fully translated.

**N1.1 — Fleet overview page** (`pages/EdrOverviewPage.tsx`). Coverage (how many endpoints have
the module, how many are missing it), health, signature freshness, detections over time,
ransomware and network-block counters, and the list of endpoints needing attention.

**N1.2 — Endpoint detail panel** (`components/EdrEndpointDrawer.tsx`), reachable both from the
overview and from the existing data sources page where agents already live
(`frontend/src/features/datasources/pages/DataSourcesPage.tsx`). Shows status, which sensors are
on, engine tuning, applied policy and drift, recent detections, quarantined files, and the
action buttons: scan, full scan, kill process, isolate, release, refresh indicators.

**N1.3 — Policy editor** (`pages/EdrPoliciesPage.tsx` plus `components/PolicyEditor.tsx`).
Create, edit, assign. Grouped sections: sensors, allowlists (paths, processes, commands,
networks), ransomware response mode and thresholds, network blocking, quarantine retention,
scheduled scan, signature source. Every dangerous setting needs an explanation in plain words —
particularly the freeze-on-launch switch, which is off for a reason, and the ransomware kill
mode.

**N1.4 — Quarantine browser** (`pages/EdrQuarantinePage.tsx`). Fleet-wide list: endpoint, path,
detection name, hash, quarantine time, restore and permanent-delete actions with a confirmation
and an audit note.

**N1.5 — Network blocking page** (`pages/EdrBlocklistPage.tsx`). Indicator counts by type, feed
freshness, recent blocked connections, and the allowlist editor.

**N1.6 — Isolation is visible and obvious.** An isolated endpoint must be unmistakable in every
list it appears in, with a one-click release and a clear note about what isolation still allows.

*Done for all of the above when:* Jose can drive a complete incident — see the detection,
inspect the endpoint, isolate it, restore a file, release it — without leaving the console and
without reading a log file.

### 6.4 Verification — Jose

**J1.1 — Windows acceptance of the control path.** Every action, every failure case, on a real
Windows machine: module absent, module present but stopped, endpoint offline mid-command,
policy conflict between local and central.

**J1.2 — Coexistence with Windows Defender.** Two real-time scanners on one machine fight.
Run the full scenario set with Defender on, and record what behaves differently and what we
must tell customers. This has never been done properly — past testing turned Defender off for
clean attribution.

**J1.3 — Response playbooks.** Write the automated responses in `definitions/soar/`: isolate on
contained ransomware, alert and collect on repeated blocked connections, notify on a sensor
turned off. Test each one end to end.

**J1.4 — False-positive pass.** Run the module for a week on ordinary working machines —
developer, accounting, backup server — and tune the built-in allowlists from what actually
fires.

---

## 7. Phase 2 — Linux parity (8–10 weeks)

Everything the Windows module does, done again with Linux mechanisms. The module already
compiles on Linux with every sensor stubbed out, so the shape exists; the sensors do not.

One genuine upgrade to call out: on Linux, `fanotify` permission events let us decide whether a
file may be opened or executed **before** it runs. On Windows the module can only detect and
then kill, with a few milliseconds of exposure. The Linux build can therefore be strictly
stronger, and the product wording should say so.

### Yadian — the sensor core

**Y2.1 — Service and layout.** systemd unit, install and uninstall verbs, install directory,
file ownership and permissions, log location. Mirror the Windows service lifecycle so the agent
controls it identically.

**Y2.2 — File watch through fanotify.** Replace the change-journal watcher. Start in notify
mode (detect and quarantine, matching Windows), then add permission mode as a policy option so
a malicious file can be refused at open or exec time. Handle mount points, bind mounts, network
file systems and container overlays deliberately — this is where the sharp edges live.

**Y2.3 — Process watch through the netlink process connector.** Replace the Windows management
interface feed. Must deliver process start with identifier, parent identifier, image path and
command line, and must not miss short-lived processes the way the current Windows poll can.

**Y2.4 — Kill and freeze.** Terminate a process tree with signals, leaves first, reusing the
existing tree logic. Freeze with the control-group freezer rather than per-process stops.

**Y2.5 — Ransomware guard on Linux.** Reuse the scorer and the decoy logic unchanged; replace
the file-activity feed with fanotify and replace the Windows recovery-tampering rules with the
Linux equivalents: deleting or shredding backup trees, stopping backup services, removing file
system snapshots, disabling journaling.

**Y2.6 — Paths, mounts and volume enumeration.** The Windows code assumes drive letters in
several places. Audit every path comparison, the allowlist matcher, quarantine storage and the
decoy placement for correct behaviour with symbolic links, hard links and case-sensitive file
systems.

### Alex — the pieces that are not sensor work

**A2.1 — Network blocking on Linux** through nftables: its own table and chain so we never
disturb the customer's firewall, the same self-lockout protection, the same fail-open behaviour
when the module stops, and the same event on a drop.

**A2.2 — Engine packaging for Linux.** Per-distribution-family and per-architecture packages,
hosted and downloaded the same way as the Windows package from A0.2.

**A2.3 — Build pipeline for Linux binaries**, published at the dependencies endpoint for both
Intel/AMD 64-bit and ARM 64-bit.

**A2.4 — Backend and interface work for anything Linux-specific** that the screens need: a
different sensor set, the permission-mode option, distribution and kernel version in the
endpoint detail.

### Andres — console

**N2.1 — Make every screen operating-system aware.** Sensors that do not exist on Linux must not
appear as "off"; they must not appear at all. The policy editor must show the Linux-only
permission-mode option and hide the Windows-only script-scanning switch.

**N2.2 — Mixed-fleet views.** Coverage and health must read correctly when a customer runs both
operating systems.

### Jose — verification

**J2.1 — Capability matrix.** Build a table of every capability against Windows and Linux with a
verified result in each cell, on real machines, not in a container. Cover at least two
distribution families and both processor architectures.

**J2.2 — Linux detection content.** Extend the parser and rules for the Linux-only fields and
behaviours, and add rules for the Linux recovery-tampering signals.

**J2.3 — Performance under load.** Measure processor and memory cost of the file watch on a
busy build server and a busy database server. Ransomware decoys and fanotify are both capable
of hurting a loaded machine; find the limits and set safe defaults.

---

## 8. Phase 3 — Production hardening (4 weeks, overlapping)

| Task | Owner | Detail |
|---|---|---|
| **H1 — Authenticode signing** | Alex | Buy or locate the certificate, sign `utmstack_edr.exe` and `utmstack_amsi.dll` for both Windows architectures as a pipeline step. The script-scanning library in particular is loaded into other processes and will be distrusted unsigned |
| **H2 — Script blocking on Intel/AMD 64-bit** | Jose | The native library has only ever run on an ARM machine. Prove it on Intel/AMD hardware |
| **H3 — Known ransomware guard defects** | Yadian | The culprit's image is sometimes empty when the process table lags; a containment event can be emitted twice; a process that finishes in under a second can slip past the process poll |
| **H4 — Known network blocking defects** | Alex | The drop event reports process identifier zero and always reports the direction as outbound; device paths are not always converted to readable paths |
| **H5 — Configuration safety** | Yadian | Hand-editing `edr.json` to switch network blocking off is silently ignored today — a trap for a support engineer. Also expose the signature fallback setting in the command-line tool |
| **H6 — Scale test** | Jose | One hundred endpoints reporting status and pulling signatures from one server. Measure the mirror's load and the event volume the behavioural sensor produces |
| **H7 — Upgrade and rollback** | Yadian | Prove an endpoint upgrades from one module version to the next without losing quarantine, configuration or the event spool, and that a failed upgrade leaves a working agent |
| **H8 — Customer-facing documentation** | Andres with Jose | What the module does, what it cannot do, how to tune a false positive, what isolation allows through. The honesty rule applies: we detect and kill quickly on Windows, we can genuinely prevent on Linux, and script scanning only sees what the script host chooses to submit |

---

## 9. Rules everyone follows

1. **Branding.** Every event field and every surfaced log line says "UTMStack EDR". The
   scanning engine's vendor name must never appear in an emitted field or a log message. There
   is a grep-based audit; keep it green.
2. **Quarantine, never delete.** Files move to a protected store with a restore path. Permanent
   deletion is only ever an explicit administrator action.
3. **No kernel driver.** Windows response uses ordinary process termination with system
   privileges; Linux uses signals and fanotify. Pre-execution blocking on Windows would need
   the built-in application control feature, which is a later phase.
4. **Fail open by default.** If the module dies, the endpoint keeps working: filters release,
   isolation lifts, files open.
5. **Honest capability claims.** Do not describe the Windows file and process path as
   prevention. It is fast detection and termination with a small window of exposure.
6. **The endpoint works offline.** Last-known policy, last-known signatures, last-known
   indicators. Stale, never empty.
7. **Test on real machines.** Every Windows-specific and Linux-specific path has produced real
   bugs that only appeared on a real machine — twelve of them so far. Check the final result
   (the quarantine store, the event at the platform), never just the log line.

---

## 10. Decisions needed from Rick before Phase 1 starts

1. **Control path shape.** Typed messages on the agent stream, as recommended above, or the
   faster route of sending command-line strings through the existing shell channel? The typed
   route costs about a week more and gives a real audit trail and real error handling.
2. **When do the two branches go to the shared repository?** They are committed locally and
   backed up privately, but not pushed. Nothing in Phase 0 can start until they are.
3. **Linux scope.** Which distribution families and which minimum kernel version? Permission-mode
   blocking needs a reasonably recent kernel.
4. **macOS.** In or out? The agent supports it; the EDR module has no macOS work at all and it is
   not in this plan.
5. **Code-signing certificate.** Who procures it, and does it cover the native library as well as
   the module?
6. **Licensing.** Is the EDR metered or licensed separately from data sources? The billing module
   exists and would need to know.

---

## 11. Risks

| Risk | Impact | What we do about it |
|---|---|---|
| Two real-time scanners on one Windows machine | Unpredictable attribution, customer confusion | J1.2 coexistence pass, and clear guidance in the documentation |
| fanotify performance on busy Linux servers | The module becomes the reason a server is slow | J2.3 measured limits before defaults are set |
| Network blocking is on and enforcing by default | A bad indicator list could cut a customer off | Self-lockout protection already exists; add an emergency central switch and make it findable |
| Signature mirror not deployed at a customer | Endpoints fall back to the public network or go stale | Fallback already built; make staleness visible in the console |
| The endpoint work sits unmerged | Drift against `release/v12.0.0` grows, and one machine holds the only copy | Phase 0 merges first; a private backup already exists |
| Behavioural telemetry volume | It can crowd out security events and inflate storage | Sensor switch already exists; measure in H6 and set a sane default |
