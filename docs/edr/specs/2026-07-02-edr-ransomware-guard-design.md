# UTMStack EDR — Ransomware Guard Design Spec

**Status:** DRAFT — design proposed, awaiting user review (decisions in §0 taken on best-judgment while user was away; confirm or redirect before planning)
**Date:** 2026-07-02
**Target:** `github.com/utmstack/UTMStack` branch `release/v12.0.0`, `agent/edr/` (Go 1.25.5), Windows amd64 + arm64
**Builds on:** Plans 1–4 (foundation, auto-detect, process-kill, scripts-telemetry). This is a **new additive capability module**, not a change to existing detection.
**Source context:** `EDR/UTMStack_Endpoint_Detection_Engineering_Plan.docx` v3.0; existing spec `2026-07-01-edr-phase1-windows-design.md`.

---

## 0. Decisions taken (confirm on review)

These were chosen on best-judgment while the user was away. They are the load-bearing choices — flag any you want changed and the spec adjusts.

1. **Architecture = unified per-process scoring engine (Approach A).** Independent detectors emit weighted, PID-tagged evidence into one fusion engine; a decaying per-process score drives a graduated escalation ladder. Rejected alternatives: independent tripwires without fusion (more false positives on fuzzy signals), and telemetry-only/platform-side detection (too slow for ransomware — the encryptor runs during the round-trip).
2. **Process attribution = ETW file-I/O feed** (`Microsoft-Windows-Kernel-File`), driverless, behind `//go:build windows`. This is the one new capability the feature genuinely needs and the biggest new piece. Consistent with the existing design, which already contemplates ETW (`Microsoft-Windows-Kernel-Process`, §4.3 of the Phase-1 spec).
3. **Scope split v1 / v2.** v1 = high-fidelity core (ETW file feed + canary tripwires + T1490 command rules + scoring engine + suspend/kill/quarantine). v2 = fuzzy file-op sensors (entropy, extension churn, rate, ransom-note), registry sensor, and VSS rollback.
4. **Default response = suspend-first, then confirm, then kill-tree + quarantine.** Configurable to `alert-only` or `kill-immediately`.

---

## 1. Purpose & scope

Add a **behavioral ransomware detection-and-response capability** to the EDR: catch **novel / unknown / packed** crypto-ransomware that slips past the signature engine, by recognizing ransomware *behavior* (mass encryption, recovery destruction, decoy tampering) and containing it fast with the EDR's existing kill/quarantine arm.

### In scope (v1)
1. **ETW file-activity feed** — per-process `(PID, path, operation)`, driverless. The attribution backbone.
2. **Canary (honeypot) tripwires** — decoy files planted across the filesystem; any tamper is a near-zero-FP, max-confidence signal.
3. **T1490 recovery-tampering command rules** — `vssadmin delete shadows`, `wbadmin delete`, `bcdedit …recoveryenabled no`, `wmic shadowcopy delete`, `reagentc /disable`, `diskshadow`. Highest-fidelity ransomware tell.
4. **Per-process scoring/fusion engine** — weighted evidence + time-decay + thresholds → escalation ladder.
5. **Response** — suspend → confirm → kill-tree + quarantine, via existing `responder`/`quarantine`.
6. **Config, events, status, cache** — new config block, `SourceBehavioral` events, `statusDoc` fields, incident persistence.

### In scope (v2, documented here, separate plan)
7. Fuzzy file-op sensors: **entropy spike**, **extension churn**, **mass-modify rate**, **ransom-note pattern**.
8. **Registry recovery-disable sensor** (ETW `Microsoft-Windows-Kernel-Registry`).
9. **VSS rollback / recovery** of encrypted files.

### Out of scope (explicitly)
- **No kernel driver / minifilter, ever** (design invariant). This makes the feature *detect-fast-then-kill-and-contain*, **not pre-block** (§7).
- **No ML model in v1.** The weighted-heuristic scorer is the pragmatic engine; an ML classifier is a possible later evolution of the same fusion interface.
- **Network-tier ransomware signals** (C2, SMB lateral spread) — platform/network correlation, not this endpoint module. USN/ETW cover local volumes only.
- **Linux** — Phase 1 is Windows-only.

---

## 2. Guiding constraints (inherited)

