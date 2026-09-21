# EDR Network Blocklist — Design Spec

**Status:** Approved (design) — 2026-07-06
**Module:** `agent/edr/netblock` (new, self-contained)
**Phase:** Phase 1 add-on (Windows-only), consistent with the existing EDR conventions in `CLAUDE.md`.

---

## 1. Goal & threat model

Block network connections **to and from** blacklisted **IP addresses, domains, and hostnames**, sourced from **ThreatWinds** threat intelligence. This closes the network arm of the EDR: the file/AMSI paths stop malicious *content*; this stops malicious *communication* — C2 callbacks, data exfiltration to known-bad infrastructure, and inbound contact from known-hostile IPs.

The feature both **enforces** (drops the connection) and **detects** (emits a branded event with process attribution) so the platform sees every contact attempt with a blacklisted indicator, blocked or not.

### 1.1 The four load-bearing decisions (resolved during brainstorming)

| Decision | Choice | Rationale |
|---|---|---|
| **IP enforcement mechanism** | **WFP user-mode dynamic filters** (`fwpuclnt.dll`) | No kernel driver (hard rule). Fast, in-memory, scoped to our own sublayer, invisible to users, scales to large IP sets, covers in+out & IPv4+IPv6. The "real EDR" path. |
| **Domain/hostname enforcement** | **DNS-Client ETW → reactive WFP block** | WFP matches IPs, not names. We watch DNS resolutions (same ETW infra the ransomware guard uses); when a blacklisted name resolves, we block the returned IP(s) and emit an event naming the domain + querying process. Self-updating as IPs rotate. |
| **Feed transport** | **UTMStack server mirror** (configurable base URL) | Platform pulls ThreatWinds with one org account and mirrors files at the agent dependencies endpoint (same model as the ClamAV signature mirror). No per-agent ThreatWinds credentials; no N-agents rate-limit exposure; works without direct threatwinds.com reachability. |
| **Default posture** | **On + enforcing**, with mandatory safety rails | Highest protection out of the box. Makes the self-lockout allowlist, fail-open-on-crash, and conservative feed-confidence defaults **mandatory** design elements, not options. |

### 1.2 Capability scope (honest claims)

**In scope**
- Block outbound & inbound connections to/from blacklisted **IPv4/IPv6** addresses and **CIDR** ranges (WFP).
- Block **domains & hostnames** reactively via DNS-Client ETW → block resolved IP(s).
- Emit a branded `network_watcher` event for every enforced block (and, when enforcement is off, every observed match), with the remote endpoint, matched indicator, and the responsible process.
- Feeds refreshed on a cadence from the UTMStack mirror: accumulative snapshot to bootstrap + daily deltas.

