# UTMStack EDR — Project Guide (CLAUDE.md)

This workspace builds a **native endpoint malware-detection & response (EDR) module** for the UTMStack agent: detect known malware, quarantine it, terminate malicious processes (including ones the user already launched), and block malicious scripts pre-execution via AMSI. **Phase 1 is Windows-only.**

Authoritative requirements: `UTMStack_Endpoint_Detection_Engineering_Plan.docx` (v3.0). Section references below (`§4.x`) point to it.

---

## Where things live

| Thing | Path |
|---|---|
| **Source spec (docx)** | `UTMStack_Endpoint_Detection_Engineering_Plan.docx` |
| **Approved design spec** | `docs/superpowers/specs/2026-07-01-edr-phase1-windows-design.md` |
| **Implementation plans (1–4)** | `docs/superpowers/plans/2026-07-01-edr-phase1-plan{1..4}-*.md` |
| **Ransomware Guard spec + plan** | `docs/superpowers/specs/2026-07-02-edr-ransomware-guard-design.md` · `docs/superpowers/plans/2026-07-03-edr-ransomware-guard-plan1-core.md` |
| **Signature-update spec + (deferred) mirror plan** | `docs/superpowers/specs/2026-07-02-edr-signature-update-mechanism.md` · `docs/superpowers/plans/2026-07-02-edr-signature-mirror-plan.md` |
| **Master-server `edr` container spec + plan** | `docs/superpowers/specs/2026-07-07-edr-server-container-design.md` · `docs/superpowers/plans/2026-07-07-edr-server-container-plan.md` (gap analysis: `docs/2026-07-07-edr-remaining-work.md`) |
| **Working code repo** | `github.com/utmstack/OpenEDR` branch `main` — full private working copy of the UTMStack v12 tree. Local full clone: `/Users/atlas/UTMStack/v12`. Older sparse clone (agent only): `utmstack-v12/` |
| **EDR module code** | `utmstack-v12/agent/edr/` (its own Go binary `utmstack_edr`) |
| **Agent-side control/relay** | `utmstack-v12/agent/{cmd,agent,dependency,serv}/` (small additions only) |
| **Build output** | `utmstack-v12/agent/dist/` |

All Go commands run from **`utmstack-v12/agent/`** (Go module `github.com/utmstack/UTMStack/agent`, Go 1.25.5).

---

## Architecture (how it's built)

The EDR is a **self-contained module the agent orchestrates** — not code fused into the agent.

- **`utmstack_edr(.exe)`** — its own Go binary, installed as a **SYSTEM** Windows service. Owns *all* detection logic: engine host (clamd), USN file watcher, scan orchestrator, verdict cache, quarantine, process watcher + responder, AMSI pipe server, ransomware guard (canary + ETW file feed + scorer). Also owns its own config surface: unified `allowlist` + per-detector `sensors`, a local admin CLI, and live `edr.json` reload. Templated on the existing `agent/updater` sub-binary.
- **`utmstack_amsi.dll`** — the one non-Go artifact: a minimal native C/C++ `IAntimalwareProvider` COM DLL (Plan 4) that forwards script buffers to the EDR service over a named pipe.
- **The agent only controls the module**: launch / enable / disable / status / uninstall (via `shared/svc` + `edr.json`) and **relays its events**.
- **Event path = agent relay.** The EDR appends normalized JSON events to a durable disk **spool** (`edr-spool/events.ndjson`); the agent's `EDRRelay` goroutine drains it into the existing `agent.LogQueue` → platform (at-least-once + acks). The EDR never talks to the platform directly.
- **Own SQLite** (`edr.db`) — separate from the agent's `logs.db` (isolated failure domain).
- **clamd** runs as a child process the EDR supervises (localhost TCP socket, INSTREAM protocol). Kept out-of-process for GPLv2-vs-AGPL licensing.
- Shared install dir (`shared/fs.GetExecutablePath()`) holds `edr.json`, `status.json`, `edr.db`, `edr-spool/`, `quarantine/`, `engine/`.

Reuses the agent's `dependency` downloader (ships the EDR binary like `updater`) and `shared/` helpers (`fs`, `http`, `exec`, `svc`, `archive`, `logger`). **Do not** reuse the agent's data stores.

---

## Non-negotiable conventions

