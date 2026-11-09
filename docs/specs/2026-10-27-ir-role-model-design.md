# Incident-Response Role Model + System-Preset Immutability — Design

**Date:** 2026-10-27
**Branch:** `fix/ir-role-model-v11.3.1` (off `v11`)
**Release:** v11.3.1
**Origin:** Artiphishell security disclosure (email thread 2026-10-06 / 2026-10-12, Jay Vadayath) — verified 2026-10-27 in `docs/2026-10-27-artiphishell-disclosure-verification.md`. Findings H1/M1/M3/M4 (incident-response authorization) were written against a wrong assumption that these routes were meant to be admin-only. This spec corrects the role model, protects the shared system presets, and records the reclassification for the reply.

## 1. Problem

- `ROLE_USER` in UTMStack is the **security analyst** role. Analysts must be able to: read any log, update alerts/incident status, connect agents, and **author and execute incident-response actions** (existing, modified, or new). That is their job, not a privilege to be gated behind `ROLE_ADMIN`.
- `ROLE_ADMIN` is for **platform configuration**: email servers, user management, instance settings, etc.
- Current state (v11 head):
  - `/api/utm-incident-jobs` and `/api/utm-incident-jobs/**` → `hasAuthority(ADMIN)` only — **wrong** for the model; blocks analysts from dispatching IR jobs.
  - `UTMIncidentCommandWebsocket.processCommand` → explicit `isCurrentUserInRole(ADMIN)` check — **wrong** for the same reason.
  - `/api/utm-incident-actions/**` and `/api/incident-action-commands/**` → resolve to `ROLE_USER` via the `/api/**` catch-all (`SecurityConfiguration:146`) — **correct**, no change needed.
  - The seeded system IR actions (`SHUTDOWN_SERVER`, `DISABLE_USER`, `BLOCK_IP`, `ISOLATE_HOST`, `RESTART_SERVER`, `KILL_PROCESS`, `UNINSTALL_PROGRAM`, `RUN_CMD`) are flagged `actionEditable=false` in the seed data (`data.sql:1087-1094`) and their edit/delete UI buttons are commented out (`incident-response-command.component.html:31-33,76`). **The flag is never enforced in the backend** — any authenticated user can `PUT`/`DELETE` a preset, and can rewrite the shell strings in `utm_incident_action_command` (e.g. `init 0`, `cmd.exe /c shutdown -s -t 0 -f`) that `UtmIncidentJobService.save` prepends to dispatched agent jobs (`UtmIncidentJobService.java:65-69`).

## 2. Goals

1. Analysts (`ROLE_USER`) can execute IR jobs and console commands; admins keep the same rights.
2. Seeded system presets (`actionEditable=false`) become **immutable via API for everyone** — no admin override. Repairs to presets are reseed/liquibase changes, not API calls.
3. Analyst-authored actions (`actionEditable=true`) remain freely creatable, editable, and deletable by any authenticated user.
4. The disclosure is answered with a documented reclassification, not a silent accept/reject.

## 3. Non-goals

- No new roles, no per-object ownership, no approved-command catalog.
- No frontend changes (preset edit/delete is already absent from the UI; the 403 is defense-in-depth behind it).
- No change to the alert-response automation (`UtmAlertResponseRuleService` / `utm-alert-response-rules` — already admin-gated, separate feature).
- No changes for C1/C2/C3/C4/M2 — tracked separately (see §6).

## 4. Changes

### 4.1 Role opening (DONE — already edited on the branch)

- `backend/src/main/java/com/park/utmstack/config/SecurityConfiguration.java:120-121`
  - `/api/utm-incident-jobs` and `/api/utm-incident-jobs/**`: `hasAuthority(ADMIN)` → `hasAnyAuthority(ADMIN, USER)`.
- `backend/src/main/java/com/park/utmstack/web/rest/incident_response/UTMIncidentCommandWebsocket.java:52`
  - Explicit check → `ADMIN || USER`; denial message updated ("requires an authenticated user").

### 4.2 Preset immutability guards (service layer)

Guard in the **service** layer (not route layer) so it holds for every caller. Grep-verified: the only callers of these service `save`/`delete` methods are the two REST resources — no internal flow is affected.

