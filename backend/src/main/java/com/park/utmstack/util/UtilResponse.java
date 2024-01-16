package com.park.utmstack.util;

import com.park.utmstack.web.rest.util.HeaderUtil;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;

public class UtilResponse {
    public static <T> ResponseEntity<T> buildErrorResponse(HttpStatus errorCode, String msg) {
        return ResponseEntity.status(errorCode)
            .headers(HeaderUtil.createFailureAlert("", "",
                msg.replaceAll("\n", " ").replaceAll("\r", "")))
            .body(null);
    }

    public static <T> ResponseEntity<T> buildInternalServerErrorResponse(String msg) {
        return buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
    }

    public static <T> ResponseEntity<T> buildBadRequestResponse(String msg) {
        return buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
    }

    public static <T> ResponseEntity<T> buildLockedResponse(String msg) {
        return buildErrorResponse(HttpStatus.LOCKED, msg);
    }

    public static <T> ResponseEntity<T> buildUnauthorizedResponse(String msg) {
        return buildErrorResponse(HttpStatus.UNAUTHORIZED, msg);
    }

    public static <T> ResponseEntity<T> buildPreconditionFailedResponse(String msg) {
        return buildErrorResponse(HttpStatus.PRECONDITION_FAILED, msg);
    }

    public static <T> ResponseEntity<T> buildNotFoundResponse(String msg) {
        return buildErrorResponse(HttpStatus.NOT_FOUND, msg);
    }
}

