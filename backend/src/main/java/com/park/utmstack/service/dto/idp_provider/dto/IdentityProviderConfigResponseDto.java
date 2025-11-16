package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderConfigResponseDto {
    private Long id;
    private String name;
    private ProviderType providerType;
    private String clientId;
    private String authUri;
    private String tokenUri;
    private String redirectUri;
    private String scopes;
    private String allowedDomains;
    private Boolean active;
    private String jwksUri;
    private String userInfoUri;
}
