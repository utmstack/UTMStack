# UTMStack EDR — Phase 1 (Windows) Design & Implementation Spec

**Status:** Approved design, ready for implementation planning
**Date:** 2026-07-01
**Target agent:** `github.com/utmstack/UTMStack` branch `release/v12.0.0`, `agent/` (Go 1.25.5)
**Source requirements:** `EDR/UTMStack_Endpoint_Detection_Engineering_Plan.docx` v3.0 (§ references below point to it)

---

## 1. Purpose & scope

Build the complete **Phase 1** capability from the engineering plan, **Windows only**, as a self-contained module the UTMStack agent orchestrates: detect known malware on the endpoint, quarantine it, **terminate malicious processes — including one the user has already launched** — and block malicious scripts/macros/fileless content pre-execution via AMSI.

### In scope (all of Phase 1, Windows)
1. Engine host — clamd + freshclam (managed)
2. File-arrival watcher — whole-volume NTFS USN journal
3. Process-creation watcher — ETW `Microsoft-Windows-Kernel-Process` (WMI fallback)
4. Scan orchestrator + file↔process correlation
5. Verdict cache — dedicated SQLite
6. Responder — quarantine + process-tree kill + suspend-on-launch
7. AMSI provider — native `IAntimalwareProvider` COM DLL
8. Behavioral forwarder (agent-side telemetry: process-creation + PowerShell script-block logs + AMSI content)
9. Event emitter (via agent relay)
10. Config & feed updater (freshclam scheduling, exclusions, fail-open/closed, suspend toggle, quarantine retention)
11. Agent orchestration — enable / disable / status / install / uninstall

### Out of scope (explicitly deferred)
- **§4.8 Allowlisting manager (WDAC)** — Phase 2.
- All **Linux** components (scheduled scan, `clamonacc`, fapolicyd, Linux responder).
- Platform-side **Sigma correlation rules** and the correlation parser for the EDR data type — these are UTMStack platform work items, tracked here only as external dependencies (§12).
- No custom **kernel driver** at any point (design invariant).

---

## 2. Objective & guiding constraints

- **Single scan engine:** ClamAV (`clamd`) for all file and in-memory buffer scanning, run as a **separate managed process** invoked over a local socket — never statically linked — to keep ClamAV's GPLv2 clear of the agent's AGPL-3.0 (§11 licensing).
- **No kernel driver.** Phase 1 is user-mode only. The process kill relies on privilege, not a driver.
- **Privileged service (precondition, not obstacle).** The EDR module installs as a **Windows SYSTEM service** (highest local privilege). SYSTEM already holds the rights to terminate a process in any user's session via `OpenProcess(PROCESS_TERMINATE)` + `TerminateProcess` — **no driver needed**. Privilege is a precondition of the design; the *obstacle* Phase 1 solves is *knowing a process exists to kill*, which the process-creation watcher provides.
- **Quarantine, not delete.** Detections move to a protected, non-executable store with a restore path; never auto-delete.
- **Go-first.** Everything is implemented in Go except the single unavoidable exception: the AMSI provider must be a native (C/C++) in-process COM server (§4.7). The Go module owns, ships, registers, and drives it.
- **Honest capability claims.** The file/process path is **detect-and-kill after the fact** — it mitigates but cannot guarantee prevention (see §10). AMSI is a hard pre-execution block for the content submitters send it. A strict pre-execution gate for unknown EXEs is Phase 2 (WDAC).
- **Branding / labeling (mandatory).** Every emitted **event and log line is branded "UTMStack EDR."** ClamAV / clamd is **never named** in any event field or surfaced log message. See §8.

---

## 3. High-level architecture

The EDR is an **independent module** the agent controls — not code fused into the agent. It ships as its own Go binary/service plus one minimal native COM DLL.

