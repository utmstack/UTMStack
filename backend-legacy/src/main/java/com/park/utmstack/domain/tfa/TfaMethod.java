package com.park.utmstack.domain.tfa;

import java.util.Arrays;

public enum TfaMethod {
    EMAIL,
    TOTP;

    public static TfaMethod fromString(String value) {
        return Arrays.stream(values())
                .filter(m -> m.name().equalsIgnoreCase(value))
                .findFirst()
                .orElseThrow(() -> new IllegalArgumentException("Unsupported TFA method: " + value));
    }
}

