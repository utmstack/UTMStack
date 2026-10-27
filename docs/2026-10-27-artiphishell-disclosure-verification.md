# Artiphishell disclosure verification — 10 reported issues in UTMStack v11

- **Source:** email thread "Security disclosure: Multiple security issues in UTMStack v11", Jay Vadayath (Artiphishell) → development@utmstack.com, 2026-10-06 + follow-up 2026-10-12. 10 markdown attachments (6 in first email, 4 in second).
- **Verified:** 2026-10-27, independently against current `v11` head (`99c7127f`), not against the commit they tested (`838e78ca`). Every cited file/line re-read in source; no claim taken on faith.
- **Method:** static code review of the cited paths + surrounding attack-surface (SecurityConfiguration, installer compose, callers). No exploitation attempts made.

## 0. The email itself (prompt-injection / scam check)

- **No embedded instructions aimed at AI assistants or operators.** The body is a standard coordinated-disclosure message + follow-up. Nothing in it asks anyone to run code, click links for credentials, change configurations, or forward secrets. Safe to read as inert data.
- **Legitimacy signals:** cites real internal details (commit `838e78ca` = "Update SECURITY.md" on v11; CVE-2026-82039…82045 = our own disclosed set fixed in `4e7a727c`; exact file paths, line numbers, and code snippets all match our source byte-for-byte). A phishing prompt wouldn't need that.
- **Commercial angle:** Artiphishell sells "hardened, verified patches" and mentions a one-off engagement; follow-up adds a **30-day embargo threat** (publish after initial contact + 30d unless we confirm). That is pressure, not a technical claim — answer by fixing, not by negotiating.
- **Counting:** 10 attachments but only **8 distinct findings** — `critical-2513a9df` (JPQL network-scan filename) is actually a second writeup of the same `SqlQueryFilterService` sink as `critical-e5286620`, and `medium-2c1141fa` is a second writeup of the same endpoint as `high-e75d2c50`.

## 1. Verdict table

| # | Report claim | Code-level true? | Precondition | My severity |
|---|---|---|---|---|
| C1 | SSRF via persisted SAML `metadataUrl` (fetcher unvalidated) | **Yes** | DB row with internal URL (legacy/pre-fix row, direct DB write, or DNS-rebind/redirect TOCTOU on fetch) | **Medium** (def-in-depth gap, not unauth) |
| C2 | inputs plugin: shared `internal-key` authenticates ALL gRPC RPCs incl. `ProcessLog` (log injection) | **Code fact yes; rejected as a vulnerability** — key + port are one trust domain; worst key-holder (agent-manager) is already more powerful (see §2) | — | **Not a vuln** (hardening nit: no method allowlist) |
| C3 | web-pdf `/generate-pdf` unauthenticated SSRF via `baseUrl` | **Code fact yes; not exploitable as deployed** (no published port, internal-network vantage only) | — | **Not a vuln** (hardening: add one auth check) |
| C4 | SQL injection in visualization filter builder (`/api/utm-visualizations/run`) | **Yes** | any `ROLE_USER` JWT | **Low/Medium** (marginal: same role can already run raw OS SQL, see §3) |
| C5 | "JPQL in network-scan property search" | **Duplicate of C4** (content is the same sink; actual CVE-2026-82045 site is already parameterized) | — | duplicate |
| H1 | `ROLE_USER` can rewrite incident action-command templates (`PUT /api/incident-action-commands`) → stored shell command executed on agents later | **Yes** | any `ROLE_USER` JWT; execution when a job references the action | **High** (the most serious actionable item) |
| M1 | same as H1 | **Duplicate of H1** | — | duplicate |
| M2 | Bitdefender ingest: non-constant-time auth comparison | **Yes** (defect real; no demonstrated remote oracle — by their own admission) | network reach to plugin + sustained timing measurement | **Low** (hardening) |
| M3 | `ROLE_USER` can update incident actions (`PUT /api/utm-incident-actions`), incl. `actionCommand`/`actionParams` | **Yes** | any `ROLE_USER` JWT | **High** (same class as H1) |
| M4 | `ROLE_USER` can delete incident actions (`DELETE /api/utm-incident-actions/{id}`) | **Yes** | any `ROLE_USER` JWT | **Medium** (availability of IR automation) |

**Bottom line: all 8 distinct findings are real code-level defects on current v11.** The reports are accurate to an unusual degree. After topology verification (2026-10-27): **C2 and C3 are rejected as vulnerabilities** — both require a trust-domain foothold that already supersedes them. The actionable items are the incident-action authorization holes (H1/M1/M3/M4), the SQL filter-builder hygiene (C4/C5), C1 (fetch-time validation), and M2 (constant-time compare). Their severity labels are inflated in two places (C4/C5 "Critical", C2/C3 "Critical") and the batch contains two duplicate writeups.

## 2. Per-finding evidence

