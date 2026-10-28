# UTMStack v11 — Security Review: Dependabot Triage (332 alerts) + Application Vulnerability Sweep

- **Date:** 2026-10-27
- **Scope:** Full Dependabot backlog — **332 open alerts** on `utmstack/UTMStack` (36 critical / 133 high / 135 medium / 28 low; 212 runtime-scoped, 120 development-scoped) — plus a manual sweep of the OWASP Top-10 / common web-app classes across the Java backend, Angular frontend, and Go components.
- **Mode:** **Read-only.** No code was modified in this phase. (The separate `fix/ir-role-model-v11.3.1` incident-response work is untouched.)
- **Standing rule applied (user):** any path a user can send/receive data through the API must be validated + sanitized as the first layer; DB, filesystem, and process boundaries need their own independent safeguards.
- **Trust boundaries:** backend serves the REST API; Angular frontend is static (nginx); Go event-processor plugins + agent/collector run gRPC; Postgres + OpenSearch are the data stores.

> **Correction note (v1 of this report said "99 alerts"):** the first pass read only page 1 of the Dependabot REST API (single `per_page=100` call; the endpoint paginates via the `Link` header, and `page=` is rejected with HTTP 400). Full pagination gives **332**. The corrected inventory and triage below supersede that draft. The draft's Part C (application-level findings F-1…F-6) is unchanged — it was not affected by the count.

---

## Part B½ — Live verification on lab 192.168.99.75 (v11.3.0, 2026-10-28)

Each finding below carries a step-by-step reproducible path with the evidence actually observed on the lab. Lab notes that shape the tests: OpenSearch speaks **HTTPS** on `node1:9200` (plain-HTTP probes get `NotSslRecordException` — use `-k` against `https://`); the panel proxies `/api` through nginx with an injected `x-shared-key`, so host-side API probes need that header (read from `/root/utmstack.yml` on the lab); a throwaway analyst account `labsec_poc` (ROLE_USER, TOTP self-enrolled with a known secret) was created for authenticated probes and is **pending cleanup**.

