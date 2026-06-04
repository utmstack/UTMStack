package com.park.utmstack.service.dto.correlation.validators;

import com.park.utmstack.service.dto.correlation.UtmCorrelationRulesDTO;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.springframework.validation.Errors;
import org.springframework.validation.Validator;

@Component
public class CorrelationRuleValidator implements Validator {

    @Override
    public boolean supports(Class<?> clazz) {
        return UtmCorrelationRulesDTO.class.equals(clazz);
    }

    @Override
    public void validate(Object target, Errors errors) {
        UtmCorrelationRulesDTO dto = (UtmCorrelationRulesDTO) target;

        if (dto.getDataTypes() == null || dto.getDataTypes().isEmpty()) {
            errors.rejectValue("dataTypes", "DataTypesEmpty", "The rule must have at least one data type.");
        }
    }
}

