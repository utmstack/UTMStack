package com.park.utmstack.service.util;

import com.park.utmstack.domain.UtmConfigurationParameter;
import org.springframework.stereotype.Service;
import org.springframework.validation.Errors;
import org.springframework.validation.Validator;

import java.util.regex.Pattern;

@Service
public class EmailValidatorService implements Validator {
    private static final String emailListRegex = "^(\\s*[\\w.-]+@[\\w.-]+\\.\\w+(\\s*,\\s*[\\w.-]+@[\\w.-]+\\.\\w+)*\\s*|\\s*[\\w.-]+@[\\w.-]+\\.\\w+\\s*)$";
    private static final Pattern pattern = Pattern.compile(emailListRegex);

    public boolean isValidEmailList(String emailList) {
        return pattern.matcher(emailList).matches();
    }

    @Override
    public boolean supports(Class<?> clazz) {
        return false;
    }

    @Override
    public void validate(Object target, Errors errors) {
        UtmConfigurationParameter utm = (UtmConfigurationParameter) target;

        if( utm.getConfParamDatatype().equals("email_list") && utm.getConfParamRegexp() != null && !Pattern.compile(utm.getConfParamRegexp()).matcher(utm.getConfParamValue()).matches()){
            errors.rejectValue("confParamValue", "customValidation.name.invalidFormat", "Invalid email address");
        }
    }
}
