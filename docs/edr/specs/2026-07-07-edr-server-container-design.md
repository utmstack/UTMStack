# UTMStack EDR — Master-Server `edr` Container (Signature Mirror + Threat-Intel Feed Mirror) — Design Spec

**Status:** Approved (design) — 2026-07-07
**Scope:** Server-side delivery of the two content channels the endpoint EDR already consumes — ClamAV signature databases and ThreatWinds network-blocklist indicators — plus the small agent-side polish that makes both work out of the box.
**Companion docs:** `2026-07-02-edr-signature-update-mechanism.md` (Option A evaluation, approved), `2026-07-06-edr-network-blocklist-design.md` (§5 feed contract), `docs/2026-07-07-edr-remaining-work.md` (gap analysis).

---

## 1. Goal

Stand up one new container on the UTMStack master server — service name **`edr`** — that:

1. **Mirrors ClamAV signatures** (full CVDs + cdiff incrementals) from the official CDN using the official `cvdupdate` tool, so the whole fleet updates from one upstream fetch. Solves CDN rate-limiting at fleet scale and gives air-gapped installs a signature path.
2. **Mirrors ThreatWinds indicator feeds** (`ip`/`domain`/`hostname` accumulative lists) using the platform's existing ThreatWinds credentials, normalized to the wire format the agent's `edr/netblock` feed client already implements and VM-validated.

Agents pull both over the **existing agent-facing plane** — `https://<server>:9001` served by `agentmanager` with the platform TLS cert. No new exposed port, no new firewall rule, no per-agent upstream credentials.

**Division of labor (confirmed by the signature-update spec §4):** the agent's `dependency` mechanism keeps shipping **binaries**; **freshclam** owns signatures (cdiff incrementals, Talos-signature verification); the **netblock feed client** owns indicators. All three ride the same host:port + TLS trust surface.

---

## 2. Architecture

```
ClamAV CDN ──────── cvdupdate (scheduled) ──┐
                                            ├─► [edr container] ── writes ──► <DataDir>/edr-mirror
ThreatWinds Feeds API ── sync + normalize ──┘        │                        (host dir, shared)
  (api-key/secret from backend                       │                              │ read-only
   /api/v1/config, INTERNAL_KEY auth)           status.json                         ▼
                                                              [agentmanager :9001] StaticFS /private/edr/**
                                                                                    ▲
      agents: freshclam (PrivateMirror)  +  edr/netblock feed client ───────────────┘
              same host, same port, same TLS cert they already use for dependencies
```

- The `edr` container is a **producer, not a server**. It maintains a static mirror tree in a shared host-dir volume. Serving stays in `agentmanager` (one added `StaticFS` route) — gin's `http.FileServer` gives Range and `If-Modified-Since` support, which freshclam uses.
- **Stale, never empty:** a failed upstream sync never deletes or replaces good artifacts. Agents keep enforcing last-known-good content and surface staleness; monitoring watches the mirror's `status.json`.

### 2.1 Mirror tree & URL space

```
<DataDir>/edr-mirror/                       →  https://<server>:9001/private/edr/
├── status.json                             →  …/private/edr/status.json
├── signatures/                             →  …/private/edr/signatures/          (freshclam PrivateMirror base)
│   ├── main.cvd  daily.cvd  bytecode.cvd
│   └── *.cdiff                                (cvdupdate-managed layout)
└── feeds/v1/download/list/
    └── level{N}/accumulative/
        ├── {name}                          →  …/feeds/v1/download/list/levelN/accumulative/{name}
        └── {name}.sha256                        name ∈ ip | domain | hostname
```

**Branding constraint on the URL space:** mirror URLs appear in agent-side error logs (e.g. a failed-fetch warning), and the non-negotiable branding rule forbids `clamav`/`clamd`/`threatwinds` in surfaced strings. Therefore the path is `signatures/` — never `clamav/` — and the feed path names no vendor. Server-side code and logs (platform-internal) may name the vendors freely.

---

## 3. The `edr` container

### 3.1 Repo layout & image

