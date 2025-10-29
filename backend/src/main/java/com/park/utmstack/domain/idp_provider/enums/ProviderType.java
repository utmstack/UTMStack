package com.park.utmstack.domain.idp_provider.enums;

import lombok.RequiredArgsConstructor;

@RequiredArgsConstructor
public enum ProviderType {
    GOOGLE("GOOGLE"),
    MICROSOFT("MICROSOFT");

    private final String type;
}
