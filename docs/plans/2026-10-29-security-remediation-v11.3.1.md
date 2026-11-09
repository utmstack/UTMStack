# Security Remediation v11.3.1 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development (recommended) or executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship all non-discarded security remediations for v11.3.1 — dependency bumps (Go cluster, jspdf cluster, P1 npm/Java) plus application fixes (C1 SAML fetch validation, C4 SQL-filter escaping, M2 constant-time compare, C2 inputs gRPC allowlist, C3 web-pdf auth, F-1 frontend XSS hardening) — each with end-to-end tests of affected functionality.

**Architecture:** Backend Java tests run in Docker (box is Java 25; `maven:3.9-eclipse-temurin-17`, offline against mounted `C:\Users\osmon\.m2`). Go modules build/test natively (Go 1.27.1 on box). Frontend builds in Node 14.16.1 Docker (node-sass 4). End-to-end verification of user-facing flows on lab 192.168.99.75 where the deployed image can be slim-rebuilt (backend WAR / frontend dist), per the established #2985 pattern.

**Working branch:** `fix/ir-role-model-v11.3.1` (main checkout, current with v11 head `720764a9`). NEVER run `git stash` (user WIP stash exists). Leave `frontend/package-lock.json`'s pre-existing dirty state and `docs/qa-v11.3.0-test-plan.md` untouched until their owning task commits them. Commits use `--no-gpg-sign`.

**Decisions locked (do not re-open):**
- jspdf → **3.0.4** (identical module shape to 2.5.2, no `exports` map for webpack 4, no dompurify/canvg deps, clears all `<= 3.0.1`-range alerts).
- `safe.pipe` keeps `html` case for **trusted** content (compliance admin text, pre-escaped console output); log-derived components switch to `escapeAndBreakLine`; default case escapes.
- C2 inputs-plugin internal-key branch → allowlist `/grpc.health.v1.Health/Check` only (plugin implements Integration + Health; agents/collectors use key/id/type branch, federation uses connection-key branch — verified, no internal-key data-plane caller exists).
- C3 web-pdf: require `X-Utm-Internal-Key` header == `INTERNAL_KEY` env; backend `PdfService` sends it; installer gives web-pdf the env.
- cel-go bump is **direct** per module (go-sdk 1.1.38 pins cel-go 0.28.1; MVS makes direct 0.29.0 win).

**Go module inventory (from go.mod scan, 2026-10-29):**
- go-sdk v1.1.38 everywhere except `plugins/compliance-orchestrator` (v1.1.8 — stale, out of scope, flag to user).
- x/crypto lagging: shared 0.42.0, agent/updater 0.47.0, installer 0.51.0 → **0.55.0**.
- x/net lagging: shared 0.44.0, agent/updater 0.49.0 → **0.55.0**.
- quic-go 0.59.0 → 0.59.1 in installer, agent/updater, compliance-orchestrator.
- otel 1.41.0 → 1.45.0 in installer (sdk + otlptracehttp; check contrib/otelhttp too).
- grpc lagging: plugins/stats 1.81.1, plugins/soc-ai 1.81.1, plugins/geolocation 1.81.1/1.83.1, plugins/alerts 1.83.1/1.83.2 → **1.84.0**.
- cel-go 0.28.1 (compliance-orchestrator 0.26.1) → **0.29.0** in all ~20 modules.

**Backend test command (all Java tasks):**
```
docker run --rm -v "C:\Users\osmon\.m2:/root/.m2" -v "D:\Projects\github.com\utmstack\UTMStack:/src" -w /src/backend maven:3.9-eclipse-temurin-17 sh -c 'mvn -B -o -Drevision=11.3.1 clean test -Dtest=<Class> -s settings.xml'
```
(user-auditor uses its own pom — same image, `-w /src/user-auditor`.)