### F-6 — NOT A FINDING (reclassified 2026-10-28, product decision)
1. From any host that can reach the panel: `curl -sk https://192.168.99.75/v3/api-docs`
2. **Observed:** `200`, `application/json`, **374,641 bytes** — full OpenAPI 3.0.1 schema of "UTMStack API v11.3.0", **252 paths**.
3. Parsed the JSON: it enumerates the public REST routes.
4. `curl -sk https://192.168.99.75/swagger-ui/index.html` → `200`, Swagger UI shell.
- **Root cause:** the panel nginx config proxies `location /v3` and `location /swagger-ui` to the backend without the `x-shared-key` gate; Spring Security does not require auth on these paths.
- **Why it's not a finding:** UTMStack is open source — the spec is generated from source that is already public on GitHub, so it discloses nothing beyond the repo. The version string it carries is already exposed by `/api/info/version` (permitAll) and the login page. No confidential information is leaked.
- **Optional hardening (only if it's ever wanted):** gate or drop the `/v3` + `/swagger-ui` nginx locations. Not scheduled.

### F-3 — NOT A FINDING (reclassified 2026-10-28, product decision: no cross-origin browser consumers exist)
1. `curl -sk -D - -o /dev/null -H "Origin: https://evil-attacker.com" http://192.168.99.75:10001/api/ping`
2. **Observed:** `Access-Control-Allow-Origin: *` + `Access-Control-Expose-Headers: Authorization, ...` — the wildcard is served.
- **Why it is inert (verified end-to-end, 2026-10-28):**
  - REST: `environment.prod.ts:3` `SERVER_API_URL: ''` → all `HttpClient` calls resolve same-origin; the panel nginx proxies `/api` on the same hostname (`installer/templates/front-end.go:48-49`). No browser ever makes a cross-origin request.
  - Websockets: `WEBSOCKET_URL: ''` → SockJS/STOMP connect to the same host (`tracker.service.ts:39`); `utm-agent-console` sets `withCredentials=false`.
  - The single `withCredentials:true` (`utm-webflux.service.ts:36`, AD-tracker `EventSource`) is also same-origin, and its token rides a `?access_token=` query param (which is why `JWTFilter` reads `access_token`), not a header — so the `Expose-Headers: Authorization` line has no effect on it.
  - The federation server consumes the API **server-to-server** (Go → HTTP); CORS is a *browser* mechanism and does not apply to non-browser clients. Same for agents/plugins.
- **Verdict:** `allowed-origins: '*'` is inert JHipster boilerplate — there is no deployment with a cross-origin browser consumer. Reclaimed the earlier "hardening" grading: no functional change needed. Optional cosmetic cleanup only.

### F-4 — CONFIRMED endpoint is live for ROLE_USER (reclassified: not a finding — it's the Log Explorer SQL feature)
1. Authenticate as a plain `ROLE_USER` → JWT `{"auth":"ROLE_USER"}`.
2. `POST /api/elasticsearch/search/sql` (panel, `x-shared-key` + Bearer JWT) body `{"query":"SELECT 1"}`.
3. **Observed:** backend log `OpenSearch.searchBySqlQuery: HTTP 400` — **Spring Security accepted the request and forwarded the SQL to the OpenSearch engine** (the 400 is the engine rejecting the SQL, not an auth block). Unauthenticated → 401.
4. `GET /api/elasticsearch/index/all` as the same ROLE_USER → `200` with full index inventory.
- **Reclassification 2026-10-28:** this endpoint backs the Log Explorer SQL mode (analyst types raw SQL in `log-analyzer-view`), so an analyst running read SQL is a designed capability. Analysts read all data by design and OS SQL is read-only → **not a vulnerability.** The C4/C5 filter-builder concat stays as optional hygiene only.

### F-1 — Stored XSS PoC: payload stored, but rendering is INERT (PoC failed to exploit — corrected 2026-10-28)
1. **Stored payload (done):** injected a synthetic alert doc into `v11-alert-2026-10-28` (via HTTPS to OS with the backend service creds): `{"_index":"v11-alert-2026-10-28","_id":"kRk-gqEBG2EDI9JlXLDl","result":"created"}`.
2. **Round-trip confirmed (done):** `_search` match on `name:"LABSEC-XSS-POC-9917"` → stored `description` = `LABSEC-XSS-POC <img src=x onerror="window.__labsecXSS=1"> done` — the markup survives storage byte-for-byte.
3. **Rendering (corrected):** the live bundle `main.d20a757d340e10b2c1d1.js` contains `bypassSecurityTrustHtml` 4× (the `safe` pipe + vs-task URL trust), **but** the alert-detail bindings I flagged (`alert-description` etc.) use **plain `[innerHTML]`**, which Angular 7's default sanitizer neutralizes (`onerror` stripped, `<img>` stays inert). The only true bypasses are the `| safe:'html'` sites (inventory in Part C½ F-1), all of which are currently fed admin-authored or pre-escaped data.
- **Verdict: NO live stored XSS.** The earlier "CONFIRMED" was wrong — it conflated "markup survives storage" with "markup executes at render," which is false under Angular's default `[innerHTML]` sanitization. The stored-payload PoC is retired (doc deleted).

### F-5 — Static only (unchanged)
`plugins/soc-ai/utils/request.go:41` `InsecureSkipVerify: true` on outbound calls — verified in source; runtime behavior depends on whether soc-ai has external endpoints configured in a given deployment. Severity stands as MEDIUM/context-dependent. No lab instance of soc-ai config to test against safely.

### F-5 — NOT A FINDING (reclassified 2026-10-28 after re-verification)
- The `InsecureSkipVerify` client in `plugins/soc-ai/utils/request.go:41` is selected **only** when `opts.SkipTLSVerify` is set, and the only caller doing that is `DoOpenSearchReq` (request.go:137-148) — i.e., internal calls to the self-signed OpenSearch endpoint. All other soc-ai traffic (including external AI-model calls via `DoReq`) uses the standard-verification `defaultClient`. Same house pattern as `inputs/`, `agent-manager/`, `compliance-orchestrator/`. **No remediation needed.**

### Lab cleanup — DONE (2026-10-28)
- Throwaway account removed: `DELETE 1` on `jhi_user_authority` (user_id=4) + `DELETE 1` on `jhi_user` (`labsec_poc`); remaining logins verified = `fsclient, admin, fstestuser` (original set).
- Poison alert doc deleted: `_doc/kRk-gqEBG2EDI9JlXLDl` → `{"result":"deleted"}`, post-delete `_search` on `name:"LABSEC-XSS-POC-9917"` → `hits.total.value: 0`.
- No other lab state changed.

---

## Part A — Executive summary

1. **The backlog is 332 alerts but only ~90 distinct packages, and the split is stark:** 236 are the frozen Angular 7 / Node 14 frontend lockfile, ~65 are Go modules (mostly the same 6 packages repeated per `go.mod`), 2 are Java (`snakeyaml` backend, `postgresql` user-auditor).
2. **~60% of the alerts are development-scoped** (webpack/karma/node-sass/postcss toolchain). Those never execute in the shipped bundle or on the server — their control is a hardened isolated build, not patching.
3. **The genuinely runtime-reachable, patchable cluster is Go, and it is one `go get` pass:** `x/crypto` (41 alerts, 21 critical — pinned low in **three** manifests: `shared` 0.42.0, `installer` 0.51.0, `agent/updater` 0.47.0; the rest of the repo is already ≥0.52), `cel-go` 0.28.1→0.29.0 (~20 modules, the rule-evaluation engine, hot path), `grpc` 1.81.1/1.83.x→1.84.0 (4 plugin modules; xDS CVEs not applicable — no xDS usage), `x/net` (shared 0.44.0, agent/updater 0.49.0 → 0.55.0), `otel` (installer 1.41.0 → 1.45.0), `quic-go` 0.59.0→0.59.1 (3 modules).
4. **The frontend runtime criticals are the jspdf cluster:** `jspdf` 2.3.1 (2 critical + 8 high, fixes up to 4.2.1) is imported in `export-to-pdf-util.ts`, and it **bundles `dompurify` 2.2.7** (critical CVE-2024-48910), `canvg`, `fflate`, `html2canvas` into the shipped app. These run in the analyst's browser, on data the app feeds them (export flows). Remediation: bump jspdf to a patched major **with an API-compat test** (jspdf 2→4 is a breaking-major jump; the code does `import jsPDF from 'jspdf'` + basic export calls — needs verification), or pin dompurify up via `overrides` if jspdf 2.x's range allows (jspdf's dep is `dompurify ^2.2.0`, so `overrides: {"dompurify": "3.4.16"}` may not satisfy jspdf's own 2.x-era calls — test required).
5. **Angular 7 framework alerts (22: core/common/compiler/router) have no patch on 7.x** — fixes start at v10.2.5 / v19.2.16. Disposition per alert: several are SSR/HttpTransferCache/i18n features UTMStack does not use (N/A dismissal with justification), the rest accepted risk pending the framework upgrade project.
6. **Two Go packages have no fixed version yet:** `github.com/docker/docker` (8 alerts, `+incompatible` v28.5.2 in installer + collector — used to drive the swarm deployment) and a few npm packages with `fixed=` blank. These need a documented monitor-or-mitigate decision, not a dismiss.
7. **The manual sweep found NO live application vulnerability once Angular's default sanitizer was properly accounted for.** Five candidates (F-1 HTML sinks, F-2 GVM render, F-3 CORS `*`, F-4 raw-OS-SQL, F-5 soc-ai TLS) were each re-verified and **reclassified as not findings** — the remaining F-1 residual (a `safe.pipe` that bypasses Angular's sanitizer, currently only fed admin-authored or pre-escaped data) is downgraded to **hygiene**. The original F-1 "stored XSS" claim was wrong: Angular 7 sanitizes plain `[innerHTML]`, and my live PoC payload rendered inertly. See Part B½ (live) + Part C (static re-verification) for the full trail.

