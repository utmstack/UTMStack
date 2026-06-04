package com.park.utmstack.util.exceptions;

public class OpenSearchIndexNotFoundException extends RuntimeException {
    public OpenSearchIndexNotFoundException() {
    }

    public OpenSearchIndexNotFoundException(String message) {
        super(message);
    }
}
