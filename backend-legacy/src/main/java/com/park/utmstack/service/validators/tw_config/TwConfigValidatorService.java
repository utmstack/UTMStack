package com.park.utmstack.service.validators.tw_config;

import com.park.utmstack.domain.UtmConfigurationParameter;
import com.park.utmstack.service.app_info.AppInfoService;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.validation.Errors;
import org.springframework.validation.Validator;

@Service
@RequiredArgsConstructor
public class TwConfigValidatorService implements Validator {

    private final AppInfoService infoService;

    @Override
    public boolean supports(Class<?> clazz) {
        return UtmConfigurationParameter.class.equals(clazz);
    }

    @Override
    public void validate(Object target, Errors errors) {
        UtmConfigurationParameter parameter = (UtmConfigurationParameter) target;

        String value = parameter.getConfParamValue();
        if (value == null) {
            errors.rejectValue("confParamValue", "null", "Value cannot be null");
            return;
        }

        String edition;
        try {
            edition = infoService.loadVersionInfo().getEdition();
        } catch (Exception e) {
            errors.reject("appInfo.error", "Could not determine application edition");
            return;
        }

        if (!Boolean.parseBoolean(value) && "community".equals(edition)) {
            errors.rejectValue("confParamValue", "forbidden", "Forbidden to disable in COMMUNITY edition");
        }
    }
}
