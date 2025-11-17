package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.jhipster.service.filter.*;

import java.io.Serializable;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderCriteria implements Serializable {
    private static final long serialVersionUID = 1L;

    public static class ProviderTypeFilter extends Filter<ProviderType> { }

    private LongFilter id;
    private StringFilter name;
    private ProviderTypeFilter providerType;
    private StringFilter redirectUri;
    private StringFilter scopes;
    private StringFilter authUri;
    private StringFilter tokenUri;
    private StringFilter userInfoUri;
    private StringFilter jwksUri;
    private InstantFilter createdDate;
    private InstantFilter lastModifiedDate;
    private BooleanFilter active;
}