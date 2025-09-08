package com.park.utmstack.advice;


import com.park.utmstack.security.TooMuchLoginAttemptsException;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.ResponseUtil;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.MDC;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import javax.servlet.http.HttpServletRequest;
import java.util.NoSuchElementException;

@Slf4j
@RestControllerAdvice
@RequiredArgsConstructor
public class GlobalExceptionHandler {

    private final ApplicationEventService applicationEventService;
    private final ExceptionLogger exceptionLogger;

    @ExceptionHandler(BadCredentialsException.class)
    public ResponseEntity<?> handleForbidden(BadCredentialsException e, HttpServletRequest request) {
        return exceptionLogger.buildResponse(e, request, HttpStatus.UNAUTHORIZED);
    }

    @ExceptionHandler(TooMuchLoginAttemptsException.class)
    public ResponseEntity<?> handleTooManyLoginAttempts(TooMuchLoginAttemptsException e, HttpServletRequest request) {
        return exceptionLogger.buildResponse(e, request, HttpStatus.LOCKED);
    }

    @ExceptionHandler(NoSuchElementException.class)
    public ResponseEntity<?> handleNotFound(NoSuchElementException e, HttpServletRequest request) {
        return exceptionLogger.buildResponse(e, request, HttpStatus.NOT_FOUND);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<?> handleGenericException(Exception e, HttpServletRequest request) {
        return exceptionLogger.buildResponse(e, request, HttpStatus.INTERNAL_SERVER_ERROR);
    }
}
