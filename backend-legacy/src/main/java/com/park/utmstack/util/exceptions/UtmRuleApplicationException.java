package com.park.utmstack.util.exceptions;

public class UtmRuleApplicationException extends Exception {
    public UtmRuleApplicationException(String message) {
        super(message);
    }

    public UtmRuleApplicationException(Throwable cause) {
        super(cause);
    }
}
