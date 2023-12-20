package com.park.utmstack.security;

import org.springframework.security.core.AuthenticationException;

public class TooMuchLoginAttemptsException extends AuthenticationException {
    public TooMuchLoginAttemptsException(String msg, Throwable cause) {
        super(msg, cause);
    }

    public TooMuchLoginAttemptsException(String msg) {
        super(msg);
    }
}
