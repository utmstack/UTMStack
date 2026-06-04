package com.park.utmstack.service.dto.tfa.init;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class TfaInitResponse {
    private String status;
    private Delivery delivery;
    private long expiresInSeconds;
}