1. **BRANDING (critical):** Every emitted **event** and every **surfaced log/status string** says **"UTMStack EDR"**. The strings `clamav`/`clamd` must **never** appear in an event field or a surfaced log line. Internal Go identifiers (`ClamdAddr`, `clamd.exe`, package `engine`) are fine.
   - Event contract: `DataType = "utmstack_edr"`; `source ∈ {file_watcher, process_watcher, amsi, edr_engine, behavioral}`; `product`/`engine = "UTMStack EDR"`.
   - Any engine version banner (which contains "ClamAV") must be sanitized before it reaches `status.json` — see `service.sigDBVersion()`.
   - There is a labeling audit (grep for `clam`) in every plan's acceptance; keep it green.
2. **Go-first.** All code is Go. The **only** exception is the AMSI COM DLL (C/C++, Plan 4) — a hard technology requirement. No other language without an equally hard reason.
3. **Own `edr.db`**, never the agent's `logs.db`.
4. **Quarantine, never delete.** Detections move to a protected store with a restore path.
5. **No kernel driver, ever.** Process kill is `OpenProcess(PROCESS_TERMINATE)` + `TerminateProcess` (SYSTEM privilege makes this work with no driver). Phase 2 blocking uses OS-native WDAC.
6. **SYSTEM service.** The EDR installs as a Windows SYSTEM service.
7. **Honest capability claims.** The file/process path is detect-and-kill *after the fact* (residual ms-level race on process start; suspend narrows it). AMSI sees only submitter-sent content. State these limits; don't oversell.
8. **Self-containment.** New EDR logic goes under `agent/edr/`. The only agent-package additions are the thin control/relay surface (`cmd/*_edr.go`, `agent/edr_relay.go`, `dependency/edr.go`, one line in `serv/service.go`).

---

## Build & test

Supported Windows architectures: **`amd64` and `arm64`** (the agent has no `deps_windows_386.go`, so 386 is out of scope). All EDR code is architecture-portable; build and vet **both** arches.

From `utmstack-v12/agent/`:

```bash
# Cross-build EDR + agent for BOTH supported Windows architectures
for arch in amd64 arm64; do
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_edr_windows_$arch.exe ./edr
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_agent_windows_$arch.exe .
  GOOS=windows GOARCH=$arch go build ./...      # whole module (windows-tagged files)
  GOOS=windows GOARCH=$arch go vet ./edr/...
done

# Run all EDR + agent unit tests (host = macOS/darwin)
go test ./edr/... ./agent/
go vet ./edr/...

# Labeling audit — must return nothing
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*clam' edr/
```

The EDR dependency entry must be registered in **both** `dependency/deps_windows_amd64.go` and `dependency/deps_windows_arm64.go`; the shared helpers live in the untagged `dependency/edr.go`. Binary names carry the arch via `runtime.GOARCH` (`EDRFile()`).

**Testing model:** pure-Go logic is TDD'd with `go test` on the dev host (macOS). Windows-native pieces (USN journal I/O, ETW/WMI, kill/suspend syscalls, the COM DLL, HKLM registration, event-log reading) **cannot** run on macOS — they are written behind `//go:build windows` with `!windows` stubs so the module builds everywhere, and are verified by compile + run on the **test VM**. Each plan's final task is VM acceptance.

**Test VM:** Windows (Parallels) at `10.211.55.12`, login `ricardovald1d15\atlas`. Provision ClamAV manually on `127.0.0.1:3310` for testing (the server does not yet host clamd — see External deps). Verify **final outputs** (quarantine store, events at the platform), not just logs.

---

## EDR package map (`agent/edr/`)

