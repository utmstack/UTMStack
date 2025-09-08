package com.park.utmstack.advice;


import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.security.TooMuchLoginAttemptsException;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.ResponseUtil;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.MDC;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

@Slf4j
@RestControllerAdvice
@RequiredArgsConstructor
public class GlobalExceptionHandler {

    private final ApplicationEventService applicationEventService;


    @ExceptionHandler(BadCredentialsException.class)
    public ResponseEntity<?> handleForbidden(BadCredentialsException e) {
        String msg = String.format("%s: %s", MDC.get("context"), e.getMessage());
        log.error(msg, e);

        return ResponseUtil.buildUnauthorizedResponse(msg);
    }

    @ExceptionHandler(TooMuchLoginAttemptsException.class)
    public ResponseEntity<?> handleTo(BadCredentialsException e) {
        String msg = String.format("%s: %s", MDC.get("context"), e.getMessage());
        log.error(msg, e);

        return ResponseUtil.buildUnauthorizedResponse(msg);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<?> handleGenericException(Exception e) {
        String msg = String.format("%s: %s", MDC.get("context"), e.getMessage());
        log.error(msg, e);

        return ResponseUtil.buildInternalServerErrorResponse(msg);
    }



}
