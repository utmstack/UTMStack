package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderConfigResponseDto {
    private Long id;
    private String name;
    private ProviderType providerType;
    private String redirectUri;
    private String clientId;
    private String clientSecret;
    private String scopes;
    private String allowedDomains;
    private Boolean active;
}