| Package | Responsibility | Plan |
|---|---|---|
| `config` | `edr.json` load/save + defaults; unified `allowlist` + `sensors` blocks (legacy-field migration); path constants; `ServiceName` | 1 |
| `event` | Normalized event model + `Spool` (durable ndjson); branding | 1 |
| `cache` | Own SQLite `edr.db`: `VerdictRecord`, `QuarantineRecord` (+purge), `USNCursor`, ransomware incidents; staleness | 1–2 |
| `engine` | clamd INSTREAM client + child supervision; **generates `clamd.conf`/`freshclam.conf`** (detection engines, heuristic alerts, scan limits) with **host-adaptive tuning** (`DeriveTuning` tiers + `SafeDefaultTuning` failsafe) per `UTMStack_ClamAV_Configuration_Guide.docx` | 1 |
| `scanner` | hash → cache → engine → verdict; malicious → quarantine + event | 1–2 |
| `service` | kardianos SYSTEM service, `startPipeline` wiring, `status.json`, sensor gating, live config reload, quarantine-retention sweep | 1–2 |
| `quarantine` | move-to-store (never delete) + restore + `Purge`/`PurgeExpired` | 2 |
| `orchestrator` | worker pool + backpressure + hot-swappable `Excluder` (path/process) | 2 |
| `watcher` | USN_RECORD parser + whole-volume USN journal reader (parent-dir exclusion, throttled cursor) | 2 |
| `feed` | freshclam scheduler + cache invalidation (official CDN default; `signature_mirror` → private mirror) | 2 |
| `proctable` | PID→info map + leaves-first tree reconstruction | 3 ✅ |
| `responder` | kill-tree (+ live-snapshot descendants, kill-once dedup) + suspend/resume | 3 ✅ |
| `guard` | launch → suspend → scan → kill/resume (§6.2 race); never suspends OS images; timeout-bounded | 3 ✅ |
| `procwatch` | WMI process-creation feed + dispatch (shared by guard, ransomware T1490, behavioral) | 3 ✅ |
| `amsi` | scan core + named-pipe server + HKLM registration + native C/C++ COM DLL (`native/`) | 4 ✅ |
| `behavioral` | process + PowerShell 4104 (script-block) telemetry forwarder | 4 ✅ |
| `ransomware` | canary tripwires + T1490 command rules + ETW file feed → decaying per-PID scorer → suspend/kill/quarantine; **on by default** (`response_mode="suspend"`) | RW ✅ |

`agent/edr/main.go` = the binary entrypoint; `agent/edr/manage.go` = the local admin CLI (`config show|get|set`, `allow path|process|command add|remove|list`, `quarantine list|restore|purge`) parsed off raw `os.Args` (no cobra, so the VM-tested service-start path is untouched). Verbs: `install`/`uninstall`/`run`/`scan`/`status`/`restore`/`config`/`allow`/`quarantine`.

**To emit an event:** build an `event.Event` (branded helpers in `event/`), `ToJSON()`, `spool.Append(js)`. The agent relay does the rest. Never write platform-facing code in the EDR.

---

## Implementation status

_All four Phase-1 plans **and** the Ransomware Guard are implemented, unit-tested (17/17 `edr` packages green on darwin), cross-build + vet clean on windows/amd64 **and** arm64, labeling audit clean — and VM-validated end-to-end (incl. a full plans-1–5 holistic run + soak). Nothing is git-committed (the user drives git); all working code is intentionally uncommitted in `utmstack-v12/`._

