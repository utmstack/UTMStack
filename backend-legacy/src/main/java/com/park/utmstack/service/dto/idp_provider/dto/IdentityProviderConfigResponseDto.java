package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

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
    private String metadataUrl;
    private String spCertificatePem;
    private Boolean active;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
