package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import java.time.LocalDateTime;

/**
 * DTO for Identity Provider configuration requests.
 * Extended for SAML providers.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderConfigRequestDto {

    private Long id;

    @NotBlank
    private String name;

    @NotNull
    private ProviderType providerType;

    @NotBlank
    private String metadataUrl;

    @NotBlank
    private String spPrivateKeyPem;

    @NotNull
    private Boolean active;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;

}
