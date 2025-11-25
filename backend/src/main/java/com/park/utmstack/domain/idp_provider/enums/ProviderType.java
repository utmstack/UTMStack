package com.park.utmstack.domain.idp_provider.enums;

public enum ProviderType {
    GOOGLE,
    KEYCLOAK,
    MICROSOFT;

    public static ProviderType from(String value) {
        return ProviderType.valueOf(value.toUpperCase());
    }
}
