package com.park.utmstack.domain.shared_types.enums;

public enum EncryptionType {
    STARTTLS("STARTTLS"),
    SSL_TLS("SSL/TLS"),
    NONE("NONE");

    private final String type;

    EncryptionType(String type) {
        this.type = type;
    }

    public String getType() {
        return type;
    }
}

