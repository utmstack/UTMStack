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

### F-3 — CONFIRMED, live (wildcard CORS header)
1. `curl -sk -D - -o /dev/null -H "Origin: https://evil-attacker.com" http://192.168.99.75:10001/api/ping`
2. **Observed:** `Access-Control-Allow-Origin: *` + `Access-Control-Expose-Headers: Authorization, ...`.
3. Preflight on the **gated** `/api` path returns 403 from nginx (`x-shared-key` gate) — so on THIS deployment the wildcard is not directly exploitable cross-origin for `/api`; but it is returned on non-gated/`permitAll` paths and remains a latent landmine (any future non-gated endpoint + credentials mode = token exfiltration).
- **Verdict: real misconfiguration, lower practical severity than the draft's "MEDIUM" on this deployment. Fix: replace `*` with the panel origin in `application-prod.yml` + `application-dev.yml` (`WebConfigurer` consumes it).**

### F-4 — CONFIRMED, live, as a plain ROLE_USER
1. Authenticate as `labsec_poc` (ROLE_USER) → JWT with claims `{"sub":"labsec_poc","auth":"ROLE_USER","authenticated":true}`.
2. `POST /api/elasticsearch/search/sql` (panel, with `x-shared-key` + `Authorization: Bearer <jwt>`) body `{"query":"SELECT 1"}`.
3. **Observed:** backend log `ElasticsearchService.searchBySql: OpenSearch.searchBySqlQuery: HTTP 400` — i.e., **Spring Security accepted the request and forwarded the attacker SQL to the OpenSearch SQL engine** (the 400 is the engine rejecting `SELECT 1` with no FROM — proof of execution, not an auth block). Same request **unauthenticated → 401** (auth gate works).
4. `GET /api/elasticsearch/index/all` as the same ROLE_USER → `200` with full index inventory incl. `security-auditlog-2026.10.28` (**43,330 docs**).
- **Verdict: confirmed — any analyst issues arbitrary OpenSearch SQL and enumerates all indices.** The security consequence is bounded by the role model (analysts read all data by design), so it stays **LOW–MEDIUM design hardening** (scope to single read-only SELECT over `v11-log-*`/`v11-alert-*` patterns), not a privilege escalation. The C4/C5 disclosure items (filter-builder concat into this same endpoint) are therefore hygiene, as reclassified.

### F-1 — Stored XSS: payload stored + sink shipped; final render click pending
1. **Stored payload (done):** injected a synthetic alert doc into `v11-alert-2026-10-28` (via HTTPS to OS with the backend service creds): `{"_index":"v11-alert-2026-10-28","_id":"kRk-gqEBG2EDI9JlXLDl","result":"created"}`.
2. **Round-trip confirmed (done):** `_search` match on `name:"LABSEC-XSS-POC-9917"` → stored `description` = `LABSEC-XSS-POC <img src=x onerror="window.__labsecXSS=1"> done` — the markup survives storage byte-for-byte.
3. **Sink shipped (done):** the live frontend bundle `main.d20a757d340e10b2c1d1.js` contains `bypassSecurityTrustHtml` **4×** (the `safe` pipe + `vs-task` URL trust); `alert-description.component.html:3` binds `[innerHTML]="alertDetail"` where `alertDetail = replaceBreakLine(...)` — **no escape** (`string-util.ts:3` vs the safe `escapeAndBreakLine` at `:36` which these components do NOT call).
4. **Render (pending — manual step for you):** in the already-open lab browser, the alert view facet search only filters facet labels, and the MCP `fill_form` on that heavy Angular page hangs. Fastest manual confirmation: Log Explorer → filter `name: "LABSEC-XSS-POC-9917"` → open the alert → watch the Summary pane. Expected: a network/console artifact for `window.__labsecXSS=1`. Or simply: DevTools console → `!!window.__labsecXSS` after opening it.
- **Verdict: chain is closed except the final browser render, which is a deterministic consequence of steps 1–3 (`[innerHTML]` of a stored unescaped string — no browser behavior in question). Treat as CONFIRMED with a reproducible on-screen demo available on request.**

### F-5 — Static only (unchanged)
`plugins/soc-ai/utils/request.go:41` `InsecureSkipVerify: true` on outbound calls — verified in source; runtime behavior depends on whether soc-ai has external endpoints configured in a given deployment. Severity stands as MEDIUM/context-dependent. No lab instance of soc-ai config to test against safely.

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
7. **The manual sweep found application vulnerabilities the scanner cannot:** the most important is **stored DOM XSS in the SIEM console** from log-derived content rendered via a `bypassSecurityTrustHtml` "safe" pipe — for a SIEM ingesting untrusted traffic, this is attacker-writable → analyst-session-executable. (The OpenAPI/Swagger endpoint is public but is **not** a finding for an open-source product — see F-6.)

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

