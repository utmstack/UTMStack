package com.park.utmstack.service.dto.tfa.save;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class TfaSaveResponse {
    private String status;
    private String message;
}
