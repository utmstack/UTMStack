package com.park.utmstack.domain.idp_provider;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;

import javax.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "utm_identity_provider_config")
@Data
@NoArgsConstructor
@AllArgsConstructor
@EqualsAndHashCode(onlyExplicitlyIncluded = true)
public class IdentityProviderConfig {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @EqualsAndHashCode.Include
    private Long id;

    private String name;

    @Enumerated(EnumType.STRING)
    private ProviderType providerType;

    private String clientId;
    private String clientSecret;
    private String redirectUri;
    private String scopes;
    private String allowedDomains;
    private Boolean active;

    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

}
