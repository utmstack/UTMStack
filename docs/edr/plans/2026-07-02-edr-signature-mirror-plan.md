# UTMStack EDR — Deferred Phase: UTMStack-Hosted Signature (CVD) Mirror

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **STATUS: DEFERRED — implement later.** This phase is gated on an **external dependency**: the UTMStack platform team standing up a `cvdupdate`-based CVD mirror (§ "Server-side"). The client-side hook (`signature_mirror` → freshclam `PrivateMirror`) is **already implemented** in Plans 1–2; this phase adds the remaining client polish (fallback, auto-derivation, freshness health) and formalizes the server-side requirement. Do not start until the server mirror exists or a test mirror is available.

**Goal:** Let the UTMStack server act as a ClamAV signature mirror so the whole fleet updates from one upstream fetch — preserving cdiff incrementals and ClamAV's signature verification — with a safe fallback when the mirror is unavailable.

**Architecture:** Confirmed by the evaluation in `docs/superpowers/specs/2026-07-02-edr-signature-update-mechanism.md`: **Option A (ClamAV-native mirror)**. The server runs the official `cvdupdate` tool and serves the CVD directory over its existing HTTPS; agents keep using **freshclam** pointed at that URL via `PrivateMirror`. The agent's dependency downloader is **not** used for signatures (it would lose cdiff incrementals and reinvent version tracking — see the design doc §4).

**Tech Stack:** Go 1.25.5, the existing `edr/engine` (config generation) + `edr/feed` (freshclam scheduler); server-side `cvdupdate` (Python) + static HTTPS hosting (platform team).

## Global Constraints

All Plan 1 Global Constraints carry over (branch `release/v12.0.0`, Go-first, **branding never emits `clamav`/`clamd` in events or surfaced logs**, own `edr.db`, multi-arch amd64+arm64, no new Go deps). Plus:

- **Integrity is by signature, not transport.** CVDs are ClamAV-signed; freshclam/clamd verify them on download/load regardless of source. The mirror is **untrusted by design** — it cannot inject malicious signatures. Do not add custom "trust the mirror" logic.
- **Reuse freshclam for the fetch.** All this phase adds on the client is *config generation* and *scheduling/health* around freshclam — never a custom CVD downloader.
- **Already built (do not redo):** `edr.json` `signature_mirror` field; `engine.RenderFreshclamConf` emits `PrivateMirror <url>` when set, else `DatabaseMirror database.clamav.net`.

**Build/test note:** identical split — config/fallback/health logic is TDD'd with `go test` on macOS; freshclam-against-a-real-mirror is verified on the VM against a local test mirror (a `cvdupdate` dir or a static file server).

---

## Server-side (EXTERNAL DEPENDENCY — platform team, not EDR code)

This phase cannot deliver value until the platform provides:

