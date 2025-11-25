package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Response DTO for Identity Provider configuration.
 * Adapted for SAML providers only.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderConfigResponseDto {
    private Long id;
    private String name;
    private ProviderType providerType;

    private String entityId;       // IdP entityID
    private String ssoUrl;         // SingleSignOnService URL
    private String sloUrl;         // SingleLogoutService URL
    private String nameIdFormat;   // NameID format (emailAddress, persistent, etc.)
    private String binding;        // HTTP-POST, Redirect, etc.

    private Boolean active;
}
