package com.park.utmstack.service.dto.jwt;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;

@AllArgsConstructor
@Data
@Builder
public class LoginResponseDTO {
    private boolean success;
    private boolean tfaConfigured;
    private String method;
    private String token;
}