```
┌────────────────────────────┐    controls: install / enable / disable /    ┌──────────────────────────────────────────┐
│  UTMStack Agent (v12.0.0)  │    status / uninstall  (shared/svc + edr.json) │  UTMStack EDR module  (SYSTEM service)    │
│                            │───────────────────────────────────────────────▶│  utmstack_edr.exe                          │
│  • ships/updates EDR via    │                                                │  ┌──────────────────────────────────────┐  │
│    Dependency downloader    │                                                │  │ Supervisor (lifecycle, config, log)   │  │
│  • starts/stops SYSTEM svc   │                                                │  ├──────────────────────────────────────┤  │
│  • cmd verbs: enable-edr,    │                                                │  │ Engine host  → clamd + freshclam      │  │
│    disable-edr, edr-status   │                                                │  │   (child proc, localhost socket)      │  │
│                             │                                                │  │ File watcher → USN journal (volumes)  │  │
│  • EDRRelay goroutine  ◀────┼────── events (durable disk spool) ─────────────┼──│ Process watcher → ETW Kernel-Process  │  │
│    drains spool → LogQueue   │                                                │  │ Orchestrator → correlate file↔proc    │  │
│  • LogQueue → platform       │                                                │  │ Verdict cache → edr.db (own SQLite)   │  │
│    (existing at-least-once)  │                                                │  │ Responder → quarantine + tree-kill    │  │
└────────────────────────────┘                                                │  │           + suspend-on-launch         │  │
                                                                               │  │ Behavioral forwarder (telemetry)      │  │
      ┌──────────────────────────────┐    Scan(buffer)                         │  │ Event forwarder → spool               │  │
      │ AMSI provider (native COM DLL)│◀──── loaded by PowerShell/WScript/      │  └───────────────┬──────────────────────┘  │
      │ utmstack_amsi.dll             │      Office/.NET; forwards buffer  ─────┼──────────────────┘ (local pipe)          │
      └──────────────────────────────┘      to EDR service → verdict           └──────────────────────────────────────────┘
```

**Data flow (one line).** watcher → orchestrator → verdict cache → clamd → verdict → responder (quarantine / kill / suspend) → event forwarder → spool → agent relay → platform. The **process watcher runs in parallel** so the orchestrator can correlate a launch with a file verdict and act on an already-running process.

**Reuse from the agent (plumbing, not data stores):** the `dependency` downloader (ships the EDR binary like the existing `updater`), `shared/` helpers (`fs`, `http`, `exec`, `svc`, `archive`), the cobra `cmd/` pattern, the kardianos service-install pattern, config/path conventions, and the logger. The EDR does **not** share the agent's `logs.db` (isolation — §5.4).

---

## 4. Component specifications

### 4.1 Engine host — clamd + freshclam
**Responsibility.** Provide malware verdicts for files and in-memory buffers with current signatures.
- On first run, ensure clamd/freshclam binaries + config are present; download the archive via `shared/http.DownloadFile` from the UTMStack server dependencies endpoint (`config.DependUrl` / `DependenciesPort`) and unzip via `shared/archive` if missing.
- **Spawn clamd as a supervised child process** bound to a **localhost TCP socket** (Windows lacks robust unix sockets); health-check periodically; restart on failure. clamd is fully owned by the EDR service and is killed on EDR stop.
- Scan files by path (`SCAN` / `CONTSCAN`) and in-memory buffers (`INSTREAM`, for AMSI) over the clamd socket.
- Run `freshclam` on a schedule (default several times/day; interval configurable) for signature updates from Cisco Talos; support a local mirror URL for scale.
- **Sizing note:** clamd loads its full signature set (>1 GB RAM) — document minimum host memory in deployment docs. Exclude pseudo/temporary and trusted paths from scanning (§4.10).
- **Labeling:** all engine log/status text says "UTMStack EDR engine" — never "clamav"/"clamd".

### 4.2 File-arrival watcher — whole-volume USN journal
**Responsibility.** Detect every new or changed file the instant it lands and hand it to the orchestrator.
- For each fixed NTFS volume, open the volume handle and read the **USN change journal** via `DeviceIoControl` with `FSCTL_QUERY_USN_JOURNAL` / `FSCTL_READ_USN_JOURNAL` (`golang.org/x/sys/windows`), filtering for file-create / data-extend / rename / close records.
- **Persist the last processed USN** (and journal ID) per volume in `edr.db` so processing survives restarts; on journal ID change (journal recreated), resync.
- Preferred over per-directory `ReadDirectoryChangesW` because it covers a whole volume and does not drop events under load.
- Emit file events `{path, operation, timestamp, volume}` onto the scan queue.
- Apply exclusions (§4.10) before enqueue.

