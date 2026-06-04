package com.park.utmstack.service.dto.tfa.verify;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.ToString;

@Getter
@Builder
@ToString
@AllArgsConstructor
public class TfaVerifyResponse {
    private final boolean valid;
    private final boolean expired;
    private final long remainingSeconds;
    private final String message;
}