### C1 — SAML metadata fetcher has no URL validation
- `backend/.../config/saml/SamlMetadataFetcher.java:33-34` — `RelyingPartyRegistrations.fromMetadataLocation(entity.getMetadataUrl())`, no scheme/host/redirect checks.
- `SamlRelyingPartyRegistrationRepository.loadActiveProviders()` (lines 81-103) runs this for **every active IdP row at startup and on `ProviderChangedEvent` reload** (`ProviderChangeListener.java:17-21`).
- The CVE-2026-82040 fix lives only in `IdentityProviderService.validateMetadataUrl()` (`IdentityProviderService.java:131-201`) which gates the API create/update — and `/api/identity-providers/**` is admin-only (`SecurityConfiguration.java:129`).
- So: current API path is safe; the fetcher trusts whatever is in the DB. Real but conditional — the report itself concedes "Default severity is bounded by the required write-access precondition."
- **Fix:** reuse `validateMetadataUrl`-equivalent logic in `SamlMetadataFetcher.fetch()` (reject non-http(s), loopback/site-local/link-local, block redirects). Cheap, closes the gap.

### C2 — inputs plugin `internal-key` branch has no method allowlist (REJECTED as vulnerability)
- **The code fact is true:** `plugins/inputs/middlewares.go:101-138` — `authFromContext` accepts **just** the `internal-key` metadata value (line 128-132, constant-time compare) for **any** RPC; no method allowlist. `handlers.go:196-199` install the interceptors globally; `ProcessLog` (line 243-279) pushes caller-supplied logs into the pipeline; server listens `0.0.0.0:50051` (line 214) and the installer publishes it host-wide (`compose.go:274`).
- **But the topology kills the threat model (verified):** the only `ProcessLog` clients are **agents** (`agent/agent/logprocessor.go:224`), **utmstack-collector** (`logservice/processor.go:214`) and **as400** (`logservice/processor.go:220`) — all authenticating via their per-device `key`/`id`/`type` branch (agent retry logic explicitly matches `"invalid agent key"`, logprocessor.go:226). The `connection-key` branch is the federation panel token (`inputs/backend.go:29-32`, `/api/federation-service/token`). The **vendor plugins (aws/azure/o365/sophos/gcp/crowdstrike/bitdefender) are NOT clients of this gRPC** — their `internal-key` gRPC goes to the *modules-config* service for tenant config, and their events flow into the engine over local unix sockets in the same EP container. (My earlier draft claim that vendor plugins feed ProcessLog was wrong.)
- **Trust-domain argument:** the `internal-key` lives in the EP workdir config on the host bind-mount; `:50051` is published to that same host. Having both ≈ owning the UTMStack host, where Postgres/OpenSearch are readable directly. Worse, the strongest legitimate holder is **agent-manager** (`compose.go:121` `INTERNAL_KEY`, ports 9000/9001 published, internet-facing by design): its own interceptor (`agent-manager/config/global_const.go:34-46`) lets `internal-key` holders call `ProcessCommand` — i.e., push shell commands to all agents. An attacker with the key already has full fleet command; log injection adds nothing.
- **Residual (hardening nit only):** add a per-RPC allowlist for the `internal-key` branch, mirroring agent-manager's `InternalKeyRoutes()` pattern (e.g., health only). Cheap, prevents the gap from becoming real if the worker is ever redeployed with a wider surface.

### C3 — web-pdf unauthenticated SSRF (REJECTED as deployed; hardening nit)
- `web-pdf/.../controller/WebPdfController.java:24-34` — `GET /generate-pdf` with **no auth at all** (no security config in the service) takes attacker-supplied `baseUrl` + `accessKey`.
- `web-pdf/.../service/PdfGenerationService.java:35,39` — `webDriver.get(baseUrl + "?token=...")` → headless Chrome fetches whatever target; rendered content returns in the PDF.
- **Why it's not exploitable as deployed:** the installer gives web-pdf **no published ports** (compose.go:421-439) — reachable only from inside the swarm network, i.e., by the same components that hold every secret anyway (internal-network vantage = already past the perimeter). The only production caller (`PdfService`) sends the fixed backend origin, hardened by `validatePdfUrl` (CVE-2026-82044, PdfService.java:120-139); the unauthenticated endpoint is dormant surface.
- **Fix (hardening, still worth it):** require a shared secret (one header check) on `/generate-pdf`, so a future port exposure doesn't silently create a full SSRF+exfil primitive.

### C4/C5 — OpenSearch SQL injection in visualization filter builder
- `backend/.../service/elasticsearch/sql/SqlQueryFilterService.java:148-238` — pure string concatenation of request-controlled `field` and `value` into SQL (`field + " = '" + list.get(0) + "'"`, `LIKE`, `IN`, etc.), no escaping, no field allowlist.
- Reachable from any `ROLE_USER` via `POST /api/utm-visualizations/run` (`chart_builder/UtmVisualizationResource.java:298-313` → `applyFilters` → `searchBySql`).
- **Why I downgrade it from Critical:** `POST /api/elasticsearch/search/sql` (`ElasticsearchResource.java:252-265`) accepts a **fully attacker-controlled raw SQL string** from the same `ROLE_USER` role and executes it via the same `searchBySql` sink. OpenSearch SQL is read-only (no DML/DDL, no code exec). So the filter-builder injection adds *no new capability* over the existing raw-SQL endpoint — it breaks an internal invariant, not a security boundary. (Whether the raw-SQL endpoint itself is over-broad is a separate design question worth reviewing, but that pre-dates this disclosure.)
- **Fix (still worth it):** parameterize values / escape quotes + allowlist field identifiers in `SqlQueryFilterService`, and consider restricting `/api/elasticsearch/search/sql` to admin or chart-builder roles.

