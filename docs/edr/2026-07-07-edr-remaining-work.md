# UTMStack EDR — Remaining Work (Gap Analysis)

**Date:** 2026-07-07
**Status of the agent side:** Phase-1 Plans 1–4, the Ransomware Guard, and the Network Blocklist (Plans 1–2) are all implemented, unit-tested, cross-built for windows/amd64+arm64, and VM-validated end-to-end. What remains is almost entirely **server-side and delivery-pipeline work** — the endpoint code is done, but nothing on the UTMStack master server yet feeds or ships it.

---

## 1. The two big gaps (this document's focus)

### 1.1 ClamAV signature updates from the master server — **missing entirely (server side)**

**Today:** every agent's `freshclam` pulls directly from the official ClamAV CDN (`database.clamav.net`). This works for pilots but:
- The CDN rate-limits and can **block** networks that poll too aggressively; ClamAV's own guidance is that any fleet beyond a handful of endpoints must run a **private mirror** (`cvdupdate`).
- Air-gapped / egress-restricted customers get **no signature updates at all**.
- The client-side hook is already built: setting `signature_mirror` in `edr.json` switches the generated freshclam config to `PrivateMirror <url>`. There is simply nothing to point it at.

**Missing (server):** a `cvdupdate`-based mirror service on the master server — one scheduled upstream fetch (CVDs + cdiff incrementals) serving the CVD directory over the server's existing agent-facing HTTPS. Designed in `docs/superpowers/specs/2026-07-02-edr-signature-update-mechanism.md` (Option A approved); never implemented.

**Missing (client polish, deferred plan already written):** `signature_fallback` policy (mirror→CDN failover), `"auto"` mirror-URL derivation from the registered server, signature-freshness health in `status.json`/events. See `docs/superpowers/plans/2026-07-02-edr-signature-mirror-plan.md`.

### 1.2 ThreatWinds threat-intelligence feeds from the master server — **missing entirely (server side)**

**Today:** the agent's network-blocklist module (`edr/netblock`) is fully implemented and VM-validated: it pulls `ip`/`domain`/`hostname` indicator lists from a configurable mirror base URL, verifies checksums, hot-swaps the in-memory blocklist, and enforces via WFP. It was validated against a **local test mirror** — in production there is nothing to pull from. The feature ships **on by default**, so a fresh install will log feed failures and enforce nothing until the server mirror exists.

**Missing (server):** a service that authenticates to the **ThreatWinds Feeds API** with the platform's credentials (the backend already auto-registers and stores an `api-key`/`api-secret` — see `plugins/feeds/internal/client/threadwinds_setup.go`), downloads the accumulative snapshots (and later daily deltas) for the configured levels/types, **normalizes** them to the agreed agent contract, and serves them to the fleet. One upstream account, one fetch, fans out to N agents — no per-agent ThreatWinds credentials, no rate-limit exposure.

**The agent contract the server must satisfy** (already implemented + VM-validated in `edr/netblock/feed.go`):

| Artifact | URL shape (relative to the mirror base) | Format |
|---|---|---|
| Accumulative list | `download/list/{level}/accumulative/{name}` (`level` ∈ `level1..3`, `name` ∈ `ip`,`domain`,`hostname`) | **plain gzip'd newline list** (not tar.gz — VM-validated wire format) |
| Checksum | `download/checksum?level=…&type=accumulative&name=…` (today; see §3 note) | bare lowercase sha256-hex of the compressed artifact bytes |
| Daily delta (future) | `download/list/{level}/daily/{name}` | ndjson `{value,type,op}` — parser exists, fetch loop not yet wired |

**Missing (client, small):** the auto-derived fallback URL is scheme-less (`<host>/feeds/v1`) and unusable — `blocklist_mirror` must currently be set by hand. Needs the same `"auto"` derivation treatment as the signature mirror once the server path is fixed.

---

## 2. Other missing pieces (complete inventory)

| # | Gap | Owner | Notes |
|---|---|---|---|
| 1 | **ClamAV signature mirror service** (server) | this project (server-side plan) | §1.1 |
| 2 | **ThreatWinds feed mirror service** (server) | this project (server-side plan) | §1.2 |
| 3 | **Hosting the EDR artifacts at the dependencies endpoint**: `utmstack_edr_windows_{amd64,arm64}.exe` + per-platform **ClamAV engine packages** (clamd + support libs + seed databases) | CI / release pipeline | Agent dependencies are baked into the `agent-manager` image at build time (`agent-manager/Dockerfile` copies `./dependencies/agent/`); the CI workflow must build/stage these artifacts the same way it already stages the collector. |
| 4 | **Engine provisioning on the endpoint** | this repo (agent/EDR code) | No code anywhere downloads the engine: today `engine/` is provisioned manually on the test VM. Needs an enable-time (or dependency-driven) download+unzip of the per-platform engine package from the dependencies endpoint. |
| 5 | **Signature-mirror client polish** | this repo (deferred plan exists) | `signature_fallback`, `"auto"` URL, failover, freshness health. |
| 6 | **Netblock feed client polish** | this repo | Fix auto-derive URL; align checksum URL with what the server serves; wire the daily-delta fetch (`applyDelta` currently has no caller — full re-bootstrap each cycle is the live behavior and is acceptable meanwhile). |
| 7 | **Authenticode signing** of `utmstack_edr.exe` + `utmstack_amsi.dll` | release / external cert | Ordinary code-signing, not the driver path. |
| 8 | **amd64-host AMSI runtime test** | this repo (VM acceptance) | arm64 runtime-proven; amd64 pending a suitable host. |
| 9 | **Platform correlation parser** for `DataType="utmstack_edr"` (incl. `network_watcher` fields: `remote_ip`, `domain`, `direction`, `indicator`) | platform team | Events already arrive at the platform; they need first-class parsing/dashboards. |
| 10 | **Sigma rules + SOAR playbooks** for behavioral / ransomware / network telemetry | platform team | |
| 11 | **Ransomware Guard minor known issues** | this repo | Empty culprit image when WMI lags; possible duplicate `contained` event; sub-second T1490 processes can miss the WMI poll window. |

