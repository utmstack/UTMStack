package com.park.utmstack.config.saml;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.saml.PemUtils;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrationRepository;

import java.security.PrivateKey;
import java.security.cert.X509Certificate;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
public class SamlRelyingPartyRegistrationRepository implements RelyingPartyRegistrationRepository {

    private volatile Map<String, RelyingPartyRegistration> registrations = new ConcurrentHashMap<>();
    private final SamlMetadataFetcher fetcher;
    private final String encryptionKey;

    public SamlRelyingPartyRegistrationRepository(IdentityProviderConfigRepository jpaProviderRepository) {

        encryptionKey = getValidatedEncryptionKey();
        fetcher = new SamlMetadataFetcher();

        loadProviders(jpaProviderRepository);
    }

    @Override
    public RelyingPartyRegistration findByRegistrationId(String registrationId) {
        return registrations.get(registrationId);
    }

    public void reloadProviders(IdentityProviderConfigRepository jpaProviderRepository) {
        try {
            registrations = loadActiveProviders(jpaProviderRepository);
            log.info("SAML providers reloaded successfully: {} providers loaded", registrations.size());
        } catch (Exception e) {
            log.error("Failed to reload SAML providers - keeping previous configuration", e);
        }
    }

    /**
     * Loads SAML providers using the specialized loader.
     * Delegates all async loading logic to SamlProvidersLoader.
     * App will start without providers if loading fails.
     */
    private void loadProviders(IdentityProviderConfigRepository jpaProviderRepository) {
        try {
            registrations = loadActiveProviders(jpaProviderRepository);
            if (registrations.isEmpty()) {
                log.warn("No active SAML2 providers found. SAML2 authentication will not be available.");
            } else {
                log.info("Successfully loaded {} SAML2 provider(s) on startup", registrations.size());
            }
        } catch (Exception e) {
            log.error("Error during SAML provider loading - app will start without SAML2 authentication: {}",
                    e.getMessage(), e);
        }
    }

    /**
     * Validates and retrieves the encryption key from environment variables.
     *
     * @return The validated encryption key
     * @throws IllegalStateException if ENCRYPTION_KEY is not configured
     */
    private String getValidatedEncryptionKey() {
        String encryptionKey = System.getenv(Constants.ENV_ENCRYPTION_KEY);
        if (encryptionKey == null || encryptionKey.isBlank()) {
            throw new IllegalStateException(
                    "Environment variable " + Constants.ENV_ENCRYPTION_KEY + " not configured");
        }
        return encryptionKey;
    }

    private Map<String, RelyingPartyRegistration> loadActiveProviders(IdentityProviderConfigRepository repo) {

        Map<String, RelyingPartyRegistration> map = new ConcurrentHashMap<>();

        List<IdentityProviderConfig> activeProviders = repo.findAllByActiveTrue();

        activeProviders.forEach(entity -> {

            PrivateKey spKey = PemUtils.parsePrivateKey(CipherUtil.decrypt(
                    entity.getSpPrivateKeyPem(),
                    encryptionKey));

            X509Certificate spCert = PemUtils.parseCertificate(entity.getSpCertificatePem());

            RelyingPartyRegistration reg = fetcher.fetch(entity, spKey, spCert);

            if (reg != null) {
                map.put(entity.getProviderType().name().toLowerCase(), reg);
            }
        });

        return map;
    }


}