**E2E on lab (192.168.99.75, v11.3.0 swarm):** slim image rebuilds — `docker build` FROM the running service image + `COPY` the new WAR/dist, then `docker service update --force --image <svc>:v11.3.0-local`. Order: backend first. Lab creds: shared_key + admin TOTP secret in `/root/utmstack.yml` (read via `sudo -n grep`); OS is **HTTPS** from the backend container with `ELASTICSEARCH_USER/PASSWORD` env; docker via `sudo -n`. Write scripts locally → `scp` → `bash` (PowerShell inline quoting through ssh mangles everything — standing rule).

---

## Execution status (verified 2026-10-30)

All tasks implemented, committed, and reviewed (implementer → spec → quality → security). E2E verified on lab 192.168.99.75 where a live path exists:

| Task | Unit/build | Lab E2E |
|---|---|---|
| A Go deps | 23/23 modules build; alerts/collector tests = baseline (no new failures) | n/a (no host-exposed Go service to probe) |
| B C1 SAML | 15/15 hermetic | **PASS** — poisoned loopback row rejected at fetch time (`SamlMetadataUrlInvalidException`), 0 outbound conns |
| C C4/C5 SQL | 16/16 | **PASS** — `POST /api/utm-visualizations/run`: `it's` → 200/0 rows (escaped), unescaped control → 400 ParserException; app emits escaped form |
| D M2 bitdefender | 5/5 behavior-pinning | n/a (plugin port not host-published) |
| E C2 inputs gRPC | 13/13 driving real auth branch | n/a (plugin port not host-published) |
| F C3 web-pdf | 10/10 + backend 67/67 | **PASS** — 401 no key / 401 wrong key / 408 with key (gate passes, render proceeds) |
| G frontend deps | Node14 build green | **PASS** — js-yaml 4.3.2 lib (merge-key vector + clean throw); jspdf 2.5.2 valid `%PDF` via exact app calls; dashboard export runs clean in-browser |
| H F-1 XSS | Node14 build green; 79-site `[innerHTML]` sweep | **PASS** — poison alert rendered inert (`LINE2`+`onerror` visible as text; `__labsecXSS`/`__labsecScript` null; no live markup) + highlight 59 `<mark>` |
| I postgres 42.7.12 | compiles green | n/a (user-auditor internal) |

Full E2E evidence: `docs/2026-10-27-security-review-dependabot-and-app-sweep.md` **Part F**.

---

## Task A — Go dependency cluster (P0, ~86 Dependabot alerts)

**Files (per module `go.mod`/`go.sum`):**
- `shared/go.mod`, `agent/updater/go.mod` — x/crypto 0.55.0, x/net 0.55.0
- `agent/go.mod`, `as400/go.mod`, `utmstack-collector/go.mod`, `plugins/{alerts,aws,azure,bitdefender,compliance-orchestrator,config,crowdstrike,events,gcp,geolocation,inputs,modules-config,o365,rule-flood-guard,soc-ai,sophos,stats}/go.mod` — cel-go 0.29.0
- `plugins/{stats,soc-ai,geolocation,alerts}/go.mod` — grpc 1.84.0
- `installer/go.mod` — x/crypto 0.55.0, x/net 0.55.0, quic-go 0.59.1, otel/sdk 1.45.0, otel exporters 1.45.0
- `agent/updater/go.mod` — quic-go 0.59.1

**Affected functionality to test:** every Go module that compiles/tests (agent, agent-manager, collector, as400, all 17 plugins, shared via consumers, installer). cel-go is the rule engine → `plugins/alerts` contract tests are the behavioral E2E.

