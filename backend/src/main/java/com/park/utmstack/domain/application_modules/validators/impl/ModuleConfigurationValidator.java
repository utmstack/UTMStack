package com.park.utmstack.domain.application_modules.validators.impl;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.factory.ModuleFactory;
import com.park.utmstack.domain.application_modules.factory.IModule;
import com.park.utmstack.domain.application_modules.validators.ValidModuleConfiguration;
import com.park.utmstack.repository.application_modules.UtmModuleRepository;
import com.park.utmstack.service.dto.application_modules.GroupConfigurationDTO;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import javax.validation.ConstraintValidator;
import javax.validation.ConstraintValidatorContext;

@Component
@RequiredArgsConstructor
public class ModuleConfigurationValidator implements ConstraintValidator<ValidModuleConfiguration, GroupConfigurationDTO> {

    private final ModuleFactory moduleFactory;
    private final UtmModuleRepository moduleRepository;

    @Override
    public boolean isValid(GroupConfigurationDTO dto, ConstraintValidatorContext context) {
        if (dto.getModuleId() == null || dto.getKeys() == null || dto.getKeys().isEmpty()) {
            return false;
        }

        try {
            UtmModule utmModule = moduleRepository.findById(dto.getModuleId())
                    .orElseThrow(() -> new IllegalArgumentException("Module not found with ID: " + dto.getModuleId()));
            IModule module = moduleFactory.getInstance(utmModule.getModuleName());
            return module.validateConfiguration(utmModule, dto.getKeys());
        } catch (Exception e) {
            context.disableDefaultConstraintViolation();
            context.buildConstraintViolationWithTemplate(e.getMessage())
                    .addConstraintViolation();
            return false;
        }
    }
}
