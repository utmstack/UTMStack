package com.park.utmstack.service.dto.tfa.enroll;

import com.park.utmstack.domain.tfa.TfaMethod;
import com.park.utmstack.service.dto.tfa.save.TfaSaveRequest;
import lombok.Data;

@Data
public class TfaEnrollRequest {
    private TfaMethod initMethod;
    private TfaMethod verifyMethod;
    private String verifyCode;
    private TfaSaveRequest completeRequest;
}
