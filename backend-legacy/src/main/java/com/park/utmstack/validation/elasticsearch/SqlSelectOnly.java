package com.park.utmstack.validation.elasticsearch;

import javax.validation.Constraint;
import javax.validation.Payload;
import java.lang.annotation.Documented;
import java.lang.annotation.Retention;
import java.lang.annotation.Target;

import static java.lang.annotation.ElementType.*;
import static java.lang.annotation.RetentionPolicy.RUNTIME;

@Documented
@Constraint(validatedBy = SqlSelectOnlyValidator.class)
@Target({ FIELD, PARAMETER })
@Retention(RUNTIME)
public @interface SqlSelectOnly {
    String message() default "Only SELECT queries are allowed";
    Class<?>[] groups() default {};
    Class<? extends Payload>[] payload() default {};
}
