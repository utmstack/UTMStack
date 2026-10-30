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
- **Verdict: NO live stored XSS.** The earlier "CONFIRMED" was wrong — it conflated "markup survives storage" with "markup executes at render," which is false under Angular's default `[innerHTML]` sanitization. The stored-payload PoC is retired (doc deleted). The follow-up hardening (P0.3) still shipped: log-derived `[innerHTML]` sources now pre-escape via `escapeAndBreakLine`, and `safe.pipe`'s default case no longer bypasses the sanitizer.

### F-5 — NOT A FINDING (reclassified 2026-10-28 after re-verification)
- The `InsecureSkipVerify` client in `plugins/soc-ai/utils/request.go:41` is selected **only** when `opts.SkipTLSVerify` is set, and the only caller doing that is `DoOpenSearchReq` (request.go:137-148) — i.e., internal calls to the self-signed OpenSearch endpoint. All other soc-ai traffic (including external AI-model calls via `DoReq`) uses the standard-verification `defaultClient`. Same house pattern as `inputs/`, `agent-manager/`, `compliance-orchestrator/`. **No remediation needed.**

### Lab cleanup — DONE (2026-10-28)
- Throwaway account removed: `DELETE 1` on `jhi_user_authority` (user_id=4) + `DELETE 1` on `jhi_user` (`labsec_poc`); remaining logins verified = `fsclient, admin, fstestuser` (original set).
- Poison alert doc deleted: `_doc/kRk-gqEBG2EDI9JlXLDl` → `{"result":"deleted"}`, post-delete `_search` on `name:"LABSEC-XSS-POC-9917"` → `hits.total.value: 0`.
- No other lab state changed.

### jspdf per-CVE reachability trace (why 2.5.2 is retained) — 2026-10-30
UTMStack's ONLY jspdf import is `frontend/src/app/shared/util/export-to-pdf-util.ts`: `new jsPDF('p','mm','a4')` (literal), `addImage(fixedLocalAsset,'PNG')`, `addImage(canvas.toDataURL('image/png'),'PNG')`, `output('blob')`, `save(timestamp+'.pdf')`. No jspdf API takes attacker input. The browser webpack build ships `dist/jspdf.es.min.js` (not the node build). Per CVE:

| jspdf CVE | Sev | Vuln path | Reachable in UTMStack? |
|---|---|---|---|
| CVE-2025-68428 | crit | **node build** `loadFile` → `fs` LFI | **No** — browser build has no `fs`; `loadFile` never called |
| CVE-2026-31938 | crit | `output(options)` HTML injection (newwindow types) | **No** — `output('blob')` string literal only |
| CVE-2026-31898 | high | `createAnnotation` color | **No** — not called |
| CVE-2026-25940 | high | AcroForm RadioButton `appearanceState` | **No** — AcroForm not used |
| CVE-2026-25755 | high | `addJS` PDF object injection | **No** — not called |
| CVE-2026-25535 | high | `addImage` malicious **GIF** OOM | **No** — only `'PNG'` passed |
| CVE-2026-24737 | high | AcroForm choice/radio JS | **No** — not used |
| CVE-2026-24133 | high | `addImage` malicious **BMP** OOM | **No** — only `'PNG'` |
| CVE-2026-24043 | med | `addMetadata` XMP injection | **No** — not called |
| CVE-2026-24040 | med | `addJS` shared-state (node) | **No** — not called / browser build |
| CVE-2025-57810 | high | `addImage` malicious **PNG** CPU DoS | **No** — PNG is a locally-drawn canvas, not attacker data |
| CVE-2025-29907 | high | `addImage` malicious **data-url** CPU DoS | **No** — dataURL is locally generated |
| CVE-2026-31898 / 25940 / 24737 (dupes) | — | — | — |

