package com.park.utmstack.service.dto.tfa.enroll;

import com.park.utmstack.domain.tfa.TfaMethod;
import com.park.utmstack.domain.tfa.TfaStage;
import com.park.utmstack.service.dto.tfa.save.TfaSaveRequest;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyRequest;
import lombok.Data;

@Data
public class TfaEnrollRequest {
    private TfaStage stage;
    private TfaMethod method;
    private String code;
    private boolean enable;

    public TfaVerifyRequest toVerifyRequest() {
        return new TfaVerifyRequest(method, code);
    }
}
