package com.park.utmstack.config.saml;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrationRepository;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
public class SamlRelyingPartyRegistrationRepository implements RelyingPartyRegistrationRepository {

    private final Map<String, RelyingPartyRegistration> registrations = new ConcurrentHashMap<>();
    private final SamlProvidersLoader providersLoader;

    public SamlRelyingPartyRegistrationRepository(IdentityProviderConfigRepository jpaProviderRepository) {

        String encryptionKey = getValidatedEncryptionKey();
        SamlMetadataFetcher metadataFetcher = new SamlMetadataFetcher();
        SamlRegistrationBuilder registrationBuilder = new SamlRegistrationBuilder(encryptionKey, metadataFetcher);
        this.providersLoader = new SamlProvidersLoader(registrationBuilder);

        // Load providers on initialization
        loadProviders(jpaProviderRepository);
    }

    @Override
    public RelyingPartyRegistration findByRegistrationId(String registrationId) {
        return registrations.get(registrationId);
    }

    public void reloadProviders(IdentityProviderConfigRepository jpaProviderRepository) {
        registrations.clear();
        loadProviders(jpaProviderRepository);
    }

    /**
     * Loads SAML providers using the specialized loader.
     * Delegates all async loading logic to SamlProvidersLoader.
     */
    private void loadProviders(IdentityProviderConfigRepository jpaProviderRepository) {
        try {
            List<IdentityProviderConfig> activeProviders = jpaProviderRepository.findAllByActiveTrue();
            Map<String, RelyingPartyRegistration> loadedRegistrations =
                    providersLoader.loadProvidersAsync(activeProviders);
            registrations.putAll(loadedRegistrations);
        } catch (Exception e) {
            log.error("Error during SAML provider loading: {}", e.getMessage(), e);
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

}