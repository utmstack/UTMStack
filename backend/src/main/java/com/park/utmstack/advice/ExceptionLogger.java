package com.park.utmstack.advice;

import com.park.utmstack.loggin.LogContextBuilder;
import com.park.utmstack.util.ResponseUtil;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.logstash.logback.argument.StructuredArguments;
import org.slf4j.MDC;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;

@Slf4j
@Component
@RequiredArgsConstructor
public class ExceptionLogger {

    private final LogContextBuilder logContextBuilder;

    public ResponseEntity<?> buildResponse(Exception e, HttpServletRequest request, HttpStatus status) {
        log(e, request);
        return ResponseUtil.buildErrorResponse(status, e.getMessage());
    }

    private void log(Exception e, HttpServletRequest request) {
        String msg = String.format("%s: %s", MDC.get("context"), e.getMessage());
        log.error(msg, e, StructuredArguments.keyValue("args",logContextBuilder.buildArgs(e, request)));
    }
}

