package com.park.utmstack.web.rest.errors;

import org.zalando.problem.AbstractThrowableProblem;
import org.zalando.problem.Status;

public class AgentOfflineException extends AbstractThrowableProblem {

    private static final long serialVersionUID = 1L;

    public AgentOfflineException() {
        super(ErrorConstants.DEFAULT_TYPE, "Agent not found", Status.BAD_REQUEST);
    }
}