- New top-level directory **`edr/`** in the UTMStack repo (peer of `agent-manager/`), its own Go module `github.com/utmstack/UTMStack/edr`. One binary: the sync daemon.
- `edr/Dockerfile`: `ubuntu:24.04` base (matches `agent-manager`), installs `python3` + `pip` + **`cvdupdate`** (the official ClamAV mirror tool — the single non-Go component, mirroring the clamd precedent on the agent), copies the Go daemon. `CMD ["/app/edr"]`.
- Image `ghcr.io/utmstack/utmstack/edr:${UTMSTACK_TAG}`, built/pushed by the v12 deployment pipeline like the other services.
- **Also serves the mirror tree over plain HTTP on port 9002** (`internal/serve`, `http.FileServer` at `/private/edr/…`, same layout as the HTTPS path) for self-signed deployments where freshclam cannot validate TLS (§7.4). Producer-plus-thin-server; `agentmanager` still serves the HTTPS view on 9001. Publishes `9002:9002` in the installer compose.

### 3.2 Signature mirror (cvdupdate)

- The daemon shells out to `cvd update` on a schedule (default hourly, ±jitter) with `CVD_HOME`/config pointed so the database directory is `/mirror/signatures`. `cvdupdate` is rate-limit-compliant and maintains exactly the CVD+cdiff layout freshclam `PrivateMirror` consumes; this one job is the fleet's only CDN contact.
- Integrity is ClamAV's own: CVDs are Talos-signed and freshclam/clamd verify on download and load. The mirror is **untrusted by design** — a torn or tampered file fails client-side verification and is retried next cycle. No extra checksum machinery for signatures.
- First run performs the initial full download; subsequent runs fetch only cdiffs. Failures log + retry next cycle (stale, never empty).

### 3.3 Feed mirror (ThreatWinds sync + normalize)

