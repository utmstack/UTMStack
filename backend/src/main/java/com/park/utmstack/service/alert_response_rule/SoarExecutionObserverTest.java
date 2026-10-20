package com.park.utmstack.service.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleExecution;
import com.park.utmstack.domain.alert_response_rule.enums.RuleExecutionStatus;
import com.park.utmstack.domain.alert_response_rule.enums.RuleNonExecutionCause;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleExecutionRepository;
import com.park.utmstack.service.grpc.CommandResult;
import org.junit.jupiter.api.Test;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Proxy;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;

/**
 * issue #2942 — the SOAR command re-sent forever because the execution only
 * turned EXECUTED in the gRPC observer's onCompleted(), which the agent-manager
 * never triggers (it sends one result and keeps the stream open). These tests
 * pin the finalization contract: first result -> EXECUTED, error -> FAILED,
 * empty completion -> safety-net EXECUTED, and no double finalization.
 */
class SoarExecutionObserverTest {

    /** Captures every save(...) call on the repository, ignores the rest. */
    private static UtmAlertResponseRuleExecutionRepository proxyRepo(List<UtmAlertResponseRuleExecution> saved) {
        InvocationHandler h = (proxy, method, args) -> {
            switch (method.getName()) {
                case "save":
                    saved.add((UtmAlertResponseRuleExecution) args[0]);
                    return args[0];
                case "findAllRuleByExecutionStatusAndRule_RuleActiveTrue":
                    return List.of();
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
        return (UtmAlertResponseRuleExecutionRepository) Proxy.newProxyInstance(
                UtmAlertResponseRuleExecutionRepository.class.getClassLoader(),
                new Class<?>[]{UtmAlertResponseRuleExecutionRepository.class}, h);
    }

    private static UtmAlertResponseRuleExecution executingRow() {
        UtmAlertResponseRuleExecution cmd = new UtmAlertResponseRuleExecution();
        cmd.setId(7L);
        cmd.setAlertId("alert-7");
        cmd.setCommand("echo ok");
        cmd.setAgent("host-1");
        cmd.setExecutionStatus(RuleExecutionStatus.EXECUTING);
        return cmd;
    }

    @Test
    void firstResultFinalizesAsExecuted() {
        List<UtmAlertResponseRuleExecution> saved = new ArrayList<>();
        UtmAlertResponseRuleExecution cmd = executingRow();
        SoarExecutionObserver o = new SoarExecutionObserver("ctx", cmd, proxyRepo(saved));

        o.onNext(CommandResult.newBuilder().setResult("exit 0").build());

        assertEquals(1, saved.size(), "result must be persisted exactly once");
        assertEquals(RuleExecutionStatus.EXECUTED, cmd.getExecutionStatus());
        assertEquals("exit 0", cmd.getCommandResult());
        assertNull(cmd.getNonExecutionCause());
    }

    @Test
    void streamErrorFinalizesAsFailed() {
        List<UtmAlertResponseRuleExecution> saved = new ArrayList<>();
        UtmAlertResponseRuleExecution cmd = executingRow();
        SoarExecutionObserver o = new SoarExecutionObserver("ctx", cmd, proxyRepo(saved));

        o.onError(new RuntimeException("agent disconnected"));

        assertEquals(1, saved.size());
        assertEquals(RuleExecutionStatus.FAILED, cmd.getExecutionStatus());
        assertEquals(RuleNonExecutionCause.UNKNOWN, cmd.getNonExecutionCause());
        assertNotNull(cmd.getCommandResult());
    }

    @Test
    void completionWithoutResultFallsBackToExecuted() {
        List<UtmAlertResponseRuleExecution> saved = new ArrayList<>();
        UtmAlertResponseRuleExecution cmd = executingRow();
        SoarExecutionObserver o = new SoarExecutionObserver("ctx", cmd, proxyRepo(saved));

        o.onCompleted();

        assertEquals(1, saved.size());
        assertEquals(RuleExecutionStatus.EXECUTED, cmd.getExecutionStatus());
    }

    @Test
    void onCompletedAfterResultIsNoOp() {
        List<UtmAlertResponseRuleExecution> saved = new ArrayList<>();
        UtmAlertResponseRuleExecution cmd = executingRow();
        SoarExecutionObserver o = new SoarExecutionObserver("ctx", cmd, proxyRepo(saved));

        o.onNext(CommandResult.newBuilder().setResult("ok").build());
        o.onCompleted(); // must not save a second time

        assertEquals(1, saved.size(), "onCompleted must not re-finalize an already EXECUTED row");
        assertEquals(RuleExecutionStatus.EXECUTED, cmd.getExecutionStatus());
    }
}