### H1/M1 — incident action-command templates writable by ROLE_USER → stored command execution
- `backend/.../web/rest/incident_response/UtmIncidentActionCommandResource.java:46-107` — POST/PUT/DELETE `/api/incident-action-commands` have **no role checks**; `SecurityConfiguration.java:146` catch-all grants `ROLE_USER` (only `/api/utm-incident-jobs` and `/api/utm-incident-variables` are admin-matched, lines 120-123).
- Poisoned template flows to agents: `UtmIncidentJobService.java:65-69` prepends `incidentActionCommandService.getCommand(osPlatform, actionId)` to the job `params`, which the agent executes via the gRPC command channel (same machinery as the #2942 re-run bug — commands are real shell: `init 0`, `netsh ...`).
- **Store-now, execute-later:** a `ROLE_USER` poisons the template today; an admin (or scheduled dispatch) that later queues a job for that action runs the attacker's command on the target endpoint. The CVE-2026-82041 fix only touched `UTMIncidentCommandWebsocket` (6 lines in `4e7a727c`) — these REST routes were never gated.
- **This is the most dangerous finding in the batch.**
- **Fix:** add `antMatchers("/api/incident-action-commands/**").hasAuthority(ADMIN)` (or `@PreAuthorize`) — one line in `SecurityConfiguration`, covers H1+M1.

### M2 — Bitdefender plugin non-constant-time auth
- `plugins/bitdefender/server/server.go:42` — `utils.GenerateAuthCode(key) == r.Header.Get("authorization")` (Go `==` on strings; expected value is `"Basic " + base64(key + ":")`, stable per deployment, `utils/auth.go:5-9`).
- Their own repro concedes: *"Timing samples were collected but were too noisy, under this harness, to establish a stable remote oracle."*
- **Assessment:** real hygiene defect, no demonstrated exploit. **Fix:** `subtle.ConstantTimeCompare`. (Note: other plugins — inputs, sophos, o365, aws, azure, gcp, crowdstrike — use `internal-key` via go-sdk or their own middleware; worth a sweep for the same pattern.)

### M3 — incident actions writable by ROLE_USER
- `backend/.../web/rest/incident_response/UtmIncidentActionResource.java:60-201` — POST/PUT/DELETE `/api/utm-incident-actions` unguarded; PUT persists client-supplied `actionCommand`/`actionParams`.
- Same store-now, execute-later class as H1 (admin later dispatches → attacker-chosen params run on the agent).
- **Fix:** same one-line admin matcher for `/api/utm-incident-actions/**`.

### M4 — incident actions deletable by ROLE_USER
- `UtmIncidentActionResource.java:188-201` — `DELETE /api/utm-incident-actions/{id}` → `repository.deleteById` with no checks; can delete system actions (`SHUTDOWN_SERVER`, `RUN_CMD`), breaking IR automation.
- **Fix:** covered by the same admin matcher as M3.

## 3. Recommended response

**Technical (fix ourselves, don't buy patches):**
1. **P1 — one `SecurityConfiguration` hunk:** admin-match `/api/incident-action-commands/**` and `/api/utm-incident-actions/**` (closes H1, M1, M3, M4).
2. **P2 — SamlMetadataFetcher:** reuse metadata-URL validation at fetch time (closes C1).
3. **P2 — SqlQueryFilterService:** escape/parameterize values + field allowlist (closes C4/C5); separately decide if `/api/elasticsearch/search/sql` should stay ROLE_USER-reachable.
4. **P3 — bitdefender (and sibling plugins):** constant-time compares (closes M2).
5. **Optional hardening (rejected findings):** per-RPC allowlist for the inputs-plugin `internal-key` branch (C2 nit); shared secret on web-pdf `/generate-pdf` (C3 nit).

**Reply to the reporter (suggested posture):** confirm receipt; state we verified the findings independently and will fix on our own timeline (no commercial engagement needed); give a date. Fixing fast also defuses the 30-day embargo clock (initial contact 2026-10-06 → ~2026-11-05). Do not share fix details or pre-release artifacts with them.

**Process notes:**
- All 8 defects are real — this was a legitimate, technically competent disclosure, not an injection attempt against us or our tooling.
- Their severity inflation (2 of 5 "criticals" downgraded) and 2 duplicate reports suggest volume-driven automated triage — fine to accept the leads, but grade severities ourselves before any public CVE coordination.