Items 1–2 (plus the thin client polish in 5–6) are what the **master-server EDR container plan** covers. Items 3–4 ride the same train (the artifacts can live in the same delivery path). Items 7–10 are external to this repo.

---

## 3. Recommended architecture: the `edr` server container

### What the agents' existing update mechanism is — and why it's the wrong tool for feeds

The agent's `dependency` mechanism (`dependency.Reconcile`) downloads **whole, versioned binary files** from `https://<server>:9001/private/dependencies/agent/<file>` (served by the `agentmanager` container's gin static server over the platform TLS cert) and runs install hooks. It is the right tool for **shipping binaries** (updater, EDR executable, engine package) and the wrong tool for **content that changes continuously**:

- **Signatures:** `daily.cvd` changes several times a day at ~60 MB. The dependency mechanism would re-ship the full file every time to every endpoint. freshclam's **cdiff incrementals** (a few hundred KB) exist precisely to avoid this, and require a ClamAV-native mirror layout. Integrity is also better: CVDs are **ClamAV-signed** and freshclam/clamd verify the signature on download and load — the mirror is untrusted by design. (Full evaluation: signature-update spec §4 — Option A approved.)
- **Threat-intel indicators:** the EDR already has its own feed client with checksum verification, hot-swap, staleness handling, and cache-backed restart survival. It only needs a URL that exists.

**So: keep one host, one port, one trust surface — but three logical channels over it:**

| Channel | Client | Path under `https://<server>:9001/` |
|---|---|---|
| Binaries (agent, updater, EDR exe, engine pkg) | `dependency.Reconcile` (unchanged) | `/private/dependencies/agent/…` |
| ClamAV signatures | `freshclam PrivateMirror` (hook already built) | `/private/edr/clamav/…` |
| ThreatWinds indicators | `edr/netblock` feed client (already built) | `/private/edr/feeds/v1/…` |

### The `edr` container

A new Swarm service in the installer's generated compose (same pattern as every other service; image `ghcr.io/utmstack/utmstack/edr:${UTMSTACK_TAG}`, manager-node placement, memory-balanced via `stack.go`). It is a **producer, not a server**:

```
ClamAV CDN ──cvdupdate (scheduled)──┐
                                    ├──► edr container ──writes──► shared volume (edr-mirror)
ThreatWinds Feeds API ──sync+normalize┘         │                        │ read-only
   (platform api-key/secret,                    │                        ▼
    from backend /api/v1/config)          status.json ─────► agentmanager :9001 serves
                                                             /private/edr/**  (existing TLS)
```

- **Sync daemon (Go, consistent with the repo):** schedules `cvdupdate` runs (the official ClamAV mirror tool, Python, installed in the image) into `/mirror/clamav/`; pulls + normalizes ThreatWinds lists into `/mirror/feeds/v1/download/list/…` with `.sha256` siblings; all writes atomic (tmp + rename); writes a `status.json` freshness document for monitoring.
- **Serving stays in `agentmanager`** — one added static route (`/private/edr` → the shared volume, read-only). No new exposed port, no new TLS listener, no new firewall rule at customer sites; agents keep talking to the single endpoint they already trust.
- **Credentials:** ThreatWinds `api-key`/`api-secret` read from the backend's config API with the `INTERNAL_KEY` — the exact flow `plugins/feeds` already uses; no new secret distribution.
- **Failure posture:** if upstream is unreachable the mirror goes **stale, never empty** (agents keep enforcing last-known-good and flag staleness); air-gapped installs can be seeded manually into the volume.

**Checksum-URL note:** the agent's current checksum request uses query parameters, which a static file tree cannot serve. Under this architecture the contract changes to a path-based sibling (`…/accumulative/{name}.sha256`) — a two-line client change, re-validated on the VM. (The alternative — the EDR container running its own HTTPS listener on a new port, e.g. 9002 — keeps the query URL and leaves `agentmanager` untouched, at the cost of a new exposed port + TLS listener at every deployment.)

### Alternatives considered

- **B — EDR container serves its own HTTPS port (9002):** fully self-contained, zero `agentmanager` change, but every customer deployment needs another open port and a second TLS listener; agents need new port config. Rejected unless single-container self-containment is valued over deployment friction.
- **C — cron jobs inside `agentmanager`:** no new container, but bloats the agent-manager image with Python/cvdupdate and couples unrelated concerns. Rejected.
- **D — ship CVDs/feeds via `dependency.Reconcile`:** rejected in the signature-update spec §4 (loses cdiff incrementals, reinvents version tracking, same server cost anyway).
