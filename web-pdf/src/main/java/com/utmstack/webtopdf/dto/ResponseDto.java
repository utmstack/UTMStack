package com.utmstack.webtopdf.dto;

import lombok.Builder;
import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
@Builder
public class ResponseDto {
    private byte[] pdfBytes;
    private String message;
}
