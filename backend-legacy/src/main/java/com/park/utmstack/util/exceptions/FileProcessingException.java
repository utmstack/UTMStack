package com.park.utmstack.util.exceptions;

import org.springframework.http.HttpStatus;

public class FileProcessingException extends ApiException{
    public FileProcessingException(String message) {
        super(message, HttpStatus.BAD_REQUEST);
    }
}
