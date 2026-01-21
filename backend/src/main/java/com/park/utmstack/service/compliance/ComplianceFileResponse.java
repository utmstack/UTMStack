package com.park.utmstack.service.compliance;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class ComplianceFileResponse {
    // Success fields
    private byte[] pdfBytes;

    // Error fields
    private boolean error;
    private String message;
    private String details;
}

