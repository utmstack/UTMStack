package com.park.utmstack.validation.saml;

import com.park.utmstack.validation.saml.impl.ValidCertificateValidator;

import javax.validation.Constraint;
import javax.validation.Payload;
import java.lang.annotation.*;

@Target({ElementType.FIELD})
@Retention(RetentionPolicy.RUNTIME)
@Constraint(validatedBy = ValidCertificateValidator.class)
@Documented
public @interface ValidCertificate {
    String message() default "The file does not contain a valid PEM certificate";
    Class<?>[] groups() default {};
    Class<? extends Payload>[] payload() default {};
}