Bundled optionals (dompurify 2.5.9, canvg 3.0.11) load ONLY via dynamic `import()` inside jspdf's `html`/`addSvgAsImage` plugins — UTMStack feeds a PNG, so neither executes. **Conclusion:** keeping jspdf 2.5.2 is a valid reachability disposition; the 12 package CVEs are closed off by the browser-build + fixed-asset + local-PNG + `output('blob')` combination. Revisit when the Angular framework upgrade makes jspdf 4.x buildable.

---

## Part A — Executive summary

1. **The backlog is 332 alerts but only ~90 distinct packages, and the split is stark:** 236 are the frozen Angular 7 / Node 14 frontend lockfile, ~65 are Go modules (mostly the same 6 packages repeated per `go.mod`), 2 are Java (`snakeyaml` backend, `postgresql` user-auditor).
2. **~60% of the alerts are development-scoped** (webpack/karma/node-sass/postcss toolchain). Those never execute in the shipped bundle or on the server — their control is a hardened isolated build, not patching.
3. **The genuinely runtime-reachable, patchable cluster is Go, and it is one `go get` pass:** `x/crypto` (41 alerts, 21 critical — pinned low in **three** manifests: `shared` 0.42.0, `installer` 0.51.0, `agent/updater` 0.47.0; the rest of the repo is already ≥0.52), `cel-go` 0.28.1→0.29.0 (~20 modules, the rule-evaluation engine, hot path), `grpc` 1.81.1/1.83.x→1.84.0 (4 plugin modules; xDS CVEs not applicable — no xDS usage), `x/net` (shared 0.44.0, agent/updater 0.49.0 → 0.55.0), `otel` (installer 1.41.0 → 1.45.0), `quic-go` 0.59.0→0.59.1 (3 modules).
4. **The frontend runtime criticals are the jspdf cluster — resolved at 2.5.2 with a written reachability argument (not a version bump to patched).** `jspdf` 2.3.1 was imported in `export-to-pdf-util.ts`. jspdf 3.x/4.x (the patched majors) **cannot build under Angular 7/webpack 4** — jspdf 3.0.4's `fast-png@6` dep uses ES2022 class-field syntax webpack 4's parser rejects (verified build failure). The original premise that 3.x "drops dompurify/canvg" was wrong: those are `optionalDependencies`, present in both 2.5.2 and 3.0.4. At 2.5.2 they resolve at **patched** versions (dompurify 2.2.7→2.5.9, canvg→3.0.11, fflate→0.8.3) and **never execute** for UTMStack's PNG-only `addImage` usage. All 12 open jspdf package CVEs are **unreachable** in UTMStack's actual API usage (browser build not node; `addImage` fed only a fixed local asset + a locally-drawn canvas PNG; `output('blob')` literal; no AcroForm/`addJS`/`addMetadata`/`addSvgAsImage`/`html`). Per-CVE trace in Part B½. **Residual:** the 12 CVEs stay open on paper until the framework upgrade makes jspdf 4.x buildable.
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
1. **Go bump pass — DONE** (commits 2359e6be, 27a44fc5, f7d6e1b2, 4cb113e1, 2f4cdda2): `x/crypto` 0.55.0 (shared/agent/updater/installer + 4 grpc plugins), `cel-go` 0.29.0 (19 modules), `grpc` 1.84.0 (stats/soc-ai/geolocation/alerts), `x/net`≥0.55 (MVS floors 0.57/0.58), `otel` 1.45.0 (installer), `quic-go` 0.59.1 (3). ~86 Dependabot alerts closed incl. 21 x/crypto criticals. x/crypto is not exercised by current code (all `// indirect`; no `x/crypto/ssh` agent-constraint usage) but ships in every binary, so it was bumped anyway.
2. **jspdf — 2.5.2 KEPT with written reachability analysis (security-approved).** 3.x/4.x are NOT buildable under Angular 7/webpack 4: jspdf 3.0.4's `fast-png@6` dep uses ES2022 class-field syntax webpack 4 cannot parse (verified build failure). The original "3.0.4 drops dompurify/canvg" premise was wrong — they are `optionalDependencies`, present in both 2.5.2 and 3.0.4. At 2.5.2: (a) the optional cluster resolves at PATCHED versions (dompurify 2.5.9, canvg 3.0.11, fflate 0.8.3) and never executes for UTMStack's PNG-only usage; (b) all 12 open jspdf package CVEs are unreachable in UTMStack's actual API usage — per-CVE trace in Part B½ (browser build not node; `addImage` fed only a fixed local asset + locally-drawn canvas PNG; `output('blob')` literal; no AcroForm/addJS/addMetadata/addSvgAsImage/html). **Residual accepted:** the 12 jspdf CVEs stay open on paper until the framework-upgrade project makes jspdf 4.x buildable; the reachability memo is the disposition.
3. **F-1 frontend HTML-sink hardening — DONE** (502ace3285 + alert-history follow-up): 6 log-derived `[innerHTML]` components → `escapeAndBreakLine`; `highlight` pipe escape-first-then-wrap; `safe.pipe` default case now escapes (no more bare-`| safe` bypass footgun). Full 79-site `[innerHTML]` sweep found no missed exploitable sink. Node14 build green.

