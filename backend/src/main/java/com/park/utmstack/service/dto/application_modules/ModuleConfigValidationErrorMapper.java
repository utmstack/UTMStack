package com.park.utmstack.service.dto.application_modules;

import com.fasterxml.jackson.databind.ObjectMapper;

public class ModuleConfigValidationErrorMapper {

    private static final ObjectMapper mapper = new ObjectMapper();

    public static ModuleConfigValidationErrorResponse parse(String errorText) {
        try {
            ObjectMapper mapper = new ObjectMapper();

            int start = errorText.indexOf("{\"meta\"");
            if (start == -1) return null;

            String innerJson = errorText.substring(start);

            return mapper.readValue(innerJson, ModuleConfigValidationErrorResponse.class);

        } catch (Exception e) {
            return null;
        }
    }

}