**Explicitly out of scope (documented limits, per the "honest capability claims" convention)**
- **URLs / full request paths** — cannot enforce without TLS MITM; not attempted. (ThreatWinds URL/hash feeds remain the file-scan / AMSI path's job.)
- **Encrypted DNS (DoH/DoT)** bypasses the DNS-Client ETW provider: a name that resolves over DoH is not caught by the *domain* path. Its IP is still blocked if that IP is independently on the IP blocklist.
- WFP blocks at the ALE connect/accept classification — the connect is **refused**, not a retroactive teardown of an already-established socket. This is a feature (pre-connection), and the residual is negligible, but we state it.
- **No kernel driver, ever** — everything is user-mode WFP + ETW.

---

## 2. Architecture

New self-contained package `agent/edr/netblock`, shaped like the `ransomware` module (interface + build-tag platform split + pure-logic matcher + `_test.go` siblings). It is orchestrated by `service.startPipeline` and emits through the existing spool; it never talks to the platform directly.

### 2.1 Package layout

| File | Responsibility | Platform |
|---|---|---|
| `feed.go` | ThreatWinds-mirror downloader + scheduler (freshclam-style: immediate `updateOnce()` then ticker). Accumulative `.tar.gz` bootstrap + daily `.ndjson` deltas; checksum verify; parse → cache → hot-swap store. | portable |
| `parse.go` | Parse the accumulative tar.gz members and the daily ndjson records into `Indicator` values. Pure, unit-tested. | portable |
| `store.go` | In-memory blocklist: IP/CIDR structure (longest-prefix) + domain/hostname suffix set. Atomic hot-swap on refresh. Backed by `cache` tables for restart survival. | portable |
| `matcher.go` | Pure membership logic: IP-in-CIDR, domain + parent-domain suffix match, allowlist-aware. No I/O. Fully unit-testable. | portable |
| `allowlist.go` | Network allowlist (never-block) resolution: built-in defaults + configured entries + `AllowPrivateRanges`. | portable |
| `enforcer.go` | Platform-agnostic enforcement API: `Blocker` interface (`Apply(diff)`, `BlockIPs([]ip, ttl)`, `RemoveAll()`), diff engine (add/remove filters vs desired set). | portable |
| `wfp_windows.go` / `wfp_other.go` | Real WFP engine session (dynamic sublayer, IPv4+IPv6 ALE connect/recv-accept BLOCK filters, transactional batch add/remove) vs. no-op stub that builds/tests on macOS. | split |
| `dnsfeed_windows.go` / `dnsfeed_other.go` | DNS-Client ETW subscription → `(queryName, resolvedIPs, pid)`; `_other` = nop feed blocking on `ctx.Done()`. | split |
| `connaudit_windows.go` / `connaudit_other.go` | WFP net-event ETW reader (`FWPM_NET_EVENT` classify-drop/block) → blocked 5-tuple + PID + app path, for detection events; `_other` = nop. | split |
| `manager.go` | Orchestrator: wires feed → store → enforcer, consumes DNS feed + conn-audit, applies allowlist, emits events, exposes health for `status.json`, re-applies filters on startup. `Run(ctx)`. | portable |
| `*_test.go` | Siblings for every portable file. | host |

### 2.2 Data flow

```
                     ┌─ UTMStack server mirror (ThreatWinds) ─┐
   feed.go ◄─────────┤ accumulative .tar.gz  +  daily .ndjson │  (blocklist_mirror base URL)
      │              └────────────────────────────────────────┘
      ▼ parse + checksum verify
   cache tables (IndicatorRecord, FeedState) ──► store.go (in-mem IP/CIDR + name set, hot-swap)
      │                                              ▲
      ▼ diff on refresh                              │ lookups (allowlist-aware)
   enforcer/WFP filters (BLOCK IPv4/IPv6, in+out, own sublayer, dynamic session)
      ▲                                              │
      │ reactive add (short TTL)                     │ match(name) → resolved IPs
   dnsfeed (ETW DNS-Client) ─────────────────────────┘
      │ blacklisted name resolved → block IP(s) + emit (names domain + PID)
      ▼
   connaudit (ETW WFP net-events) ─► block occurred ─► event.NewNetworkEvent(...)
                                          │  source = "network_watcher"
                                          ▼  ToJSON → spool.Append → agent EDRRelay → platform
```

### 2.3 Lifecycle & wiring

In `service.startPipeline` (mirrors the ransomware wiring), gated by config:

```go
if cfg.Blocklist.Enabled {
    nb := netblock.NewManager(netblock.Deps{
        Cfg: cfg, Cache: c, Spool: sp, Allow: ex, /* unified allowlist */
    })
    p.netblock = nb // handle stored on program for status.json
    goSafe("netblock", func() { nb.Run(ctx) })
}
```

`nb.Run(ctx)`:
1. Load cached indicators → build store → (if `Enforce`) **re-apply** WFP filters from the last-known blocklist (startup convergence).
2. Start the feed scheduler (`feed.Run`).
3. If Windows + enforcing: start the DNS-Client ETW feed and the WFP net-event conn-audit reader.
4. Serve lookups and drive the enforcer on each feed refresh (diff-apply).
5. On `ctx.Done()`: tear down (dynamic filters auto-release; see fail-open below).

---

## 3. Enforcement & fail-safety (WFP)

- **Dynamic session filters.** The WFP engine is opened with `FWPM_SESSION_FLAG_DYNAMIC`; all our filters/sublayer live in that session. When the engine handle closes (service stop **or crash/kill**), WFP auto-removes every filter. ⇒ **fail-open**: a dead EDR never leaves a machine blackholed. On (re)start the manager re-derives and re-applies filters from the cached blocklist, so steady state is identical.
- **Own sublayer.** A dedicated `FWPM_SUBLAYER` (fixed GUID) isolates our filters — clean teardown, no interference with the Windows Firewall's own filters, no UI clutter.
- **Layers.** BLOCK filters at `FWPM_LAYER_ALE_AUTH_CONNECT_V4/V6` (outbound) and `FWPM_LAYER_ALE_AUTH_RECV_ACCEPT_V4/V6` (inbound), conditioned on `FWPM_CONDITION_IP_REMOTE_ADDRESS` (address or mask). Direction set by `cfg.Blocklist.Direction` (`both` default; `out` / `in` to narrow).
- **Batching.** Feed refreshes apply as a diff inside a single `FwpmTransactionBegin/Commit` so large add/remove sets are atomic and fast; steady-state churn is only the delta.
- **Reactive domain filters** carry a short TTL (e.g. resolved-IP block expires after N minutes unless re-observed) so DNS-derived IP blocks self-clean as intel/rotation changes.
- **Kill-switch.** `blocklist.enabled=false` (or `enforce=false`) on next reload → `RemoveAll()` clears the sublayer. Disabling is total and immediate at reload.

### 3.1 Self-lockout prevention (mandatory, because default is enforcing)

An always-applied **network allowlist** is checked before any indicator becomes a filter. Built-in, non-removable defaults:
- Loopback (`127.0.0.0/8`, `::1`), link-local (`169.254/16`, `fe80::/10`), multicast/broadcast.
- The configured **UTMStack platform / dependencies endpoint(s)** (`cfg.Server`, mirror host) — the agent must never lose contact with its own management plane.
- The host's configured **DNS servers** and default gateway (resolved at startup) — never blackhole name resolution or the LAN egress.

Configurable:
- `AllowPrivateRanges` (**default true**): allowlist RFC1918 (`10/8`, `172.16/12`, `192.168/16`, `fc00::/7`) so intra-LAN and management traffic is never blocked by a bad indicator.
- The unified `allowlist` block gains network entries (IP/CIDR/domain) so an admin can exempt a specific false-positive without disabling the feature.

A blacklisted indicator that intersects the allowlist is **dropped with a logged reason** (surfaced in status), never turned into a filter.

---

## 4. Domain / hostname handling (DNS-Client ETW)

- Subscribe to **`Microsoft-Windows-DNS-Client`** ETW (query-completed events carrying the queried name, the returned A/AAAA records, and the querying PID).
- On each resolution, `matcher.MatchName(queryName)` tests the name and its parent domains (suffix match) against the store, allowlist-aware.
- On a hit **and** `Enforce`: add short-TTL WFP block filters for each resolved IP, then emit a `network_watcher` event (`action=blocked`, `verdict=malicious`) carrying `domain`, `remote_ip`(s), and the `process` (from PID). On a hit with enforcement off: emit `action=detected`.
- Because enforcement is at the IP layer, this also transparently covers a later direct re-connect to the same IP (the filter is already in place for its TTL).
- **DoH/DoT limit** (see §1.2) is documented in status and events are only generated for names actually seen by the provider.

---

## 5. Feed ingestion (ThreatWinds via UTMStack mirror)

### 5.1 Source format (ThreatWinds Feeds API, mirrored)

Endpoint shape (mirrored by the platform; the EDR hits the mirror base URL, not threatwinds.com directly):
`…/feeds/v1/download/list/{level}/{type}/{name}`

- **level** ∈ `level1|level2|level3` — accuracy/confidence tiers. **Default: `level1` only** (highest confidence, lowest false-positive blast radius — required given the enforcing default). Config widens to include `level2`/`level3`.
  *Assumption:* level1 = highest confidence. To be confirmed against ThreatWinds semantics with the platform team; the config is a list so the ordering choice is not load-bearing in code.
- **type** ∈ `accumulative` (full snapshot, `.tar.gz`, `application/gzip`) | `daily` (delta, `.ndjson`, `application/x-ndjson`).
- **name** = indicator category. **In scope: `ip`, `domain`, `hostname`.** (`md5`/`sha256` = file path's concern; `url` = out of scope.)
- **Integrity:** the `…/download/checksum` endpoint is fetched and verified before a downloaded artifact is trusted.

### 5.2 Refresh strategy

- **Bootstrap / self-heal:** on first run (or when local state is missing/stale beyond a threshold), pull the **accumulative** snapshot for each in-scope `(level, name)` and rebuild the store.
- **Incremental:** on the cadence (`RefreshHours`, default 6), pull the **daily** ndjson delta(s), apply adds/removes, hot-swap the store, diff-apply the enforcer.
- **Periodic full re-sync** (e.g. weekly) to converge against drift.
- Reuses `shared/http.Download` with custom headers (mirror needs none / or an agent token, TBD by platform) and honors `cfg.SkipCertValidate → SkipTLSVerify`. Scheduler is the freshclam `Run(ctx)` loop pattern (immediate + `time.Ticker`, `ctx.Done()` aware).

### 5.3 Staleness

Feed age is tracked in `FeedState`; if the newest successful sync exceeds a staleness threshold, status reports `feed_stale=true`. Enforcement continues on the last-known-good set (stale intel still blocks; we simply flag it). We never auto-flush the blocklist on a failed fetch.

---

## 6. Data model

### 6.1 Cache tables (`cache/netblock.go`, added to `AutoMigrate`)

```go
// One row per active indicator in the local blocklist (survives restart).
type IndicatorRecord struct {
    Value     string    `gorm:"primaryKey"` // "1.2.3.4", "10.0.0.0/8", "evil.com"
    Type      string    `gorm:"primaryKey"` // "ip" | "cidr" | "domain" | "hostname"
    Level     int       // 1..3 (source confidence tier)
    ListName  string    // feed category it came from
    FirstSeen time.Time
    LastSeen  time.Time
    Removed   bool      // tombstone for delta removals
}

// Audit trail of enforced blocks / observed matches (bounded; purged like quarantine).
type NetBlockRecord struct {
    ID          uint      `gorm:"primaryKey"`
    Time        time.Time
    Direction   string    // "outbound" | "inbound"
    RemoteIP    string
    RemotePort  int
    PID         int
    ProcessPath string
    Matched     string    // the indicator that matched
    MatchedType string
    Level       int
    Action      string    // "blocked" | "detected"
    Domain      string    // set when the match came via DNS
}

// Feed sync bookkeeping, one row per (level,name) feed.
type FeedState struct {
    Feed          string    `gorm:"primaryKey"` // e.g. "level1/ip"
    LastFullSync  time.Time
    LastDelta     time.Time
    Checksum      string
    IndicatorNum  int
}
```

Accessors follow the existing mutex-guarded GORM pattern (`RLock` reads, `Lock` writes, seed zero timestamps). `NetBlockRecord` gets a `Purge`/`PurgeExpired` like quarantine so the audit table stays bounded.

### 6.2 Event model additions (`event/event.go`)

```go
// new source const
SourceNetwork = "network_watcher"

// new omitempty fields on Event
RemoteIP   string `json:"remote_ip,omitempty"`
RemotePort int    `json:"remote_port,omitempty"`
Domain     string `json:"domain,omitempty"`
Direction  string `json:"direction,omitempty"`  // "outbound" | "inbound"
Indicator  string `json:"indicator,omitempty"`  // the matched blocklist value

// new constructor
func NewNetworkEvent(action, direction, remoteIP string, port int, domain, indicator string, p *ProcInfo) Event
```

- `action` ∈ `blocked` | `detected` (both consts already exist).
- `verdict` = `"malicious"`; `source` = `network_watcher`; `product`/`engine` default to `"UTMStack EDR"` via `ToJSON()`.
- `process` carries PID/image from the DNS-Client PID or the WFP net-event PID when available.

### 6.3 Branding

Per the non-negotiable branding rule: every surfaced event/log/status string says **UTMStack EDR**. The intel *source* is surfaced as **"UTMStack Threat Intelligence"** — the string `threatwinds` must **never** appear in an event field or a surfaced log/status line (internal Go identifiers, config keys, and the mirror URL may reference it, exactly like `clamd`/`ClamdAddr`). The labeling audit is extended with a `threatwinds` grep alongside the existing `clam` grep.

---

## 7. Config, CLI, status

### 7.1 `config/config.go`

New nested block (RansomwareConfig template), on by default:

```go
type BlocklistConfig struct {
    Enabled            bool     `json:"enabled"`              // default true
    Enforce            bool     `json:"enforce"`              // default true (false = detect-only)
    MirrorBaseURL      string   `json:"blocklist_mirror"`     // empty = derive from cfg.Server
    Levels             []int    `json:"levels"`               // default [1]
    IndicatorTypes     []string `json:"indicator_types"`      // default ["ip","domain","hostname"]
    Direction          string   `json:"direction"`            // "both"(default)|"out"|"in"
    RefreshHours       int      `json:"refresh_hours"`        // default 6
    AllowPrivateRanges bool     `json:"allow_private_ranges"` // default true
    DomainBlockTTLMin  int      `json:"domain_block_ttl_min"` // reactive-IP filter TTL, default 30
}
```

- Added to `EDRConfig` as `Blocklist BlocklistConfig json:"blocklist"`.
- Seeded in `Default()` (all safety defaults above); overlaid in `Load()` (presence-detected like `Ransomware`, copied wholesale).
- Path constant `BlocklistDir = filepath.Join(InstallDir, "blocklist")` for feed staging.
- Network allowlist entries live in the unified `Allowlist` (extend with `IPs []string` / reuse a `Networks` list) so `allow` CLI verbs manage them.

### 7.2 `manage.go` (local admin CLI)

- `config set blocklist.enabled|enforce|refresh_hours|allow_private_ranges|direction … ` cases in `applySet` (with validation).
- Extend `allow` to accept network indicators: `allow network <ip|cidr|domain> add|remove|list` (feeds `Allowlist`).
- New `block <ip|cidr|domain> add|remove|list` verb for **manual local blocks** (admin-supplied, merged with the feed set).
- `status` already prints the status doc; blocklist health shows there (below).

### 7.3 `status.json` (`service.writeStatus`)

Add: `blocklist_enabled`, `blocklist_enforce`, `blocklist_indicators` (count by type), `blocklist_feed_age`, `blocklist_feed_stale`, `blocklist_active_filters` (WFP filter count), `blocklist_blocks_recent` (blocks since last status), `blocklist_allowlisted_dropped` (indicators skipped by allowlist). Populated from the `netblock.Manager` handle stored on `program`.

---

## 8. Testing strategy

Follows the established model: pure-Go logic TDD'd with `go test` on macOS; Windows-native pieces behind `//go:build windows` with `!windows` stubs, verified on the VM.

**Host (darwin) unit tests**
- `parse_test.go` — accumulative tar.gz members + daily ndjson records → indicators (fixtures).
- `matcher_test.go` — IP-in-CIDR (v4+v6), domain/parent-suffix match, hostname vs domain, allowlist precedence.
- `store_test.go` — hot-swap atomicity, longest-prefix, delta add/remove/tombstone.
- `allowlist_test.go` — built-in defaults, `AllowPrivateRanges`, platform/DNS/gateway auto-allow, blacklisted∩allowlist drop.
- `feed_test.go` — scheduler loop (injected fetch fn), checksum-fail rejects artifact, staleness flagging, bootstrap-vs-delta selection.
- `enforcer_test.go` — diff engine (desired vs applied → add/remove sets) against a fake `Blocker`.
- `config_test.go` — defaults, load/overlay, presence detection.
- `event_test.go` — `NewNetworkEvent` branding, field mapping.
- Extend the labeling audit: `grep clam` **and** `grep threatwinds` on `edr/` surfaced strings must return nothing.

**VM acceptance (Windows 10.211.55.12)** — verify final outputs, not just logs:
1. Seed a known blacklisted IP; confirm `curl`/connect to it **fails**, a `network_watcher` `blocked` event reaches the platform, and a `NetBlockRecord` is written.
2. Seed a blacklisted domain; DNS resolve + connect **fails**, event names the domain + querying process.
3. Allowlist a would-be-blocked IP → connection succeeds, indicator reported as allowlisted-dropped.
4. **Fail-open:** stop/kill the EDR service → verify all WFP filters are gone (blocked IP reachable again); restart → filters re-applied from cache.
5. Feed refresh: apply a daily delta that removes an indicator → its filter is removed; that adds one → filter appears.
6. Self-lockout guard: confirm the platform endpoint, DNS servers, and (default) RFC1918 are never blocked even if injected as indicators.
7. Cross-build + vet clean on windows/amd64 **and** arm64; all `edr` unit tests green on darwin.

---

## 9. Suggested plan phasing

The module is large; suggested split for `writing-plans` (final call made there):

- **Plan 1 — Feed + IP enforcement (core):** config block, event model, cache tables, feed downloader/parser (mirror, accumulative+daily, checksum), store + matcher + allowlist, WFP enforcer (`wfp_windows.go`/`_other.go`), conn-audit ETW block events, manager wiring, status, CLI, fail-open + re-apply, unit tests, VM acceptance for the IP path.
- **Plan 2 — Domain/hostname reactive blocking:** DNS-Client ETW feed (`dnsfeed_windows.go`/`_other.go`), name matcher integration, reactive short-TTL IP filters, domain detection events, VM acceptance for the domain path.

Both plans are Windows-native at the edges, portable in the core, and end with a VM acceptance task per the project convention.

---

## 10. Open items to confirm with other teams (not blocking design)

1. **Platform mirror** must host the ThreatWinds `ip`/`domain`/`hostname` feeds (accumulative + daily + checksum) at the agent dependencies endpoint — External-deps analog to the ClamAV mirror. Auth model (anonymous vs agent token) TBD by platform.
2. **ThreatWinds level semantics** (which of level1/2/3 is highest confidence) — confirm; config is order-agnostic.
3. **Platform correlation parser** for `source=network_watcher` events (new fields: `remote_ip`, `domain`, `direction`, `indicator`).
4. **Branding sign-off** on surfacing the intel source as "UTMStack Threat Intelligence" (hiding `threatwinds` in surfaced strings, mirroring the `clamav` rule).
