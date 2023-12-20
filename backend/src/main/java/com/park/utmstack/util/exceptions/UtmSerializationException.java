package com.park.utmstack.util.exceptions;

public class UtmSerializationException extends Exception {
    public UtmSerializationException(String message) {
        super(message);
    }

    public UtmSerializationException(String message, Throwable cause) {
        super(message, cause);
    }
}
