package com.utmstack.webtopdf.dto;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class ErrorResponse {
    private boolean error;
    private String message;
    private String details;
}

