package com.park.utmstack.domain.application_modules.validators;

import com.park.utmstack.domain.application_modules.validators.impl.ModuleConfigurationValidator;

import jakarta.validation.Constraint;
import jakarta.validation.Payload;
import java.lang.annotation.*;

@Documented
@Constraint(validatedBy = ModuleConfigurationValidator.class)
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
public @interface ValidModuleConfiguration {
    String message() default "Invalid module configuration";
    Class<?>[] groups() default {};
    Class<? extends Payload>[] payload() default {};
}

