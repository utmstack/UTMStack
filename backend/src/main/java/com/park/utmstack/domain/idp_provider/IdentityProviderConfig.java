package com.park.utmstack.domain.idp_provider;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.Type;

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

    @Column(nullable = false)
    private String name;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private ProviderType providerType;

    /**
     * Metadata URL of the IdP (Keycloak, Okta, Azure, etc.)
     * Example: https://localhost:8443/realms/UTMSTACK/protocol/saml/descriptor
     */
    @Column(name = "metadata_url", nullable = false, length = 512)
    private String metadataUrl;

    /**
     * Service Provider private key in PEM format
     * Used to sign AuthnRequests and other outgoing SAML messages
     */
    @Type(type = "text")
    @Column(name = "sp_private_key_pem", nullable = false, columnDefinition = "TEXT")
    private String spPrivateKeyPem;

    /**
     * Service Provider public certificate in PEM format
     * Shared with IdP so it can validate signed requests from the SP
     */
    @Type(type = "text")
    @Column(name = "sp_certificate_pem", nullable = false, columnDefinition = "TEXT")
    private String spCertificatePem;

    @Column(name = "sp_entity_id", nullable = false, length = 512)
    private String spEntityId;

    @Column(name = "sp_acs_url", nullable = false, length = 512)
    private String spAcsUrl;

    /**
     * Flag to enable or disable this IdP configuration
     */
    @Column(nullable = false)
    private Boolean active;

    /**
     * Timestamp when the record was created
     */
    @Column(nullable = false)
    private LocalDateTime createdAt;

    /**
     * Timestamp when the record was last updated
     */
    @Column(nullable = false)
    private LocalDateTime updatedAt;
}
