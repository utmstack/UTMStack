# ThreatIntel wiring report

End-to-end wiring of the UTMStack `threat-intel` feature from the frontend page
through a new UTMStack backend passthrough module, through the CustomersManager
(CM) proxy, and out to `apis.threatwinds.com`.

---

## 1. Context / research

Three research passes preceded any code.

### 1.1 CustomersManager proxy (existing)

Reverse proxy in front of `apis.threatwinds.com`. Sits in `CustomersManager/proxy/`.

| Concern | Value |
|---|---|
| Runtime port | `APP_PORT` (default 8080) |
| Router | Gin |
| Routes | `GET /proxy/usage`, `ANY /proxy/api/*path` |
| Auth headers in | `id`, `key` (instance credentials) |
| Auth cache | Redis `instance:{id}` → `{Key, Salt, Edition}`; SHA256(`key`+salt) match |
| Credential cache | Redis `intelligence:creds:{id}` → `{api_key, api_secret}` |
| Rate limits | Redis `quota:{id}:{prefix}` counter, 24h window, tiered by edition |
| Upstream | `${THREATWINDS_BASE_URL}` |
| Allowed prefixes | `/api/ingest/`, `/api/ai/`, `/api/feeds/`, `/api/analytics/`, `/api/search/` |
| Error surface | `401` bad auth · `429` over quota · `503` creds not ready · `502` upstream fail |

Provisioning (one-time + refresh) is the CM backend's job:
`POST /api/v1/intelligence/register` → CM backend uses `INTELLIGENCE_ADMIN_API_KEY/SECRET`
against ThreatWinds `POST /api/auth/v2/partners/user` → gets per-instance
`api-key/api-secret` → AES-encrypted in `intelligence_instance_auths`
table, cached in Redis. A goroutine `ReconcileIntelligenceCredentials()`
refreshes before expiry.

### 1.2 UTMStack frontend `threat-intel` feature (before)

- Single file: `src/features/threat-intel/pages/ThreatIntelPage.tsx` (1156 lines)
- 15 UI components defined inline
- All data mocked: hardcoded `IOCS`, `ACTORS`, `FEEDS` arrays
- No `services/`, no `hooks/`, no `domain/` folder
- Route mounted at `/threat-intelligence` in `src/app/routes/index.tsx`
- Referenced in `Sidebar.tsx` and `SocAiProvider.tsx` only

### 1.3 ThreatWinds public API (docs.threatwinds.com)

Auth: `api-key` + `api-secret` headers (confirmed).

| Section | Purpose | Frontend-useful endpoints |
|---|---|---|
| `/api/search/` | Primary IOC lookup | `POST /v1/entities/simple`, `GET /v1/entity/{id}`, `GET /v1/entity/{id}/relations` |
| `/api/analytics/` | Entity detail / relations graph | `GET /v1/entity/{id}/details`, `GET /v1/entity/{id}/relations` |
| `/api/feeds/` | IOC feeds catalog | `GET /v1/list`, `GET /v1/download` (20/h) |
| `/api/ai/` | OpenAI-compatible chat | `POST /v1/chat/completions` |
| `/api/ingest/` | Write-only community submit | not used in UI |

### 1.4 Existing CM integration in UTMStack

- `plugins/feeds/internal/client/cm_client.go` **already reads** `/updates/instance-config.yml` and expects `{server, instance_id, instance_key}`. Old flow: register → get raw TW creds → call TW directly.
- `installer/updater/*` also passes `id`+`key` headers to the CM backend on license / heartbeat / register calls.
- The UTMStack backend `modules/billing/usecase/version.go` reads the same YAML file but only for the `server` + `instance_id` fields (public info).
- **No existing consumer of the CM proxy** in UTMStack backend or frontend.

---

## 2. Architecture

The frontend must never see instance secrets. Chose a **UTMStack backend
passthrough** that reads the existing `instance-config.yml`, injects
`id`+`key` headers, and reverse-proxies to the CM proxy.

```
┌──────────────┐                    ┌─────────────────┐                    ┌───────────────┐                    ┌──────────────────┐
│              │  JWT (user auth)   │ UTMStack        │ id + key           │ CM proxy      │ api-key+secret     │                  │
│   browser    │───────────────────▶│ backend         │───────────────────▶│               │───────────────────▶│ apis.threatwinds │
│  (React SPA) │                    │ /api/v1/threat- │                    │ /proxy/api/*  │                    │      .com        │
│              │                    │  intel/*        │                    │               │                    │                  │
└──────────────┘                    └────────┬────────┘                    └───────┬───────┘                    └──────────────────┘
                                             │                                     │
                                    reads on start                          Redis lookups
                                    instance-config.yml                     instance:{id}, intelligence:creds:{id}
                                    (server, id, key)                       quota:{id}:{prefix}
```

