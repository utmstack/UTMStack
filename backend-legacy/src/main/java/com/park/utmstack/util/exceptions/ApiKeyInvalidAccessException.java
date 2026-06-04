package com.park.utmstack.util.exceptions;

import org.springframework.security.core.AuthenticationException;

public class ApiKeyInvalidAccessException extends AuthenticationException {
    public ApiKeyInvalidAccessException(String message) {
        super(message);
    }
}
