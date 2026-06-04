package com.park.utmstack.service.dto.application_modules;

import lombok.Data;

import java.util.List;

@Data
public class ModuleConfigValidationErrorResponse {
    private Meta meta;
    private List<CSError> errors;
}