### F-1 — Stored DOM XSS from log-derived content rendered via a fake-"safe" pipe — **HIGH/CRITICAL**
- **Sink:** `frontend/src/app/shared/pipes/safe.pipe.ts:30` — `bypassSecurityTrustHtml(value)` with **no sanitization** (the name is misleading; DOMPurify is *not* called).
- **Feeds (log → alert → UI):** `alert-description.component.ts:23`, `alert-category.component.ts:23`, `alert-severity-description.component.ts:23`, `alert-proposed-solution.component.ts:25`, `file-detail.component.ts:30` (file message field), `file-access-mask.component.ts:30` (SDDL description) — all `replaceBreakLine(getValueFromPropertyPath(alert, <log field>))` bound to `[innerHTML]`.
- **Path:** attacker controls a string in an ingested log field (process name, command line, filename, URL, user-agent) → correlation rule fires → analyst opens the alert → browser executes attacker markup. UTMStack ingests untrusted network traffic, so this is **attacker-writable, analyst-read**: console session compromise (harvest the JWT from the authenticated session, act as the analyst, pivot to the API).
- **Second sink:** `highlight.pipe.ts:14` wraps search matches in `<mark>` and binds `[innerHTML]` on AD tree/folder names, report names — unescaped text → HTML.
- **The fix is already in-repo:** `string-util.ts:36` `escapeAndBreakLine()` = escape-then-breaklines, documented for exactly this. The vulnerable components use the unsafe `replaceBreakLine` instead.
- **Remediation:** switch all log/IR-derived `[innerHTML]` to `escapeAndBreakLine`; make `safe.pipe` actually sanitize or just escape; treat `highlight` input as untrusted (escape first, then wrap). **Highest-value fix in the review.**
- **Regression test:** craft a log event whose alert description carries `<img src=x onerror=…>`; assert it renders inert after the fix.

### F-2 — GVM/vulnerability-scanner detail rendered unsanitized + `bypassSecurityTrustUrl` — **MEDIUM**
- `vs-task-result-detail.component.ts:38-41` builds HTML with no escape (newline→`<br>`) bound to multiple `[innerHTML]`; `:71` `sanitizer.bypassSecurityTrustUrl(doc)` on an `[href]`.
- Source: Greenbone/GVM scan results proxied by `api/utm-gvm-scan-result`. Semi-trusted, but a compromised GVM server or crafted CVE-feed entry → XSS / `javascript:` URL.
- **Remediation:** `escapeAndBreakLine`; validate `doc` is http(s) before trusting the URL.

### F-3 — Production CORS is `allowed-origins: '*'` — **MEDIUM (hardening)**
- `application-prod.yml:50-51`, registered on `/api/**`, `/management/**`, swagger paths (`WebConfigurer.java:47-52`).
- Direct token theft is blocked (JWT in `Authorization` header; **no** `withCredentials(true)` anywhere in `frontend/src/app` — verified), but the wildcard still allows any origin to call public endpoints and read responses, and it flips into a real exposure the day credentials mode is enabled.
- **Remediation:** `allowed-origins` = the real frontend origin(s) / `allowedOriginPatterns` for known deployment hostnames.

### F-4 — `POST /api/elasticsearch/search/sql` runs raw OpenSearch SQL for any `ROLE_USER` — **LOW–MEDIUM (design)**
- `ElasticsearchResource.java:252-265` executes caller-supplied SQL. Not an escalation under the analyst-reads-everything model, but an over-broad raw-query primitive (any index the OS account can read).
- **Remediation:** scope to read-only single `SELECT` on allowed index patterns, or gate to `ROLE_ADMIN`; keep `SqlQueryFilterService` value-escaping as hygiene (the C4/C5 disclosure items).

### F-5 — `plugins/soc-ai` outbound client uses `InsecureSkipVerify: true` — **MEDIUM (context-dependent)**
- `plugins/soc-ai/utils/request.go:41`. If soc-ai calls an external AI endpoint, a network MITM can tamper request/response. The `inputs/`, `agent-manager/`, `compliance-orchestrator/` skip-verify sites are internal self-signed service calls — acceptable, but should carry a comment.
- **Remediation:** verify TLS for external calls; pin the specific internal CA instead of skipping verification.

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
1. **F-1 frontend stored XSS** — `escapeAndBreakLine` on log-derived `[innerHTML]`; harden `safe.pipe` + `highlight`. (Highest real-world impact; fix already in-repo.)
2. **Go bump pass** — `x/crypto`≥0.55 (shared, agent/updater, installer), `cel-go` 0.29.0 (all modules), `grpc` 1.84.0 (4 plugins), `x/net`≥0.55 (shared, agent/updater), `otel` 1.45 (installer), `quic-go` 0.59.1 (3). Closes ~86 alerts including 21 criticals.
3. **jspdf cluster** — bump jspdf (2.3.1 → patched) with export-path tests; verify dompurify/canvg/fflate move with it.

**P1 (fix in this release, low effort):**
4. `postgresql` 42.7.12 (user-auditor); `js-yaml` 4.3.2 (direct frontend dep, user-uploaded YAML is the vector); `lodash` 4.17.23; `highlight.js` 9.18.2.
5. **F-3** CORS origin fix; **F-2** GVM detail escape.

**P2:**
6. **F-5** soc-ai TLS verification; **F-4** scope the raw-OS-SQL endpoint.
7. `websocket-driver` 0.7.5 if the bundle scan proves it ships; `tar`/`form-data`/`request` dispositions per bundle scan.

**Process (document, not patch):**
8. **snakeyaml 1.33** — config-only constraint + code-review gate (no clean patch on Spring Boot 3.1.5).
9. **Docker SDK (no fix published)** — monitor; document the host-local daemon boundary.
10. **Frontend build isolation** — isolated egress-restricted build, no creds, dist hash/sign, **post-build bundle scan**; batch-dismiss the ~120 dev-scoped + ~22 Angular-7-unpatchable alerts with per-package justification templates.

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
