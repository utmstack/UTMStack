package com.park.utmstack.service.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleExecution;
import com.park.utmstack.domain.alert_response_rule.enums.RuleExecutionStatus;
import com.park.utmstack.domain.alert_response_rule.enums.RuleNonExecutionCause;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleExecutionRepository;
import com.park.utmstack.service.grpc.CommandResult;
import io.grpc.stub.StreamObserver;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Translates the gRPC result stream of a single incident-response command into
 * the terminal status of its execution record (issue #2942).
 *
 * <p>Why this is non-obvious: the agent-manager opens the stream, sends exactly
 * ONE {@link CommandResult} per command, and then loops back on {@code Recv()}
 * waiting for the next command — it never closes the stream. gRPC therefore
 * delivers {@code onNext} (and, on the server's 5-minute timeout, {@code
 * onError}) but never {@code onCompleted} while the client keeps the stream
 * open. The old code only marked the execution EXECUTED in {@code onCompleted},
 * so the record stayed PENDING and the 30-second scheduler re-sent the same
 * command forever.
 *
 * <p>So the record is finalized on the first {@code onNext} (or {@code
 * onError}), and {@code onCompleted} is kept only as a safety net for the
 * case where the stream is closed without a result.
 */
final class SoarExecutionObserver implements StreamObserver<CommandResult> {

    private static final Logger log = LoggerFactory.getLogger(SoarExecutionObserver.class);

    private final String logContext;
    private final UtmAlertResponseRuleExecution cmd;
    private final UtmAlertResponseRuleExecutionRepository repository;
    private final StringBuilder results = new StringBuilder();

    SoarExecutionObserver(String logContext,
                          UtmAlertResponseRuleExecution cmd,
                          UtmAlertResponseRuleExecutionRepository repository) {
        this.logContext = logContext;
        this.cmd = cmd;
        this.repository = repository;
    }

    @Override
    public void onNext(CommandResult commandResult) {
        // Exactly one result per command: the execution is complete here.
        results.append(commandResult.getResult());
        finalizeAs(RuleExecutionStatus.EXECUTED, results.toString(), null);
    }

    @Override
    public void onError(Throwable throwable) {
        log.error(logContext + ": command stream failed for alert {}: {}",
                cmd.getAlertId(), throwable.getMessage(), throwable);
        finalizeAs(RuleExecutionStatus.FAILED, String.valueOf(throwable.getLocalizedMessage()),
                RuleNonExecutionCause.UNKNOWN);
    }

    @Override
    public void onCompleted() {
        // Safety net only: stream closed with no result/error. If a result
        // already arrived (onNext) the status is EXECUTED and this is a no-op.
        if (cmd.getExecutionStatus() == RuleExecutionStatus.EXECUTING) {
            finalizeAs(RuleExecutionStatus.EXECUTED, results.toString(), null);
        }
    }

    private void finalizeAs(RuleExecutionStatus status, String result, RuleNonExecutionCause cause) {
        cmd.setCommandResult(result);
        cmd.setExecutionStatus(status);
        cmd.setNonExecutionCause(cause);
        repository.save(cmd);
    }
}