### 4.3 Process-creation watcher — the kill enabler (double duty)
**Responsibility.** Provide a real-time feed of every process start so the orchestrator can act on code the user has already launched. Without this, a malicious verdict tells us the *file* is bad but not whether it is *running*.
- Subscribe to the **`Microsoft-Windows-Kernel-Process`** ETW provider via a real-time session (start trace → enable provider → consume) for process start/stop, yielding **PID, parent PID, image path, command line, session, user**. **WMI `Win32_ProcessStartTrace`** is a simpler fallback with higher latency (wider race).
- Maintain an in-memory `PID → {PPID, image, start-time, session, user}` map for tree reconstruction; disambiguate recycled PIDs by start time.
- **Single source, two subscribers (double duty):** the same event stream feeds (a) the **orchestrator/responder** (launch correlation, expedite, suspend, tree-kill) and (b) the **behavioral forwarder** (§4.8). One ETW session, no duplication.
- Output: process-start events to the orchestrator and behavioral forwarder.

### 4.4 Scan orchestrator + correlation
**Responsibility.** Decide what to scan, in what priority, and what to do with each verdict.
- Consume file events (§4.2) and process-start events (§4.3); look up the verdict cache by hash; dispatch unknowns to clamd.
- **Priority & correlation:** when a process-start references an image whose verdict is unknown or pending, **expedite** that image's scan and — per policy — instruct the responder to **suspend** the process until the verdict returns (§4.6).
- Maintain a `image-path/hash → pending-PIDs` correlation table so that when a scan completes malicious, all running PIDs launched from that image are killed.
- Manage queues, worker concurrency, and backpressure so scanning never stalls the endpoint.

### 4.5 Verdict cache
**Responsibility.** Avoid re-scanning unchanged, already-known files (the inventory).
- Key on **SHA-256**. Skip re-scan when the hash is unchanged and the entry's signature-DB version matches the current one.
- On signature update, mark clean entries **stale**; re-scan lazily on next access and via a low-priority sweep. **Malicious verdicts persist** for blocking/reporting.
- Store in the EDR's **own** SQLite database `edr.db` (gorm + `glebarez/sqlite`, same libraries the agent uses, **separate file and connection** — §5.4). Schema in §5.2.

### 4.6 Responder — quarantine, process-tree kill & suspend
**Responsibility.** Act on a malicious verdict: contain the file and stop the code, including a process the user already started.
- **Kill (the core ask).** The EDR service runs as **SYSTEM**, which holds the rights to terminate a process in a user session — no driver. `OpenProcess(PROCESS_TERMINATE)` + `TerminateProcess`.
- **Kill the whole process tree**, not one PID — malware spawns children. Build the tree from the process watcher's PID/PPID map, terminate **leaves-first**.
- **Suspend-on-launch (optional, race-narrowing).** When the watcher reports a launch of an unknown image, suspend it while the expedited scan runs (`NtSuspendProcess` from `ntdll`, or suspend every thread via `Thread32First`/`SuspendThread`), then **resume if clean, kill if malicious**. Guarded by a scan timeout after which the default (fail-open) resumes the process to avoid hanging launches. Configurable on/off (§4.10).
- **Quarantine.** Move the file to a protected, non-executable store with metadata (original path, hash, detection, timestamp) and a restore workflow (`utmstack_edr restore <id>`). **Never auto-delete.**
- **Honest limit (documented, load-bearing).** This is detect-and-kill *after the fact* — ETW notifications arrive after the process's first thread has begun, so some code runs before the kill/suspend. Killing removes running code but does not guarantee undoing damage already done (injection, file changes, persistence). You cannot hold a launch to scan-then-decide from user mode — that is Phase 2 allowlisting.

