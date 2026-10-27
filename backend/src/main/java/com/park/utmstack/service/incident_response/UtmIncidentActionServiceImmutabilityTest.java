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
        long[] nextId = { 1000L };
        InvocationHandler h = (proxy, method, args) -> {
            switch (method.getName()) {
                case "findById":
                    return Optional.ofNullable(store.rows.get((Long) args[0]));
                case "save":
                    UtmIncidentAction a = (UtmIncidentAction) args[0];
                    if (a.getId() == null) {
                        a.setId(nextId[0]++);
                    }
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
    void bodyEditableFlagIsIgnoredWhenStoredRowIsPreset() {
        RepoStore store = new RepoStore();
        store.rows.put(1L, action(1L, false));
        UtmIncidentActionService svc = new UtmIncidentActionService(proxyRepo(store));

        UtmIncidentAction incoming = action(1L, false);
        incoming.setActionEditable(true); // attacker flips the flag in the request body
        incoming.setActionCommand("EVIL_CMD");
        ApiException ex = assertThrows(ApiException.class, () -> svc.save(incoming));
        assertEquals(HttpStatus.FORBIDDEN, ex.getStatus());
        assertEquals("SHUTDOWN_SERVER", store.rows.get(1L).getActionCommand());
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
        assertFalse(store.deleteCalled, "guard must reject before any repository write");
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
