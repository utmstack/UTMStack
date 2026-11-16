package com.park.utmstack.validation.saml;

import com.park.utmstack.validation.saml.impl.ValidPrivateKeyValidator;

import jakarta.validation.Constraint;
import jakarta.validation.Payload;
import java.lang.annotation.*;

@Target({ElementType.FIELD})
@Retention(RetentionPolicy.RUNTIME)
@Constraint(validatedBy = ValidPrivateKeyValidator.class)
@Documented
public @interface ValidPrivateKey {
    String message() default "The file does not contain a valid PEM private key";
    Class<?>[] groups() default {};
    Class<? extends Payload>[] payload() default {};
}