- **No kernel driver.** User-mode only. Attribution via ETW, not a minifilter. Response via SYSTEM privilege (`OpenProcess(PROCESS_TERMINATE)`), already built.
- **Go-first.** All new code is Go. ETW consumption uses native syscalls behind `//go:build windows` (+ `!windows` stubs), same triad pattern as `responder`/`procwatch`/`watcher`. No new non-Go artifact (the AMSI DLL remains the only one).
- **Quarantine, never delete.** Reuse `quarantine.Store`.
- **Honest capability claims.** Detect-and-kill after the fact; some files encrypt in the ms–seconds before the kill; suspend-first narrows it; VSS rollback (v2) recovers the rest. Events and docs say this plainly — no "prevents ransomware" overclaim (§7).
- **Branding.** Every event/log says "UTMStack EDR"; `clam*` never appears in an event field or surfaced log. `ToJSON()` already forces `Product`/`Engine = "UTMStack EDR"`.

---

## 3. High-level architecture

The feature is a new **detection brain** (`edr/ransomware/`) bolted onto the existing **response body** (`responder` + `quarantine` + `proctable` + `event.Spool`). It runs as one supervised goroutine (`goSafe("ransomware", …)`) inside `startPipeline`.

```
  SENSORS (emit PID-tagged, weighted Evidence)                 SCORING BRAIN                RESPONSE (existing)
  ┌─────────────────────────────────────────┐                ┌───────────────────┐       ┌────────────────────┐
  │ Canary tripwire   ── file touch on decoy │──┐             │  per-PID scorer    │       │ responder.Suspend  │
  │ T1490 cmd rule    ── proc cmdline match  │──┤   Evidence  │  weighted sum with │  esc- │ responder.KillTree │
  │ (v2) entropy      ── sampled file bytes  │──┼────────────▶│  time-decay window │──────▶│ quarantine.Store   │
  │ (v2) ext churn    ── rename ops          │──┤             │  + thresholds      │ alate │ event.Spool.Append │
  │ (v2) mass-rate    ── writes/sec per PID  │──┤             │  + escalation ladder│      │ (v2) VSS rollback  │
  │ (v2) ransom-note  ── new README_* files  │──┤             └─────────┬─────────┘       └────────────────────┘
  │ (v2) registry     ── recovery-key writes │──┘                       │
  └─────────────────────────────────────────┘                          │
        ▲                        ▲                                      ▼
        │ (PID,path,op)          │ ProcStart{PID,PPID,Image,Cmdline}   proctable.Get(pid) → image/parent for kill + quarantine
  ┌─────┴──────────────┐   ┌─────┴───────────────┐
  │ ETW Kernel-File    │   │ procwatch (WMI)     │   ← existing feed; dispatch extended to also call guard.OnProcStart
  │ feed (NEW, driverless)  │ (Plan 3)            │
  └────────────────────┘   └─────────────────────┘
```

**Data flow (one line).** ETW file event / process start → sensor emits `Evidence{PID, kind, weight}` → scorer accumulates (decaying) per PID → threshold crossed → escalation ladder (suspend → confirm → kill-tree + quarantine + branded event → v2 rollback).

**Reuse (verified against code):** `responder.Responder{KillTree, Suspend, Resume}`, `proctable.Table{Get, Descendants}`, `quarantine.Store{Quarantine}`, `scanner.SHA256File` / `scanner.ScanFile`, `event.Spool.Append`, `procwatch` dispatch, `orchestrator.Excluder` (for canary/self exclusions). **No agent-package changes** beyond what Plans 1–4 already add.

---

## 4. Component specifications

New package **`edr/ransomware/`**. Pure logic in untagged files (unit-tested on macOS); native bits behind the `_windows.go`/`_other.go` triad.