**P1 (fix in this release, low effort):**
4. **DONE** — `postgresql` 42.7.12 (user-auditor, 008cc834); `js-yaml` 4.3.2 + `lodash` 4.17.23 + `highlight.js` 9.18.5 (b4f7847a).
5. **DONE** — **C1** SAML fetch-time URL validation (eb8dd421 + e56017ca); **C4/C5** SqlQueryFilterService escape + field allowlist (f8f8ff9a + 7c2d6e77).
6. **DONE** — **M2** bitdefender constant-time compare (fdc193fb); **C2** inputs-plugin internal-key allowlist (36a55d3b); **C3** web-pdf internal-key gate (9a6a317d).

**P2 (hygiene / conditional):**
7. `websocket-driver` 0.7.5 if the bundle scan proves it ships; `tar`/`form-data`/`request` dispositions per bundle scan.
8. GVM detail bindings get `escapeAndBreakLine` only if that integration is ever revived (F-2, dead code).

**Not scheduled as fixes (reclassified not-findings after deep re-verification, 2026-10-28):**
- **F-2** GVM render — dead code, backend endpoints don't exist in this repo.
- **F-3** CORS `*` — no cross-origin browser consumers exist (all same-origin; FS/agents are server-to-server).
- **F-4** `/api/elasticsearch/search/sql` — it IS the Log Explorer SQL-mode feature; analysts run read SQL by design.
- **F-5** soc-ai `InsecureSkipVerify` — internal self-signed OpenSearch calls only; external AI calls verify TLS.
- **F-6** public OpenAPI/Swagger — open-source product; the schema discloses nothing not already public.
- **F-1 original "stored XSS" claim — retracted** (Angular 7's default sanitizer neutralizes plain `[innerHTML]`; live PoC payload rendered inert). The hardening above (P0.3) is defense-in-depth + footgun removal, not a vuln fix.

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
- Per-alert GitHub dispositions (dismiss with reason / mark as fixed) not performed - that is an external-state action to do only with your approval, per bucket in B4.

---

## Part F — End-to-end verification on lab 192.168.99.75 (2026-10-30)

Each affected user-facing flow was driven through a real browser (Playwright headless Chromium) or the live API on the lab, against the fixes deployed as slim images (backend `utmstack-backend-local:v11.3.1fix`, frontend `utmstack-frontend-local:v11.3.1`). Throwaway analyst `labsec_e2e` (ROLE_USER + TOTP) was created for the runs and **deleted** afterward; both poison alert docs were deleted (confirmed 0 hits).

| Item | E2E evidence (live) | Result |
|---|---|---|
| **C1** SAML fetch-time URL validation | Inserted a poisoned legacy row (valid encrypted SP keypair + cert, `metadata_url=http://127.0.0.1:9/...`) directly in `utm_identity_provider_config`; force-restarted backend. Startup log: `Metadata fetch failed for provider 'LABSEC-C1-LOOPBACK': SamlMetadataUrlInvalidException: Metadata URL host must be publicly resolvable`. **0** outbound connections to `127.0.0.1:9` (grep count 0). Row deleted after. | PASS — fetch-time validator rejects before any request; no over-blocking of the loader |
| **C3** web-pdf internal-key gate | `GET /generate-pdf` with no `X-Utm-Internal-Key` → **401**; wrong key → **401**; backend's key → **408** (proceeds to render, fails on the fake token) — i.e. the gate passes with the correct key and the backend→web-pdf header path works end-to-end. | PASS |
| **C4/C5** SqlQueryFilterService escape | **App-level:** `POST /api/utm-visualizations/run` (SQL viz) with `audit_cluster_name IS 'utmstack'` → **200, 5 rows**; with value `it's` → **200, 0 rows** (escaped, parses + executes, no match). **Raw-engine differential:** the same value **unescaped** (`= 'it's'`) → **400 ParserException** ("Illegal SQL expression"); **escaped** (`= 'it''s'`) → **200**. `SELECT 1` control → 200 with `[[1]]`. | PASS — quote-doubling is the difference between a parse failure and clean execution; the app emits the escaped form |
| **F-1** stored-XSS hardening | Indexed a poison alert (`id` unique, `description` = `LINE1\nLINE2 <img src=x onerror=...> <script>...`), opened `/data/alert/detail/<id>` in the browser. `LINE2` and the literal `onerror` text are **visible** (description rendered, non-vacuous) and `window.__labsecXSS`/`__labsecScript` are **null**, no live `<img src=x>` / `<script>` in DOM. | PASS — payload stored, rendered inert as text, no handler executed |
| **highlight** pipe (highlight.js 9.18.5) | Active Directory search for `a` → **59 `<mark>`** elements rendered (escape-first-then-wrap still marks matches). | PASS |
| **js-yaml** 4.3.2 (rule YAML parse) | The exact installed lib parses a UTMStack-shaped rule with merge-keys/anchors (the CVE-2026 vector) correctly and throws a clean `YAMLException` on malformed input. (The rule-editor page itself throws a pre-existing `Unexpected token 'export'` on a lazy chunk — proven **not** a regression: the baseline v11.3.0 build reproduces the identical error.) | PASS (lib level; page error pre-existing & documented) |
| **jspdf** 2.5.2 (PDF export) | In-browser E2E with the app's REAL bundled libs: jspdf 2.5.2 + html2canvas 1.0.0-rc.7 → `new jsPDF('p','mm','a4')`+`addImage(html2canvas.toDataURL('image/png'),'PNG')`+`output('blob')` produced a valid **245,360-byte `%PDF`** (`application/pdf`) in headless Chromium. Also valid `%PDF` via Node on the exact installed lib. The live dashboard `Preview`/`Save to PDF` buttons hang in `html2canvas` on this lab because the dashboard charts don't render (0 echarts) — **A/B-proven pre-existing on stock v11.3.0** (identical `Unexpected token 'export'` + "No data found" with none of my changes), so it is not a jspdf regression. | PASS |

**Not re-verified live (documented, not gaps in the fixes):** the dashboard `Save to PDF` download event is flaky in headless Chromium (the build completes and reverts the button; the isolated lib-level PDF generation proves jspdf 2.5.2 emits valid bytes). The M2 (bitdefender) and C2 (inputs gRPC) changes are Go plugin authz with no host-published port to probe from this box; they are covered by their unit/behavior-pinning tests and the security-review trace.

**Lab state after this pass:** backend + web-pdf + frontend run the local fix images; the `labsec_e2e` user and both poison alert docs are removed; the OS SQL plugin is confirmed functional (`SELECT 1` → 200).
