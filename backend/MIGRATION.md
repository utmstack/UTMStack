# UTMStack Backend Migration — Java/Spring → Go

Tracking document for porting `backend-legacy/` (Java 11 / Spring Boot 2 / JHipster) to `backend/` (Go / Gin / GORM).

> **How to use this doc.** When you start working on a module flip its row in the matrix to 🟡 and tick endpoints as you ship them. Add the Go path under "Go location" so reviewers can jump straight to the code. Treat this file as the source of truth for migration scope — if a legacy endpoint is intentionally being dropped, mark it 🚫 with a one-line reason.

---

## Status legend

| Symbol | Meaning |
|---|---|
| ✅ | Migrated — feature parity (or close) with legacy |
| 🟡 | In progress — skeleton or partial endpoints |
| ❌ | Not started |
| 🚫 | Intentionally dropped (with reason) |
| ⚠️ | Behavior changed on purpose (re-architected for Go) |

---

## At a glance

| Metric | Legacy | Go (current) |
|---|---|---|
| Functional modules | 35+ | 14 consolidated modules (+ `internal/mail`) |
| Migration progress | — | **~85% by feature** |

Current Go modules: `iam` · `audit` · `appconfig` · `compliance` · `alerts` · `soar` · `collectors` · `eventprocessing` · `integrations` · `opensearch` · `incidents` · `notifications` · `socai` · `datasources` (+ `internal/mail`).

**Phase 1 (Foundation)** ✅ essentially complete (federation client-registry side pending). **Phase 2 (SIEM core)** ✅ complete: Alerts · Correlation+Logstash (→ **eventprocessing**) · Index Mgmt+ISM (→ **opensearch**) · OpenSearch Gateway · Collectors & Agents. (Asset Metrics 🚫 and Data Input Status 🚫 dropped — dead/derived tables.) **Phase 3 (SOAR)** ✅ complete: Incidents · Alert Response + Incident Response (→ **soar**) · SOC AI · Threat/Adversary (→ **alerts**). **Phase 4 (Compliance)** ✅ standards/sections/controls + OpenSearch evaluation/report-config/schedules (→ **compliance**). **Phase 5 partial:** Notifications ✅ · Mail sender ✅ · TFA ✅ · API keys ✅ · Integrations ✅ · Network scan → **datasources** ✅.

**Remaining:** Generic reports + PDF export · auditor users (AD audit → plugin) · app-info (future **billing**) · images CRUD (**inside appconfig**). **🚫 Not migrating:** menus · federation service · getting-started · schedules (unless needed). _(Dashboards/visualizations — previously 🚫 — are now ✅ ported to **dashboards** since there is no OpenSearch Dashboards to lean on.)_

> **Module consolidation (2026-06).** The original per-resource modules were merged into bounded contexts. Many deep-dive sections below still use the old names; the matrix and update log are the source of truth. Mappings: `correlation` + `logstash` + data-types/regex/tenant-config → **`eventprocessing`** (rules now **YAML-direct**, no DB) · `index_pattern` + `index_policy` → **`opensearch`** · `alert_response_rules` + `incident_response` → **`soar`** · `tfa` + `api_keys` → **`iam`** · `threat_management` (adversary) → **`alerts`** · `network_scan` + `datainput` → **`datasources`** (`utm_data_input_status` dropped — liveness derived from OpenSearch; `datainput` module removed).

---

## Migration matrix

> Sorted by suggested phase. Click into the deep-dive section for the per-endpoint checklist.

### Phase 1 — Foundation