- For each configured `(level, name)` pair (defaults: level `1`; names `ip,domain,hostname`), on a schedule (default every 6 h, matching the agent's default `refresh_hours`):
  1. Download the **accumulative** artifact from the ThreatWinds Feeds API. **Confirmed against the live docs (2026-07-08):** the path is `{TW_API_URL}/api/feeds/v1/download/list/{level}/accumulative/{name}` — note the **`/api` prefix**; `{level}` is the string `level1`/`level2`/`level3` (not numeric); `{type}` ∈ `accumulative`(`.tar.gz`, `application/gzip`) / `daily`(`.ndjson`); headers `api-key`/`api-secret`. Upstream integrity is a single `…/api/feeds/v1/download/checksum` → `bases.sum` of **MD5** hashes (server→ThreatWinds hop only; not the agent contract).
  2. **Normalize** to the agent contract: extract the tar.gz members → one indicator per line → dedupe → sort → gzip. The normalizer is the absorption layer for upstream format drift.
  3. Write atomically (tmp file + rename) to `feeds/v1/download/list/level{N}/accumulative/{name}`, then the sibling `{name}.sha256` — **bare lowercase sha256-hex of the compressed artifact bytes** (the agent verifies exactly this, VM-validated). Checksum written after the artifact.
  4. Per-artifact failure isolation: one failed feed never blocks the others and never removes the previous good artifact.
- **Daily deltas — server-computed by diffing accumulatives.** Rather than consume ThreatWinds' `daily` endpoint (its ndjson record schema is undocumented — entities are hash-IDed with the raw value in attributes), the mirror computes the delta itself: each cycle it diffs the **previous** published accumulative value-list against the **new** one and writes `feeds/v1/download/list/level{N}/daily/{name}` as **plain ndjson** — one `{"value","type","op":"add"|"del"}` per line (+ `.sha256` sibling = sha256 of the ndjson bytes). This gives real adds **and** removes with zero dependency on an undocumented upstream schema, and since the mirror is one server fetching for the whole fleet, diffing the full snapshot server-side costs nothing the model didn't already pay. The `type` per value follows the accumulative rule (`ip`→`cidr` if the value has `/` else `ip`; `domain`/`hostname`→the name). First cycle (no previous) → all-adds; agents bootstrap from the accumulative regardless.

### 3.4 ThreatWinds credentials

- Reuse the platform's existing flow (`plugins/feeds`): the backend auto-registers with ThreatWinds and stores `api-key`/`api-secret` retrievable via `GET {UTM_HOST}/api/v1/config/<key>` authenticated with `INTERNAL_KEY`. The daemon polls until credentials exist (the backend owns registration; the `edr` container never registers).
- Env overrides `TW_API_KEY`/`TW_API_SECRET` for dev/test rigs without a backend.

### 3.5 `status.json` (mirror-root freshness document)

Written after every sync cycle (atomic):

```jsonc
{
  "signatures": { "last_success": "…", "last_attempt": "…", "last_error": "", "databases": {"daily.cvd": "27461", …} },
  "feeds":      { "level1/ip": { "last_success": "…", "indicators": 12345, "sha256": "…", "last_error": "" }, … },
  "updated": "…"
}
```

Served at `…/private/edr/status.json` — the platform's single point to monitor mirror freshness (a stale mirror ⇒ fleet-wide stale detection, so this is a feature: one place to watch).

### 3.6 Configuration (env)

| Var | Default | Purpose |
|---|---|---|
| `INTERNAL_KEY` | — (required) | backend config-API auth |
| `UTM_HOST` | `http://backend:8080` | backend base URL |
| `EDR_MIRROR_DIR` | `/mirror` | mirror tree root |
| `EDR_SIG_SYNC_MINUTES` | `60` | cvdupdate cadence |
| `EDR_FEED_SYNC_HOURS` | `6` | feed cadence |
| `EDR_FEED_LEVELS` | `1` | comma list |
| `EDR_FEED_NAMES` | `ip,domain,hostname` | comma list |
| `TW_API_URL` | `https://apis.threatwinds.com` | upstream override (dev: `apis.dev.threatwinds.com`) |
| `TW_API_KEY` / `TW_API_SECRET` | empty | dev/test credential override (skips backend fetch) |

### 3.7 Failure posture / air-gap

- Upstream unreachable → mirror stale, `status.json` records the error, agents keep last-known-good. No auto-flush, ever.
- Air-gapped sites: seed `<DataDir>/edr-mirror/` manually (documented); the daemon tolerates pre-seeded content and simply fails upstream syncs.

---

## 4. `agentmanager` change (one route)

`agent-manager/updates/updates.go`: alongside the existing dependencies route, add

```go
group.StaticFS("/edr", http.Dir(config.EDRMirrorFolder)) // config: EDRMirrorFolder = "/edr-mirror"
```

If the volume isn't mounted (older installer), the route simply 404s and agents fall back — no startup failure.

---

## 5. Installer changes

`installer/docker/stack.go`:
- `stackConfig.EDRMirror = utils.MakeDir(0777, cnf.DataDir, "edr-mirror")`.
- `Services` gains `{Name: "edr", Priority: 3, MinMemory: 150, MaxMemory: 512}` (memory-balanced like the rest).

`installer/docker/compose.go` — new service entry, standard shape:

```yaml
edr:
  image: ghcr.io/utmstack/utmstack/edr:${UTMSTACK_TAG}
  volumes: [ "<DataDir>/edr-mirror:/mirror" ]
  environment: [ "INTERNAL_KEY=…", "UTM_HOST=http://backend:8080" ]
  depends_on: [ backend ]
  deploy: { placement: manager, resources from ServiceResources["edr"] }
  logging: json-file 50m
```

and `agentmanager` gains the read-only mount `"<DataDir>/edr-mirror:/edr-mirror:ro"`.

---

## 6. CI changes

`.github/workflows/v12-deployment-pipeline.yml`: a `build_edr` job — `go build` in `edr/`, `docker/build-push-action` with context `./edr`, tag `ghcr.io/utmstack/utmstack/edr:${tag}`; added to `all_builds_complete`. (Same reusable pattern as backend/frontend.)

---

## 7. Client-side changes (agent repo — small, this workspace)

### 7.1 Checksum fetch: query → path

`edr/netblock/feed.go`: checksum URL becomes `{base}/download/list/{level}/accumulative/{name}.sha256` (a static tree cannot serve query-param URLs). Format unchanged (bare lowercase sha256-hex of compressed bytes).

### 7.2 Auto-derivation + new defaults

New constants in `edr/config`: mirror HTTPS port `9001`, mirror HTTP port `9002`, base path `/private/edr`.

- **Blocklist:** `blocklist_mirror` empty (default) → `https://<Server>:9001/private/edr/feeds/v1` (fixes the currently scheme-less, unusable fallback). The netblock feed uses the agent's own skip-verify HTTP client, so it stays on 9001-HTTPS in both cert modes. Explicit URL still wins.
- **Signatures — `signature_mirror` semantics redefined** (safe: nothing is shipped/committed yet):
  - `""` (default) → **auto**, keyed off `skip_cert_validate` (see §7.4).
  - `"cdn"` → official CDN only (the old `""` behavior, now explicit).
  - explicit URL → that mirror.
- `signature_fallback`: `"cdn"` (default) | `"none"` (air-gapped).

Fresh installs therefore prefer the server mirror and degrade cleanly to the CDN when the server predates the `edr` container.

### 7.3 Signature failover, freshness health, status (folds in the deferred mirror plan)

- Mirror→CDN failover: after `mirrorFailThreshold` (3) consecutive freshclam failures in mirror mode with `signature_fallback=="cdn"`, write the CDN freshclam.conf, run one cycle, restore mirror config (retry mirror next cycle).
- `Feed` records `lastSuccess`; staleness emits a branded `health` event; `status.json` gains `signature_source` (`"utmstack-mirror"`/`"official-cdn"` — never the engine name) and `signature_last_update`.

### 7.4 freshclam TLS trust — flag-driven transport (mirrors the agent's two cert modes)

**Corrected after VM acceptance.** The agent imports no certificate anywhere; every channel is driven by one flag, `SkipCertValidation` (YAML `insecure`, install arg): `false` = full TLS validation against the OS trust store, `true` = skip-verify (self-signed). The EDR already receives it as `skip_cert_validate`. freshclam is the sole component that cannot skip-verify, and on the ClamAV Windows build it validates via the **Windows cert store and ignores `CURL_CA_BUNDLE`** (VM-proven). So an earlier `CURL_CA_BUNDLE`/TOFU-capture approach was built and then **removed** — it cannot work on this engine build, and importing into the Windows root store would be stricter than any other agent channel.

Instead, `engine.ResolveMirrorURL` keys the signature transport off `skip_cert_validate`, exactly paralleling the agent's own two modes:
- **`skip_cert_validate=false`** → `PrivateMirror https://<Server>:9001/private/edr/signatures`. freshclam validates the server cert against the OS store natively (the mode's precondition is an OS-trusted cert). No new port, no cert handling.
- **`skip_cert_validate=true`** → `PrivateMirror http://<Server>:9002/private/edr/signatures`. The `edr` container serves the mirror tree over **plain HTTP** on 9002; freshclam needs no cert. CVD integrity is guaranteed by ClamAV's Talos **signature**, verified on download and load regardless of transport (§8) — HTTP is safe for this public, signed payload. **No Windows-root-store modification, ever.**

CDN fallback (§7.3) is unchanged and orthogonal. This deletes the §7.4-original cert-capture logic entirely.

### 7.5 Out of scope (client)

Engine-package download/provisioning (gap item 4 — separate plan), Authenticode signing. (Daily-delta wiring — previously deferred — is now **implemented**: the agent runs a periodic full accumulative re-sync, `full_resync_hours` default 24, and fetches the server-computed daily ndjson delta on the intervening cycles via its existing `ParseDaily`/`applyDelta`, total-failure-safe. See §3.3.)

---

## 8. Security & integrity

| Channel | Transport | Content integrity |
|---|---|---|
| Signatures | HTTPS:9001 when `skip_cert_validate=false` (freshclam validates via the OS trust store); HTTP:9002 when `true` (self-signed — freshclam can't validate, so transport is unauthenticated by design) | **ClamAV/Talos signature** on every CVD/cdiff, verified client-side on download and load — mirror is untrusted by design, so transport authenticity is not the guarantee |
| Indicators | HTTPS (agent honors `SkipCertValidate`) | sha256 sibling verified before swap; failed verify = artifact skipped, last-good retained |
| Endpoint auth | Anonymous read, same as the existing `/private/dependencies` route | Server-side secrets (`INTERNAL_KEY`, TW creds) never leave the Docker network |

The mirror content is public-by-nature (signature DBs, threat indicators); anonymous read matches the existing dependencies posture and adds no new exposure.

---

## 9. Compatibility matrix

| Agent | Server | Behavior |
|---|---|---|
| old | new | Unaffected (new routes unused; dependencies path unchanged) |
| new | old (no `edr` container) | Signatures: mirror probe fails → CDN fallback (default). Blocklist: feed fails → stale/empty + status flag (same as today) |
| new | new | Everything works with zero config: signatures from server mirror (CDN fallback), indicators from server mirror |
| air-gapped | new | `signature_fallback:"none"`; mirror seeded/synced centrally; blocklist from mirror |

---

## 10. Testing strategy

**Server (host, `go test` in `edr/`):** normalizer (fixture tar.gz → gzip newline list, dedupe), atomic writer + checksum-sibling ordering, scheduler loop (injected clock/fetch), backend credential client (httptest), status.json writer, cvdupdate invocation wrapper (command runner injected).

**Container (local Docker):** build the image; run against a fake ThreatWinds upstream (local httptest server with fixture artifacts) + a real (or fixture-seeded) cvdupdate run; verify the mirror tree layout, checksums, atomicity, status.json.

**Client (existing model):** unit tests on darwin for checksum-path change, URL derivation, failover, health; both Windows arches build+vet; labeling audit (`clam` + `threatwinds`) stays green — including the new URL constants.

**VM acceptance (final task, per project convention — verify final outputs):** run the `edr` container on the Mac host with the mirror served at `https://10.211.55.2:9001/private/edr/**` (static TLS front mimicking agentmanager). On the VM: fresh `edr.json` auto-derives both URLs; freshclam pulls **and signature-verifies** from the mirror (tamper test → rejected), pulls a **cdiff** on the next change; netblock enforces an indicator planted in the mirrored feed; stop the mirror → signature CDN failover fires + blocklist goes stale-with-flag; restart → mirror mode restored. Labeling audit on live events.

**Installer changes** are verified by compose-generation unit coverage where present and by code review; full swarm deploy is validated in the platform's staging pipeline (outside this repo's test rig).

---

## 11. Plan phasing

One coordinated implementation plan (user decision), ordered so the server container exists before the client is pointed at it:

1. Server: `edr/` module — feed normalizer, sync daemon, cvdupdate wrapper, status writer (TDD).
2. Server: Dockerfile + local container acceptance.
3. Platform wiring: `agentmanager` route, installer compose/stack, CI job.
4. Client: checksum path, URL derivation + defaults, failover, health, CA bundle.
5. End-to-end VM acceptance.

**Mechanical prerequisite:** the local checkout is sparse to `agent/`+`shared/` — widen it (`git sparse-checkout add agent-manager installer .github`) and create `edr/`.

---

## 12. Open items (confirm during implementation; absorbed by the normalizer either way)

1. ~~Exact ThreatWinds Feeds API paths / auth-header names / artifact container format~~ **RESOLVED 2026-07-08** against the live docs: path `…/api/feeds/v1/download/list/{level}/{type}/{name}` (the `/api` prefix was added), `{level}`=`level1..3`, `{type}`=`accumulative`(.tar.gz)/`daily`(.ndjson), headers `api-key`/`api-secret`. The undocumented `daily` ndjson schema is sidestepped entirely — the mirror computes deltas by diffing accumulatives (§3.3). One remaining live-only check: whether `accumulative` members are always plain value-lists for every `name` (the normalizer tolerates both tar.gz and plain gzip); confirm once real credentials are available end-to-end.
2. ThreatWinds level semantics (level1 = highest confidence assumed; config is order-agnostic).
3. Whether the platform wants third-party ClamAV feeds (Sanesecurity etc.) mirrored through the same cvdupdate instance — deferred until the base mirror is proven (signature spec §9).
