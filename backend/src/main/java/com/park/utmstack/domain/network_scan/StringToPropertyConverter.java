package com.park.utmstack.domain.network_scan;

import com.park.utmstack.domain.network_scan.enums.PropertyFilter;
import com.park.utmstack.domain.network_scan.enums.CompositePropertyFilter;
import org.jetbrains.annotations.NotNull;
import org.springframework.core.convert.converter.Converter;
import org.springframework.stereotype.Component;

@Component
public class StringToPropertyConverter implements Converter<String, Property> {
    @Override
    public Property convert(@NotNull String source) {
        try {
            return convertToProperty(source);
        } catch (IllegalArgumentException e) {
            return null;
        }
    }

    private Property convertToProperty(String source) {
        try {
            return PropertyFilter.valueOf(source);
        } catch (IllegalArgumentException e1) {
            try {
                return CompositePropertyFilter.valueOf(source);
            } catch (IllegalArgumentException e2) {
                return null;
            }
        }
    }
}