---

## Part B — Dependabot triage (332 alerts, full inventory)

### B1. Go cluster — patchable in one pass (≈65 alerts)

| Package | Affected manifests (current → required) | Alerts | Reachable? | Action |
|---|---|---|---|---|
| `golang.org/x/crypto` | `shared` 0.42.0, `agent/updater` 0.47.0, `installer` 0.51.0 → **≥0.55.0** (rest of repo already 0.52–0.57) | 41 (21 critical: -39830/31/32/33/34, -42508, -46595; 6 high; 14 medium) | **No** in current code — all affected modules carry it as `// indirect`; the vulnerable surface is `ssh` agent-keyring constraint handling, and no Go module here uses `x/crypto/ssh` for agent constraints. But it ships in every binary. | **Bump** the three lagging manifests (`go get golang.org/x/crypto@v0.55.0` + tidy in each). Closes the largest single alert group and all 21 criticals. |
| `github.com/google/cel-go` | ~20 modules at 0.28.1 (`compliance-orchestrator` at 0.26.1) → 0.29.0 | 20 (all medium, "JSON private fields exposed via NativeTypes/Parser") | **Yes** — cel-go is the rule-evaluation engine, on the hot path. | **Bump all modules to 0.29.0.** |
| `google.golang.org/grpc` | `plugins/{stats,soc-ai}` 1.81.1, `plugins/{geolocation,alerts}` 1.83.1/1.83.2 → **1.84.0** | 16 (12 high: CVE-2026-84304 HTTP/2 OOM, CVE-2026-84445 + 84303 xDS; 4 medium) | **Partly** — HTTP/2 DATA-frame OOM reachable on any exposed gRPC listener (50051/8000/9000…); xDS CVEs not reachable (no xDS service-mesh features). | **Bump to 1.84.0** in the four modules. |
| `golang.org/x/net` | `shared` 0.44.0, `agent/updater` 0.49.0 → 0.55.0 | 2 (medium) | Partly (http2 stack is used). | Bump with the x/crypto pass. |
| `go.opentelemetry.io/otel/sdk` (+ otlptracehttp) | `installer` 1.41.0 → 1.45.0 (collector already 1.45.0) | 4 (1 high, 1 medium, 2 low — endpoint/credential logging leak) | Yes (telemetry). | Bump in `installer/go.mod`. |
| `github.com/quic-go/quic-go` | `installer`, `agent/updater`, `compliance-orchestrator` 0.59.0 → 0.59.1 | 3 (medium) | Partly. | Bump in the three modules. |
| `github.com/docker/docker` | `installer`, `utmstack-collector` at v28.5.2+incompatible | 8 (4 high, 4 medium) | **Yes** — the installer drives the Docker swarm; the collector talks to the daemon. | **No fixed version published** at alert time. Action: track the advisory, apply the vendor patch when released; meanwhile confirm daemon access is host-local only (it is — the installer runs on the host managing its own swarm) and document that network exposure of 2375/2376 is the mitigation boundary. |

