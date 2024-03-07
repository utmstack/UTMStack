package com.park.utmstack.service.util;

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

        UtmConfigurationParameter utm = (UtmConfigurationParameter) target;
        String emailListRegex = utm.getConfParamRegexp();

        Pattern pattern = Pattern.compile(emailListRegex);

        if(!pattern.matcher(utm.getConfParamValue()).matches()){
            errors.rejectValue("confParamValue", "customValidation.name.invalidFormat", "Invalid email address");
        }
    }
}
