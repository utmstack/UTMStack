package com.park.utmstack.util.exceptions;

import org.springframework.http.HttpStatus;

public class SamlMetadataUrlInvalidException extends ApiException {
    public SamlMetadataUrlInvalidException(String message) {
        super(message, HttpStatus.BAD_REQUEST);
    }
}
