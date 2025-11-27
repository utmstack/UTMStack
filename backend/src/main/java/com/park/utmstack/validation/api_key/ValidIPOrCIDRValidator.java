package com.park.utmstack.validation.api_key;


import javax.validation.ConstraintValidator;
import javax.validation.ConstraintValidatorContext;
import java.util.regex.Pattern;

public class ValidIPOrCIDRValidator implements ConstraintValidator<ValidIPOrCIDR, String> {

    private static final Pattern IPV4_PATTERN = Pattern.compile(
        "^(?:(?:25[0-5]|2[0-4]\\d|[01]?\\d?\\d)\\.){3}(?:25[0-5]|2[0-4]\\d|[01]?\\d?\\d)$"
    );

    private static final Pattern IPV4_CIDR_PATTERN = Pattern.compile(
        "^(?:(?:25[0-5]|2[0-4]\\d|[01]?\\d?\\d)\\.){3}(?:25[0-5]|2[0-4]\\d|[01]?\\d?\\d)/(\\d|[1-2]\\d|3[0-2])$"
    );
    private static final Pattern IPV6_PATTERN = Pattern.compile(
        "^(?:[\\da-fA-F]{1,4}:){7}[\\da-fA-F]{1,4}$"
    );

    private static final Pattern IPV6_CIDR_PATTERN = Pattern.compile(
        "^(?:[\\da-fA-F]{1,4}:){7}[\\da-fA-F]{1,4}/(\\d|[1-9]\\d|1[01]\\d|12[0-8])$"
    );

    @Override
    public boolean isValid(String value, ConstraintValidatorContext context) {
        // Allow null or empty values; use @NotNull/@NotEmpty to enforce non-null if needed.
        if (value == null || value.trim().isEmpty()) {
            return true;
        }
        String trimmed = value.trim();
        if (IPV4_PATTERN.matcher(trimmed).matches() || IPV4_CIDR_PATTERN.matcher(trimmed).matches()) {
            return true;
        }
        return IPV6_PATTERN.matcher(trimmed).matches() || IPV6_CIDR_PATTERN.matcher(trimmed).matches();
    }
}
