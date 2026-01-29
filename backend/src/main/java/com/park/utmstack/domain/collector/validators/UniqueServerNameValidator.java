package com.park.utmstack.domain.collector.validators;

import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;

import javax.validation.ConstraintValidator;
import javax.validation.ConstraintValidatorContext;
import java.util.List;
import java.util.stream.Collectors;

public class UniqueServerNameValidator implements ConstraintValidator<UniqueServerName, List<UtmModuleGroupConfiguration>> {

    @Override
    public boolean isValid(List<UtmModuleGroupConfiguration> keys, ConstraintValidatorContext context) {

        if (keys == null || keys.isEmpty()) return false;

        long duplicates = keys.stream()
                .filter(k -> "Hostname".equals(k.getConfName()))
                .collect(Collectors.groupingBy(UtmModuleGroupConfiguration::getConfValue, Collectors.counting()))
                .values().stream()
                .filter(count -> count > 1)
                .count();

        return duplicates == 0;

    }
}