### 4.1 ETW file-activity feed — `fileactivity_windows.go` / `_other.go`
**Responsibility.** Provide driverless **per-process** file telemetry — the attribution USN can't give.
- Consume the **`Microsoft-Windows-Kernel-File`** ETW provider (GUID `{EDD08927-9CC4-4E65-B970-C2560FB5C289}`) via a real-time session. Emit `FileEvent{PID int; Path string; Op FileOp}` where `Op ∈ {Create, Write, SetInfo, Rename, Delete, Close}`.
- **Volume:** this provider is high-rate. Mitigations: enable only write/rename/delete/setinfo/create keywords (drop reads); in-process filter to (a) canary paths and (b) user-data roots; drop EDR's own PID and self-paths (`InstallDir`, `EngineDir`, `SpoolDir`, `QuarantineDir`) via the existing `Excluder`; coalesce per (PID,path).
- **Fallback if ETW is rejected/too costly (see §12 open decision):** detect via the existing **USN journal** (extend the `Sink` to carry `USNChange.Reason` — a watcher-package change) for the *what*, with **coarse attribution** (correlate change bursts against recent `procwatch` starts). Weaker; canary + command rules still work because their attribution comes from ETW-independent paths.
- Native ETW behind `//go:build windows`; `_other.go` stub returns a closed channel so the module builds/tests on macOS. Likely one new Go dep (an ETW consumer lib), analogous to Plan 3's `go-ole`.

### 4.2 Canary manager — `canary.go` (+ small windows helper for hidden/system attrs)
**Responsibility.** Plant and track decoy files whose only purpose is to be tampered with.
- **Placement:** each fixed-volume root, user-profile dirs (Desktop, Documents, Downloads, Pictures), and each subtree the config names. Names sort *early* in a directory (e.g. `~$aaa_accounts.xlsx`, `_00_backup.docx`) so an alphabetical encryptor hits them before real data; realistic extensions (`.xlsx/.docx/.pdf`); Hidden+System attributes so users don't see/edit them. Content is plausible-but-inert bytes.
- **Registry:** persist the canary set (path + hash + placement time) in `edr.db` (§5.3) so a touch can be validated against the known set and false "user edited it" cases excluded (users never see them).
- **Detection:** the ETW feed reports a write/rename/delete on a canary path → the canary sensor emits **max-weight Evidence** for that PID. Elastic measured ~12s-class detection this way.
- **Regeneration:** re-plant on start and after any consumed canary; optionally rotate names so malware can't learn a fixed skip-list.
- Pure placement/naming/validation logic is untagged and unit-tested; only the file-attribute call is windows-tagged.

### 4.3 T1490 command-rule sensor — `rules.go`
**Responsibility.** Fire on recovery/backup-destruction commands — the highest-fidelity, lowest-FP ransomware tell.
- A rule table matched against `procwatch.ProcStart{Image, Cmdline}`: `vssadmin … delete shadows`, `vssadmin … resize shadowstorage`, `wmic shadowcopy delete`, `wbadmin delete catalog|backup`, `bcdedit … recoveryenabled no`, `bcdedit … bootstatuspolicy ignoreallfailures`, `reagentc /disable`, `diskshadow`. Case-insensitive, normalized-arg matching.
- **Attribution nuance:** these run as short-lived child processes; the *encryptor* is usually the **parent**. On match, emit max-weight Evidence against **`ProcStart.PPID`** (fall back to PID) using `proctable` to resolve the tree.
- Pure table + matcher; unit-tested with a corpus of real and benign command lines (guard against, e.g., legit admin/backup tooling — configurable allowlist).
- **Hook:** extend `procwatch` dispatch — currently `NewDispatch(tab, g)` calls `guard.OnStart`; add a second handler so each `ProcStart` also reaches `ransomware.Guard.OnProcStart`. No change to the WMI reader itself.

### 4.4 Scoring / fusion engine — `scorer.go` (pure Go, the heart)
**Responsibility.** Turn a stream of weighted, PID-tagged `Evidence` into a defensible per-process risk score and an escalation decision.
- **Model:** `map[pid] → decaying score`. Each `Evidence{PID, Kind, Weight, TS}` adds weight; scores **decay** over a sliding window (e.g. exponential, half-life configurable) so brief benign bursts don't accumulate forever.
- **Thresholds → ladder (see §6):** `observe < suspendT < killT`. Max-weight sensors (canary, T1490) alone exceed `killT` → instant action; fuzzy sensors accumulate toward it.
- **Kind-diversity bonus:** N *different* signal kinds on one PID scores higher than N repeats of one kind (fusion — the anti-FP mechanism).
- **Recycled-PID safety:** key evidence by `(pid, proctable.Proc.StartTS)` so a reused PID can't inherit a dead process's score.
- Deterministic and clock-injectable → fully unit-testable on macOS with synthetic evidence sequences.

