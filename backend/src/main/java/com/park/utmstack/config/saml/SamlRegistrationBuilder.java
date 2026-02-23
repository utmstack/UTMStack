package com.park.utmstack.config.saml;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.saml.PemUtils;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.saml2.core.Saml2X509Credential;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrations;

import java.security.PrivateKey;
import java.security.cert.X509Certificate;

/**
 * Responsible for building SAML registration objects.
 * Separates credential handling and registration building logic.
 */
@Slf4j
public class SamlRegistrationBuilder {

    private final String encryptionKey;
    private final SamlMetadataFetcher metadataFetcher;

    public SamlRegistrationBuilder(String encryptionKey, SamlMetadataFetcher metadataFetcher) {
        this.encryptionKey = encryptionKey;
        this.metadataFetcher = metadataFetcher;
    }

    /**
     * Builds a complete SAML registration with timeout protection and error handling.
     * Returns null if any step fails.
     *
     * @param entity Provider configuration
     * @return RelyingPartyRegistration, or null if build fails
     */
    public RelyingPartyRegistration buildRegistration(IdentityProviderConfig entity) {
        try {
            // Step 1: Fetch metadata with timeout
            RelyingPartyRegistration baseRegistration = metadataFetcher.fetchMetadataWithTimeout(entity);
            if (baseRegistration == null) {
                log.debug("Skipping provider '{}' - metadata fetch failed", entity.getName());
                return null;
            }

            // Step 2: Load and validate credentials
            PrivateKey spKey = loadAndDecryptPrivateKey(entity);
            if (spKey == null) {
                return null;
            }

            X509Certificate spCert = loadCertificate(entity);
            if (spCert == null) {
                return null;
            }

            // Step 3: Build final registration with credentials
            return buildWithCredentials(baseRegistration, entity, spKey, spCert);

        } catch (Exception e) {
            log.error("Unexpected error building SAML registration for provider '{}': {}",
                    entity.getName(), e.getMessage(), e);
            return null;
        }
    }

    /**
     * Loads, decrypts and validates the SP private key.
     * Returns null if decryption or parsing fails.
     */
    private PrivateKey loadAndDecryptPrivateKey(IdentityProviderConfig entity) {
        try {
            String decryptedKey = CipherUtil.decrypt(entity.getSpPrivateKeyPem(), this.encryptionKey);
            return PemUtils.parsePrivateKey(decryptedKey);
        } catch (Exception e) {
            log.error("Failed to load/decrypt SP private key for provider '{}': {}",
                    entity.getName(), e.getMessage(), e);
            return null;
        }
    }

    /**
     * Loads and validates the SP certificate.
     * Returns null if parsing fails.
     */
    private X509Certificate loadCertificate(IdentityProviderConfig entity) {
        try {
            return PemUtils.parseCertificate(entity.getSpCertificatePem());
        } catch (Exception e) {
            log.error("Failed to load SP certificate for provider '{}': {}",
                    entity.getName(), e.getMessage(), e);
            return null;
        }
    }

    /**
     * Configures the registration with SP credentials and custom settings.
     */
    private RelyingPartyRegistration buildWithCredentials(
            RelyingPartyRegistration baseRegistration,
            IdentityProviderConfig entity,
            PrivateKey spKey,
            X509Certificate spCert) {

        try {
            // Note: RelyingPartyRegistration from metadata is already built
            // We need to configure it with our SP credentials
            // For now, return the base registration built from metadata
            // Additional configuration can be done via Spring Security configuration
            return RelyingPartyRegistrations
                    .fromMetadataLocation(entity.getMetadataUrl())
                    .registrationId(entity.getProviderType().name().toLowerCase())
                    .assertionConsumerServiceLocation(entity.getSpAcsUrl())
                    .signingX509Credentials(c -> {
                        c.add(Saml2X509Credential.signing(spKey, spCert));
                    })
                    .build();
        } catch (Exception e) {
            log.error("Failed to configure SAML registration for provider '{}': {}",
                    entity.getName(), e.getMessage(), e);
            return null;
        }
    }
}

