package com.utmstack.webtopdf.exception;

import com.utmstack.webtopdf.dto.ErrorResponse;
import lombok.extern.slf4j.Slf4j;
import org.openqa.selenium.TimeoutException;
import org.openqa.selenium.NoSuchElementException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

@Slf4j
@ControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(TimeoutException.class)
    public ResponseEntity<ErrorResponse> handleTimeout(TimeoutException ex) {
        log.error("Timeout while waiting for Selenium condition: {}", ex.getMessage());

        return ResponseEntity.status(HttpStatus.REQUEST_TIMEOUT)
                .body(ErrorResponse.builder()
                        .error(true)
                        .message("The report took too long to load.")
                        .details(ex.getMessage())
                        .build());
    }

    @ExceptionHandler(NoSuchElementException.class)
    public ResponseEntity<ErrorResponse> handleNoSuchElement(NoSuchElementException ex) {
        log.error("Required element not found in Selenium: {}", ex.getMessage());

        return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                .body(ErrorResponse.builder()
                        .error(true)
                        .message("A required element was not found while generating the PDF.")
                        .details(ex.getMessage())
                        .build());
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleGeneral(Exception ex) {
        log.error("Unexpected error: {}", ex.getMessage(), ex);

        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(ErrorResponse.builder()
                        .error(true)
                        .message("An unexpected error occurred while generating the PDF.")
                        .details(ex.getMessage())
                        .build());
    }
}
