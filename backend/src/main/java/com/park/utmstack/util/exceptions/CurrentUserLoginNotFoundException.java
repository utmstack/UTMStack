package com.park.utmstack.util.exceptions;

public class CurrentUserLoginNotFoundException extends RuntimeException {
    public CurrentUserLoginNotFoundException(String message) {
        super(message);
    }
}