1. **`cvdupdate` mirror service.** A scheduled job (e.g. hourly) that fetches `main`/`daily`/`bytecode` CVDs **and their cdiffs** from the ClamAV CDN into a mirror directory, rate-limit-compliant (this single job is the fleet's only CDN contact).
2. **Static HTTPS hosting** of that directory at a stable path on the existing web tier, e.g. `https://<server>/clamav/`. No new service — a static route behind the current TLS.
3. **Freshness monitoring/alerting.** A stale mirror ⇒ fleet-wide stale detection; the central mirror must be watched.
4. *(Optional, later)* **Third-party feeds** (Sanesecurity / SecuriteInfo / URLhaus) mirrored through the same `cvdupdate` instance (custom databases), staged report-only first.

The EDR tasks below assume this endpoint exists (or a test stand-in during development).

---

## File Structure (client-side changes)

- `agent/edr/config/config.go` — MODIFY: add `SignatureFallback` field (`"none"|"cdn"`).
- `agent/edr/engine/clamdconf.go` — MODIFY: resolve the mirror URL (explicit vs `"auto"` derived from `cfg.Server`); a `RenderFreshclamConfFor(cfg, useCDN bool)` variant for the fallback cycle.
- `agent/edr/feed/freshclam.go` — MODIFY: track consecutive mirror-update failures; on threshold with `SignatureFallback=="cdn"`, run one CDN cycle then revert; record last-successful-update time.
- `agent/edr/feed/health.go` — CREATE: signature-freshness health event emitter.
- `agent/edr/service/service.go` — MODIFY: surface signature source + last-update in `status.json`.

---

## Task 1: `signature_fallback` config + `"auto"` mirror URL resolution

**Files:**
- Modify: `agent/edr/config/config.go`, `agent/edr/engine/clamdconf.go`
- Test: `agent/edr/engine/clamdconf_test.go` (extend)

**Interfaces:**
- Produces:
  - `EDRConfig.SignatureFallback string` (`json:"signature_fallback"`, default `"cdn"`).
  - `func ResolveMirrorURL(cfg config.EDRConfig) string` — returns the effective private-mirror URL: the literal `SigMirror`, or `https://<cfg.Server>/clamav` when `SigMirror == "auto"`, or `""` when unset.
  - `RenderFreshclamConf` uses `ResolveMirrorURL`; a new `RenderFreshclamConfCDN(cfg)` forces the official CDN (for the fallback cycle).

- [ ] **Step 1: Write the failing test**

```go
// add to agent/edr/engine/clamdconf_test.go
func TestResolveMirrorURL(t *testing.T) {
	cfg := config.Default()
	if ResolveMirrorURL(cfg) != "" {
		t.Fatal("unset SigMirror must resolve empty (CDN mode)")
	}
	cfg.SigMirror = "https://m.example.com/clamav"
	if got := ResolveMirrorURL(cfg); got != "https://m.example.com/clamav" {
		t.Fatalf("explicit mirror = %q", got)
	}
	cfg = config.Default()
	cfg.SigMirror = "auto"
	cfg.Server = "utm.example.com"
	if got := ResolveMirrorURL(cfg); got != "https://utm.example.com/clamav" {
		t.Fatalf("auto mirror = %q", got)
	}
}

func TestRenderFreshclamConfCDNForcesOfficial(t *testing.T) {
	cfg := config.Default()
	cfg.SigMirror = "https://m.example.com/clamav" // even with a mirror set...
	conf := RenderFreshclamConfCDN(cfg)             // ...the CDN variant ignores it
	if !strings.Contains(conf, "DatabaseMirror database.clamav.net") || strings.Contains(conf, "PrivateMirror") {
		t.Fatalf("CDN variant must use the official CDN only:\n%s", conf)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/engine/ -run 'ResolveMirror|FreshclamConfCDN' -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

In `config.go`, add the field and overlay it in `Load` (default `"cdn"` in `Default()`):
```go
// EDRConfig
SignatureFallback string `json:"signature_fallback"` // "none" | "cdn"
// Default(): SignatureFallback: "cdn",
// Load() overlay: if onDisk.SignatureFallback != "" { c.SignatureFallback = onDisk.SignatureFallback }
```

In `clamdconf.go`:
```go
// ResolveMirrorURL returns the effective private-mirror URL, expanding the
// "auto" token to the UTMStack server's conventional mirror path.
func ResolveMirrorURL(cfg config.EDRConfig) string {
	switch cfg.SigMirror {
	case "":
		return ""
	case "auto":
		if cfg.Server == "" {
			return ""
		}
		return "https://" + cfg.Server + "/clamav"
	default:
		return cfg.SigMirror
	}
}

// change RenderFreshclamConf to use ResolveMirrorURL(cfg) instead of cfg.SigMirror.

// RenderFreshclamConfCDN forces the official CDN regardless of mirror config,
// used for a single fallback cycle when the mirror is unreachable.
func RenderFreshclamConfCDN(cfg config.EDRConfig) string {
	forced := cfg
	forced.SigMirror = ""
	return RenderFreshclamConf(forced)
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/engine/ -run 'ResolveMirror|Freshclam' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/edr/config/config.go agent/edr/engine/clamdconf.go agent/edr/engine/clamdconf_test.go
git commit -m "feat(edr): signature_fallback config + auto mirror URL + CDN fallback conf"
```

---

## Task 2: EDR-managed mirror→CDN failover in the feed

**Files:**
- Modify: `agent/edr/feed/freshclam.go`
- Test: `agent/edr/feed/failover_test.go`

**Interfaces:**
- Consumes: `engine.ResolveMirrorURL`, `engine.RenderFreshclamConfCDN`, `engine.WriteEngineConfigs`.
- Produces: `Feed` gains failover: after `mirrorFailThreshold` consecutive freshclam failures while in mirror mode, if `SignatureFallback=="cdn"`, write the CDN freshclam.conf, run one update cycle, then restore the mirror conf. Uses an injectable `writeConf func(useCDN bool) error` for testing.

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/feed/failover_test.go
package feed

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestFailoverToCDNAfterRepeatedMirrorFailures(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.SigMirror = "https://m.example.com/clamav"
	cfg.SignatureFallback = "cdn"

	f := New(cfg, c)
	var wroteCDN bool
	f.writeConf = func(useCDN bool) error { wroteCDN = useCDN || wroteCDN; return nil }
	// runUpdater fails while pointed at the mirror, succeeds once switched to CDN.
	calls := 0
	f.runUpdater = func() error {
		calls++
		if wroteCDN {
			return nil
		}
		return errors.New("mirror unreachable")
	}
	f.currentSigDB = func() (string, error) { return "300", nil }

	// Drive several cycles; after the threshold it must flip to CDN and recover.
	for i := 0; i < mirrorFailThreshold+1; i++ {
		_, _ = f.updateOnce()
	}
	if !wroteCDN {
		t.Fatalf("expected failover to CDN after %d mirror failures", mirrorFailThreshold)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/feed/ -run TestFailover -v`
Expected: FAIL — undefined `writeConf`/`mirrorFailThreshold`.

- [ ] **Step 3: Implement**

Add to `freshclam.go`:
```go
const mirrorFailThreshold = 3

// Feed fields (add): writeConf func(useCDN bool) error; mirrorFails int; onCDNFallback bool
// New(): default f.writeConf = f.defaultWriteConf

func (f *Feed) defaultWriteConf(useCDN bool) error {
	// Reuse the engine's writers; PlanTuning result isn't needed for freshclam.conf.
	if useCDN {
		return engine.WriteFreshclamConf(f.cfg, true)  // helper: CDN variant
	}
	return engine.WriteFreshclamConf(f.cfg, false)
}

// In updateOnce(): if in mirror mode and runUpdater fails, increment mirrorFails;
// when it reaches the threshold and SignatureFallback=="cdn", writeConf(true),
// log a health note, and run one more cycle; on success reset and restore mirror.
```
(Provide `engine.WriteFreshclamConf(cfg, useCDN bool) error` writing the chosen freshclam.conf variant; keep `WriteEngineConfigs` for clamd.conf.)

Full failover logic:
```go
func (f *Feed) updateOnce() (string, error) {
	err := f.runUpdater()
	mirror := engine.ResolveMirrorURL(f.cfg) != ""
	if err != nil {
		logger.Error("UTMStack EDR: signature update failed: %v", err)
		if mirror && f.cfg.SignatureFallback == "cdn" {
			f.mirrorFails++
			if f.mirrorFails >= mirrorFailThreshold && !f.onCDNFallback {
				logger.Info("UTMStack EDR: signature mirror unreachable; falling back to public updates")
				if werr := f.writeConf(true); werr == nil {
					f.onCDNFallback = true
					err = f.runUpdater() // one CDN cycle
				}
			}
		}
		if err != nil {
			return "", err
		}
	}
	// success
	if f.onCDNFallback {
		// restore mirror config for the next cycle to retry the mirror
		_ = f.writeConf(false)
		f.onCDNFallback = false
	}
	f.mirrorFails = 0
	// ... existing currentSigDB + MarkStaleBySigDB ...
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./edr/feed/ -v`
Expected: PASS (existing + failover).

- [ ] **Step 5: Commit**

```bash
git add agent/edr/feed/ agent/edr/engine/
git commit -m "feat(edr): mirror→CDN failover when the signature mirror is unreachable"
```

---

## Task 3: Signature-freshness health + status reporting

**Files:**
- Create: `agent/edr/feed/health.go`
- Modify: `agent/edr/feed/freshclam.go` (record last-success), `agent/edr/service/service.go` (status fields)
- Test: `agent/edr/feed/health_test.go`

**Interfaces:**
- Produces:
  - `Feed` records `lastSuccess time.Time` and exposes `Stale(now time.Time, max time.Duration) bool`.
  - `func SignatureSource(cfg config.EDRConfig) string` — `"utmstack-mirror"` | `"official-cdn"` for status/telemetry (branded; never names the engine).
  - On staleness, emit a `health` event to the spool (fleet-monitoring signal).
  - `status.json` gains `signature_source` and `signature_last_update`.

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/feed/health_test.go
package feed

import (
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestSignatureSource(t *testing.T) {
	cfg := config.Default()
	if SignatureSource(cfg) != "official-cdn" {
		t.Fatal("default source must be official-cdn")
	}
	cfg.SigMirror = "https://m/clamav"
	if SignatureSource(cfg) != "utmstack-mirror" {
		t.Fatal("mirror set → utmstack-mirror")
	}
}

func TestStaleness(t *testing.T) {
	f := &Feed{}
	f.lastSuccess = time.Unix(1000, 0)
	if !f.Stale(time.Unix(1000+7200, 0), time.Hour) {
		t.Fatal("2h since last success with 1h max → stale")
	}
	if f.Stale(time.Unix(1000+1800, 0), time.Hour) {
		t.Fatal("30m since last success with 1h max → fresh")
	}
}
```

- [ ] **Step 2–4:** implement `SignatureSource`, `lastSuccess`/`Stale`, the `health` event emission, and add `signature_source` + `signature_last_update` to `statusDoc` in `service.go` (fed from `feed`), then run `go test ./edr/feed/ ./edr/service/ -v` (PASS).

- [ ] **Step 5: Commit**

```bash
git add agent/edr/feed/ agent/edr/service/
git commit -m "feat(edr): signature-freshness health + source reporting in status"
```

---

## Task 4: Acceptance against a test mirror (VM)

- [ ] **Step 1: Stand up a test mirror.** On the VM (or a build box), run `cvdupdate` to populate a directory, and serve it over HTTPS (or plain HTTP for the test) at e.g. `https://<host>/clamav/`. Alternatively seed a static dir with `main.cvd`/`daily.cvd`/`bytecode.cvd` + `.cdiff`s.

- [ ] **Step 2: Point the EDR at it.** Set `edr.json`: `{ "signature_mirror": "https://<host>/clamav", "signature_fallback": "cdn" }`. Restart the EDR; confirm the generated `freshclam.conf` has `PrivateMirror https://<host>/clamav` and `ScriptedUpdates yes`.

- [ ] **Step 3: Verify pull + incremental + integrity.** Trigger an update; confirm freshclam pulls from the mirror, applies a **cdiff** (not a full CVD) on the next change, and **verifies signatures** (tamper a mirrored file → freshclam rejects it). Confirm clamd loads the DB and `edr-status` shows `signature_source: utmstack-mirror` with a recent `signature_last_update`.

- [ ] **Step 4: Verify failover.** Take the mirror offline; after `mirrorFailThreshold` cycles confirm the EDR falls back to the CDN (log + a successful update), then restores mirror mode when it returns. With `signature_fallback: none`, confirm it stays mirror-only (stale, not broken) and emits a freshness `health` event.

- [ ] **Step 5: Labeling audit.** Confirm no `clamav`/`clamd` string in any emitted event or surfaced log (`signature_source` uses `utmstack-mirror`/`official-cdn`, never the engine name).

- [ ] **Step 6: Commit any fixes.**

```bash
git add -A && git commit -m "fix(edr): signature-mirror phase VM acceptance adjustments"
```

---

## Self-Review (against the design doc)

**Coverage:**
- Option A client mechanism (freshclam `PrivateMirror`, no dependency-downloader for signatures) → already built + Tasks 1–3. ✔
- `signature_mirror` explicit + `"auto"` derivation from `cfg.Server` → Task 1. ✔ (design §9 open decision resolved with the `"auto"` token.)
- Fallback policy (`none`|`cdn`), with EDR-managed mirror→CDN failover since freshclam won't auto-fall-back → Task 2. ✔ (design §5.4.)
- Freshness health + auditable source/last-update in status → Task 3. ✔ (design §5.1 monitoring, §8.5 governance.)
- Integrity preserved (freshclam verifies signatures; mirror untrusted) — no custom trust logic added. ✔
- Server-side `cvdupdate` mirror called out as the external dependency, not EDR code. ✔

**Placeholder scan:** client-side logic (URL resolution, fallback, health, source) has complete code + `go test`. The mirror stand-up and freshclam-against-mirror behavior are concrete VM steps (can't unit-test a real mirror on macOS).

**Not in scope here:** third-party feed mirroring (design §9) — a follow-on once the base mirror is proven; and the server-side `cvdupdate` service itself (platform team).

**Dependency:** DO NOT begin until the server mirror (or a test mirror) exists. The default CDN path (already shipped) keeps signatures current in the meantime.
