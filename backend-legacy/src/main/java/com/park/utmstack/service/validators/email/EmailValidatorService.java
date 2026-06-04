package com.park.utmstack.service.validators.email;

import com.park.utmstack.domain.UtmConfigurationParameter;
import org.springframework.stereotype.Service;
import org.springframework.validation.Errors;
import org.springframework.validation.Validator;

import java.util.regex.Pattern;

@Service
public class EmailValidatorService implements Validator {

    @Override
    public boolean supports(Class<?> clazz) {
        return UtmConfigurationParameter.class.equals(clazz);
    }

    @Override
    public void validate(Object target, Errors errors) {

        UtmConfigurationParameter parameter = (UtmConfigurationParameter) target;
        String emailListRegex = parameter.getConfParamRegexp();

        Pattern pattern = Pattern.compile(emailListRegex);

        if(!pattern.matcher(parameter.getConfParamValue()).matches()){
            errors.rejectValue("confParamValue", "customValidation.name.invalidFormat", String.format("Invalid %s", parameter.getConfParamShort()));
        }
    }
}
