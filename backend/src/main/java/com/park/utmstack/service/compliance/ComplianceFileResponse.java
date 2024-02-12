package com.park.utmstack.service.compliance;

public class ComplianceFileResponse {
    byte [] pdfBytes;

    public ComplianceFileResponse(byte[] pdfBytes) {
        this.pdfBytes = pdfBytes;
    }
    public ComplianceFileResponse(){}

    public byte[] getPdfBytes() {
        return pdfBytes;
    }

    public void setPdfBytes(byte[] pdfBytes) {
        this.pdfBytes = pdfBytes;
    }
}