### 4.7 AMSI provider (native C/C++ COM DLL — the sole Go exception)
**Responsibility.** Hard, real-time, pre-execution block of the script / macro / fileless class.
- Implement a COM in-process server exporting `DllGetClassObject` / `DllCanUnloadNow`, a class factory, and **`IAntimalwareProvider`** (`Scan`, `CloseSession`, `DisplayName`), registered under `HKLM\SOFTWARE\Microsoft\AMSI\Providers\{CLSID}` and the CLSID's InprocServer32. Windows submits dynamic content (PowerShell, WSH, JScript/VBScript, Office VBA, .NET) to `Scan` before execution.
- **Kept minimal.** The DLL's only job: marshal the submitted buffer to the **EDR service** over a local named pipe, receive the verdict, and return `AMSI_RESULT_DETECTED` (block) or a clean result (allow). Routing through the EDR service (rather than clamd directly) centralizes eventing, labeling, and the fail-open/closed policy in Go and keeps the native surface tiny. The EDR service scans the buffer via clamd `INSTREAM` and **emits the `amsi` event** (labeled UTMStack EDR).
- **Availability policy (configurable):** if the EDR service/clamd is unreachable, **fail-open** (allow + alert) by default to avoid breaking script hosts; **fail-closed** (block) optional.
- **Signing:** ordinary **Authenticode** code signing of the DLL (and the EDR binary) — *not* the kernel-driver EV/attestation path.
- **Limits:** AMSI sees only content submitters send (a plain `.exe` launch does not go through it), and the provider runs inside the host process, so it is not tamper-proof.

### 4.8 Behavioral forwarder (agent-side telemetry)
**Responsibility.** A light behavioral layer that reuses the platform, with no embedded engine.
- Collect **process-creation events** (the §4.3 feed — double duty), **PowerShell script-block logs** (`Microsoft-Windows-PowerShell/Operational` event 4104, via the Windows event log / ETW), and the **deobfuscated content AMSI already surfaces**; normalize and forward as UTMStack EDR telemetry events.
- Detection logic lives in a small set of **Sigma-style correlation rules** on the UTMStack platform (Office→PowerShell, encoded commands, LOLBin chains) with SOAR responses attached — **out of agent scope** (§12). A true runtime behavioral engine (eBPF / kernel sensor) is out of scope.

### 4.9 Event forwarder & agent relay
**Responsibility.** Deliver normalized events (§5.1) to the platform through the agent.
- The EDR appends each normalized event as newline-delimited JSON to a **durable disk spool** (`edr-spool/events.ndjson`, size-capped + rotated) under the shared install directory.
- The agent runs an **`EDRRelay`** goroutine (`p.goSafe("EDRRelay", …)` in `serv/service.go`) that tails/drains the spool, validates each line (`entities.ValidateString`), wraps it as `*plugins.Log{DataType: "utmstack_edr", Raw: <json>, DataSource: <host>}`, and pushes to the existing `agent.LogQueue` — reusing the agent's proven at-least-once delivery, persistence, and acks. The agent advances a persisted spool offset only after enqueue.
- **Why a spool, not a socket:** disk-durable (events survive an agent restart, brief agent-down window, or EDR restart), no listener/port management, loose coupling (matches the isolation requirement). Latency is a ~1s drain tick — acceptable for this event class. If the agent is down, the spool grows to its cap then drops oldest with an explicit logged warning (no silent loss).

### 4.10 Config & feed updater
**Responsibility.** All operational configuration and feed scheduling.
- `edr.json` (read by EDR, writable by the agent / central management) with **defaults**: `watch_volumes` = all fixed NTFS volumes, `exclusions` = pseudo/temp + trusted caches, `fail_mode` = `open`, `suspend_on_launch` = `true` (with a `suspend_timeout` = `5s` after which fail-open resumes), `quarantine_dir` = `<install>/quarantine`, `quarantine_retention` = 30 days, `clamd_addr` = `127.0.0.1:3310`, `scan_concurrency` = `min(4, NumCPU)`, `sig_update_interval` = `4h`, `scoped_realtime` toggles = off (whole-volume USN is the default coverage).
- freshclam scheduling; signature/config distribution from central management; quarantine store lifecycle and retention.
- The EDR reconciles operational config changes on a short poll (v12 dropped `fsnotify`; a poll of `edr.json` mtime is sufficient and dependency-free). On/off is handled by the agent starting/stopping the service, not by the poll.

### 4.11 Agent orchestration
- **CLI verbs** in `cmd/` (self-register via `func init(){ rootCmd.AddCommand(...) }`, guarded by `requireInstalled`, mirroring `change_retention.go`):
  - `enable-edr` → ensure EDR installed (dependency-shipped) + **start** the SYSTEM service via `shared/svc`; set state.
  - `disable-edr` → **stop** the service via `shared/svc`; set state.
  - `edr-status` → read EDR `status.json`; print running state / engine health / signature-DB version / last scan / quarantine count.
