package com.park.utmstack.advice;


import com.park.utmstack.security.TooMuchLoginAttemptsException;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.ResponseUtil;
import com.park.utmstack.util.exceptions.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
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

    @ExceptionHandler(TfaVerificationException.class)
    public ResponseEntity<?> TfaVerificationException(TfaVerificationException e, HttpServletRequest request) {
        return ResponseUtil.buildErrorResponse(HttpStatus.PRECONDITION_FAILED, e.getMessage());
    }

    @ExceptionHandler(BadCredentialsException.class)
    public ResponseEntity<?> handleForbidden(BadCredentialsException e, HttpServletRequest request) {
        return ResponseUtil.buildUnauthorizedResponse(e.getMessage());
    }

    @ExceptionHandler(TooMuchLoginAttemptsException.class)
    public ResponseEntity<?> handleTooManyLoginAttempts(TooMuchLoginAttemptsException e, HttpServletRequest request) {
        return ResponseUtil.buildLockedResponse(e.getMessage());
    }

    @ExceptionHandler({NoSuchElementException.class,
                       ApiKeyNotFoundException.class})
    public ResponseEntity<?> handleNotFound(Exception e, HttpServletRequest request) {
        return ResponseUtil.buildNotFoundResponse(e.getMessage());
    }

    @ExceptionHandler(TooManyRequestsException.class)
    public ResponseEntity<?> handleTooManyRequests(TooManyRequestsException e, HttpServletRequest request) {
        return ResponseUtil.buildErrorResponse(HttpStatus.TOO_MANY_REQUESTS, e.getMessage());
    }

    @ExceptionHandler({NoAlertsProvidedException.class})
    public ResponseEntity<?> handleNoAlertsProvided(Exception e, HttpServletRequest request) {
        return ResponseUtil.buildErrorResponse(HttpStatus.BAD_REQUEST, e.getMessage());
    }

    @ExceptionHandler({IncidentAlertConflictException.class,
                       ApiKeyExistException.class})
    public ResponseEntity<?> handleConflict(Exception e, HttpServletRequest request) {
        return ResponseUtil.buildErrorResponse(HttpStatus.CONFLICT, e.getMessage());
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<?> handleGenericException(Exception e, HttpServletRequest request) {
        return ResponseUtil.buildInternalServerErrorResponse(e.getMessage());
    }
}
