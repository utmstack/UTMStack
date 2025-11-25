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

    private String name;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private ProviderType providerType;

    @Column(nullable = false, length = 512)
    private String entityId; // IdP entityID

    @Column(nullable = false, length = 512)
    private String ssoUrl; // SingleSignOnService URL

    @Column(name = "nameid_format", length = 255)
    private String nameIdFormat; // e.g. emailAddress

    @Column(length = 512)
    private String sloUrl; // SingleLogoutService URL (optional)

    @Type(type = "text")
    @Column(name = "cert_pem", nullable = false, columnDefinition = "TEXT")
    private String certPem;

    @Type(type = "text")
    @Column(name = "sp_private_key_pem", columnDefinition = "TEXT")
    private String spPrivateKeyPem;

    @Type(type = "text")
    @Column(name = "sp_certificate_pem", columnDefinition = "TEXT")
    private String spCertificatePem;

    @Column(length = 50)
    private String binding; // HTTP-POST, Redirect, etc.

    @Column(nullable = false)
    private Boolean active;

    @Column(nullable = false)
    private LocalDateTime createdAt;

    @Column(nullable = false)
    private LocalDateTime updatedAt;

}
