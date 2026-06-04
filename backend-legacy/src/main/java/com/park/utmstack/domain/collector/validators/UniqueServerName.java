package com.park.utmstack.domain.collector.validators;

import javax.validation.Constraint;
import javax.validation.Payload;
import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Target({ ElementType.FIELD })
@Retention(RetentionPolicy.RUNTIME)
@Constraint(validatedBy = UniqueServerNameValidator.class)
public @interface UniqueServerName {
    String message() default "Server name must be unique.";
    Class<?>[] groups() default {};
    Class<? extends Payload>[] payload() default {};
}

