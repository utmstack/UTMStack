package com.park.utmstack.util.exceptions;

import org.springframework.http.HttpStatus;

public class CurrentUserLoginNotFoundException extends ApiException {
    public CurrentUserLoginNotFoundException(String message) {
        super(message, HttpStatus.NOT_FOUND);
    }
}
