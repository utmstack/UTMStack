package com.park.utmstack.domain.idp_provider;

import com.park.utmstack.domain.idp_provider.enums.ClientAuthMethod;
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
    @Column(nullable = false)
    private ProviderType providerType;

    @Column(nullable = false)
    private String clientId;

    @Column(nullable = false)
    private String clientSecret;

    @Column(nullable = false)
    private String authUri;

    @Column(nullable = false)
    private String tokenUri;

    @Column(nullable = false)
    private String redirectUri;

    private String userInfoUri;
    private String jwksUri;

    @Enumerated(EnumType.STRING)
    private ClientAuthMethod clientAuthMethod;

    private String scopes;
    private String allowedDomains;

    @Column(nullable = false)
    private Boolean active;

    @Column(nullable = false)
    private LocalDateTime createdAt;

    @Column(nullable = false)
    private LocalDateTime updatedAt;

}
