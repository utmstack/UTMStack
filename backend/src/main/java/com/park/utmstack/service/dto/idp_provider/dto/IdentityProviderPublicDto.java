package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Minimal identity-provider projection served to anonymous callers by the
 * login-page listing (GET /api/utm-providers).
 *
 * <p>Only the fields needed to render an SSO button are exposed: the full
 * {@link IdentityProviderConfigResponseDto} (metadata URL, SP certificate,
 * timestamps) stays behind the admin-only /api/identity-providers resource.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderPublicDto {

    private Long id;
    private String name;
    private ProviderType providerType;
    private Boolean active;
}