### 4.5 Guard / orchestration — `guard.go`
**Responsibility.** Own `Run(ctx)`; wire sensors → scorer → response; execute the escalation ladder.
- Constructed in `startPipeline` with `(cfg, sp, resp *responder.Responder, tab *proctable.Table, store *quarantine.Store, sc *scanner.Scanner)`.
- Subscribes to the ETW feed and receives `OnProcStart`. Feeds sensors, ticks the scorer, and on threshold calls the response arm (§6). Emits a branded event at each ladder stage.
- Response calls only go through the existing `TreeResponder` interface (`{KillTree; Suspend; Resume}`) and `quarantine.Store` — no new syscalls for response.

### 4.6 (v2) Fuzzy sensors & rollback — `sensors.go`, `rollback_windows.go`
- **entropy:** sample bytes of files a suspect PID rewrites (reuse `scanner.SHA256File`'s read path); Shannon entropy ≈ 8 bits/byte over a growth window; intermittent-encryption aware (sample multiple offsets, not just whole-file).
- **extension churn / mass-rate / ransom-note:** derived from the ETW rename/write stream per PID.
- **registry:** ETW `Microsoft-Windows-Kernel-Registry` writes to SystemRestore/WinRE/BCD keys.
- **rollback:** proactively create a VSS snapshot; on containment, restore tampered user files from the snapshot. Driverless via WMI/`vssadmin`. Complex → its own plan.

---

## 5. Data model

### 5.1 Config — new block in `config.EDRConfig`
Add a nested struct (nested structs are copied wholesale by `Load()`, so a single overlay line suffices — cleaner than flat fields, per the recon):
```go
type RansomwareConfig struct {
    Enabled        bool     `json:"enabled"`
    ResponseMode   string   `json:"response_mode"`   // "alert" | "suspend" | "kill" ; default "suspend"
    CanaryDirs     []string `json:"canary_dirs"`     // extra subtrees to seed; defaults = volume roots + user profile
    CanaryPerDir   int      `json:"canary_per_dir"`  // default 1
    SuspendThreshold int    `json:"suspend_threshold"`
    KillThreshold    int    `json:"kill_threshold"`
    DecayHalfLifeMs  int    `json:"decay_half_life_ms"`
    CommandAllowlist []string `json:"command_allowlist"` // benign admin tooling exempt from T1490 rules
    UseETW           bool   `json:"use_etw"`         // default true; false → USN-fallback attribution
}
```
- `EDRConfig` gains `Ransomware RansomwareConfig \`json:"ransomware"\``.
- `Default()` sets sane defaults (Enabled=false initially so rollout is opt-in; ResponseMode="suspend").
- `Load()` gets **one overlay line** copying the sub-struct when present.

### 5.2 Events — reuse `SourceBehavioral`, add actions
- No event-contract change (source stays `behavioral` → no platform-parser churn). Add `Action` consts: `ransomware_suspected`, `ransomware_contained`. Add a constructor `NewRansomwareEvent(action string, p event.ProcInfo, signal, severity string) event.Event` alongside the existing ones.
- Emit at each ladder stage: suspect (score crossed `suspendT`), contained (killed+quarantined). Fields carry the offending `ProcInfo`, the contributing `signal` (e.g. `"canary:~$aaa_accounts.xlsx"`, `"t1490:vssadmin_delete_shadows"`), and `severity`.
- `Product`/`Engine`/`OS`/`Timestamp`/`HostID` auto-branded by `ToJSON()`. Labeling audit stays green.

### 5.3 Cache — new record types (+ `AutoMigrate`)
Following the `USNCursor` template (own file `cache/ransomware.go`, mutex-guarded accessors, added to the `AutoMigrate` list in `Open`):
- `CanaryRecord{Path (pk), SHA256, PlacedAt, Volume}` — the decoy registry.
- `RansomwareIncident{ID (pk), PID, Image, Cmdline, Signals, Score, Action, DetectedAt, QuarantineID}` — audit trail of contained incidents.

### 5.4 Status — new `statusDoc` fields
`writeStatus` adds: `RansomwareEnabled`, `CanaryCount`, `ETWActive` (or `AttributionMode`), `LastIncidentAt`. Consumed by `main.go status`.

---

## 6. Detection & response logic (the escalation ladder)

```
Evidence stream ─▶ scorer(pid) ─▶ score
   score ≥ suspendThreshold  and ResponseMode ≥ "suspend":
        responder.Suspend(pid)            # freeze damage, narrow the encryption race
        emit ransomware_suspected
        confirm: re-check score over a short grace window / optional scanner.ScanFile(image)
   score ≥ killThreshold (or confirmed):
        pids := proctable.Descendants(pid)         # leaves-first tree
        responder.KillTree(pid, "ransomware:<top-signal>")
        quarantine.Store.Quarantine(image, sha256, "ransomware:<top-signal>")
        write RansomwareIncident; emit ransomware_contained
        (v2) rollback tampered files from VSS snapshot
   ResponseMode == "alert":  emit events only, never suspend/kill (rollout/first-trust mode)
```
- **Max-weight sensors** (canary, T1490) reach `killThreshold` in one hit → immediate suspend+kill (subject to `ResponseMode`).
- **Fuzzy sensors** (v2) accumulate; the **kind-diversity bonus** means "high entropy + extension churn + mass-rate on one PID" crosses the threshold while any one alone stays under it → low FP.
- **False-positive guards:** command allowlist for legit admin/backup tooling; canary-touch is inherently near-zero-FP (users never see decoys); scores decay; `alert` mode for cautious rollout.

---

## 7. Honest capability statement (mandatory, per constraint #7)

- **This is detection-and-response, not prevention.** No kernel minifilter → the EDR cannot veto a write before it lands. It observes the encrypt via ETW/canary *just after* it starts and kills the process within seconds. **Files touched before the kill are encrypted**; suspend-first shrinks that set; VSS rollback (v2) recovers what was lost. Docs/events state this — never "blocks ransomware."
- **Canary + T1490 are high-confidence and fast; fuzzy heuristics are corroborating and tunable.** The product claim is "rapid behavioral detection and containment of unknown ransomware," bounded honestly.
- **Attribution depends on ETW.** If ETW is unavailable/disabled/tampered, file-based sensors degrade to coarse attribution; canary and command rules remain.

---

## 8. Branding & labeling

- All new events use branded fields via `ToJSON()`; signals are described in EDR terms (`"canary"`, `"t1490"`), never engine internals.
- `clam*` must not appear in any new event field or surfaced log. The existing labeling audit (`grep -rniE '…clam' edr/`) must stay green after this feature.

---

## 9. Testing strategy

- **Pure Go (macOS `go test`):** scorer (evidence sequences, decay, thresholds, diversity bonus, recycled-PID), rules matcher (real + benign command corpus, allowlist), canary planner (naming/placement/validation), config load/overlay round-trip, event constructors + labeling audit.
- **Windows VM acceptance (Parallels `10.211.55.12`):** ETW feed emits real `(PID,path,op)`; canary tamper by a synthetic encryptor triggers suspend+kill within seconds; a benign `vssadmin delete shadows` (max weight) is contained; verify **final outputs** — process killed, binary in quarantine store, `RansomwareIncident` row, branded events at the spool → platform. Build+vet both amd64 and arm64.
- **Safety:** the synthetic encryptor operates only on a sandboxed test tree of throwaway files; never real user data.

---

## 10. Phasing

- **Plan A (v1 core):** package skeleton + config/cache/event plumbing → scorer (TDD) → rules sensor + procwatch dispatch hook → canary manager → ETW file feed (windows) + stub → guard wiring in `startPipeline` + status → VM acceptance.
- **Plan B (v2):** fuzzy sensors (entropy/extension/rate/ransom-note) → registry sensor → VSS rollback → tuning + FP corpus → VM acceptance.

---

## 11. External dependencies

- **Platform correlation parser** must handle the new `ransomware_suspected` / `ransomware_contained` actions under `DataType=utmstack_edr`, `source=behavioral`.
- **ETW consumer Go library** selection (or hand-rolled) — see §12.
- Sigma/SOAR rules for the new events (platform-side).

---

## 12. Open decisions for review

1. **ETW vs USN-fallback for attribution** (§4.1). ETW gives real per-process file attribution but adds a Go dep + windows-native complexity; USN-fallback avoids the dep but gives weak attribution. Recommend **ETW**.
2. **v1 scope** — core-only (recommended) vs everything-in-v1.
3. **Default `ResponseMode`** — `suspend` (recommended) vs `alert` vs `kill`.
4. **Package name** — `edr/ransomware/` (proposed) vs folding into `edr/behavioral/`.
5. **New Go dependency** for ETW is acceptable? (Precedent: Plan 3's `go-ole`.)
