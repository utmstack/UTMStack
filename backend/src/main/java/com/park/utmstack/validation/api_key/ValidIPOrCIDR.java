package com.park.utmstack.validation.api_key;


import javax.validation.Constraint;
import javax.validation.Payload;
import java.lang.annotation.Documented;
import java.lang.annotation.Retention;
import java.lang.annotation.Target;

import static java.lang.annotation.ElementType.*;
import static java.lang.annotation.RetentionPolicy.RUNTIME;

@Documented
@Constraint(validatedBy = ValidIPOrCIDRValidator.class)
@Target({FIELD, METHOD, PARAMETER, ANNOTATION_TYPE, TYPE_USE})
@Retention(RUNTIME)
public @interface ValidIPOrCIDR {
    String message() default "Invalid IP address or CIDR notation";

    Class<?>[] groups() default {};

    Class<? extends Payload>[] payload() default {};
}