- [ ] **Step 1: shared** — `cd shared; go get golang.org/x/crypto@v0.55.0 golang.org/x/net@v0.55.0; go mod tidy`. Verify: `go build ./...` (shared has no main; build via a consumer — `agent` in Step 2 covers it). Note: `shared` is consumed via replace directives, so its own tidy must keep the `replace` targets consistent.
- [ ] **Step 2: agent + agent/updater + as400 + utmstack-collector** — in each: `go get golang.org/x/crypto@v0.55.0 golang.org/x/net@v0.55.0` (updater also `github.com/quic-go/quic-go@v0.59.1`); `go mod tidy`; `go build ./...`; `go test ./...`.
- [ ] **Step 3: installer** — `go get golang.org/x/crypto@v0.55.0 golang.org/x/net@v0.55.0 github.com/quic-go/quic-go@v0.59.1 go.opentelemetry.io/otel/sdk@v1.45.0 go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.45.0`; then `go mod tidy`; if tidy re-lowers any otel sibling (otel core, metric, trace) below 1.45, `go get` them to 1.45.0 explicitly (they must move together). `go build ./...`; `go test ./...`.
- [ ] **Step 4: the cel-go modules (all except `plugins/compliance-orchestrator`)** — compliance-orchestrator is pinned to go-sdk v1.1.8 (stale, out of scope — flagged to user); bumping its cel-go independently risks MVS churn against the old go-sdk. In each in-scope module: `go get github.com/google/cel-go@v0.29.0`; `go mod tidy`; `go build ./...`.
- [ ] **Step 5: the 4 grpc modules** — `plugins/{stats,soc-ai,geolocation}`: `go get google.golang.org/grpc@v1.84.0`; `plugins/alerts`: same (it also got cel-go in Step 4 — one tidy). `go mod tidy`; `go build ./...`.
- [ ] **Step 6: behavioral tests** — `cd plugins/alerts && go test ./...` (contract + netflow rule tests = cel engine E2E); `cd agent && go test ./...`; `cd utmstack-collector && go test ./...`; `cd shared && go build ./...`.
- [ ] **Step 7: verify resolved versions** — for each modified module: `go list -m golang.org/x/crypto github.com/google/cel-go google.golang.org/grpc` (only the ones it has) and confirm ≥ patched (0.55.0 / 0.29.0 / 1.84.0). Record output.
- [ ] **Step 8: commit** — one commit per module group (shared+agents, installer, cel-go set, grpc set) with message `chore(deps): bump <pkg> to <ver> across <modules> (dependabot <CVE list>)`.

**Expected test evidence:** `go test ./...` green in alerts/agent/collector; `go build ./...` green in every module; version table from Step 7.

---

## Task B — C1: SAML metadata fetch-time validation (backend)

**Why:** `SamlMetadataFetcher.fetch()` (line 33-34) passes the persisted `metadataUrl` straight to `RelyingPartyRegistrations.fromMetadataLocation()` — no scheme/host/redirect validation. The CVE-2026-82040 fix only guards the API write path; a legacy/direct-DB row pointing at loopback/link-local makes the backend issue that request at every startup and provider reload.

**Files:**
- Modify: `backend/src/main/java/com/park/utmstack/config/saml/SamlMetadataFetcher.java`
- Modify: `backend/src/main/java/com/park/utmstack/service/idp_provider/IdentityProviderService.java` (delegate to the shared validator)
- Test: `backend/src/main/java/com/park/utmstack/config/saml/SamlMetadataFetcherValidationTest.java` (create)

**Affected functionality to test E2E:** SAML provider create/update (API path must still validate — regression), provider reload on config change, app startup with SAML rows.

- [ ] **Step 1: Extract shared validator.** In `SamlMetadataFetcher`, add a package-visible static validator with an injectable resolver seam:
  - `public static void assertFetchableMetadataUrl(String url)` — parses URL, requires http/https, resolves via `InetAddress.getByName`, rejects loopback/site-local/link-local/any-local/multicast, throws `SamlMetadataUrlInvalidException`.
  - `static boolean isFetchableHost(InetAddress addr)` — the address-class check, unit-testable without DNS.
  - `static void assertFetchableMetadataUrl(String url, Function<String, InetAddress> resolver)` — overload used by the test with a fake resolver (no real DNS in the test env).
  `IdentityProviderService.validateMetadataUrl` keeps its live-GET reachability check but replaces its inline `isMetadataUrlAllowed` host logic with a call to `SamlMetadataFetcher.isFetchableHost` (behavior-identical: same rejection set).
