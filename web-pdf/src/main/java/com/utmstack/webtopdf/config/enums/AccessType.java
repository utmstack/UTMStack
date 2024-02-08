package com.utmstack.webtopdf.config.enums;

import lombok.Getter;

public enum AccessType {
    UTM_TOKEN("Utm_Token", "?token=", "&url=dashboard/overview"),
    UTM_INTERNAL_KEY("Utm_Internal_Key", "?key=", "");

    @Getter
    private final String type;
    private final String keyPrefix;
    private final String additionalPart;

    AccessType(String type, String keyPrefix, String additionalPart) {
        this.type = type;
        this.keyPrefix = keyPrefix;
        this.additionalPart = additionalPart;
    }

    public String buildUrlPart(String accessKey) {
        return keyPrefix + accessKey + additionalPart;
    }
}
