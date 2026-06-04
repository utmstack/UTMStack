package com.park.utmstack.util.exceptions;

import com.park.utmstack.aop.logging.NoLogException;
import lombok.Getter;
import lombok.Setter;
import org.springframework.http.HttpStatus;

@Getter
@Setter
@NoLogException
public class ApiException extends RuntimeException {
    private final String message;
    private final HttpStatus status;

    public ApiException(String message, HttpStatus status) {
        super(message);
        this.message = message;
        this.status = status;
    }
}