- [ ] **Step 2: Guard the fetcher.** In `SamlMetadataFetcher.fetch()`, first line of the lambda: `assertFetchableMetadataUrl(entity.getMetadataUrl());` before `fromMetadataLocation`. A reject → logged + returns null (provider skipped), same as a failed fetch. This prevents the outbound request entirely (validate BEFORE Spring's client resolves/redirects — note: re-validating after DNS can't fully defeat TOCTOU rebind without redirect-control, which Spring's SAML client doesn't expose; document that residual).
- [ ] **Step 3: Failing test** `SamlMetadataFetcherValidationTest` (plain JUnit, no Spring): `assertFetchableMetadataUrl` throws for `http://127.0.0.1:9999/x`, `http://169.254.169.254/latest/meta-data`, `http://10.0.0.5/m.xml`, `file:///etc/passwd`, `http://[::1]/x`; does NOT throw for `https://idp.example.com/m.xml` (DNS resolves to public — use a resolvable public host or a `InetAddress`-stubbing seam; if DNS is unavailable in the test env, inject a resolver function parameter and test with a fake). Run → FAIL (method missing).
- [ ] **Step 4: Implement**, run → PASS.
- [ ] **Step 5: Regression test** that `IdentityProviderService` still rejects the same URLs on create/update — reuse existing service tests if present; if not, add a focused test on the delegated method.
- [ ] **Step 6: Full backend suite** (Docker command) → 36+ new, BUILD SUCCESS.
- [ ] **Step 7: E2E lab check (if SAML rows exist on lab)** — `sudo psql` inspect `utm_identity_provider_config` for active rows; if none, create one with a PUBLIC metadata URL via the admin API, confirm startup/reload fetches it (log "Metadata fetch failed" only if unreachable); then a second row with `http://127.0.0.1:9/…` → confirm the fetcher logs the rejection and issues NO outbound request (no new log line from the fetch thread; verify via `tcpdump` on backend container if feasible, else via the absence of the Spring SAML client's connection attempt in logs). Cleanup: delete both test rows.
- [ ] **Step 8: Commit** `fix(security): validate IdP metadata URL at fetch time (C1, Artiphishell)`.

---

## Task C — C4/C5: SqlQueryFilterService escaping + field allowlist (backend)

**Why:** `SqlQueryFilterService.toSqlCondition`/`joinQuoted`/`buildTimestampCondition` string-concat request-controlled field+value into OpenSearch SQL (CWE-89 hygiene — impact bounded by the analyst read-all model, but violates the boundary-validation rule).

**Files:**
- Modify: `backend/src/main/java/com/park/utmstack/service/elasticsearch/sql/SqlQueryFilterService.java`
- Test: `backend/src/main/java/com/park/utmstack/service/elasticsearch/sql/SqlQueryFilterServiceTest.java` (create)

**Affected functionality to test E2E:** chart-builder "run" for SQL visualizations (every operator branch: IS, IS_NOT, CONTAIN, CONTAIN_ONE_OF, IS_ONE_OF, IS_BETWEEN, IS_GREATER_THAN, START_WITH, IS_IN_FIELDS, @timestamp variants) — a filter that previously produced valid SQL must still produce valid SQL; only quote-bearing values change shape.

- [ ] **Step 1: Failing tests** — `applyFilters("SELECT * FROM t", [IS on field 'name' value "o'brien"])` must produce `name = 'o''brien'` (doubled quote). `value "' OR 1=1 --"` must be neutralized. Field `name; DROP` (semicolon/non-identifier chars) → `IllegalArgumentException`. Run → FAIL.
- [ ] **Step 2: Implement** — add `escapeSqlValue(String v) = v.replace("'","''")` applied in every value-interpolation site (`toSqlCondition` all branches, `joinQuoted`, `toSqlTime` absolute branch); add `validateField(String f)` requiring `f` to match `^[A-Za-z_][A-Za-z0-9_.\- ]*$` (OS SQL identifiers incl. dotted paths) else `IllegalArgumentException`; call it at the top of `toSqlCondition` and for `@timestamp` (already literal-safe, but the field name is fixed there).
- [ ] **Step 3: Pass + full suite** (Docker) → BUILD SUCCESS.
- [ ] **Step 4: E2E** — via the lab: with an analyst JWT (recreate throwaway account pattern from F-4 test: `labsec_poc` style, delete after), POST `/api/utm-visualizations/run` with a SQL query + one filter each of: IS (plain value), IS (value containing `'`), CONTAIN_ONE_OF (2 values), @timestamp IS_BETWEEN (now-24h/now), IS_ONE_OF (list) — confirm 200 + correct row filtering for the plain cases and correct escaping (no 500) for the quote case. (This exercises `UtmVisualizationResource.run` → `applyFilters` → `searchBySql`, the exact C4/C5 path.)
- [ ] **Step 5: Commit** `fix(security): escape + allowlist SqlQueryFilterService inputs (C4/C5 hygiene)`.

---

## Task D — M2: bitdefender constant-time auth compare (Go)

**Files:** `plugins/bitdefender/server/server.go:42`, test `plugins/bitdefender/server/server_test.go` (create if absent).

**Affected functionality:** Bitdefender GZ log ingest auth (module active → valid header 200, invalid 401).

- [ ] **Step 1: Failing test** — httptest server with `ModuleActive=true` + known ConnectionKey: valid `GenerateAuthCode(key)` header → 200; wrong same-length → 401. Asserts don't depend on timing (just that compare is `subtle.ConstantTimeCompare` — grep-level, plus behavior). Run → currently PASSES behaviorally (test pins behavior, not the flaw) — so this task's test is a regression pin; the fix is the `==` → `subtle.ConstantTimeCompare([]byte(expected), []byte(got)) == 1` with a length pre-check (return 401 on length mismatch, constant-time otherwise).
- [ ] **Step 2: Implement**, `cd plugins/bitdefender && go build ./... && go test ./...` → green.
- [ ] **Step 3: Commit** `fix(security): constant-time connection-key compare in bitdefender ingest (M2)`.

---

## Task E — C2: inputs-plugin internal-key method allowlist (Go)

**Files:** `plugins/inputs/middlewares.go` (authFromContext), test `plugins/inputs/middlewares_test.go` (create).

**Affected functionality:** `Integration/ProcessLog` for agents+collectors (key/id/type branch — MUST still work), federation connection-key branch (MUST still work), health check via internal-key (MUST still work), and the NEW deny: internal-key on any non-Health RPC.

- [ ] **Step 1: Failing test** — build the gRPC server from `handlers.go` startGRPCServer pieces (or call `authFromContext` directly with synthetic metadata): (a) `internal-key` + method `/grpc.health.v1.Health/Check` → nil error; (b) `internal-key` + method `/plugins.Integration/ProcessLog` → PermissionDenied; (c) valid `key/id/type` + ProcessLog → nil; (d) valid `connection-key` + ProcessLog → nil. Requires `authFromContext` to see the method — change signature to take `*grpc.UnaryServerInfo`/`*grpc.StreamServerInfo` (already available in the interceptors, currently ignored `_`).
- [ ] **Step 2: Implement** — add `internalKeyAllowedMethods = map[string]bool{"/grpc.health.v1.Health/Check": true}`; in the internal-key branch, check `info.FullMethod` (unary) / `ss.Method()` (stream); reject otherwise with `PermissionDenied`.
- [ ] **Step 3:** `cd plugins/inputs && go build ./... && go test ./...` → green.
- [ ] **Step 4: E2E reasoning check** — grep ALL `internal-key` senders in-repo (done: sophos/soc-ai/bitdefender/o365/gcp/azure/aws/crowdstrike → modules-config; inputs health.go → agent-manager; inputs auth.go → agent-manager). NONE target the inputs plugin's own gRPC with internal-key. Document this grep output in the commit message.
- [ ] **Step 5: Commit** `fix(security): allowlist internal-key to Health only on inputs gRPC (C2 hardening)`.

---

## Task F — C3: web-pdf endpoint auth (Java web-pdf + backend + installer)

**Files:**
- `web-pdf/src/main/java/com/utmstack/webtopdf/controller/WebPdfController.java` — require header `X-Utm-Internal-Key` == env `INTERNAL_KEY` (constant-time via `MessageDigest.isEqual`); 401 otherwise; add env-less dev fallback = reject all (fail closed).
- `backend/.../service/util/PdfService.java:63-67` — add `headers.add("X-Utm-Internal-Key", System.getenv("INTERNAL_KEY"))` (backend already has the env, compose.go:210).
- `installer/docker/compose.go` web-pdf service — add `"INTERNAL_KEY=" + conf.InternalKey` to Environment.
- Test: `web-pdf` has no test infra; cover with a focused unit test if one exists, else rely on lab E2E.

**Affected functionality to test E2E:** ALL PDF exports — dashboard export (`report-export`), compliance export/print, alert detail PDF (if the route exists — verify `/data/alert/detail` is a real frontend route before claiming it; if not, drop it from `ALLOWED_PDF_PATH_PREFIXES` too).

- [ ] **Step 1: web-pdf controller test** (if any test harness; else skip to E2E) — missing header → 401, wrong → 401, right → proceeds.
- [ ] **Step 2: Implement** controller + PdfService + compose.
- [ ] **Step 3: Build both WARs** — web-pdf WAR (Java 11 image: `maven:3.9-eclipse-temurin-11` if that's its toolchain — check web-pdf/pom.xml java.version) + backend WAR (17).
- [ ] **Step 4: Lab E2E** — slim-rebuild web-pdf + backend images, force-update services; then as analyst: trigger dashboard PDF export from API (`PdfGeneratorResource` route) → expect a valid PDF (proves the backend→web-pdf header works end-to-end); direct `GET /generate-pdf` with NO header → 401 (proves the gate); with the internal key header → 200 (proves the key path). Verify the installer compose change by diffing the generated compose (or by confirming the env reaches the running service: `docker service ps web-pdf` + `docker exec <web-pdf> env | grep INTERNAL_KEY`).
- [ ] **Step 5: Commit** `fix(security): authenticate web-pdf /generate-pdf with the internal key (C3 hardening)`.

---

## Task G — Frontend P1 bumps (package.json + lock)

**Files:** `frontend/package.json` (ranges), `frontend/package-lock.json` (regenerated by npm i under Node 14).

- [ ] **Step 1:** bump ranges — `jspdf ^2.3.1` → `~3.0.4` (pin tight: 3.x line), `js-yaml ^4.1.0` → `~4.3.2`, `highlight.js 9.12.0` → `~9.18.2` (+ `@types/highlight.js` if it constrains), `lodash ^4.17.21` → `^4.17.23` (or latest 4.17.x).
- [ ] **Step 2:** Node 14 Docker: `npm install` (regenerates lock; 8 GB heap if it OOMs), then `NODE_OPTIONS=--max_old_space_size=8192 npm run build` → **must be green** (this is the compile E2E for the whole frontend).
- [ ] **Step 3:** bundle verification — grep `dist/utm-stack` for: jspdf version marker (3.0.4), NO dompurify/canvg/fflate strings (the cluster is gone from the bundle), js-yaml 4.3.2. Record the grep output.
- [ ] **Step 4:** functional E2E (lab): slim-rebuild frontend image from the new dist, force-update; in the lab browser (user-driven if MCP hangs): (a) open a dashboard → Export PDF → confirm a PDF downloads (jspdf 3 path: `new jsPDF()`, `addImage`, `output('blob')`); (b) open Log Explorer → confirm syntax highlighting still renders (highlight.js); (c) open a rule with YAML (rule view) → confirm js-yaml parse still works.
- [ ] **Step 5: Commit** `chore(frontend): bump jspdf/js-yaml/highlight.js/lodash (dependabot runtime cluster)`.

**Risk gate:** if jspdf 3.0.4 breaks the build or the PDF export, FALL BACK to jspdf 2.5.2 (still clears the dompurify/canvg/fflate cluster — they're absent from 2.5.2's deps too; check which jspdf alerts 2.5.2 satisfies and note the remainder).

---

## Task H — F-1: frontend XSS hygiene (log-derived [innerHTML])

**Files:**
- `alert-description.component.ts` — `replaceBreakLine` → `escapeAndBreakLine`
- `alert-category.component.ts` — same
- `alert-severity-description.component.ts` — same
- `alert-proposed-solution.component.ts` — same
- `file-detail.component.ts` — same
- `file-access-mask.component.ts` — same
- `highlight.pipe.ts` — escape the input text BEFORE wrapping matches in `<mark>` (escape `&`, `<`, `>` in both the `match` and the surrounding text: escape first, then regex-replace on the escaped string with the escaped match token)
- `safe.pipe.ts` — default case: `escapeHtml`-style output (return plain string), keep explicit `html`/`style`/`url` cases for the trusted callers; add a class-level doc comment enumerating the allowed use (trusted/admin-authored or pre-escaped content only)

**Affected functionality to test E2E:** every consumer of the 7 touched components — alert detail view (description/category/severity/solution panes), file-management detail (event message, SDDL access mask tooltip), AD tree + folder list + report template name highlight (search term highlighting must still visually mark matches).

- [ ] **Step 1:** apply the six `escapeAndBreakLine` swaps (import from `shared/util/string-util`).
- [ ] **Step 2:** `highlight.pipe` — new transform: escape text, then replace escaped-search-term occurrences with `<mark>`-wrapped escaped term. (Search term is also user input — escape it the same way in the pattern building; the existing regex-escape of special chars stays.)
- [ ] **Step 3:** `safe.pipe` — default case escapes; `html` case unchanged (trusted callers: compliance components, agent-console which pre-escapes).
- [ ] **Step 4:** Node 14 Docker build → green (compile E2E).
- [ ] **Step 5:** Lab E2E (frontend slim-rebuild, user-driven browser): (a) open an alert with multi-line description → confirm line breaks still render as `<br>` and NO HTML from the field renders (use the stored-payload pattern from the earlier PoC: re-inject a `LABSEC` alert with `<img onerror>` in description, open it, confirm inert text + `!!window.__labsecXSS === false` in console, then delete the doc); (b) AD tree: search a node name → confirm `<mark>` highlighting still appears; (c) dashboard PDF export still works (interacts with highlight-free paths but must not regress).
- [ ] **Step 6: Commit** `fix(security): escape log-derived [innerHTML] sources; safe.pipe default now escapes (F-1)`.

---

## Task I — user-auditor postgres bump (P1, 1 alert)

**Files:** `user-auditor/pom.xml` 42.7.11 → 42.7.12.
- [ ] **Step 1:** edit pom; **Step 2:** Docker build+test: `docker run --rm -v "C:\Users\osmon\.m2:/root/.m2" -v "D:\Projects\github.com\utmstack\UTMStack:/src" -w /src/user-auditor maven:3.9-eclipse-temurin-11 sh -c 'mvn -B -o clean test -s settings.xml'` (check user-auditor's java target first — Spring Boot 2.7 = Java 11) → green. **Step 3:** commit `chore(deps): postgresql 42.7.12 (user-auditor)`.

---

## Execution order & review gates

A (Go, fastest evidence) → B (C1) → C (C4) → D (M2) → E (C2) → F (C3) → G (frontend deps) → H (F-1 frontend) → I (postgres). G and H can share one Node14 build (do both, then one build) — if sequenced that way, run G+H changes together and a single build/lab pass, splitting commits.

Every task: implementer subagent → spec-reviewer → code-quality-reviewer → **security-reviewer** (all of these touch authn/authz, injection sinks, or untrusted input). Critical/High security findings block the task. Lab E2E steps require the lab reachable — if the lab is down at a step, complete code+unit tests, mark the E2E pending, and surface it.
