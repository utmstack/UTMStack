package com.park.utmstack.service.dto.tfa.verify;

import com.park.utmstack.domain.tfa.TfaMethod;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class TfaVerifyRequest {
    private TfaMethod method;
    private String code;
}