- **`UtmIncidentActionService`** (`service/incident_response/UtmIncidentActionService.java`):
  - `save(entity)`: if `entity.getId() != null` (update path), load existing row; if existing `actionEditable == false` (or `null`, fail closed) → throw `ApiException("System incident action <id> is not editable", HttpStatus.FORBIDDEN)`. Create path untouched.
  - `delete(id)`: same check before `deleteById`.
- **`UtmIncidentActionCommandService`** (`service/incident_response/UtmIncidentActionCommandService.java`):
  - `save(entity)`: on update path, load the parent action via `UtmIncidentActionRepository.findById(entity.getActionId())`; if parent not editable → `ApiException(403)`. On create path (`id == null`), enforce parent-exists **and** parent-editable — command rows are per-(action, osPlatform), so a "new" command row always belongs to an existing action; presets' command rows are therefore locked too. (This protects the actual shell strings under `SHUTDOWN_SERVER`/`RUN_CMD`.)
  - `delete(id)`: load the command row → parent action → same check.
  - Constructor gains `UtmIncidentActionRepository` (already a Spring bean; no schema change).
- **Exception propagation:** both resources wrap handlers in `catch (Exception e) → 500`. Add the house pattern (same as `ElasticsearchResource.java:195-198`, from #2909): `catch (ApiException e) { throw e; }` **before** the generic catch, in every handler of both resources that touches these service methods, so `GlobalExceptionHandler`'s `@ExceptionHandler(ApiException.class)` (`advice/GlobalExceptionHandler.java:69`) returns the 403.

### 4.3 Tests (backend, tests live in `src/main/java` per repo convention)

- **`UtmIncidentActionServiceImmutabilityTest`** — JDK Proxy repository stubs (repo has no Mockito; pattern from #2942 `SoarExecutionObserver` tests; surefire is pointed at the main output dir, so `mvn test` picks these up):
  - update on `actionEditable=false` → `ApiException` 403; update on `true` → saved.
  - delete on `false` → 403; delete on `true` → deleted.
  - create (null id) → saved regardless of flag.
  - existing row with `actionEditable=null` → 403 (fail closed).
- **`UtmIncidentActionCommandServiceImmutabilityTest`**:
  - update of a command row whose parent action is a preset → 403; parent editable → saved.
  - create under a preset parent → 403; under an editable parent → saved; unknown parent → 403.
  - delete under preset parent → 403.

### 4.4 Verification doc + disclosure reply

- Update `docs/2026-10-27-artiphishell-disclosure-verification.md`:
  - H1/M1/M3/M4 → **reclassified: by-design role behavior**, not missing authorization. The residual defect (unprotected system presets) is fixed in this spec (§4.2).
  - C4/C5 → **code hygiene** (string concat into OpenSearch SQL; no tenant boundary exists because analysts read all logs by design). Backlog item, not a disclosure acceptance.
  - C2/C3 → rejected (trust-domain preconditions), as already documented.
  - C1 → accepted as minor hardening (fetch-time metadata-URL validation in `SamlMetadataFetcher`).
  - M2 → optional hygiene (constant-time compare).
- Draft reply to the reporter: receipt confirmed, findings independently verified, IR findings reclassified per the product role model with preset immutability shipped as a product fix, remaining items folded into our own fix cycle on our timeline, no commercial engagement.

## 5. Data model

No schema change. `utm_incident_actions.action_editable` (nullable=false) is the single authority. `utm_incident_action_command` inherits protection through `action_id`.

## 6. Sequencing relative to other v11.3.x items

- This change is self-contained: backend only, no DB migration, no frontend rebuild.
- C1 (SAML fetch-time validation), C4/C5 (SQL filter escaping), M2 (constant-time) are **separate work items** — same release train (v11.3.1) or later, per their own review/PRs. This spec does not include them.
- v11.3.0 remains the pre-release string in flight; per the standing decision, fixes land on the release line before it is scheduled to customers, so v11.3.1 is a fresh version string cut from this branch's merged state.

## 7. Risk notes

- Opening IR dispatch to `ROLE_USER` broadens a powerful primitive (console shell to any agent in the instance). Accepted explicitly per product decision: that is the analyst role.
- Preset immutability has no override: a genuinely broken preset requires a reseed change. Accepted explicitly (user choice: "Enforce immutability").
- The `catch (ApiException) rethrow` additions touch all four mutating handlers in both resources; missing one handler means a 403 arrives as a 500 — the same-class error a previous dashboard bug produced. Each handler is listed in the plan as its own checklist item.