- **Dependency entry** in `dependency/deps_windows_amd64.go` ships/updates `utmstack_edr.exe` (and registers/installs it) like the existing `updater` (`Configure` = `utmstack_edr install`; `Uninstall` = `utmstack_edr uninstall`; `Critical:false`).
- **Remote control for free:** because enable/disable are CLI verbs and the platform already pushes shell commands via `IncidentResponseStream`, the SOC can enable/disable EDR remotely with no extra plumbing.

---

## 5. Data models

### 5.1 Normalized event schema (branded UTMStack EDR)
Emitted as JSON in `plugins.Log.Raw`, `DataType = "utmstack_edr"`.

| Field | Description |
|---|---|
| `timestamp` / `host_id` / `os` | Event time, asset id, `windows` |
| `product` | Constant `"UTMStack EDR"` |
| `source` | `file_watcher` \| `process_watcher` \| `amsi` \| `edr_engine` (**never `clamav`**) |
| `action` | `detected` \| `quarantined` \| `killed` \| `suspended` \| `resumed` \| `blocked` \| `allowed` \| `scan_complete` \| `restored` \| `health` |
| `process` | PID, PPID, image path, command line, session, user (where applicable) |
| `object_path` / `file_hash` | Target file path and SHA-256 |
| `verdict` / `signature` | `clean` \| `malicious` \| `unknown`; signature/rule name |
| `engine` | Constant `"UTMStack EDR"` (the detecting engine label) |
| `policy_mode` / `severity` | e.g. detect/enforce; normalized severity |

### 5.2 Verdict cache record (`edr.db`)
| Field | Description |
|---|---|
| `sha256` (key) | File content hash |
| `verdict` | `clean` \| `malicious` \| `unknown` |
| `signature` | Detection name, if malicious |
| `sigdb_version` | Signature-DB version at scan time |
| `first_seen` / `last_seen` | Timestamps |
| `stale` | True when signatures update; triggers lazy re-scan |

### 5.3 Quarantine record (`edr.db`)
| Field | Description |
|---|---|
| `quarantine_id` | Unique id / stored filename in the protected store |
| `original_path` / `sha256` | Origin and hash |
| `detection` / `engine` | Signature name; `engine = "UTMStack EDR"` |
| `quarantined_at` / `restorable` | Timestamp and restore eligibility |

### 5.4 Storage isolation
The EDR uses its **own** SQLite file `edr.db` and connection — **not** the agent's `logs.db`. Rationale (per requirement): the log DB has a different write pattern and lifecycle (the agent runs retention/vacuum sweeps against it), and mixing EDR verdict/quarantine data would couple two failure domains and complicate troubleshooting. Separate DB = isolated failure domain, independent retention, regenerable cache (worst case: rebuild).

### 5.5 On-disk layout (shared install dir)
```
<agent-install>/
  utmstack_edr.exe        # EDR service binary (dependency-shipped)
  utmstack_amsi.dll       # native AMSI provider (signed)
  edr.json                # operational config
  status.json             # EDR-written status
  edr.db                  # EDR's own SQLite (verdict cache + quarantine meta + USN cursors)
  edr-spool/events.ndjson # durable event spool (EDR writes, agent drains)
  quarantine/             # protected, non-executable store
  engine/                 # clamd/freshclam binaries + signature DB
```

---

## 6. Key runtime flows

### 6.1 File arrival & scan (file not yet run)
File lands → USN watcher fires → orchestrator hashes it, checks the verdict cache (known-clean + unchanged + current sig-DB → skip) → unknown dispatched to clamd → verdict cached → **malicious → responder quarantines the file** → event emitted.

### 6.2 Race: user executes the file before the scan finishes (the core scenario)
1. `a.exe` lands → USN watcher queues it for scanning.
2. User launches it → the **process watcher fires instantly** with PID, PPID, image path.
3. Orchestrator sees a process started from an unknown/unscanned image → records the PID, **expedites that image's scan**, and (if `suspend_on_launch`) instructs the responder to **suspend** the process.
4. Scan returns **malicious** → responder **terminates the process and its whole child tree** (leaves-first) and quarantines the file; events emitted. If **clean** → the process is resumed / left running.

