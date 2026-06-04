package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.jhipster.service.filter.*;

import java.io.Serializable;

/**
 * Criteria class for filtering IdentityProviderConfig entities.
 * Adapted for SAML providers only.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderCriteria implements Serializable {
    private static final long serialVersionUID = 1L;

    public static class ProviderTypeFilter extends Filter<ProviderType> { }

    private LongFilter id;
    private StringFilter name;
    private ProviderTypeFilter providerType;
    private BooleanFilter active;
    private InstantFilter createdDate;
    private InstantFilter lastModifiedDate;
}
