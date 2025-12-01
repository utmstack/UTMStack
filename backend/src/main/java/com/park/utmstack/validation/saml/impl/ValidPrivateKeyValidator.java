package com.park.utmstack.validation.saml.impl;

import com.park.utmstack.util.saml.PemUtils;
import com.park.utmstack.validation.saml.ValidPrivateKey;

import javax.validation.ConstraintValidator;
import javax.validation.ConstraintValidatorContext;

public class ValidPrivateKeyValidator implements ConstraintValidator<ValidPrivateKey, String> {

    @Override
    public boolean isValid(String value, ConstraintValidatorContext context) {
        if (value == null || value.isBlank()) {
            return false;
        }
        try {
            PemUtils.parsePrivateKey(value);
            return true;
        } catch (Exception e) {
            addConstraintViolation(context);
            return false;
        }
    }

    private void addConstraintViolation(ConstraintValidatorContext context) {
        context.disableDefaultConstraintViolation();
        context.buildConstraintViolationWithTemplate("The file does not contain a valid PEM private key").addConstraintViolation();
    }
}