### B2. Java (2 alerts)

| Package | Where | Alert | Reachable? | Action |
|---|---|---|---|---|
| `org.yaml:snakeyaml` 1.33 | `backend/pom.xml` | CVE-2022-1471 (high, deserialization RCE) | **Conditional** — exploitable only if SnakeYAML `Constructor` parses **untrusted** YAML; UTMStack uses it for boot config parsing (trusted). | **No clean patch** — SnakeYAML 2.0 crashes Spring Boot 3.1.5 (`Representer()` no-arg ctor removed; this exact trap was hit and reverted, #2691). Keep 1.33; hardening process = (a) document "config-only, never parse untrusted YAML," (b) add a code-review gate forbidding `new Yaml(new Constructor()).load(...)` on request data, (c) re-test 2.0 when Spring Boot moves to 3.2+. |
| `org.postgresql:postgresql` 42.7.11 | `user-auditor/pom.xml` | CVE-2026-54291 (high, channel-binding downgrade) | **No** — requires `channelBinding=require`; JDBC URL is bare `jdbc:postgresql://…`. | **Bump to 42.7.12.** Cheap, closes the alert. |

### B3. Frontend npm (236 alerts) — by class

**Class 1 — development-scoped toolchain (120 alerts): webpack 4 / karma / node-sass / postcss / browserslist / brace-expansion / tar / adm-zip / node-forge / socket.io(-parser) / engine.io / webpack-dev-server / proxy-addr / path-to-regexp / minimatch / loader-utils / @babel/traverse / sha.js / cipher-base / pbkdf2 / elliptic / browserify-sign / bn.js / ws / ansi-regex / postcss / follow-redirects / http-proxy-middleware / serialize-javascript / http-cache-semantics / scss-tokenizer / trim-newlines / ip / cookie / on-headers / serve-static / send / express / body-parser / flatted / tmp / ini / semver / yargs-parser / mem / micromatch / npm-registry-fetch / braces / ansi-html / sockjs / diff / xml2js / webpack-subresource-integrity / js-yaml(3.x dev copy).**
- None of these executes in the shipped `dist/utm-stack` or on the server.
- **Hardening process (the actual remediation for this class):**
  1. Isolated, egress-restricted, credential-free build environment (build cannot reach internal networks or hold cloud/GitHub secrets).
  2. Dist artifact integrity: commit/sign the built bundle hash so a poisoned dep tree can't silently change shipped JS.
  3. **Post-build bundle scan** (e.g. `webpack-bundle-analyzer` output or a grep of `dist/` for the listed package names) to *prove* dev-only packages aren't shipped.
  4. Dismiss on GitHub with that justification per package (one template, batched).
- Note: the dev-only crypto/`babel-traverse`/`loader-utils` criticals (8 of the 36 total criticals) are all in this class.

**Class 2 — runtime-scoped, bundled into the app (≈100+ alerts):**
| Package (lock ver) | Alerts | In-app usage | Action |
|---|---|---|---|
| `jspdf` 2.3.1 | 12 (2 critical) | `export-to-pdf-util.ts` imports it (client-side PDF export) | **Bump to patched major (4.2.1) with API-compat tests**, or evaluate 3.0.x. This single bump also pulls patched `dompurify`/`canvg`/`fflate`/`html2canvas` inside its tree if the ranges allow — verify via lock diff. |
| `dompurify` 2.2.7 (via jspdf) | 20 (1 critical, CVE-2024-48910) | not imported directly; executes in-browser inside jspdf's SVG pipeline | Resolved with jspdf bump; if jspdf can't move, add `overrides` and test the export path. |
| `tar` 2.2.2 | 15 (1 critical, 10 high) | node-sass install pipeline pulls it; also flagged runtime by Dependabot | Verify it doesn't ship (bundle scan); if shipped by any dev-time extraction of runtime-relevant assets, pin 7.5.21. |
| `websocket-driver` 0.7.4 | 2 (1 critical, CVE-2026-54466) | flagged **runtime** — check the bundle; likely via sockjs-client/webstomp-client path | If shipped: the 0.7.5 patch is a drop-in (same major) — bump. If dev-only: Class 1 treatment. |
| `@angular/core` 7.2.0 | 8 (4 high) | the framework | **No patch on 7.x** (fixes start v10.2.5). Per-CVE triage: SSR/transfer-cache findings = N/A (no SSR, no `HttpTransferCache`); XSS-class findings = accepted risk, tracked under the framework-upgrade project. |
| `@angular/common` 7.2.0 | 7 (6 high) | framework | Same as above (fixes at v19.2.16 — unreachable). |
| `@angular/compiler` 7.2.0 | 6 (3 high) | framework | Same. |
| `@angular/router` 7.2.0 | 1 (high, SSR DoS — no fix listed) | framework | N/A (no SSR) — dismiss with justification. |
| `js-yaml` 4.1.0 (runtime dep) | 10 high (3.x + 4.x copies) | **imported at runtime** by `rule-view.component.ts` and `utm-file-upload.component.ts` (parses rule YAML + uploaded files) | Direct dep `^4.1.0` — **bump to 4.3.2** (same major, drop-in). This is one of the few clean runtime wins on the frontend. The quadratic-CPU merge-key/omap findings are reachable via user-uploaded YAML — real DoS surface on an authenticated endpoint. |
| `lodash` 4.17.21 | 3 (1 high) | runtime dep, widely used | 4.17.23+ available (same major) — **bump**. |
| `moment` 2.31.0 | 1 (medium, path traversal via crafted locale) | runtime dep | 2.31.0 is the latest 2.x; no further patch — accepted risk / monitor. |
| `highlight.js` 9.12.0 | 2 (medium) | runtime dep (syntax highlighting of rule definitions) | 9.18.2 same major — **bump** (and it feeds the same `[innerHTML]` surfaces as F-1 — coordinate with the XSS fix). |
| `echarts` 4.9.0 | 1 (medium) | runtime (charts) | 5.x is breaking; 4.9.0 is the last 4.x — accepted risk / monitor. |
| `html2canvas` 1.0.0-rc.7 | (part of jspdf tree) | runtime via jspdf | Moves with the jspdf bump. |
| `form-data` (runtime) | 2 (1 critical) | check bundle — likely via `request` (deprecated) | Identify the parent (`request` is runtime, 1 alert, no fix — `request` is deprecated; find who pulls it and replace or verify it's inert). |
| `cross-spawn` 6.x | 1 (high) | flagged runtime — verify bundle; likely dev-only via karma | bundle scan decides. |
| `uuid`, `tough-cookie`, `es5-ext`, `ajv`, `decode-uri-component`, `qs`, `fflate`, `moment-timezone`, `canvg`, `zrender`, `@babel/runtime-corejs3`, `picomatch`, `socket.io-client`(absent at top), `node-sass` 4.0.0 (1 medium — node-sass is EOL, the root cause of the whole frozen tree) | remainder | mixed | Disposition per bundle-scan result: shipped → bump within major if possible; not shipped → Class 1. `node-sass` itself: no fix on 4.x (EOL); it is the reason the whole toolchain is frozen — the strategic fix is the framework/Node upgrade project, already tracked. |

**Class 3 — strategic:** the *reason* 236 alerts live here is the Angular 7 / Node 14 / node-sass-4 freeze. Every future CVE in this tree will land the same way. The dependency-upgrade project (framework + Node + sass) is the durable remediation; until then, isolated build + dist integrity + bundle scan + per-package justified dismissals is the operating model.

### B4. Disposition totals (proposed)

| Bucket | Count (approx) | Disposition |
|---|---|---|
| Go bumps (x/crypto, cel-go, grpc, x/net, otel, quic-go) | ~86 | **Fix now** — `go get` + tidy per module |
| snakeyaml | 1 | Keep + document (no clean patch) |
| postgresql (user-auditor) | 1 | **Fix now** — 42.7.12 |
| Frontend runtime clean bumps (js-yaml, lodash, highlight.js, websocket-driver if shipped) | ~15 | **Fix now** — same-major |
| Frontend runtime major bumps (jspdf cluster: jspdf, dompurify, canvg, fflate, html2canvas, form-data?, request?) | ~40 | **Fix with testing** — jspdf 2→4 is breaking; staged |
| Angular 7 framework, no patch on 7.x | ~22 | **Dismiss with per-CVE justification** (N/A feature or accepted risk until framework upgrade) |
| Dev-scoped toolchain | ~120 | **Dismiss (build-time only)** + hardening process (isolated build, dist integrity, bundle scan) |
| No fixed version yet (docker SDK, `request`, others) | ~12 | **Monitor / documented mitigation** — re-check at next advisory cycle |

---

## Part C — New application vulnerabilities (manual sweep — OWASP Top-10 pass)

Not in the scanner output. Ranked by severity.

### F-1 — Frontend HTML sinks (CORRECTED 2026-10-28: NOT a live XSS — downgraded to latent-risk + hygiene)
- **Original claim was wrong and is corrected here.** I initially reported that log-derived alert fields rendered through `replaceBreakLine` + `[innerHTML]` were unescaped XSS sinks. That misread Angular 7's behavior: **the framework's default sanitizer sanitizes plain `[innerHTML]`** (strips `<script>`, `on*` handlers), so `alert-description` / `alert-category` / `alert-severity-description` / `alert-proposed-solution` / `file-detail` / `file-access-mask` are **safe by default**. My live PoC alert (`<img src=x onerror=…>` in a stored description) renders inertly in the UI — the payload does NOT execute.
- **The only bindings that bypass Angular's sanitizer** are the 17 sites using the `safe:'html'` pipe (`safe.pipe.ts:30` = `bypassSecurityTrustHtml`, zero sanitization) and two explicit `bypassSecurityTrustUrl` calls. Full inventory and live-status:
  - `compliance-*` components (7 sites): admin-authored compliance control text → admin self-XSS only (low).
  - `utm-agent-console.html:65`: the TS **already escapes** via `escapeAndBreakLine` before the binding (`utm-agent-console.component.ts:180`) → double-safe.
  - `utm-terminal-input.html:10`: TS escapes at `:26` → safe.
  - `logstash-filters.html:11`: binding is **commented out** → dead.
  - `vs-task-result-detail.html` (10 sites + 1 URL trust): GVM/openvas backend endpoints don't exist in this repo → dead (see F-2).
- **Verdict: no live unauthenticated/attacker-writable XSS found.** Residual risk = the `safe.pipe` exists as a one-liner that silently disables Angular's sanitizer for any future caller; the codebase is inconsistent (some sites pre-escape, some rely on the pipe).
- **Remediation (hygiene, not P0):** (a) delete or rewrite `safe.pipe` to escape-then-render (it is now misnamed "safe" doing the opposite), or at minimum document "input MUST be pre-escaped"; (b) CI grep-gate for new `| safe:'html'` usages on non-admin data.
- **F-1 live PoC artifacts (already cleaned):** injected alert `_id kRk-gqEBG2EDI9JlXLDl` deleted; `labsec_poc` account deleted.

### F-2 — GVM detail unsanitized render (RECLASSIFIED: dead code — backend endpoints don't exist in this repo)
- `vs-task-result-detail.component.ts:38-41` builds HTML with no escape (newline→`<br>`) bound to ~10 `[innerHTML]`; `:71` `sanitizer.bypassSecurityTrustUrl(doc)` on an `[href]`.
- **Provenance check 2026-10-28:** the component's data comes from `api/utm-gvm-tasks` and `api/openvas/tasks/quickScanTask` — **no backend controller for either exists in this repo** (searched Java + Go); only legacy table stubs (`utm_gvm_task`/`utm_gvm_scan_result`) remain in liquibase `tables.sql`. So in this codebase the feature is **unreachable** — the sink exists but has no live data source.
- **Verdict:** not a live vulnerability. If the GVM integration is ever revived (endpoint added), the same `escapeAndBreakLine` fix from F-1 should be applied to these bindings first. Keep on the list as a "fix-if-revived" note, not a scheduled item.

### F-3 — Production CORS `allowed-origins: '*'` (RECLASSIFIED: not a finding — no cross-origin browser consumers)
- `application-prod.yml:50-51`, registered on `/api/**`, `/management/**`, swagger paths (`WebConfigurer.java:47-52`).
- **Why it's inert (verified 2026-10-28):** every browser consumer is same-origin — `environment.prod.ts` `SERVER_API_URL:''`/`WEBSOCKET_URL:''` make all REST/STOMP calls relative to the panel hostname, and the panel nginx proxies `/api` + `/ws` on that same host (`installer/templates/front-end.go`). The single `withCredentials:true` (AD-tracker `EventSource`) is same-origin and passes its token as a `?access_token=` query param, so the `Expose-Headers: Authorization` line affects nothing. Federation-server, agent, and plugin API consumers are server-to-server (non-browser), where CORS does not apply.
- **Remediation:** none required. Optional cosmetic cleanup of the JHipster-default wildcard.

### F-4 — `POST /api/elasticsearch/search/sql` runs raw OpenSearch SQL (RECLASSIFIED: not a finding — it's the Log Explorer SQL feature)
- `ElasticsearchResource.java:252-265` executes caller-supplied SQL. **This is the backend for the Log Explorer "SQL mode"** — `log-analyzer-view.component.ts:289` sends the analyst-typed `this.sqlQuery` free-form box straight to this endpoint. So "analyst runs raw read SQL" is a designed capability, not a hidden primitive.
- **Why it's not a vulnerability (same reasoning as F-3/F-6):** analysts read all log data by design (no tenant/row boundary to escape); the OpenSearch SQL engine is **read-only** (no DML/DDL/code-exec). The endpoint doing its job is not a missing gate.
- **Only residual (optional hygiene, not a finding):** `SqlQueryFilterService.java:148+` builds SQL by string-interpolating filter values without escaping (C4/C5 disclosure items) — CWE-89 in principle, but low-impact since the same user already has the raw-SQL box one screen over.
- **Remediation:** none required for the endpoint. Escape filter values in `SqlQueryFilterService` if we want the hygiene.

### F-5 — `plugins/soc-ai` TLS (RECLASSIFIED: not a finding — skip-verify is internal OpenSearch only)
- Re-verified 2026-10-28: the `InsecureSkipVerify: true` client (`utils/request.go:41`) is **only** selected when `opts.SkipTLSVerify` is set, and the sole caller is `DoOpenSearchReq` (`request.go:137-148`) — i.e., the plugin's calls to the **internal self-signed OpenSearch** endpoint. All other soc-ai traffic (incl. the external AI-model calls via `DoReq`/`DoParseReq`) uses `defaultClient` with **standard TLS verification** (`request.go:81-84`).
- **Verdict: not a finding.** Self-signed internal-DB TLS-skip is the house pattern (same in `inputs/`, `agent-manager/`, `compliance-orchestrator/`) and is acceptable inside the deployment's trust boundary. The only actionable nit is the misleading comment on line 35 ("for OpenSearch" — it already says that; keep as-is).
- The other `InsecureSkipVerify` sites in the repo are the same class: internal self-signed service calls between UTMStack components.

### F-6 — Public OpenAPI/Swagger (RECLASSIFIED: not a finding)
- `springdoc-openapi-ui` is in `backend/pom.xml`; `/v3/api-docs` + `/swagger-ui` are served unauthenticated (verified live: 200, 374 KB, 252 paths). **Product decision 2026-10-28: UTMStack is open source — the spec is generated from public source and discloses no confidential information; the version string is already exposed via `/api/info/version` and the login page.** No fix scheduled; optional nginx gating if ever wanted.

### Verified clean / low-risk (checked, no finding)
- **SQLi (DB):** native `createNativeQuery` sites are parameterized or allowlisted — `CollectorService.buildSearchQuery` (bound params + `validateSortColumn` allowlist), `UtmMenuServiceImpl` concatenates only a `Long`, `UtmAssetGroupService` uses the CVE-2026-82039 fix. No request-text concatenation into Postgres SQL found.
- **Command injection (backend):** zero `Runtime.exec`/`ProcessBuilder` in `backend/src/main/java`; shell execution is isolated to the agent side by design (IR).
- **Path traversal (backend):** `AppInfoService`/`DefinitionSyncService` read fixed paths; CVE-2026-82040/82044 URL validations are in place.
- **Java deserialization:** no request-path `ObjectInputStream`; SnakeYAML config-only (see B2).
- **Broken access control (backend core):** `SecurityConfiguration` admin-gates match the role model (identity-providers, utm-providers, alert-response-rules, encrypt, delete-index, management/logs, federation-service, incident-jobs/variables — the latter two now correctly opened to analysts on `fix/ir-role-model-v11.3.1`).

---

## Part D — Recommended sequencing for v11.3.1

**P0 (fix in this release):**
1. **Go bump pass** — `x/crypto`≥0.55 (shared, agent/updater, installer), `cel-go` 0.29.0 (all modules), `grpc` 1.84.0 (4 plugins), `x/net`≥0.55 (shared, agent/updater), `otel` 1.45 (installer), `quic-go` 0.59.1 (3). Closes ~86 alerts including 21 criticals.
2. **jspdf cluster** — bump jspdf (2.3.1 → patched) with export-path tests; verify dompurify/canvg/fflate move with it.

**P1 (fix in this release, low effort):**
3. `postgresql` 42.7.12 (user-auditor); `js-yaml` 4.3.2 (direct frontend dep, user-uploaded YAML is the vector); `lodash` 4.17.23; `highlight.js` 9.18.2.

**P2 (hygiene):**
4. **F-1 residual:** delete or harden `safe.pipe` (it's a one-line `bypassSecurityTrustHtml` trap; currently only fed admin-authored or pre-escaped data) + CI grep-gate on new `| safe:'html'` usages. `SqlQueryFilterService` value-escaping (C4/C5). GVM detail bindings get `escapeAndBreakLine` only if that integration is ever revived (F-2).
5. `websocket-driver` 0.7.5 if the bundle scan proves it ships; `tar`/`form-data`/`request` dispositions per bundle scan.

**Not scheduled as fixes (reclassified not-findings after deep re-verification, 2026-10-28):**
- **F-1** stored XSS — NOT exploitable as shipped: Angular 7's default sanitizer neutralizes plain `[innerHTML]` (live PoC payload rendered inert); only the 17 `| safe:'html'` / `bypassSecurityTrustUrl` sites bypass it, and all currently receive admin-authored or pre-escaped data. Residual = hygiene (P2 item 4).
- **F-2** GVM render — dead code, backend endpoints don't exist in this repo.
- **F-3** CORS `*` — no cross-origin browser consumers exist (all same-origin; FS/agents are server-to-server).
- **F-4** `/api/elasticsearch/search/sql` — it IS the Log Explorer SQL-mode feature; analysts run read SQL by design.
- **F-5** soc-ai `InsecureSkipVerify` — internal self-signed OpenSearch calls only; external AI calls verify TLS.
- **F-6** public OpenAPI/Swagger — open-source product; the schema discloses nothing not already public.

**Process (document, not patch):**
6. **snakeyaml 1.33** — config-only constraint + code-review gate (no clean patch on Spring Boot 3.1.5).
7. **Docker SDK (no fix published)** — monitor; document the host-local daemon boundary.
8. **Frontend build isolation** — isolated egress-restricted build, no creds, dist hash/sign, **post-build bundle scan**; batch-dismiss the ~120 dev-scoped + ~22 Angular-7-unpatchable alerts with per-package justification templates.

**Verification to add:**
- Post-build bundle scan proving dev-only packages are absent from `dist/utm-stack`.
- Stored-XSS regression test (F-1).
- jspdf export smoke test (PDF export from a dashboard) after the bump.

---

## Part E — What I did not do (read-only)

- No files modified in this phase; no dependency bumps applied — all of Part D is recommended, not executed.
- Bundle scan not yet run (needs a frontend build; flagged as a verification step).
- F-6 resolved by product decision (open source — public schema is not information disclosure); no runtime verification needed.
- Per-alert GitHub dispositions (dismiss with reason / mark as fixed) not performed — that is an external-state action to do only with your approval, per bucket in B4.
