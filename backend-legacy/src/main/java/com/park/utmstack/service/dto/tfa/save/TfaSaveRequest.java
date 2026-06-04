package com.park.utmstack.service.dto.tfa.save;

import com.park.utmstack.domain.tfa.TfaMethod;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class TfaSaveRequest {
    private TfaMethod method;
    private boolean enable;
}
