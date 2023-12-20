package com.park.utmstack.web.rest.errors;

import org.zalando.problem.AbstractThrowableProblem;
import org.zalando.problem.Status;

public class AgentNotfoundException extends AbstractThrowableProblem {

    private static final long serialVersionUID = 1L;

    public AgentNotfoundException() {
        super(ErrorConstants.DEFAULT_TYPE, "Agent not found", Status.NOT_FOUND);
    }
}
