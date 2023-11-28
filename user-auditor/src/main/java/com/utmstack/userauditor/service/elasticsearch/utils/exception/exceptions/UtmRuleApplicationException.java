package com.utmstack.userauditor.service.elasticsearch.utils.exception.exceptions;

public class UtmRuleApplicationException extends Exception {
    public UtmRuleApplicationException(String message) {
        super(message);
    }

    public UtmRuleApplicationException(Throwable cause) {
        super(cause);
    }
}