| # | Module | Legacy entrypoint | Go location | Status | Notes |
|---|---|---|---|---|---|
| 1 | [Authentication & sessions](#1-authentication--sessions) | `web/rest/UserJWTController.java` | `modules/iam/handler/auth.go` | 🟡 | JWT login/refresh/logout ✅ · TFA ✅ · federation, password reset pending |
| 2 | [User management](#2-user-management) | `web/rest/UserResource.java` | `modules/iam/handler/users.go` | 🟡 | CRUD ✅ · advanced search/email validation pending |
| 3 | [Role & permission (RBAC)](#3-role--permission-rbac) | `web/rest/AuthorityResource.java` | `modules/iam/handler/roles.go` | ✅ | CRUD + `RequirePermission` middleware ✅ |
| 4 | [Account / profile](#4-account--profile) | `web/rest/AccountResource.java` | `modules/iam/handler/auth.go` | 🟡 | `me` GET/PUT, change-password ✅ · password reset pending |
| 5 | [Audit logging](#5-audit-logging) | `web/rest/AuditResource.java`, `aop/logging/AuditEvent` | `modules/audit/` | ✅ | Read endpoints ✅ · `AuditEvent` decorator + hash chain + IP/UA/session capture ✅ |
| 6 | [App configuration & secrets](#6-app-configuration--secrets) | `web/rest/UtmConfigurationParameterResource.java` | `modules/appconfig/` | 🟡 | CRUD + AES rotate ✅ · sections grouping pending |
| 7 | [Health checks](#7-health-checks) | actuator + `LogsResource.java` | `modules/health/` | 🟡 | Liveness only · DB/ES/Kafka deep checks pending |
| 8 | [Connection key (federation)](#8-connection-key-federation) | `service/federation_service/` | `modules/iam/` | ⚠️ | Connection-key validate + token issuance/rotation ✅ (internal-service auth, needed) · cross-instance **federation service** client registry 🚫 not migrating |

### Phase 2 — SIEM core

| # | Module | Legacy entrypoint | Go location | Status | Notes |
|---|---|---|---|---|---|
| 9 | [Alerts](#9-alerts) | `web/rest/UtmAlertResource.java` | `modules/alerts/` | ✅ | 22 endpoints · status/notes/tags/convert/count ✅ · alert logs CRUD ✅ · tags CRUD ✅ · tag rules CRUD ✅ · scheduler ✅ |
| 10 | [Alert tags & tag rules](#10-alert-tags--tag-rules) | `web/rest/UtmAlertTagResource.java` | `modules/alerts/` | ✅ | Folded into module #9 |
| 11 | [Alert response rules](#11-alert-response-rules) | `web/rest/alert_response_rule/` | `modules/soar/` | ✅ | Merged into **soar**. CRUD rules/templates/history/executions ✅ · evaluation engine ✅ · gRPC dispatch scheduler ✅ |
| 12 | [Correlation rules](#12-correlation-rules) | `web/rest/correlation/` | `modules/eventprocessing/` | ✅ | Merged into **eventprocessing**. Rules **YAML-direct** (no DB) ✅ · data types + sync ✅ · tenant config ✅ · regex patterns ✅ |
| 13 | [Data input ingestion status](#13-data-input-ingestion-status) | `web/rest/UtmDataInputStatusResource.java` | — | 🚫 | **Dropped.** `datainput` module + `utm_data_input_status`(+checkpoint) removed — was a materialized cache of OpenSearch `v11-statistics-*` with no consumer. Source liveness derived from OpenSearch directly. |
| 14 | [Logstash filters & pipelines](#14-logstash-filters--pipelines) | `web/rest/logstash_filter/`, `web/rest/logstash_pipeline/` | `modules/eventprocessing/` | ✅ | Merged into **eventprocessing**. filter groups CRUD ✅ · filters CRUD + audit ✅ · pipelines read/validate/delete + stats ✅ |
| 15 | [Index management (ISM)](#15-index-management-ism) | `web/rest/index_pattern/`, `web/rest/index_policy/` | `modules/opensearch/` | ✅ | Merged into **opensearch**. index pattern CRUD ✅ · ISM policy GET/PUT ✅ · registry bootstrap ✅ · snapshot repo ✅ |
| 16 | [Elasticsearch / OpenSearch gateway](#16-elasticsearch--opensearch-gateway) | `web/rest/elasticsearch/` | `modules/opensearch/` | ✅ | 11 endpoints · search/generic-search/count/CSV/SQL ✅ · property values ✅ · index list/delete ✅ · cluster status ✅ · 22-operator FilterType DSL ✅ |
| 17 | [Collectors & agents](#17-collectors--agents) | `web/rest/collectors/`, `web/rest/agent_manager/` | `modules/collectors/` | ✅ | 8 endpoints · collectors list+sync+delete ✅ · agents list/commands/by-hostname/can-run-command/update-attrs ✅ |
| 18 | [Asset metrics](#18-asset-metrics) | `web/rest/UtmAssetMetricsResource.java` | — | 🚫 | Dropped — `utm_asset_metrics` was dead (only the un-ported network_scan services wrote it). Reintroduce with network_scan if needed. |

### Phase 3 — SOAR & response

| # | Module | Legacy entrypoint | Go location | Status | Notes |
|---|---|---|---|---|---|
| 19 | [Incidents](#19-incidents) | `web/rest/incident/` | `modules/incidents/` | ✅ | 16 endpoints · incident CRUD + add-alerts + change-status ✅ · incident-alerts CRUD + bulk status ✅ · notes CRUD ✅ · history list/count/get ✅ |
| 20 | [Incident response (playbooks)](#20-incident-response-playbooks) | `web/rest/incident_response/` | `modules/incident_response/` | ✅ | 20 endpoints · actions CRUD ✅ · action-commands CRUD ✅ · jobs CRUD ✅ · variables CRUD + AES-256-GCM encryption ✅ · WebSocket command streaming (coder/websocket + gRPC bidi) ✅ |
| 21 | [SOC AI / enrichment](#21-soc-ai--enrichment) | `web/rest/soc_ai/` | `modules/socai/` | ✅ | 1 endpoint · POST /soc-ai/analyze → HTTP passthrough to SOC_AI_BASE_URL ✅ |
| 22 | [Threat management / adversaries](#22-threat-management--adversaries) | `web/rest/threat_management/` | `modules/alerts/` (adversary.go) | ✅ | Merged into **alerts**. Adversary alerts aggregation (3-level OpenSearch agg tree) ✅ |

### Phase 4 — Compliance, reporting, dashboards

| # | Module | Legacy entrypoint | Go location | Status | Notes |
|---|---|---|---|---|---|
| 23 | [Compliance standards & controls](#23-compliance-standards--controls) | `web/rest/compliance/` | `modules/compliance/` | ✅ | Standards + sections + control-config CRUD ✅ · OpenSearch control evaluation + history ✅ |
| 24 | [Compliance reports & schedules](#24-compliance-reports--schedules) | `web/rest/compliance/UtmComplianceReportSchedule*` | `modules/compliance/` | ✅ | report-config CRUD ✅ · schedules CRUD ✅ |
| 25 | [Reports (generic)](#25-reports-generic) | `web/rest/reports/`, `util/PdfGeneratorResource.java` | `modules/compliance/` (partial) | 🟡 | **Remaining** — generic section builder + PDF export. Compliance report-config/schedules landed; standalone reporting + PDF still pending. |
| 26 | [Dashboards & visualizations](#26-dashboards--visualizations) | `web/rest/chart_builder/` | `modules/dashboards/` | ✅ | Chart-builder ported: dashboards/visualizations/layout CRUD (definitions only) + audit + swagger. `utm_dashboard_authority` 🗑️ dropped. Runtime data + build UX live in the frontend (ECharts + OpenSearch gateway). |

### Phase 5 — Integrations & advanced

| # | Module | Legacy entrypoint | Go location | Status | Notes |
|---|---|---|---|---|---|
| 27 | [Notifications (in-app + email + SMS)](#27-notifications-in-app--email--sms) | `web/rest/notification/` | `modules/notifications/` | ✅ | Migrated |
| 28 | [Mail sending](#28-mail-sending) | `service/mail_sender/` | `internal/mail/` | ✅ | SMTP sender wired (used by password reset, incidents, notifications) |
| 29 | [TFA / MFA](#29-tfa--mfa) | `web/rest/tfa/` | `modules/iam/handler/tfa.go` | ✅ | TOTP + email challenges ✅ |
| 30 | [Identity providers (SAML)](#30-identity-providers-saml--done-saml2) | `web/rest/idp_provider/`, `config/saml/` | `iam` (idp + saml) | ✅ | Config CRUD + live SP-initiated SAML2 flow (login/ACS) via `crewjam/saml`. No JIT. OIDC not needed (legacy SAML2-only) |
| 31 | [API keys](#31-api-keys) | `web/rest/api_key/` | `modules/iam/handler/api_keys.go` | ✅ | Merged into **iam**. Hashed keys + auth middleware ✅ |
| 32 | [Integrations (Slack, Jira, …)](#32-integrations-slack-jira-) | `web/rest/UtmIntegrationResource.java`, `application_modules/` | `modules/integrations/` | ✅ | Integrations + `utm_module` bounded context ✅ |
| 33 | [Network scanning & assets](#33-network-scanning--assets) | `web/rest/network_scan/` | `modules/datasources/` | ⚠️ | Renamed `network_scan` → **datasources** (merge target for assets + data-input). Assets/groups + asset-sync (OpenSearch `v11-statistics-*`, no checkpoint) ✅. ⚠️ WIP: `utm_asset_types` dropped (→ free-text `label`) and `utm_ports` dropped, but Go code + `utm_network_scan` recreate path (000003) not yet reconciled. |
| 34 | [Server / module management](#34-server--module-management) | `web/rest/UtmServerResource.java`, `application_modules/` | `modules/integrations/` | ✅ | Module mgmt covered by **integrations** (`utm_module`) |
| 35 | [Schedules / dynamic tasks](#35-schedules--dynamic-tasks) | `web/rest/UtmScheduleResource.java` | — | 🚫 | Not migrating unless a concrete need appears |
| 36 | [Log analyzer / saved queries](#36-log-analyzer--saved-queries) | `web/rest/log_analyzer/` | `modules/loganalyzer/` | ✅ | Own module. Saved-query CRUD + `top-x-values` + `chart-view` (aggregations via the OpenSearch client + shared FilterType builder). Log rows served by the `opensearch` search gateway. |
| 37 | [Configuration sections / menus](#37-configuration-sections--menus) | `web/rest/UtmMenuResource.java` | — | 🚫 | Not migrating — navigation lives in the React frontend |
| 38 | [Getting started / onboarding](#38-getting-started--onboarding) | `web/rest/getting_started/` | — | 🚫 | Not migrating |
| 39 | [Auditor users (AD audit)](#39-auditor-users) | `web/rest/user_auditor/` | — | ❌ | **Remaining** — Active Directory user auditing (winlogbeat). Heavy lifting is the standalone **`user-auditor`** microservice → migrate to a **Go plugin**; backend keeps only a thin proxy (2 endpoints). NOT an iam/role feature. |
| 40 | [App info / version](#40-app-info--version) | `web/rest/app_info/AppInfoResource.java` | — | ❌ | **Remaining** — planned inside a future **billing** module |
| 41 | [Images / media](#41-images--media) | `web/rest/UtmImagesResource.java` | `uploads/` (static only) | 🟡 | Static serving exists; CRUD pending — will fold **into `appconfig`** |
| 42 | [Client / stack info](#42-client--stack-info) | `web/rest/UtmClientResource.java` | — | 🚫 | `utm_client` dropped (license read from file); tenant identity implicit |

---

## Module deep-dives

### Phase 1 — Foundation

#### 1. Authentication & sessions

- **Legacy:** `web/rest/UserJWTController.java`, `security/jwt/`, `service/login_attempts/`, `service/tfa/`, `domain/User.java`, `repository/UserRepository.java`
- **Go:** `modules/iam/handler/auth.go`, `modules/iam/usecase/auth.go`, `pkg/jwt/`, `pkg/ratelimit/`
- **Endpoints:**
  - [x] `POST /api/authenticate` → `POST /api/v1/auth/login`
  - [x] `POST /api/v1/auth/refresh` (new — refresh token rotation)
  - [x] `POST /api/v1/auth/logout`
  - [x] `GET /api/v1/auth/sessions` — list active sessions
  - [x] `DELETE /api/v1/auth/sessions/{id}` — revoke session
  - [x] `DELETE /api/v1/auth/sessions` — revoke all other sessions
  - [x] `POST /api/authenticateFederationServiceManager` — federation token login
  - [x] TFA challenge issue/verify (see [#29](#29-tfa--mfa))
- **Key features:** rate limiting (9 fails / 10min, IP-bucketed) ✅ · refresh token storage ✅ · TFA pending · federation auth pending
- **External deps:** PostgreSQL (users, refresh_tokens), `pkg/secret` (bcrypt)

---

#### 2. User management

- **Legacy:** `web/rest/UserResource.java`, `service/UserService.java`, `domain/User.java`, `domain/Authority.java`
- **Go:** `modules/iam/handler/users.go`, `modules/iam/usecase/user.go`
- **Endpoints:**
  - [x] `GET /api/v1/users` — list (paginated)
  - [x] `GET /api/v1/users/{id}` — get one
  - [x] `POST /api/v1/users` — create
  - [x] `PUT /api/v1/users/{id}` — update
  - [x] `DELETE /api/v1/users/{id}` — deactivate
  - [x] `PUT /api/v1/users/{id}/roles` — assign roles
  - [ ] Search across login/email/firstname/lastname (legacy: `?login.contains=…`)
  - [x] Email uniqueness validation on create/update
  - [ ] `POST /api/users/reset-password` (admin-triggered)
- **Notes:** Go uses numeric IDs; legacy uses `login` as path param. Frontend already adapted.

---

#### 3. Role & permission (RBAC)

- **Legacy:** `web/rest/AuthorityResource.java`, `service/AuthorityService.java`
- **Go:** `modules/iam/handler/roles.go`, `pkg/http/middleware/permission.go`
- **Endpoints:**
  - [x] `GET /api/v1/roles`
  - [x] `GET /api/v1/roles/{id}`
  - [x] `POST /api/v1/roles`
  - [x] `PUT /api/v1/roles/{id}`
  - [x] `DELETE /api/v1/roles/{id}`
  - [x] `PUT /api/v1/roles/{id}/permissions`
  - [x] `GET /api/v1/permissions`
- **Notes:** ⚠️ Re-architected — Go has explicit `permissions` table (legacy used hard-coded `@PreAuthorize("ROLE_ADMIN")` strings). The fine-grained permission model is an improvement, not a port.

---

#### 4. Account / profile

- **Legacy:** `web/rest/AccountResource.java`
- **Go:** `modules/iam/handler/auth.go` (`Me`, `UpdateMe`, `ChangePassword`, `UploadAvatar`)
- **Endpoints:**
  - [x] `GET /api/account` → `GET /api/v1/auth/me`
  - [x] `POST /api/account` → `PUT /api/v1/auth/me`
  - [x] `POST /api/account/change-password` → `POST /api/v1/auth/change-password`
  - [x] `POST /api/v1/auth/me/avatar` (new)
  - [x] `DELETE /api/v1/auth/me/avatar` (new)
  - [x] `POST /api/account/reset-password/init`
  - [x] `POST /api/account/reset-password/finish`
  - [x] `GET /api/authenticate` (idempotent auth-state check)
- [x] **Blockers:** password reset needs mail sender ([#28](#28-mail-sending)).

---

#### 5. Audit logging

- **Legacy:** `web/rest/AuditResource.java`, `service/AuditEventService.java`, `aop/logging/AuditEvent` (annotation), `domain/PersistentAuditEvent.java`
- **Go:** `modules/audit/` (handler, usecase, middleware, dto)
- **Endpoints:**
  - [x] `GET /management/audits` → `GET /api/v1/audit`
  - [x] `GET /management/audits/{id}` → `GET /api/v1/audit/{id}`
  - [x] Date-range filter (`?from=…&to=…`)
- **Critical work pending:**
  - [x] **Write path** — currently middleware logs HTTP method/path/status/duration but does **not** capture business intent (`AUTH_SUCCESS`, `ALERT_UPDATE_SUCCESS`, etc.). Decide whether to: (a) emit business events explicitly from each usecase, (b) build a Go-friendly equivalent of `@AuditEvent` (codegen / wrapper), or (c) accept the simpler middleware-only audit.
        - Path taken @AuditEvent decorator for handlers
  - [x] Capture IP + user-agent + session id on writes (in middleware)
  - [x] Hash chain integrity (legacy entries are hash-chained — preserve in Go)
- **External deps:** `audit_logs` table.

---

#### 6. App configuration & secrets

- **Legacy:** `web/rest/UtmConfigurationParameterResource.java`, `web/rest/UtmConfigurationSectionResource.java`, `service/ApplicationPropertyService.java`
- **Go:** `modules/appconfig/`, `pkg/secret/cipher.go`
- **Endpoints:**
  - [x] `GET /api/v1/config` — list all
  - [x] `GET /api/v1/config/{key}` — get one
  - [x] `PUT /api/v1/config/{key}` — upsert (encrypted secrets)
  - [x] `POST /api/v1/config/{key}/rotate` — rotate encryption
  - [x] `DELETE /api/v1/config/{key}`
  - [ ] `GET /api/utm-configuration-sections` — UI grouping metadata
  - [x] `POST /api/checkEmailConfiguration` — validate SMTP
- **Notes:** Sections endpoint is purely for legacy admin UI grouping; new React frontend organizes settings client-side, so this may be 🚫 droppable.

---

#### 7. Health checks

- **Legacy:** Spring Actuator (`/management/health`), `web/rest/LogsResource.java` (log levels)
- **Go:** `modules/health/`
- **Endpoints:**
  - [x] `GET /api/v1/health-info`
  - [x] `GET /api/v1/health` (liveness)
  - [ ] Deep checks: DB connectivity, OpenSearch, Kafka, mail
  - [ ] `GET /management/logs`, `PUT /management/logs` (runtime log level — likely 🚫)

---

#### 8. Connection key (federation)

- **Legacy:** `service/federation_service/UtmFederationServiceClientService.java`, `domain/federation_service/UtmFederationServiceClient.java`, `web/rest/federation_service/UtmFederationServiceClientResource.java`, `web/rest/UserJWTController.java#authorizeFederationServiceManager`
- **Go:** `modules/connectionkey/handler.go`, `modules/federationservice/`, `modules/iam/usecase/auth.go#FederationServiceLogin`
- **Endpoints:**
  - [x] `POST /api/v1/connection-key/validate` — validate shared secret from internal services (agent-manager, event-processor, web-pdf)
  - [x] `POST /api/v1/auth/authenticateFederationServiceManager` — federation token → access/refresh token pair
  - [x] `GET /api/v1/federation-service/token` — admin-only, audited; returns current token (404 if not initialised)
  - [x] `POST /api/v1/federation-service/generate-token` — admin-only, audited; (re)generates token and rotates `fsclient` password
  - [x] Boot-time init — if `utm_federation_service_client` is empty on startup, token is auto-generated (legacy `UtmFederationServiceClientService.init()` parity)
  - [x] Audit hooks (admin endpoints wrapped in `auditmw.AuditEvent` with `CONFIG_CHANGED`)
  - [ ] Federation client registration (cross-instance data sharing — registry side still pending)
- **Notes:** Admin gate uses new `middleware.RequireAdmin()`; JWT claims now carry `Roles`. Token format mirrors Java: `Base64("SERVER_NAME|encrypt(password, ENCRYPTION_KEY)")`. `fsclient` user is upserted via `iam.UserUsecase.EnsureFederationUser`.

---

### Phase 2 — SIEM core

#### 9. Alerts

- **Legacy:** `web/rest/UtmAlertResource.java`, `web/rest/UtmAlertLogResource.java`, `web/rest/UtmAlertTagResource.java`, `web/rest/UtmAlertTagRuleResource.java`, `service/UtmAlertService.java`, `service/UtmAlertTagRuleService.java`, `aop/logging/AlertLoggingAspect.java`
- **Go:** `modules/alerts/` (handler, usecase, repository, domain, dto, connectors, errors)
- **Endpoints:**
  - [x] `POST /api/utm-alerts/status` — bulk status update with note
  - [x] `POST /api/utm-alerts/notes` — analyst notes (raw string body + `?alertId=`)
  - [x] `POST /api/utm-alerts/tags` — replace tags on alerts
  - [x] `POST /api/utm-alerts/convert-to-incident`
  - [x] `GET /api/utm-alerts/count-open-alerts`
  - [x] `POST /api/utm-alert-logs` — create log entry
  - [x] `PUT /api/utm-alert-logs` — update log entry
  - [x] `GET /api/utm-alert-logs` — list (paginated + filters)
  - [x] `GET /api/utm-alert-logs/count`
  - [x] `GET /api/utm-alert-logs/{id}`
  - [x] `DELETE /api/utm-alert-logs/{id}`
  - [x] `POST /api/utm-alert-tags`
  - [x] `PUT /api/utm-alert-tags`
  - [x] `GET /api/utm-alert-tags` — list (paginated + filters)
  - [x] `GET /api/utm-alert-tags/{id}`
  - [x] `DELETE /api/utm-alert-tags/{id}`
  - [x] `POST /api/alert-tag-rules`
  - [x] `PUT /api/alert-tag-rules`
  - [x] `GET /api/alert-tag-rules` — list (paginated + filters)
  - [x] `GET /api/alert-tag-rules/by-ids?ids=1,2,3`
  - [x] `GET /api/alert-tag-rules/{id}`
  - [x] `DELETE /api/alert-tag-rules/{id}` — soft delete
- **Scheduler:** `UtmAlertTagRuleService.automaticReview()` ported as `usecase.Scheduler` — 30s ticker, feature-flagged via `ALERTS_SCHEDULER_ENABLED` env var (default: false)
- **AlertStatus enum:** `AUTOMATIC_REVIEW=1`, `OPEN=2`, `IN_REVIEW=3`, `COMPLETED=5` (no code 4)
- **Storage:** ⚠️ Alert documents live in **OpenSearch** (index `.alerts-*`). PG stores `utm_alert_log`, `utm_alert_tag`, `utm_alert_tag_rule`, `utm_alert_last`.
- **External deps:** OpenSearch (`pkg/opensearch/` using `opensearch-go/v4`), PostgreSQL.

---

#### 10. Alert tags & tag rules

- **Go:** Folded into module #9 (`modules/alerts/`). See above for all endpoints.
- **Status:** ✅ Complete

---

#### 11. Alert response rules

- **Legacy:** `web/rest/alert_response_rule/UtmAlertResponseRuleResource.java`, `…ExecutionResource.java`, `…HistoryResource.java`, `…ActionTemplateResource.java`, `service/UtmAlertResponseRuleService.java`
- **Go:** `modules/alert_response_rules/` · `pkg/agentmanager/` (gRPC client)
- **Endpoints:**
  - [x] `POST /api/utm-alert-response-rules` — create rule
  - [x] `PUT /api/utm-alert-response-rules` — update rule (writes history)
  - [x] `GET /api/utm-alert-response-rules` — list (paginated + filters)
  - [x] `GET /api/utm-alert-response-rules/{id}` — get one
  - [x] `GET /api/utm-alert-response-rules/resolve-filter-values` — distinct platforms + users
  - [x] `GET /api/utm-alert-response-action-templates` — list templates
  - [x] `GET /api/utm-alert-response-rule-executions` — list executions
  - [x] `GET /api/utm-alert-response-rule-histories` — list history
- **Rule evaluation engine**: pure Go, 3 operators (IS / IS_ONE_OF / IS_NOT_ONE_OF), `$(field.path)` command substitution. Triggered from alert ingestion (wiring pending — alert ingestion service not yet ported).
- **Dispatch scheduler**: 5-min ticker, dispatches PENDING executions to agents via gRPC `PanelService.ProcessCommand` (single-shot bidi stream). Retry logic: up to 5 retries on AGENT_OFFLINE, FAILED on AGENT_NOT_FOUND.
- **Tables**: `utm_alert_response_rule`, `utm_response_action_template`, `utm_alert_response_rule_execution`, `utm_alert_response_rule_history`, `utm_alert_response_rule_template` (join).
- **External deps**: PostgreSQL, agent-manager gRPC (`pkg/agentmanager/` using `google.golang.org/grpc v1.81.1`).

---

#### 12. Correlation rules

- **Legacy:** `web/rest/correlation/config/UtmCorrelationRulesResource.java`, `UtmTenantConfigResource.java`, `UtmRegexPatternResource.java`, `UtmDataTypesResource.java`
- **Go:** `modules/correlation/`
- **Endpoints:**
  - [x] `POST/PUT/GET/GET/:id/DELETE /api/correlation-rule` — correlation rules CRUD + activate/deactivate + search-by-filters + search-property-values
  - [x] `POST/PUT/GET/GET/:id/DELETE /api/data-types` + `PUT /api/data-types/include-exclude-list`
  - [x] `POST/PUT/GET/GET/:id/DELETE /api/tenant-config`
  - [x] `POST/PUT/GET/GET/:id/DELETE /api/regex-pattern`
- **Scheduler:** 60s data-type sync + YAML DefinitionSyncService (seeds system rules from `./utmstack/rules`)
- **⚠️ Known**: `search-property-values` queries `utm_correlation_rules` instead of `utm_network_scan` (TODO module-33).
- **External deps:** PostgreSQL.

---

#### 13. Data input ingestion status — 🚫 DROPPED

- **Legacy:** `web/rest/UtmDataInputStatusResource.java`, `service/UtmDataInputStatusService.java`
- **Go:** — (module removed)
- **Decision:** Not migrated. `utm_data_input_status` was a materialized cache of OpenSearch `v11-statistics-*` (latest event timestamp per source+data_type) and `utm_data_input_status_checkpoint` only held the sync poll cursor. Nothing in the Go stack read them (no frontend, no cross-module consumer; the `median`/`isDown` liveness fields were unused). The `datainput` module, its 6 CRUD endpoints, and the `network_scan` writer were removed; both tables are dropped in `000001_init.up.sql`.
- **Replacement:** Source liveness is derived from OpenSearch directly. The asset-sync in **datasources** still reads `v11-statistics-*` to reconcile assets (no checkpoint — idempotent upserts over a fixed lookback). The `dataTypes` asset-search filter that joined this table was removed (to be reintroduced against OpenSearch in the liveness redesign).

---

#### 14. Logstash filters & pipelines

- **Legacy:** `web/rest/logstash_filter/UtmFilterResource.java`, `UtmLogstashFilterGroupResource.java`, `web/rest/logstash_pipeline/UtmLogstashPipelineResource.java`
- **Go:** `modules/logstash/`
- **Endpoints:**
  - [x] `POST/PUT/GET/GET/:id/DELETE /api/utm-logstash-filter-groups` + count
  - [x] `POST/PUT/GET/GET/by-pipelineid/GET/:id/DELETE /api/utm-filters` (+ audit events)
  - [x] `GET/GET/stats/GET/:id/POST/validate/DELETE/:id /api/logstash-pipelines`
- **Scheduler:** 20s pipeline status ticker (reads `v11-statistics-*` from OpenSearch)
- **Seeder:** FilterDefinitionSyncService (seeds system filters from `./utmstack/filters`)
- **⚠️ Known**: `utm_module` table not yet ported — pipeline scheduler falls back to returning all.
- **External deps:** PostgreSQL, OpenSearch (`v11-statistics-*`).

---

#### 15. Index management (ISM)

- **Legacy:** `web/rest/index_pattern/UtmIndexPatternResource.java`, `web/rest/index_policy/IndexPolicyResource.java`, `domain/index_policy/` (~20 ISM classes)
- **Go:** `modules/indexpattern/`
- **Endpoints:**
  - [x] `POST/PUT/GET/GET/fields/GET/:id/DELETE /api/utm-index-patterns`
  - [x] `GET/PUT /api/index-policy/policy`
- **Bootstrap:** loads `utm_index_pattern` at startup → populates `IndexPatternRegistry` (LOGS=`v11-log-*`, ALERTS=`v11-alert-*`, LOGS_WINDOWS=`v11-log-wineventlog-*`). Creates ISM policy + snapshot repo in OpenSearch if not present. `os.Exit(-1)` on failure.
- **Seed:** 51 system patterns with `v11-` prefix (post Liquibase `20241227001`). `setval` to 10000.
- **External deps:** PostgreSQL, OpenSearch ISM API (`_plugins/_ism/policies`).

---

#### 16. Elasticsearch / OpenSearch gateway

- **Legacy:** `web/rest/elasticsearch/ElasticsearchResource.java`, `service/elasticsearch/ElasticsearchService.java`
- **Go:** `modules/opensearch/`
- **Endpoints:**
  - [x] `GET /api/elasticsearch/property/values` — distinct field values (terms agg)
  - [x] `POST /api/elasticsearch/property/values-with-count` — values + counts
  - [x] `GET /api/elasticsearch/index/properties` — field schema
  - [x] `GET /api/elasticsearch/index/all` — list indices (filters system indices)
  - [x] `POST /api/elasticsearch/index/delete-index` — bulk delete
  - [x] `POST /api/elasticsearch/search` — FilterType DSL search + includeChildren enrichment
  - [x] `POST /api/elasticsearch/search/csv` — same as search, streamed as CSV
  - [x] `POST /api/elasticsearch/search/sql` — SQL with 11-rule validator
  - [x] `POST /api/elasticsearch/generic-search` — search without enrichment
  - [x] `GET /api/elasticsearch/cluster/status` — cluster health
  - [x] `POST /api/elasticsearch/count` — returns Boolean (exists check, NOT numeric)
- **FilterType DSL:** 22 operators translated to OpenSearch BoolQuery (`shared/common_models/filter_type.go`).
- **⚠️ Known**: `/count` returns `Boolean` (Java quirk preserved). Cluster disk stats deferred.
- **External deps:** OpenSearch (`pkg/opensearch/` using `opensearch-go/v4`).

---

#### 17. Collectors & agents

- **Legacy:** `web/rest/collectors/UtmCollectorResource.java`, `web/rest/agent_manager/AgentManagerResource.java`, `service/agent_manager/AgentService.java`, `service/collectors/CollectorService.java`
- **Go:** `modules/collectors/` · `pkg/agentmanager/` (gRPC client extended)
- **Endpoints:**
  - [x] `GET /api/v1/collectors` — list + sync: calls gRPC `ListCollector`, upserts into `utm_collectors`, marks missing rows OFFLINE, returns gRPC page slice as `{collectors, total}`
  - [x] `DELETE /api/v1/collectors/{id}` — local DB only (gRPC delete commented out in Java — ported faithfully)
  - [x] `GET /api/v1/agent-manager/agents` — paginated list from gRPC, `agentKey` masked as `"SECRET"`
  - [x] `GET /api/v1/agent-manager/agents-with-commands` — same as agents list (command enrichment deferred to module #20)
  - [x] `GET /api/v1/agent-manager/agent-commands` — paginated commands list with nested agent (masked key)
  - [x] `GET /api/v1/agent-manager/agent-by-hostname` — first agent matching hostname
  - [x] `GET /api/v1/agent-manager/can-run-command` — returns bool: agent status == ONLINE; 500 if not found
  - [x] `POST /api/v1/agent-manager/update-agent-attrs` — updates ip + hostname via gRPC `UpdateAgent` with per-RPC metadata
- **Key decisions:**
  - `utm_collectors.id` is NOT autoincrement — PK comes from agent-manager
  - `OnConflict` on upsert uses explicit column list — `group_id` intentionally excluded to preserve user assignments
  - `last_seen` parsed as `"2006-01-02 15:04:05"` (agent-manager format) with RFC3339 fallback
  - `agentKey` masked at a single point in `agentToDTO()` — no way to leak via any endpoint
- **Deferred:** `synchronizeAgents` scheduler (commented out in Java), `agents-with-commands` full enrichment (needs module #20)
- **External deps:** PostgreSQL (`utm_collectors` table — pre-exists in production), agent-manager gRPC (`CollectorService` + `AgentService`).

---

#### 18. Asset metrics

- **Legacy:** `web/rest/UtmAssetMetricsResource.java`, `service/UtmAssetMetricsService.java`, `domain/UtmAssetMetrics.java`, `repository/UtmAssetMetricsRepository.java`
- **Go:** `modules/asset_metrics/`
- **Endpoints:**
  - [x] `POST /api/v1/utm-asset-metrics` — create (body must have id == "", returns 201)
  - [x] `PUT /api/v1/utm-asset-metrics` — update (body must have id != "")
  - [x] `GET /api/v1/utm-asset-metrics` — list all (unpaginated — matches Java `findAll()`)
  - [x] `GET /api/v1/utm-asset-metrics/{id}` — get one (404 if missing)
  - [x] `DELETE /api/v1/utm-asset-metrics/{id}` — delete
- **Key decisions:**
  - PK is caller-assigned `string` (no autoincrement, no UUID gen) — matches Java `@Id` without `@GeneratedValue`
  - UNIQUE constraint on `(asset_name, metric)` — enforced via composite GORM uniqueIndex
  - No auth on any endpoint — matches Java (no `@PreAuthorize`)
  - Frontend never calls these endpoints directly — metrics are consumed embedded in network_scan (#33) responses
  - `FindAllByAssetName` and `FindAllByAssetNameIn` exported on `Module` as reader seam for module #33
- **External deps:** PostgreSQL only.

---

### Phase 3 — SOAR & response

#### 19. Incidents

- **Legacy:** `web/rest/incident/UtmIncidentResource.java`, `…UtmIncidentAlertResource.java`, `…UtmIncidentNoteResource.java`, `…UtmIncidentHistoryResource.java`, `service/incident/UtmIncidentService.java`
- **Go:** `modules/incidents/`
- **Endpoints:**
  - [x] `POST /api/v1/utm-incidents` — create (alertList with name/severity, severity=max, conflict→409, history, fail-soft mail) → 200
  - [x] `POST /api/v1/utm-incidents/add-alerts` — add alerts to existing incident (conflict→409, history) → 201
  - [x] `PUT /api/v1/utm-incidents/change-status` — change status + cascade to alerts DB + OpenSearch sync via alerts gateway → 200
  - [x] `GET /api/v1/utm-incidents` — list with filters (name, status, assignedTo, date range, sort) + pagination
  - [x] `GET /api/v1/utm-incidents/users-assigned` — distinct assigned users `[{id, login}]` via IAM gateway
  - [x] `GET /api/v1/utm-incidents/{id}` — get one (404 if missing)
  - [x] `POST /api/v1/utm-incident-alerts` — create link → 201
  - [x] `POST /api/v1/utm-incident-alerts/update-status` — bulk status update + history
  - [x] `PUT /api/v1/utm-incident-alerts` — update link
  - [x] `GET /api/v1/utm-incident-alerts` — list (filter by incidentId, alertStatus)
  - [x] `DELETE /api/v1/utm-incident-alerts/{id}` — delete link
  - [x] `POST /api/v1/utm-incident-notes` — create note (sendDate=now, sendBy=currentUser, max 1000 chars, history) → 201
  - [x] `GET /api/v1/utm-incident-notes` — list (filter by incidentId)
  - [x] `GET /api/v1/utm-incident-histories` — list (filter by incidentId, actionType)
  - [x] `GET /api/v1/utm-incident-histories/count` — count
  - [x] `GET /api/v1/utm-incident-histories/{id}` — get one
- **Key decisions:**
  - One alert → one incident enforced at DB level (UNIQUE on `alert_id`) + batch conflict check → 409
  - `ChangeStatus` cascades to `utm_incident_alert` + syncs OpenSearch via `AlertsGateway`
  - `severity = max(alertList[].alertSeverity)` computed on create
  - `IncidentMailer` interface with noop default — wire real impl when mail module (#28) lands
  - `IAMGateway` backed by real `userRepo.List(IDs=[...])` — extended `ListUsersFilter` with `IDs []int64`
  - History stores both human label (`action`) and enum code (`action_type`)
  - Status labels: OPEN="Open", IN_REVIEW="In review", COMPLETED="Completed", MERGED="Merged"
- **Tables:** `utm_incident`, `utm_incident_alert`, `utm_incident_note`, `utm_incident_history`
- **External deps:** PostgreSQL, alerts module (OpenSearch sync via gateway), IAM module (user lookup).

---

#### 20. Incident response (playbooks)

- **Legacy:** `web/rest/incident_response/UtmIncidentActionResource.java`, `UtmIncidentActionCommandResource.java`, `UtmIncidentJobResource.java`, `UtmIncidentVariableResource.java`, `websocket/UTMIncidentCommandWebsocket.java`
- **Go:** `modules/incident_response/`
- **Endpoints:**
  - [x] `POST/PUT/GET/GET/:id/DELETE /api/v1/utm-incident-actions` — action templates CRUD
  - [x] `POST/PUT/GET/GET/:id/GET/:id/result/DELETE /api/v1/utm-incident-action-commands` — command executions CRUD + result
  - [x] `POST/GET/GET/:id/DELETE /api/v1/utm-incident-jobs` — jobs CRUD (writes incident history on create)
  - [x] `POST/PUT/GET/GET/:id/DELETE /api/v1/utm-incident-variables` — variables CRUD with AES-256-GCM encryption
  - [x] `GET /api/v1/ws/incident-command/:hostname` — WebSocket streaming (coder/websocket + gRPC bidi)
- **Key decisions:**
  - Variables: secrets encrypted with `pkg/secret/cipher.go` (AES-256-GCM, same as appconfig). Masked as `"****"` in all API responses.
  - Variable interpolation: `$[variables.NAME]` → plaintext value (or `$[NAME:<encryptedBlob>]` for secrets forwarded to agent)
  - WebSocket: raw RFC 6455 via `coder/websocket` (NOT STOMP/SockJS — frontend is new React). JWT validated before upgrade (header or `?token=`).
  - gRPC: bidi stream via `ProcessCommandStream` — context cancelled on WS disconnect → gRPC stream closed automatically
  - Secret masking in WS output: **fail-closed** — if masking fails, connection closes (never leaks raw secrets)
  - Agent ONLINE check before opening gRPC stream
  - Job create: writes `utm_incident_history` entry. TODO: OS platform lookup + command prefix (blocked on module #33)
- **Deferred:** OS-specific command prefix in `POST /utm-incident-jobs` (depends on module #33 network_scan for asset OS lookup)
- **Tables:** `utm_incident_action`, `utm_incident_action_command`, `utm_incident_job`, `utm_incident_variable`
- **External deps:** PostgreSQL, agent-manager gRPC (`ProcessCommandStream`), `pkg/secret/cipher.go`, `pkg/jwt` (WS auth), incidents module (history write).

---

#### 21. SOC AI / enrichment

- **Legacy:** `web/rest/soc_ai/UtmSocAiResource.java`, `service/soc_ai/SocAIService.java`
- **Go:** `modules/socai/`
- **Endpoints:**
  - [x] `POST /api/v1/soc-ai/analyze` — validates alert.id present (400 if missing), forwards full alert JSON to `{SOC_AI_BASE_URL}/api/v1/analyze` with `X-Internal-Key` header, returns 202 `{status, alertId, message}`
- **Key decisions:**
  - Pure HTTP passthrough — no DB, no parsing of external service response
  - TLS skip-verify (internal service)
  - `utm_alert_socai_processing_request` table NOT ported — dead code in Java (entity/repo/service with zero references)
  - Alerts-scheduler hook (module #34 dependency) deferred
- **External deps:** `SOC_AI_BASE_URL` env var (already in config), `INTERNAL_KEY` env var.

---

#### 22. Threat management / adversaries

- **Legacy:** `web/rest/threat_management/AdversaryAlertsResource.java`, `service/threat_management/AdversaryAlertsService.java`
- **Go:** `modules/threat_management/`
- **Endpoints:**
  - [x] `POST /api/v1/adversary/alerts` — 3-level OpenSearch aggregation: group by adversary.host.keyword → get Side object + parent alerts (no parentId) + child alerts (by parentId). Returns `[{adversary: Side, alerts: [{alert, children}]}]`. 204 on empty/missing index.
- **Key decisions:**
  - No DB — pure OpenSearch read module
  - `Side` struct extended to full Java + proto parity (74 new fields) in `modules/alerts/domain/alert.go`
  - `RawSearch` + `IndexExists` added to `pkg/opensearch/client.go`
  - `@Hidden` in Java not replicated — endpoint visible in Swagger (improvement for new frontend)
- **Also fixed in this batch:** `pkg/opensearch` search endpoint now correctly passes `sortBy`/`sortOrder` to repo and removed artificial `total > top` cap
- **External deps:** OpenSearch (`v11-alert-*` index pattern).

---

### Phase 4 — Compliance, reporting, dashboards

#### 23. Compliance standards & controls

- **Legacy:** `web/rest/compliance/UtmComplianceStandardResource.java`, `…StandardSectionResource.java`, `…ControlConfigResource.java`, `…ControlEvaluationHistoryResource.java`, `…ControlLatestEvaluationResource.java`, `web/rest/compliance/HipaaResource.java`, `web/rest/compliance/CustomComplianceResource.java`
- **Endpoints (target):**
  - [ ] `GET /api/compliance-standards` (+ sections)
  - [ ] CRUD `/api/compliance-control-configs`
  - [ ] `POST /api/compliance-control-configs/{id}/evaluate`
  - [ ] `GET /api/compliance-control-evaluations`
  - [ ] CRUD `/api/custom-compliances`
- **Complex:** control-to-query translation, evidence collection.

---

#### 24. Compliance reports & schedules

- **Legacy:** `web/rest/compliance/UtmComplianceReportConfigResource.java`, `UtmComplianceReportScheduleResource.java`, `service/compliance/UtmComplianceReportScheduleService.java`, `service/compliance/ComplianceMailService.java`
- **Endpoints (target):**
  - [ ] CRUD `/api/compliance-report-configs`
  - [ ] CRUD `/api/compliance-report-schedules`
  - [ ] `POST /api/compliance-reports`
  - [ ] `GET /api/compliance-reports/{id}` (PDF/HTML)

---

#### 25. Reports (generic)

- **Legacy:** `web/rest/reports/UtmReportResource.java`, `UtmReportSectionResource.java`, `CustomReportsResource.java`, `web/rest/util/PdfGeneratorResource.java`
- **Endpoints (target):**
  - [ ] CRUD `/api/utm-reports`
  - [ ] CRUD `/api/utm-report-sections`
  - [ ] `POST /api/utm-reports/{id}/generate`
  - [ ] `POST /api/pdf-generator` (generic HTML→PDF)
- **External deps:** PDF engine — legacy uses external `web-pdf` service. Likely keep it and just call out.

---

#### 26. Dashboards & visualizations → dashboards

- **Legacy:** `web/rest/chart_builder/UtmDashboardResource.java`, `UtmVisualizationResource.java`, `UtmDashboardVisualizationResource.java`, `UtmDashboardAuthorityResource.java`
- **Go:** `modules/dashboards/` (one file per entity across domain/dto/repository/usecase/handler)
- **Status:** ✅ chart-builder ported. There is no OpenSearch Dashboards deployed to lean on, so the custom chart-builder is migrated rather than dropped.
- **Endpoints:**
  - [x] CRUD `/dashboards` (+ `/:id`)
  - [x] CRUD `/visualizations` (+ `/:id`)
  - [x] CRUD `/dashboard-layouts` (viz↔dashboard placement/layout; legacy `utm-dashboard-visualizations`)
- **Tables:** `utm_dashboard`, `utm_visualization`, `utm_dashboard_visualization` (AutoMigrate). `utm_dashboard_authority` 🗑️ **dropped** — no per-role ACL; access via the `dashboards.read/write` permission.
- **Definitions only.** Tables store the *recipe* (query/aggregation/filters/chart_config/layout). The chart **data is fetched from OpenSearch at runtime**, not stored.
- **Deliberately NOT ported (decisions, not gaps):**
  - The `/run` execute-query subsystem (legacy `RequestDsl` + per-chart-type response parsers) — runtime data should be fetched by the frontend via the existing `/opensearch/search` + `/search/sql` gateway and rendered with ECharts; no need to port the parser engine.
  - import/batch endpoints; system-owner ID-range generator + prebuilt-content seeding (the "out-of-the-box" dashboards — a data seed, pending).
  - No `system_owner` edit guard yet (mirrors legacy: system content is editable).
- **Frontend note:** raw fields (query/aggregation/chart_config) are a power-user model — the new frontend should expose a *guided* builder + ship *prebuilt* content, rendering via `echarts-for-react`.

---

### Phase 5 — Integrations & advanced

#### 27. Notifications (in-app + email + SMS)

- **Legacy:** `web/rest/notification/UtmNotificationResource.java`, `service/notification/UtmNotificationService.java`
- **Endpoints (target):**
  - [ ] CRUD `/api/utm-notifications`
  - [ ] Per-user delivery preferences

---

#### 28. Mail sending

- **Legacy:** `service/mail_sender/MailService.java`, `service/mail_config/MailConfigService.java`, `config/MailSenderConfig.java`
- **Pending:**
  - [ ] SMTP client (`net/smtp` or `gomail`)
  - [ ] Template engine (text/html)
  - [ ] `POST /api/checkEmailConfiguration` — validate SMTP
  - [ ] Bulk send (password reset, alert notification, report delivery)
- **Note:** Config is already stored in `appconfig` — only the sender side is missing.

---

#### 29. TFA / MFA

- **Legacy:** `web/rest/tfa/TfaResource.java`, `web/rest/tfa/TfaEnrollmentResource.java`, `service/tfa/TfaService.java`, `config/TfaCacheConfig.java`
- **Endpoints (target):**
  - [ ] `POST /api/tfa/setup-init`
  - [ ] `POST /api/tfa/setup-verify`
  - [ ] `POST /api/tfa/setup-cancel`
  - [ ] `POST /api/tfa/disable`
  - [ ] `GET /api/tfa/status`
- **External deps:** TOTP (e.g. `pquerna/otp`), in-memory cache (Redis or local) for challenges, optional SMS gateway.

---

#### 30. Identity providers (SAML) — ✅ Done (SAML2)

- **Legacy:** `web/rest/idp_provider/IdentityProviderResource.java`, `IdentityProviderConfigResource.java`, `config/saml/OAuth2ClientConfig.java`, `SamlRelyingPartyRegistrationRepository.java`, `SamlMetadataFetcher.java`, `ProviderChangeListener.java`, `security/saml/Saml2LoginSuccessHandler.java`, `Saml2LoginFailureHandler.java`
- **Go:** inside `iam` — `domain/idp.go`, `dto/idp.go`, `repository/idp.go`, `usecase/idp.go` (config CRUD, SP key encrypted at rest), `usecase/saml.go` (live flow), `handler/idp.go` + `handler/saml.go`.
- **Endpoints:**
  - [x] Admin CRUD `/api/v1/identity-providers` (`idp.read`/`idp.write`)
  - [x] Public `GET /api/v1/idp-providers` (login-page IdP list, no auth)
  - [x] `GET /api/v1/sso/saml/:name/login` — SP-initiated AuthnRequest redirect
  - [x] `POST /api/v1/sso/saml/:name/acs` — assertion consumer (configure as `sp_acs_url`)
- **Design:** SP built **per request** from DB config (key decrypted via cipher, cert parsed from PEM, IdP metadata fetched live via `samlsp.FetchMetadata`). Response validated with `sp.ParseXMLResponse(raw, nil, sp.AcsURL)` — `currentURL` pinned to the configured public ACS URL so signature/Destination/Recipient checks pass behind nginx. NameID → `userRepo.FindByLogin` (**no JIT provisioning**; user must pre-exist + be activated) → JWT via `IssueTokenPair` → browser redirected to `/?token=…` (or `/?error=saml2`). Audited as `AUTH_*`.
- **External dep:** `github.com/crewjam/saml v0.5.1` (direct).
- **OIDC:** not implemented (legacy is SAML2-only). `coreos/go-oidc` left for a future provider type.
- **Follow-ups (phase 3, optional):** cache built SPs / IdP metadata with TTL + reload-on-config-change (currently fetched each login); `AllowIDPInitiated=true` skips strict `InResponseTo` binding — tighten with a signed request-ID cookie (SameSite=None; Secure) if strict replay protection is required; needs validation against a real IdP (Okta/Azure AD/Keycloak).

---

#### 31. API keys

- **Legacy:** `web/rest/api_key/ApiKeyResource.java`, `service/api_key/ApiKeyService.java`, `domain/api_keys/ApiKey.java`, `ApiKeyUsageLog.java`, `security/api_key/ApiKeyFilter.java`
- **Endpoints (target):**
  - [ ] `GET /api/api-keys` (paginated, per user)
  - [ ] `POST /api/api-keys` — generate
  - [ ] `DELETE /api/api-keys/{id}` — revoke
  - [ ] `GET /api/api-keys/{id}/usage` — usage stats
- **Notes:** Key auth middleware sits alongside JWT middleware — both feed into the same `current_user` context.

---

#### 32. Integrations (Slack, Jira, …)

- **Legacy:** `web/rest/UtmIntegrationResource.java`, `web/rest/UtmIntegrationConfResource.java`, `service/web_clients/WebClientService.java`
- **Endpoints (target):**
  - [ ] `GET /api/utm-integrations` (template catalog)
  - [ ] CRUD `/api/utm-integration-configs`
  - [ ] `POST /api/utm-integration-configs/{id}/test`

---

#### 33. Network scanning & assets → datasources

- **Legacy:** `web/rest/network_scan/UtmNetworkScanResource.java`, `UtmAssetGroupResource.java`, `UtmAssetTypesResource.java`, `UtmPortsResource.java`
- **Go:** `modules/datasources/` (renamed from `network_scan`; the bounded context that absorbs assets + the former data-input concern)
- **Endpoints:**
  - [x] `/api/network-scans` — assets CRUD + search/criteria/count + update-type/update-group + report
  - [x] `/api/asset-groups` — groups CRUD + search
  - [x] `/api/network-scans/probe` — probe scan/ping/check-interface
- **Scheduler:** `AssetSync` reconciles assets from OpenSearch `v11-statistics-*` over a fixed 12h lookback (checkpoint removed with `utm_data_input_status`). Feature-flagged via `NETWORK_SCAN_SCHEDULER_ENABLED`.
- **⚠️ WIP / not reconciled:**
  - `utm_asset_types` dropped "for good" in `000001_init.up.sql` (→ free-text `label` on `utm_network_scan`), but `datasources/…/asset_types.go` (domain/handler/repo/usecase + `/api/asset-types` route) and the `label` column are not done yet.
  - `utm_ports` dropped in `000001_init.up.sql`, but `datasources/…/ports.go` still models it.
  - `000001` drops `utm_network_scan` (CASCADE) expecting a slim recreate in `000003`, but `000003_*.sql` is currently missing and the asset entities aren't in AutoMigrate → no CREATE path right now.

---

#### 34. Server / module management

- **Legacy:** `web/rest/UtmServerResource.java`, `web/rest/UtmServerModuleResource.java`
- **Endpoints (target):**
  - [ ] CRUD `/api/utm-servers`
  - [ ] CRUD `/api/utm-server-modules`

---

#### 35. Schedules / dynamic tasks

- **Legacy:** `web/rest/UtmScheduleResource.java`, `service/dynamic_schedules/`
- **Endpoints (target):**
  - [ ] CRUD `/api/utm-schedules`
  - [ ] `POST /api/utm-schedules/{id}/run`
- **External deps:** Quartz → Go (`robfig/cron/v3` or similar).

---

#### 36. Log analyzer / saved queries → loganalyzer

- **Legacy:** `web/rest/log_analyzer/LogAnalyzerResource.java`, `service/log_analyzer/`, `domain/log_analyzer/LogAnalyzerQuery.java`
- **Go:** `modules/loganalyzer/` (own module — not folded into eventprocessing; it's log *exploration*, not rules)
- **What it is:** the "Discover"-style log explorer — saved searches + field top-values + a chart/timeline aggregation.
- **Endpoints:**
  - [x] CRUD `/log-analyzer/queries` (+ `/:id`) — saved searches (`utm_log_analyzer_query`: index + filters + columns)
  - [x] `POST /log-analyzer/top-x-values/{indexPattern}/{field}/{top}` — terms + value_count aggregation
  - [x] `POST /log-analyzer/chart-view` — date_histogram (when `interval` set) or terms aggregation
- **Implementation:** aggregations run via `osdk.RawSearch` + the shared `pkg/common_models.FiltersToQuery` (22-operator FilterType builder); no parser engine re-implemented. The actual **log rows** come from the `opensearch` search gateway (`/opensearch/search`), not this module. Definitions only in `utm_log_analyzer_query`; data fetched from OpenSearch at runtime.
- Permissions `loganalyzer.read/write` (seeded), swagger on all handlers, audit on the saved-query writes.

---

#### 37. Configuration sections / menus

- **Legacy:** `web/rest/UtmConfigurationSectionResource.java`, `web/rest/UtmMenuResource.java`, `web/rest/UtmMenuAuthorityResource.java`
- **Status:** ⚠️ Probably 🚫 droppable — new React frontend defines the navigation client-side. Keep only if multi-tenant role-based menu tailoring is still required server-side.

---

#### 38. Getting started / onboarding

- **Legacy:** `web/rest/getting_started/UtmGettingStartedResource.java`
- **Endpoints (target):**
  - [ ] `POST /api/getting-started/init`
  - [ ] `PUT /api/getting-started/{step}`
  - [ ] `GET /api/getting-started/status`

---

#### 39. Auditor users (Active Directory audit)

- **What it is:** auditing of **Active Directory / Windows users** — an inventory of AD accounts (name, `sid`, source/DC, attributes) built from **winlogbeat** events in OpenSearch. NOT "users with the auditor role" (the earlier note was wrong).
- **Legacy — two pieces:**
  - **`user-auditor`** (repo-root microservice, Spring Boot + own Postgres + Liquibase): connects to Elasticsearch, scans winlogbeat, and maintains users/sources/attributes/source-scans. Does the heavy lifting.
  - **`web/rest/user_auditor/UtmAuditorUsersResource.java`**: a **thin proxy** in the backend that forwards `/winlogbeat-info-by-filter` and `/utm-auditor-users-by-src` to the microservice (`ENV_AD_AUDIT_SERVICE`).
- **Migration plan (later):** port **`user-auditor` → a Go plugin** (`plugins/user-auditor/`, same pattern as `compliance-orchestrator` / the cloud pullers — standalone, reads OpenSearch, keeps its own state). Backend keeps only the thin proxy (or the frontend calls the plugin directly). So this is mostly a *plugin* migration, not backend.

---

#### 40. App info / version

- **Legacy:** `web/rest/app_info/AppInfoResource.java`
- **Endpoints (target):**
  - [ ] `GET /api/app-info` — version, build hash, edition, license
- **Note:** Frontend `AboutPage` already consumes this shape.

---

#### 41. Images / media

- **Legacy:** `web/rest/UtmImagesResource.java`, `domain/UtmImages.java`
- **Endpoints (target):**
  - [ ] `GET /api/images/all`
  - [ ] `GET /api/images/{shortName}`
  - [ ] `PUT /api/images`
  - [x] `GET /uploads/{path}` — static serve (already in Go)

---

#### 42. Client / stack info

- **Legacy:** `web/rest/UtmClientResource.java`, `web/rest/UtmStackResource.java`
- **Endpoints (target):**
  - [ ] `GET /api/utm-clients`
  - [ ] `GET /api/utm-stack`
- **Note:** Tenant identity is currently implicit; if multi-tenancy goes formal in Go, this becomes the metadata API.

---

## Cross-cutting concerns

| Concern | Legacy | Go status | Notes |
|---|---|---|---|
| **AOP / declarative audit** | `aop/logging/AuditEvent` annotation | 🟡 | Go has no AOP. Pattern chosen: explicit `HistoryRecorder` interface injected into usecases (alerts). `auditmw.AuditEvent` wrapper for HTTP-level events. |
| **Exception translation** | `web/rest/errors/ExceptionTranslator.java` (`@ControllerAdvice`) | 🟡 | `pkg/http/middleware` recovers panics; needs typed-error → HTTP code mapping. |
| **Validation** | `validation/`, `checks/` (custom validators + Bean Validation) | ❌ | Add `go-playground/validator` + per-DTO rules. |
| **Crypto** | `util/CipherUtil.java` (AES) | ✅ | `pkg/secret/cipher.go` (AES-256-GCM). |
| **Logging context** | `loggin/LogContextBuilder.java` (request IP, UA, ts) | 🟡 | `pkg/logger` is structured but doesn't enrich with request metadata yet. |
| **Spring Data Auditing** (createdBy/createdDate auto-fill) | enabled globally | ❌ | Implement via GORM hooks or middleware that injects `current_user` into context. |
| **DB migrations** | Liquibase XML changelogs | ⚠️ | Go uses raw `.sql` in `migrations/`. Plan: don't re-engineer schema — port Liquibase changesets in order. |
| **Config binding** | `@ConfigurationProperties` | ✅ | `pkg/env` reads `.env`. |
| **Cache** | Spring Cache | ❌ | Needed for TFA challenges, IdP metadata. Pick Redis or local LRU. |
| **WebSocket / STOMP** | Spring WebSocket | ❌ | Required only for incident command streaming ([#20](#20-incident-response-playbooks)). |
| **gRPC clients** | `grpc/client/CollectorServiceClient.java`, `PanelCollectorServiceClient.java` | ✅ | `pkg/agentmanager/` ✅ (PanelService.ProcessCommand + AgentService.ListAgents/UpdateAgent/DeleteAgent/ListAgentCommands + CollectorService.ListCollector). |
| **OpenAPI / Swagger** | springdoc-openapi | ✅ | `swaggo/swag` + `gin-swagger` already wired in `server.go`. |

---

## External dependencies

| Dependency | Used by | Go strategy |
|---|---|---|
| **PostgreSQL** | All modules | GORM ✅ |
| **OpenSearch / Elasticsearch** | Alerts, dashboards, log analyzer, index mgmt | `opensearch-go/v4` ✅ integrated via `pkg/opensearch/` |
| **Kafka** | Event streaming (alerts, logs) | `segmentio/kafka-go` or `confluent-kafka-go` |
| **gRPC** | Agent / collector / incident commands | `google.golang.org/grpc v1.81.1` ✅ integrated via `pkg/agentmanager/` |
| **SMTP** | Notifications, password reset, reports | `gomail.v2` / `net/smtp` |
| **SAML 2.0** | IdP login | ✅ `crewjam/saml` (in `iam`) |
| **OIDC / OAuth2** | IdP login | `coreos/go-oidc`, `oauth2` |
| **TOTP** | TFA | `pquerna/otp` |
| **PDF generation** | Reports, compliance | External `web-pdf` service (Node/Puppeteer) — keep, call via HTTP |
| **Quartz scheduler** | Schedules, compliance schedules | `robfig/cron/v3` |

---

## Gotchas worth flagging up front

1. **Domain sprawl.** The Java model has 150+ entities. Don't 1:1 port — flatten aggregates where it makes sense. E.g. `UtmIncident` + `UtmIncidentNote` + `UtmIncidentHistory` can probably live in a single Go package with three tables.
2. **`@AuditEvent` is everywhere.** Roughly every state-changing endpoint in Java has it. Decide the Go pattern *before* porting Phase 2 or you'll thrash.
3. **Alerts live in OpenSearch, not Postgres.** Don't try to "migrate the alerts table" — there isn't one. Tags/notes/links to incidents live in PG; the alert document itself is in OS.
4. **Liquibase → Go migrations.** Don't redesign the schema. Port changesets verbatim into `.sql` files. The data is production data.
5. **WebSocket + gRPC streaming for incidents.** This is the single hardest item. Start it with a design doc, not code.
6. **SAML2 dynamic registration.** Spring Security re-builds the relying-party registry whenever an IdP config changes. The Go equivalent must do the same — failing to handle this leads to "I updated the IdP and nothing took effect" bugs.
7. **TFA challenges in cache only.** No DB persistence — they live in Spring Cache for ~5 minutes. Use Redis or an in-process LRU; do not stuff them in Postgres.
8. **Federation auth.** `connection-key` already validates tokens, but the **registration** side (cross-instance client registry) is missing. Required for distributed deployments.
9. **`CustomAuditEventRepository`.** Legacy has a hash-chain audit log already (each entry references `prev_hash`). The Go `audit_logs` table doesn't yet — preserve this property.
10. **Two RBAC styles.** Legacy uses string-based `@PreAuthorize("ROLE_ADMIN")`. Go has explicit `permissions` table and `RequirePermission(...)` middleware. Don't mix the two — pick the Go model and translate legacy `ROLE_ADMIN` checks into the new permission set.

---

## Roadmap (suggested)

| Phase | Scope | Rough endpoints | Status |
|---|---|---|---|
| **1 — Foundation** | Auth, users, roles, account, audit, appconfig, health, connectionkey | ~30 | ✅ Essentially complete (federation client-registry pending) |
| **2 — SIEM core** | Alerts, response rules, correlation, ingestion status, logstash, ISM, OS gateway, collectors | ~25 | ✅ Complete (consolidated into eventprocessing/opensearch/soar; asset metrics 🚫 dropped) |
| **3 — SOAR** | Incidents + incident response + SOC AI + threat/adversary | ~15 | ✅ Complete |
| **4 — Compliance & reporting** | Compliance standards/controls/reports, generic reports (+PDF) | ~25 | 🟡 Compliance ✅ (→ **compliance**); Dashboards/visualizations ✅ chart-builder ported (→ **dashboards**); generic reports + PDF ❌ remaining |
| **5 — Integrations & advanced** | ✅ Notifications · Mail · TFA · API keys · Integrations · Network scan · Server/module mgmt (integrations) — ❌ remaining: IdP/SSO (iam), auditor users (AD audit → plugin), app-info (billing), images (appconfig) | ~40 | 🟡 Partial |

---

## Update log

<!-- Add an entry every time you flip a status or complete a row -->

- _2026-06-07_ — ♻️ **datainput eliminated · network_scan → datasources · compliance confirmed migrated.**
  - **🚫 Data Input Status (#13) dropped.** Removed the `datainput` module (6 CRUD endpoints, zero consumers) and the `network_scan` writer/gateway. `utm_data_input_status` + `utm_data_input_status_checkpoint` dropped in `000001_init.up.sql`; AutoMigrate entries removed. Asset-sync now reads OpenSearch `v11-statistics-*` over a fixed 12h lookback (no checkpoint). The `dataTypes` asset-search filter (joined that table) was removed — reintroduce against OpenSearch later.
  - **♻️ `network_scan` → `datasources`.** Module renamed/merged into the `datasources` bounded context (assets + former data-input). Wiring updated (`modules.go`, `server.go`, `main.go`); build green.
  - **🚫 `utm_asset_types` dropped "for good"** (→ free-text `label` on `utm_network_scan`) and **`utm_ports` dropped** in `000001_init.up.sql`. ⚠️ WIP: the corresponding Go code (`asset_types.go`, `ports.go`) and the `label` column / `utm_network_scan` recreate path (`000003`) are **not yet reconciled** — see deep-dive #33.
  - **✅ Compliance (#23/#24) confirmed migrated** (doc was stale): `modules/compliance/` ships standards + sections + control-config CRUD, OpenSearch control evaluation + history, report-config CRUD and report schedules.
  - **🚫 Liquibase tables** `databasechangelog` / `databasechangeloglock` dropped in `000001_init.up.sql` (Go uses golang-migrate).
  - **📋 DB tracking sheet** added at repo root `tables.md` (per-table migrated / dropped / YAML / pending status).
  - **✅ Log analyzer (#36) ported** to a new `modules/loganalyzer/` (its own bounded context, not eventprocessing). Saved-query CRUD (`utm_log_analyzer_query`) + `top-x-values` + `chart-view` aggregations (via `osdk.RawSearch` + shared `common_models.FiltersToQuery`); log rows still served by the `opensearch` gateway. Permissions `loganalyzer.read/write` seeded; swagger + audit on writes.
  - **✅ Dashboards/visualizations (#26) ported — reversal of the "🚫 not migrating" decision.** Rationale: there is no OpenSearch Dashboards deployed to lean on, so the legacy chart-builder is migrated. New `modules/dashboards/` ships CRUD for `utm_dashboard` / `utm_visualization` / `utm_dashboard_visualization` (definitions only; one file per entity; swagger + audit; `dashboards.read/write` perms). `utm_dashboard_authority` 🗑️ dropped (no per-role ACL). Runtime chart data + the build UX are frontend concerns (ECharts + the existing OpenSearch search gateway); the legacy `/run` parser engine is intentionally not ported.
- _2026-06-06_ — ♻️ **Module consolidation + scope decisions.**
  - **Consolidated** into bounded contexts: `correlation` + `logstash` (+ data-types/regex/tenant-config) → **`eventprocessing`** (rules **YAML-direct**, no DB — read/written as YAML under a shared volume; legacy `utm_correlation_rules` read once by the bootstrap then dropped); `index_pattern` + `index_policy` → **`opensearch`**; `alert_response_rules` + `incident_response` → **`soar`**; `tfa` + `api_keys` → **`iam`**; `threat_management` (adversary) → **`alerts`**; server/module mgmt → **`integrations`** (`utm_module`).
  - **Confirmed migrated** since last log: Notifications ✅ · Mail sender (`internal/mail`) ✅ · Integrations ✅ · Network scan ✅ · API keys ✅.
  - **Dropped 🚫:** Asset Metrics (dead table), `utm_client`.
  - **Not migrating 🚫:** dashboards/visualizations · menus · **federation service** (cross-instance MSSP client) · **getting-started** · **schedules** (unless a concrete need appears).
  - **Future placement decisions:** IdP/SSO → inside **iam** · Reports → inside **compliance** (when compliance lands) · App-info → inside a future **billing** module · Images → inside **appconfig** · Log analyzer → likely inside **eventprocessing** (TBD).
  - **Remaining real work:** Compliance (+ reports + PDF), Log analyzer, SSO/IdP, auditor users.
  - **Planned next (data-source liveness re-design):** drop `utm_data_input_status` (+ checkpoint); compute source status from OpenSearch `v11-statistics-*`; add a `heartbeat` topic (the **inputs plugin** emits per live agent/collector stream; cloud plugins self-report) routed through the existing **stats plugin** → green/yellow/red semaphore. gRPC keepalive on the inputs plugin makes "connected" a valid liveness signal without agent changes.
- _2026-06-02_ — ✅ **Phase 1 RBAC (#3) + Audit (#5) flipped to complete** — verified all 7 RBAC endpoints wired in `modules/iam/handler/roles.go` + `RequirePermission` middleware in `pkg/http/middleware/auth.go`. Audit module verified: read endpoints, `AuditEvent` decorator (used across iam/alerts/correlation/federation/logstash), IP/UA/session capture in `auditctx`, SHA256 hash chain with `prev_hash` linking + serialization lock in `repository/audit_log.go`.
- _2026-06-02_ — ✅ **Phase 1 Federation service issuance complete** — `modules/federationservice/` added with `GET /federation-service/token` + `POST /federation-service/generate-token` (admin-only, audited). Boot-time `EnsureInitialized` mirrors legacy `UtmFederationServiceClientService.init()` — fresh installs auto-seed the token + `fsclient` user. New `iam.UserUsecase.EnsureFederationUser` upserts the user; JWT claims gain `Roles` and `middleware.RequireAdmin()` is wired for admin-only routes. `utm_federation_service_client` added to GORM AutoMigrate.
- _2026-05-02_ — Initial inventory drafted from `backend-legacy/` analysis (35+ functional modules; ~75 legacy endpoints; ~30 Go endpoints already shipped in Phase 1).
- _2026-05-28_ — ✅ **Phase 2 OpenSearch Gateway complete** — 11 endpoints shipped across `modules/opensearch/`. Includes: search/generic-search/count/CSV/SQL endpoints, property values, index list/delete/properties, cluster status. 22-operator FilterType DSL in `shared/common_models/`. SQL validator with 11 rules. `pkg/opensearch/client.go` extended with 7 methods.
- _2026-05-28_ — ✅ **Phase 2 Index Management complete** — 8 endpoints, index pattern CRUD + ISM policy GET/PUT, OpenSearch registry bootstrap (v11- prefixed patterns), snapshot repo, `alertIndex` corrected to `v11-alert-*`.
- _2026-05-28_ — ✅ **Phase 2 Logstash Filters & Pipelines complete** — 17 endpoints shipped across `modules/logstash/`. Includes: filter groups CRUD, filters CRUD + audit events, pipelines read/validate/delete + stats, 20s OpenSearch scheduler, filter definition sync. `pkg/opensearch/` extended with `SearchWithSort`. swagger annotations added to all handlers.
- _2026-05-28_ — 🔧 **Fixes**: `autoIncrement` added to sequence-backed PKs (regex pattern, logstash filter/group domains); `utm_regex_pattern_id_seq` reset added to migration seed.
- _2026-05-27_ — ✅ **Phase 2 Data input ingestion status complete** — 6 endpoints shipped across `modules/datainput/`. Includes: CRUD + `findImportantDatasource` filtered list + JHipster criteria count. Real `DataInputStatusReader` wired into correlation 60s scheduler.
- _2026-05-27_ — ✅ **Phase 2 Correlation rules complete** — 18 endpoints shipped across `modules/correlation/`. Includes: correlation rules CRUD (JSON *_def columns, ManyToMany data types), data types CRUD + 60s sync scheduler, tenant config CRUD, regex patterns CRUD + system seed (26 patterns). New infra: YAML DefinitionSyncService seeds system-owned rules from `./utmstack/rules` on startup. ⚠️ `search-property-values` queries `utm_correlation_rules` instead of `utm_network_scan` (TODO module-33).
- _2026-05-26_ — ✅ **Phase 2 Alert response rules complete** — 8 endpoints shipped across `modules/alert_response_rules/`. Includes: CRUD rules/templates/history/executions, pure-Go evaluation engine (3 operators), gRPC dispatch scheduler (5-min, retry logic). New infra: `pkg/agentmanager/` gRPC client (`google.golang.org/grpc v1.81.1`). Legacy bugs BUG-ARR-001..003 ported faithfully — documented in `docs/alert-response-rules-legacy-bugs.md`. ⚠️ EvaluateRules not yet wired to alert ingestion — pending alert ingestion service port.
- _2026-05-26_ — ✅ **Phase 2 Alerts complete** — 22 endpoints shipped across `modules/alerts/`. Includes: 5 alert mutation endpoints, 6 alert-log CRUD endpoints, 5 alert-tag CRUD endpoints, 6 alert-tag-rule CRUD endpoints, automatic-review scheduler (feature-flagged). New infra: `pkg/opensearch/` client (`opensearch-go/v4`), `shared/common_models/filter_type.go`. Legacy bugs BUG-001..006 ported faithfully — documented in `docs/alerts-legacy-bugs.md`.
- _2026-06-02_ — ✅ **Phase 3 Threat Management complete** — 1 endpoint, pure OpenSearch aggregation. Side struct extended to full proto parity (74 new fields). RawSearch + IndexExists added to pkg/opensearch. Also fixed: opensearch search endpoint sortBy/sortOrder + total count cap removed.
- _2026-06-01_ — ✅ **Phase 3 SOC AI complete** — 1 endpoint, pure HTTP passthrough to external SOC AI service. No DB. Dead code (utm_alert_socai_processing_request) not ported.
- _2026-06-01_ — ✅ **Phase 3 Incident Response complete** — 20 REST endpoints + WebSocket streaming. 4 tables. Variable encryption AES-256-GCM (pkg/secret/cipher). WS uses coder/websocket (raw RFC 6455, not STOMP). gRPC bidi ProcessCommandStream with context-cancel on disconnect. Secret masking fail-closed. JWT auth before WS upgrade. TODO: OS command prefix in job create (blocked on module #33).
- _2026-05-31_ — ✅ **Phase 3 Incidents complete** — 16 endpoints shipped across `modules/incidents/`. 4 tables: utm_incident, utm_incident_alert, utm_incident_note, utm_incident_history. One alert→one incident enforced at DB+app level (409). ChangeStatus cascades to OpenSearch via AlertsGateway. severity=max(alerts). IAMGateway backed by real userRepo. HistoryAction stores both label and type. IncidentMailer noop (wire when #28 lands). Extended IAM ListUsersFilter with IDs.
- _2026-05-30_ — ✅ **Phase 2 Asset Metrics complete** — 5 endpoints shipped across `modules/asset_metrics/`. Pure CRUD, Postgres only. Caller-assigned string PK. Composite unique index on `(asset_name, metric)`. No auth (Java parity). Reader seam exported (`FindAllByAssetName`, `FindAllByAssetNameIn`) for module #33.
- _2026-05-30_ — ✅ **Phase 2 Collectors & Agents complete** — 8 endpoints shipped across `modules/collectors/`. Includes: collector list+sync (gRPC → upsert → mark OFFLINE) + delete, agent list/commands/by-hostname/can-run-command/update-attrs. `agentKey` masked at single point in `agentToDTO()`. `utm_collectors.id` non-autoincrement PK. `OnConflict` excludes `group_id` to preserve user assignments. `last_seen` parsed with Java format `"2006-01-02 15:04:05"`. `pkg/agentmanager/` extended with `CollectorService` + 4 new gRPC methods. Proto stubs (`collector.pb.go`, `collector_grpc.pb.go`) copied from agent-manager.
- _2026-05-02_ — ⚠️ New module **`workspace`** added (no legacy equivalent — replaces the implicit single-tenant assumption in `UtmClientResource`). Tables: `workspaces`. Endpoints: `GET /workspaces`, `GET /workspaces/{id}`, `POST /workspaces`, `PUT /workspaces/{id}`, `DELETE /workspaces/{id}`. Permissions: `workspaces.{read,write,delete}`. A "Default" workspace is seeded on first install. Memberships are **out of scope** for this iteration — workspace ownership of resources will be enforced by `workspace_id` columns on each resource as Phase 2 modules land.
