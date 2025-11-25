package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

/**
 * DTO for Identity Provider configuration requests.
 * Adapted for SAML providers only.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderConfigRequestDto {

    private Long id;

    @NotBlank
    private String name;

    @NotNull
    private ProviderType providerType; // KEYCLOAK, GOOGLE, OKTA, etc.

    @NotBlank
    private String entityId; // IdP entityID from metadata

    @NotBlank
    private String ssoUrl; // SingleSignOnService URL

    private String sloUrl; // SingleLogoutService URL (optional)

    @NotBlank
    private String certPem; // PEM formatted certificate

    private String nameIdFormat; // e.g. emailAddress

    private String binding; // HTTP-POST, Redirect, etc.

    @NotNull
    private Boolean active;
}