*Design note.* The process watcher is what makes this possible; without it, the agent would know the file is bad but have no handle on the running process. The residual ms-level race (code runs before the event arrives) is inherent to user mode — Phase 2 allowlisting removes it for unknown EXEs by blocking the launch outright.

### 6.3 Script / fileless execution (AMSI)
Script host is about to execute dynamic content → Windows submits the content to the AMSI provider's `Scan` → provider forwards the buffer to the EDR service → EDR scans via clamd `INSTREAM`, emits an `amsi` event, returns the verdict → on malicious the provider returns `AMSI_RESULT_DETECTED` and the host refuses to execute — a hard, pre-execution block, no process to kill.

### 6.4 Signature update → cache invalidation → rescan
freshclam updates signatures → clean cache entries marked `stale` → stale files re-scanned lazily on next access and in a low-priority sweep; malicious verdicts persist.

### 6.5 Enable / disable lifecycle
`enable-edr` → agent starts the SYSTEM service → EDR `run`: ensure engine → spawn clamd + schedule freshclam → migrate `edr.db` → start USN watcher(s) + process watcher + orchestrator workers + behavioral forwarder → register AMSI DLL → write `status.json`.
`disable-edr` → agent stops the service → EDR graceful stop: unregister/quiesce AMSI, stop watchers, drain scanner, **kill clamd child**, flush spool, close `edr.db`, update status.

---

## 7. Error handling & resilience

- **clamd unreachable/unhealthy:** scanner keeps files `pending`, retries with backoff, emits a `health` event; EDR restarts the clamd child. A missing verdict defers (never lets malware run *because of* EDR, never breaks the host).
- **freshclam failure:** keep existing signatures, alert, retry on interval.
- **USN journal overflow / journal recreated:** detect journal-ID change, resync from current USN, and trigger a scoped rescan sweep so nothing is silently missed.
- **ETW session drop:** auto-restart the session; fall back to WMI if ETW is unavailable; log the degraded state.
- **Spool back-pressure (agent down):** spool grows to a configured cap, then drops oldest with an explicit logged warning; resumes on agent return. No silent loss.
- **`edr.db` errors/corruption:** isolated to EDR; cache is regenerable → rebuild. No impact on the agent's `logs.db`.
- **Crash isolation:** EDR is a separate process — a crash cannot take down the agent; the Windows SCM restarts the EDR service; EDR restarts clamd.
- **AMSI EDR-unreachable:** apply configured `fail_mode` (default fail-open + alert).
- **Quarantine move fails (locked file):** retry with backoff, log; **never delete**.

---

## 8. Labeling & branding rules (mandatory, cross-cutting)

- All emitted **events**: `DataType = "utmstack_edr"`; `product`/`engine = "UTMStack EDR"`; `source ∈ {file_watcher, process_watcher, amsi, edr_engine}`. The string `clamav`/`clamd` must **never** appear in an event.
- All **surfaced logs** (EDR service log, `status.json`, `edr-status` output, agent relay logs): refer to "UTMStack EDR" / "UTMStack EDR engine" / "scan engine" — never "clamav"/"clamd".
- clamd/ClamAV naming is permitted **only** inside source code identifiers and low-level debug traces that are not surfaced; production log lines must not name it.
- **Acceptance:** an automated test greps emitted events and the service log for `clamav`/`clamd` (case-insensitive) and fails if found (§9).

---

## 9. Acceptance tests & definition of done

Run on the Windows test VM (Parallels, `10.211.55.12`, `ricardovald1d15\atlas`). Verify **final outputs** (quarantine store contents, events at the platform), not just logs.

| Capability | Acceptance test |
|---|---|
| Signature detection | An EICAR file on disk is detected and quarantined; event reaches the platform. |
| Scan on arrival (USN) | A freshly written test file is scanned within target latency without a manual trigger. |
| Verdict cache | Re-writing an unchanged known-clean file does not re-scan; a signature update marks it stale and it re-scans. |
| **Process kill** | A running test process **and a child it spawns** are terminated on malicious verdict; both PIDs are gone. |
| **Race (run-before-scan)** | Launching a malicious test file immediately after it lands results in the process (and tree) being killed and the file quarantined. |
| **Suspend-on-launch** | With suspend enabled, an unknown image is suspended at launch and resumed if clean / killed if malicious. |
| **AMSI block** | The AMSI test string (or an EICAR-style script) is blocked before execution by the provider. |
| Behavioral forwarding | A PowerShell encoded-command / Office→PowerShell chain surfaces process-creation + script-block telemetry at the platform. |
| False-positive handling | A quarantined benign file restores from the protected store. |
| Orchestration | `enable-edr` launches it; `disable-edr` stops it (clamd child gone too); `edr-status` reports accurately. |
| **Labeling** | No emitted event or surfaced log contains `clamav`/`clamd` (automated grep). |
| Performance | With exclusions applied, steady-state CPU/I/O stays within target on the reference host. |

