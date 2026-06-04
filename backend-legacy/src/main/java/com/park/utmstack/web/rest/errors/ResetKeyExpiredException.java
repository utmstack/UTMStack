package com.park.utmstack.web.rest.errors;

import com.park.utmstack.util.exceptions.ApiException;
import org.springframework.http.HttpStatus;

public class ResetKeyExpiredException extends ApiException {

    public ResetKeyExpiredException(String message) {
        super(message, HttpStatus.BAD_REQUEST);
    }
}
