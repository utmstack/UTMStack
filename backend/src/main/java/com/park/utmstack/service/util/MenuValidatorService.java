package com.park.utmstack.service.util;

import com.park.utmstack.domain.UtmMenu;
import org.springframework.stereotype.Service;
import org.springframework.validation.Errors;
import org.springframework.validation.Validator;

@Service
public class MenuValidatorService implements Validator {

    @Override
    public boolean supports(Class<?> clazz) {
        return UtmMenu.class.equals(clazz);
    }

    @Override
    public void validate(Object target, Errors errors) {

        UtmMenu menu = (UtmMenu) target;

        if (menu.getParentId() != null && menu.getUrl() == null) {
            errors.rejectValue("url", "urlField", "Url field  can't be null");
        }
    }
}