- **Plan 1 — Foundation:** ✅ implemented + VM-validated (on-demand scan → branded event at platform; enable/disable; verdict cache; engine child-supervision).
- **Plan 2 — Auto-detect:** ✅ implemented + VM-validated (USN watcher → quarantine → branded event → platform; restore; sig-feed invalidation). **4 real bugs** found+fixed on the VM (dead OpenFileById struct, duplicate events, `\\?\` path, dead cache invalidation) + 4 more during holistic testing (handle-reuse, parent-exclusion, cursor feedback loop, cache-hit quarantine).
- **Plan 3 — Process kill:** ✅ implemented + VM-validated (WMI procwatch → guard → whole-tree kill leaves-first; suspend-on-launch). Adds `github.com/go-ole/go-ole` (WMI). Fixed a COM double-free crash, redundant-kill noise, drop-and-run survival, live-snapshot tree, and a **dangerous suspend-on-launch default** (froze the VM → now OFF, rails added).
- **Plan 4 — Scripts & telemetry:** ✅ implemented + VM-validated incl. a **real AMSI block** (native C/C++ COM DLL built with MSVC on the VM, arm64 runtime-proven; fixed a fatal factory/object refcount crash). Behavioral proc + PowerShell-4104 forwarder (fixed an engine-process branding leak). **Pending:** Authenticode-sign the DLL+binary for production; an amd64-host AMSI runtime test.
- **Ransomware Guard (v1 core):** ✅ implemented + VM-validated (canary tripwire + T1490 recovery-tampering rules → decaying scorer → suspend/kill/quarantine). Adds `github.com/0xrawsec/golang-etw` (ETW file feed; **windows-only import**). **On by default** (`ransomware.enabled=true`, `response_mode="suspend"` — reversible) as of 2026-07-06; the original plan shipped it opt-in. An explicit `ransomware` block in `edr.json` still wins (a user who set `enabled=false` stays off; only fresh/unspecified configs get the on default). Own spec+plan. Minor known issues remain (empty culprit image when WMI lags; possible duplicate `contained` event; sub-second T1490 procs miss the `WITHIN 1` WMI poll).
- **Signature auto-mirror + fallback:** 📝 **designed, not implemented.** Today freshclam defaults to the official CDN (`database.clamav.net`); setting `signature_mirror` → `PrivateMirror`. Approved next step (deferred plan in this repo): default the source to the UTMStack server's ClamAV service, auto-derived as `https://<server>:9001/private/clamav/` (reuse the 9001 dependencies tier); **probe each update cycle** and fall back to the official repos when unreachable; capture the server cert into a `CURL_CA_BUNDLE` so freshclam trusts self-signed on-prem certs (freshclam has no skip-verify). Server-side `cvdupdate` mirror is an external (platform-team) dependency.

Phase 1 (Windows) = Plans 1–4 (+ the Ransomware Guard feature). **Excluded from Phase 1:** the §4.8 WDAC allowlisting manager (Phase 2) and all Linux components.

## Runtime capabilities beyond the four plans

Implemented on top of the plans (all Go, TDD'd, both arches, labeling-clean):

- **Host-adaptive engine tuning** (`engine/tuning.go`, `profile.go`): profiles RAM/cores/free-temp → tier (`constrained`/`standard`/`server`), derives threads/queue/size-caps; static security bounds (MaxRecursion 16 / MaxFiles 10000) never scaled; `SafeDefaultTuning` failsafe on profiling failure; a viability gate skips a resident daemon on a too-small host. Generates `clamd.conf` + `freshclam.conf`; effective tuning is reported in `status.json` (§8.5). VM-validated (8 GB/4-core → Standard tier).
- **Unified allowlist** (`config.Allowlist{Paths,Processes,Commands}`, `json:"allowlist"`): the single FP-tuning surface. `Load()` migrates the deprecated scattered lists (`exclusions`, `trusted_processes`, `ransomware.command_allowlist`) into it and `Save()` stops writing them. User entries are **additive** to built-in defaults (self-exclusions; a conservative backup/imaging/DR/sync + VSS/Search trusted-process set). Boundary-aware, case/separator-insensitive matching (path subtree / basename / glob).
- **Sensor toggles** (`config.Sensors{FileWatcher,ProcessGuard,AMSI,Behavioral *bool}`, nil = ON): disable one noisy detector (esp. chatty behavioral) without disabling the module. Structural (restart to apply). The single WMI procwatch feed runs if *any* of guard / ransomware / behavioral needs it.
- **Local admin CLI** (`edr/manage.go`): `config show|get|set` (validated scalar set), `allow path|process|command …`, `quarantine list|restore|purge`. Every mutation = Load→validate→Save; `edr.json` stays hand-editable.
- **Live reload** (`service.maybeReload`, polls `edr.json` mtime every 3 s): hot-applies allowlists (`Excluder.Set`) + ransomware `response_mode`/command-allowlist (`Guard.SetPolicy`). Structural settings still need disable/enable — the reload log says which.
- **Quarantine retention** (`quarantine_retention_days`, 0 = keep forever): a 6-hourly sweep purges expired items. `quarantine purge` is the deliberate admin exception to "never auto-delete".
- **Supervision:** the engine (clamd) is restarted if it dies (it can be OOM-killed under load); the ransomware ETW feed is supervised with a `ransomware_feed_healthy` status; all pipeline goroutines run under `goSafe` panic-recovery.

## Learnings & gotchas (VM-hard-won — read before touching Windows-native code or testing)

- **VM testing is mandatory and finds real bugs.** Every USN/WMI/kill/COM/ETW path had bugs that only surfaced on the VM. Verify **final outputs** (quarantine store, events at the platform), never just logs. Details live in the auto-memory files.
- **Reliable VM redeploy** (a lingering old binary silently invalidates a test round): `sc stop` → wait (`ping -n 10`) → `taskkill /F` → verify `tasklist … || echo GONE` → download → **verify hash** → `sc start`. Ship big binaries over an HTTP channel (Mac `python3 -m http.server` on `10.211.55.2`; ICMP host→VM is blocked but TCP works); small files via chunked base64 + `certutil -decode` (delete target first — it won't overwrite).
- **`prlctl exec` runs as SYSTEM / session 0:** Parallels shared folders are invisible; PowerShell stdout is **not** captured (write results to a `C:\` file, read via `cmd /c type`); `%errorlevel%` is parse-time-expanded in one-liners (use `cmd /v:on` + `!errorlevel!` or `&&`/`||`) — this trap once masked a crash as "fixed".
- **A Windows service's stderr is discarded** → a Go `fatal error`/access-violation stack is invisible in the log. Capture it by running the binary in console mode: `utmstack_edr run > out.txt 2> err.txt` (detached), then read `err.txt`.
- **COM refcount discipline is critical** (two separate crashes): in go-ole (`procwatch`), `ToIDispatch()` shares the VARIANT's ref — `Release()` it, don't *also* `Clear()` (double-free `0xc0000005`); in the native DLL, keep the class-factory and per-object refcounts **separate** (a shared count let Windows unload the DLL mid-call → `0xC0000409`). `runtime.LockOSThread()` around COM; wrap pipeline goroutines in `goSafe`.
- **PowerShell sends AMSI content as UTF-16LE** — clamd signatures must be UTF-16 (an ASCII `.ndb` won't match).
- **EICAR embedded in a PE is NOT flagged** by clamd (only the standalone EICAR file). To test process-kill, build a real test EXE with a unique marker **appended as an overlay** + a custom `.ndb`/YARA rule; stage it in a file-watch-excluded dir (the process guard scans the image regardless of file exclusions).
- **ETW device paths** (`\Device\HarddiskVolumeN\…`) must be mapped to drive letters via `QueryDosDevice` or canary membership never matches and the ransomware tier silently never fires — the hard gate for enabling ransomware.
- **Always self-exclude the EDR's own tree** (`InstallDir`/`EngineDir`/`SpoolDir`/`QuarantineDir` + the engine dir): otherwise the watcher scans the engine's own `clamav-*.tmp` extraction artifacts → spurious events **and** a literal "clamav" in `object_path` = a branding-rule violation in an emitted field.
- **Branding audit is self-referential on a live telemetry box:** any `prlctl exec` command containing "clam" is itself captured as a behavioral event → false leak. Stop the EDR, *then* audit. (Engine helper processes are excluded from telemetry + guard so they never leak the engine name.)
- **Windows Defender fights the EDR during tests** (two real-time AVs). Its AMSI provider **blocks writing EICAR via PowerShell** (use `certutil -decode` from base64 — the reason for the base64/certutil transfer above), and its real-time protection **races the EDR to quarantine** test samples. For clean attribution — especially process-kill and AMSI, where two killers / two AMSI providers act non-deterministically — **disable Defender real-time protection** (or `Add-MpPreference -ExclusionPath` the test dirs for lighter isolation, enough for the file/quarantine path). Gotcha: with **Tamper Protection ON, `Set-MpPreference -DisableRealtimeMonitoring` is silently ignored** — it must be toggled off in the Windows Security GUI first (a manual step `prlctl exec` can't do). Do one coexistence pass with Defender **ON**, since real customers run both.
- **Suspend-on-launch is dangerous** — enabling it once froze the whole VM (every launch suspended, unbounded, nothing resumed on stop). Rails: never suspend images under `%SystemRoot%`; bound every suspend by `SuspendTimeoutMs`; default **OFF**.
- **Deploy skew:** rebuild+redeploy **both** the agent and the EDR binary — a stale agent's `enable-edr` Load/Save strips config fields it doesn't know (`engine_tier_override`, `signature_mirror`, …).
- **`github.com/0xrawsec/golang-etw` is a windows-only import** — do **not** `go mod tidy` on the macOS host or it gets dropped from `go.mod`.

---

## Master-server `edr` container & the agent TLS trust model

The server-side half of signature + threat-intel delivery is a new master-server container, service name **`edr`** (own top-level Go module at repo root `edr/`, image `ghcr.io/utmstack/utmstack/edr`, deployed by the installer as a Swarm service). It runs `cvdupdate` (ClamAV signature mirror) and syncs ThreatWinds indicator feeds into a shared volume, normalized to the agent's wire contract. Spec + plan in the table above; gap analysis in `docs/2026-07-07-edr-remaining-work.md`. Built on branch `feature/edr-server-container` in a **separate full clone at `/Users/atlas/UTMStack/v12`** (not the sparse `utmstack-v12/`), because it touches `agent-manager/`, `installer/`, `.github/`.

**The load-bearing finding (verify before touching signature transport):** the agent has **no certificate import/pinning anywhere**. Every encrypted channel — gRPC SOAR/commands (9000), gRPC log shipping (50051), HTTPS dependency downloads (9001) — is driven by one flag `SkipCertValidation` (YAML `insecure`, set at install from `install <server> <key> <yes/no>`, `agent/config/config.go`). Two modes:
- **`insecure: no`** → full TLS validation **against the OS/system trust store** (relies on the server presenting an OS-trusted cert; nothing is imported).
- **`insecure: yes`** → skip-verify everywhere (self-signed servers).

The EDR module already receives this flag (`cmd/enable_edr.go` copies it → `edr.json` `skip_cert_validate`).

**freshclam is the one component that cannot skip-verify.** On the ClamAV Windows build it validates via the **Windows cert store and ignores `CURL_CA_BUNDLE`** (VM-proven — an earlier CA-bundle approach was built and then removed). So the signature-mirror transport **mirrors the agent's two modes**, keyed off `skip_cert_validate` in `engine.ResolveMirrorURL`:
- `skip_cert_validate=false` → `PrivateMirror https://<server>:9001/private/edr/signatures` (freshclam validates via the OS store, natively — served by `agentmanager`).
- `skip_cert_validate=true` → `PrivateMirror http://<server>:9002/private/edr/signatures` (the `edr` container serves the mirror tree over **plain HTTP** on 9002; CVD integrity is guaranteed by ClamAV's Talos **signature**, not transport — the spec's "integrity by signature, not transport" principle, so HTTP is safe for public signed data). No Windows-root-store modification, ever.

Threat-intel feeds + binaries still ride 9001-HTTPS with the agent's own skip-verify client; only freshclam needs the HTTP alternative. The URL path `/private/edr/…` is identical on both ports; branding rule holds (no `clamav`/`threatwinds` in the path).

---

## Repo topology (updated 2026-09-21)

**`github.com/utmstack/OpenEDR` (PRIVATE) is the working home for this project.** It is a
full working copy of the UTMStack v12 tree, not an EDR-only repo, because almost every
remaining task touches `backend/`, `frontend/`, `agent-manager/`, `installer/` or
`definitions/` as well as `agent/edr/`.

| Branch in OpenEDR | What it is |
|---|---|
| `main` | **Default. Team works here.** `release/v12.0.0` with both EDR branches merged in |
| `edr-phase1` | Endpoint module: all of `agent/edr/` + the thin agent hooks |
| `feature/edr-server-container` | Master-server `edr` container + agent-manager/installer/CI wiring |
| `release/v12.0.0` | Mirror of the upstream base, so merging back upstream stays a normal pull request |
| `snapshot-archive` | The old flat safety snapshot (superseded; kept for history) |

`github.com/utmstack/UTMStack` (public) stays the eventual upstream: when a slice of work is
ready, it goes up as a pull request into `release/v12.0.0` there. Drop `docs/edr/` from that
pull request — it is internal planning.

Local clones: `/Users/atlas/UTMStack/v12` is the **full** clone (everything; `main`
lives here) and `/Users/atlas/UTMStack/EDR/utmstack-v12` is the older sparse clone
(`agent/` + `shared/` only, branch `edr-phase1`). Both have `origin` = UTMStack and
`openedr` = OpenEDR. Prefer the full clone for new work.

## Git & workflow policy

- **The user drives git.** Do **not** run `git commit`/`push`, do **not** create branches, and **never** use git to revert unless the user asks in the current turn. Write and test code; leave commits to the user by default.
- Push to `openedr`, never to `origin` (the public repo) without being asked.
- Follow the plans task-by-task (TDD: failing test → implement → pass). When executing a plan surfaces a flaw in the plan's code, fix it and note the deviation (e.g. the `sigDBVersion` branding fix in Plan 1).

---

## External dependencies (other teams / not in this repo)

1. **Server hosting** of per-platform ClamAV builds + the `utmstack_edr` binary at the agent dependencies endpoint (same model as `updater`/beats).
2. **ClamAV signature mirror + ThreatWinds feed mirror** — now built as the master-server **`edr` container** (see the dedicated section above), serving `/private/edr/signatures` (freshclam mirror) and `/private/edr/feeds/v1` (netblock indicators) off a shared volume via `agentmanager` (9001-HTTPS) and its own 9002-HTTP surface. Preserves cdiff incrementals + ClamAV signature verification. Until deployed, agents fall back to the official CDN (signatures) / stale-but-enforcing last-good set (feeds).
3. **Platform correlation parser** for the `utmstack_edr` data type.
4. **Sigma correlation rules** + SOAR for the behavioral + ransomware telemetry.
5. **Authenticode code-signing cert** for the EDR binary and the AMSI DLL (ordinary signing, not the driver path).
