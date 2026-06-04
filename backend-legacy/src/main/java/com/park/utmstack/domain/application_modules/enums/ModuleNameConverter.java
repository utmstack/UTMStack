package com.park.utmstack.domain.application_modules.enums;

import javax.persistence.AttributeConverter;
import javax.persistence.Converter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

@Converter(autoApply = true)
public class ModuleNameConverter implements AttributeConverter<ModuleName, String> {
    private static final Logger log = LoggerFactory.getLogger(ModuleNameConverter.class);

    @Override
    public String convertToDatabaseColumn(ModuleName attribute) {
        return attribute == null ? null : attribute.name();
    }

    @Override
    public ModuleName convertToEntityAttribute(String dbData) {
        if (dbData == null) {
            return null;
        }
        try {
            return ModuleName.valueOf(dbData);
        } catch (IllegalArgumentException e) {
            log.warn("Unknown ModuleName in database: '{}'. Mapping to null to prevent crash. ---", dbData);
            return null;
        }
    }
}