Provisioning path (already done by feeds plugin, not touched here):

```
UTMStack backend ─── POST /api/v1/intelligence/register ───▶ CM backend
                     (id + key)                                     │
                                                                    ▼
                                                    creates TW user, encrypts
                                                    creds in Postgres, caches
                                                    in Redis for the proxy
```

---

## 3. Endpoint / path mapping

| Frontend action | Frontend URL | UTMStack backend route | CM proxy path | TW upstream |
|---|---|---|---|---|
| IOC search | `POST /api/v1/threat-intel/search` | `POST /api/v1/threat-intel/search` | `POST /proxy/api/search/v1/entities/simple` | `POST /api/search/v1/entities/simple` |
| Entity detail | `GET /api/v1/threat-intel/entity/:id` | `GET /api/v1/threat-intel/entity/:id` | `GET /proxy/api/analytics/v1/entity/{id}/details` | `GET /api/analytics/v1/entity/{id}/details` |
| Entity relations | `GET /api/v1/threat-intel/entity/:id/relations` | same | `GET /proxy/api/analytics/v1/entity/{id}/relations` | `GET /api/analytics/v1/entity/{id}/relations` |
| Feeds catalog | `GET /api/v1/threat-intel/feeds` | same | `GET /proxy/api/feeds/v1/list` | `GET /api/feeds/v1/list` |
| AI brief | `POST /api/v1/threat-intel/ai/chat` | same | `POST /proxy/api/ai/v1/chat/completions` | `POST /api/ai/v1/chat/completions` |
| Usage / quota | `GET /api/v1/threat-intel/usage` | same | `GET /proxy/usage` *(no `/api` prefix — proxy meta)* | — |

Query strings and request/response bodies pass through verbatim. Upstream
status codes are relayed.

---

## 4. Backend module

New Go module `backend/modules/threatintel/` following the existing
`modules/billing` / `modules/socai` conventions.

### 4.1 Files

| File | Purpose |
|---|---|
| `module.go` | `Module` struct + `NewModule(cmProxyURL, updatesDir)` factory |
| `routes.go` | `RegisterRoutes(api, m, userAuth)` — mounts 6 routes with per-route path prefix |
| `handler/proxy.go` | `ReverseProxyHandler` — `httputil.ReverseProxy` with Director (URL rewrite + header inject) + ErrorHandler (JSON 502) |
| `handler/proxy_test.go` | Unit test: `TestDirectorRewritesPathAndHeaders` — path rewrite + header injection |
| `internal/instanceconfig.go` | `sync.Once`-cached reader of `${UPDATES_DIR}/instance-config.yml` |

### 4.2 Wiring

- `modules.go` — construct `threatIntelMod := threatintel.NewModule(env.String("CM_PROXY_URL", "", false), env.String("UPDATES_DIR", "/updates", false))`, add to `modules` struct
- `server.go` — `threatintel.RegisterRoutes(api, m.threatIntel, userAuth)` alongside the other module route registrations

### 4.3 Design choices

- **Pure passthrough** — no DTOs, no request/response mapping. `httputil.ReverseProxy` from stdlib handles the whole thing.
- **No new dependencies** — `net/http`, `net/http/httputil`, existing `gin`, existing `yaml.v3` reused.
- **Config cached on first call** with `sync.Once`. Instance config is immutable at runtime.
- **Auth** — existing UTMStack `userAuth` middleware (JWT). Same protection as the rest of `/api/v1`.
- **Failure modes** —
  - Missing/empty `CM_PROXY_URL` → 503 `{"error":"cm proxy not configured"}`
  - Missing/empty instance-config.yml fields → 503 `{"error":"instance not configured"}`
  - Upstream unreachable → 502 via ErrorHandler
  - Upstream `401/429/503` → passed through verbatim

### 4.4 Build / test

```
go build ./...          # exit 0, clean
go test ./modules/threatintel/handler/... -v   # PASS: TestDirectorRewritesPathAndHeaders
```

---

## 5. Frontend

Structure per user's global CLAUDE.md (`features/x/{domain,services,hooks}`).
**Minimal split** chosen — 15 inline components in `ThreatIntelPage.tsx`
stayed put; only new folders were added and data sources were swapped.

### 5.1 Files created

