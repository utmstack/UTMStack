package com.park.utmstack.domain.idp_provider.enums;

import org.springframework.security.oauth2.core.ClientAuthenticationMethod;

public enum ClientAuthMethod {
    CLIENT_SECRET_BASIC,
    CLIENT_SECRET_POST,
    NONE,
    PRIVATE_KEY_JWT;

    public ClientAuthenticationMethod toSpringMethod() {
        return new ClientAuthenticationMethod(this.name());
    }

    public static ClientAuthMethod from(String value) {
        return ClientAuthMethod.valueOf(value.toUpperCase());
    }
}

