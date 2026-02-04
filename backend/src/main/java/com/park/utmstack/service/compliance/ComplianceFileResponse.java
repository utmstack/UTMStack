package com.park.utmstack.service.compliance;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ComplianceFileResponse {
    private byte[] pdfBytes;
    private boolean error;
    private String message;
    private String details;
}


