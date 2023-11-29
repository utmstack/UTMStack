package com.utmstack.userauditor.service.elasticsearch.utils.exception.exceptions;

public class UtmSerializationException extends Exception {
    public UtmSerializationException(String message) {
        super(message);
    }

    public UtmSerializationException(String message, Throwable cause) {
        super(message, cause);
    }
}