---

## 10. Risks, caveats & honest limits

- **File detect-and-kill is after the fact.** Phase 1 mitigates but cannot guarantee prevention of unknown EXEs; Phase 2 WDAC is the true gate.
- **Suspend narrows, not closes, the race.** The launch event arrives after the first thread starts.
- **AMSI limits.** Submitter-sent content only, in-process tamper risk, not executables.
- **Detection ceiling & performance.** Signatures miss novel/evasive threats; scope real-time watching to high-risk paths and exclude noisy directories; clamd needs >1 GB RAM.
- **Privileged tamper.** A local administrator can stop the service — a control against unknown code, not a compromised admin.

---

## 11. Dependencies & environment

- **ClamAV:** clamd, clamdscan, freshclam (current stable); Windows build bundled/managed by the EDR module. Signature-DB memory (>1 GB) sized into host requirements.
- **Windows build:** Go 1.25.5 for the EDR module and agent components; **C/C++ toolchain** for the AMSI COM DLL; access to ETW/WMI, USN journal, `ntdll` suspend APIs; installs as a SYSTEM service. **No WDK / driver toolchain.**
- **Code signing:** Authenticode-sign the EDR binary and the AMSI DLL with a standard code-signing certificate.
- **Platform:** the existing agent event channel (`LogQueue` → correlation) for telemetry; central management for feeds/config distribution.

---

## 12. External dependencies (other teams / out of this spec)
1. **Server hosting** of per-platform ClamAV builds and the `utmstack_edr.exe` binary at the dependencies endpoint (same model as `updater`/beats today).
2. **Correlation parser** on the UTMStack platform for the `utmstack_edr` data type (normalize §5.1 events).
3. **Sigma correlation rules** + SOAR responses for the behavioral telemetry (Office→PowerShell, encoded commands, LOLBin chains).
4. **Code-signing certificate** provisioning for CI.

---

## 13. Licensing
- clamd runs as a **separate process over its socket — not linked** — keeping ClamAV's GPLv2 clear of the agent's AGPL-3.0.
- The AMSI provider and EDR binary use ordinary **Authenticode** signing — no kernel-driver signing anywhere.
- *Not legal advice; confirm component licensing and combination with qualified counsel before release.*

---

## 14. Work breakdown & sequencing (each step independently verifiable)
1. Scaffold `utmstack_edr`: SYSTEM service `install`/`uninstall`/`run`, config load, logger, `status.json` (reuse `shared`).
2. Engine host: download + supervise clamd/freshclam; socket client → prove EICAR verdict via `INSTREAM`/`SCAN`.
3. Verdict cache: `edr.db` model + lookup/store + sig-DB versioning.
4. File watcher: whole-volume USN journal reader + persisted cursors + exclusions.
5. Orchestrator/scanner: wire watcher → hash → cache → clamd → verdict; worker pool + backpressure.
6. Quarantine + restore workflow.
7. Process-creation watcher: ETW Kernel-Process (WMI fallback) + PID/PPID map.
8. Responder: process-tree kill + suspend-on-launch + correlation with pending scans (the §6.2 flow).
9. AMSI provider: native COM DLL + registration + EDR-service pipe + `amsi` eventing + fail-open/closed.
10. Behavioral forwarder: process-creation + PowerShell 4104 + AMSI content → telemetry events.
11. Event forwarder → spool; agent `EDRRelay` drain → `LogQueue`.
12. Agent orchestration: dependency entry + `enable-edr`/`disable-edr`/`edr-status` verbs.
13. Config & feed updater: freshclam scheduling, exclusions, retention, config poll.
14. End-to-end acceptance (§9) on the VM, including the labeling grep.
