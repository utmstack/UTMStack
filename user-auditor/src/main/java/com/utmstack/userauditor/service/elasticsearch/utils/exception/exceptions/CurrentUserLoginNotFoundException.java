package com.utmstack.userauditor.service.elasticsearch.utils.exception.exceptions;

public class CurrentUserLoginNotFoundException extends RuntimeException {
    public CurrentUserLoginNotFoundException(String message) {
        super(message);
    }
}