```
frontend/src/features/threat-intel/
  domain/
    threatwinds.model.ts    # TWEntity, TWEntityDetail, TWSearchRequest/Response,
                            # TWChatMessage, TWChatResponse, TWFeed
    feed.model.ts           # ThreatFeed (UI), TWFeedResponse (wire)
    usage.model.ts          # ThreatIntelUsage
  services/
    threat-intel.api.ts     # createApiClient() wrapper — search, entity,
                            # entityRelations, feeds, aiChat, usage
    mappers.ts              # mapTWEntityToIOC — reputation → severity, etc.
  hooks/
    useIocSearch.ts         # useQuery
    useIocDetail.ts         # useQuery, enabled when drawer id present
    useFeeds.ts             # useQuery
    useThreatIntelUsage.ts  # useQuery
    useAiBrief.ts           # useMutation
```

### 5.2 File modified

`pages/ThreatIntelPage.tsx` — removed `IOCS` and `FEEDS` hardcoded arrays,
swapped for hook calls; `IocDrawer` calls `useAiBrief` on open; tab counts
now derived from hook data; loading/error states inline. Inline components
kept untouched.

### 5.3 Type-safety

- No `any`. Loose parts of TW responses (`extended_metadata`,
  relations graph) typed `unknown` and narrowed at usage sites.
- `tsc --noEmit` exit 0.

### 5.4 Mapping decisions (TW → UI type)

| UI field | TW source | Rule |
|---|---|---|
| `severity` | `reputation` | ≤−80 critical · ≤−50 high · ≤−20 medium · else low |
| `feeds` | `source.includes` | fallback `['ThreatWinds']` |
| `tags` | `tags` | direct |
| `matchesInEnv` | — | 0 (no local correlation yet) |
| `firstSeenInEnv`, `lastSeenInEnv` | — | `undefined` |
| `attributedActor` | `attributes` (if actor entity linked) | else `undefined` |

Placed in `services/mappers.ts` with `ponytail:` comments naming what the
placeholders wait on.

---

## 6. Configuration

### 6.1 Backend

| Var | Required | Purpose |
|---|---|---|
| `CM_PROXY_URL` | yes | Base URL of the CM proxy (e.g. `https://cm-proxy.utmstack.com`) |
| `UPDATES_DIR` | no (default `/updates`) | Where `instance-config.yml` lives |

### 6.2 `instance-config.yml` (already required by feeds plugin & installer)

```yaml
server:        https://cm.utmstack.com     # CM backend URL (unused by this module)
instance_id:   <uuid>                       # sent as `id` header to CM proxy
instance_key:  <secret>                     # sent as `key` header to CM proxy
```

### 6.3 Frontend

Nothing new. Uses existing `VITE_API_URL` and shared axios client.

---

## 7. Prerequisites (already satisfied elsewhere)

For the runtime path to succeed on a given instance:

1. The instance must have registered as an intelligence reporter
   (`POST /api/v1/intelligence/register` on the CM backend) — the feeds
   plugin already does this at startup.
2. CM Redis must have `instance:{id}` (auth blob) and
   `intelligence:creds:{id}` (TW creds). Auto-provisioned by the CM backend
   register flow + `ReconcileIntelligenceCredentials()` goroutine.
3. `CM_PROXY_URL` must resolve from the UTMStack backend pod/container
   and be reachable.

If (1) or (2) is missing, the CM proxy returns `503 credentials_not_ready`
and the UI shows an error state.

---

## 8. Deferred (`ponytail:` debt)

| Marker | Where | What waits |
|---|---|---|
| Actors tab uses hardcoded fallback | `ThreatIntelPage.tsx` | TW exposes an actors endpoint, or we derive client-side from IOC groupings by `attributedActor` |
| `matchesInEnv` fixed to 0 | `services/mappers.ts` | Local IOC correlation against event indices is wired |
| Feed shape has placeholder `itemsTotal`, `status` | `services/mappers.ts` | Backend enriches feed metadata (TW only returns `name/type/accuracy`) |
| AI model hardcoded to `gpt-4` | `useAiBrief.ts` | Model selection surfaces in settings, or TW `GET /api/ai/v1/models` is exposed to pick |

---

## 9. Verification checklist

- [x] `go build ./...` — exit 0
- [x] `go test ./modules/threatintel/...` — PASS
- [x] `tsc --noEmit` — exit 0
- [x] Backend routes wired in `server.go`
- [x] Backend module constructed in `modules.go`
- [x] Frontend folder structure matches CLAUDE.md convention
- [x] No `any` in new frontend code
- [ ] Runtime smoke test against a live CM proxy — not run in this session

Suggested next step: point a dev UTMStack backend at a dev CM proxy
(`CM_PROXY_URL=<dev-cm-proxy>`), open the ThreatIntel page, search for a
known IOC (e.g. `8.8.8.8`), confirm results and the AI brief on drawer open.
