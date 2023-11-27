package com.utmstack.userauditor.controller.utils;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;

public class UtilResponse {
    public static <T> ResponseEntity<T> buildErrorResponse(HttpStatus errorCode, String msg) {
        return ResponseEntity.status(errorCode)
            .headers(HeaderUtil.createFailureAlert("", "",
                msg.replaceAll("\n", " ").replaceAll("\r", "")))
            .body(null);
    }

}

