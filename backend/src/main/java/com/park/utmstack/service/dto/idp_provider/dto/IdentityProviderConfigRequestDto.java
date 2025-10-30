package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderConfigRequestDto {

    @NotBlank
    private String name;

    @NotNull
    private ProviderType providerType;

    @NotBlank
    private String clientId;

    @NotBlank
    private String clientSecret;

    @NotBlank
    private String authUri;

    @NotBlank
    private String tokenUri;

    @NotBlank
    private String redirectUri;

    private String scopes;

    private String allowedDomains;

    private Boolean active;
}
