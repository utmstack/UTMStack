package com.park.utmstack.config.saml;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.exceptions.ApiException;
import com.park.utmstack.util.saml.PemUtils;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.security.saml2.core.Saml2X509Credential;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrationRepository;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrations;

import java.security.PrivateKey;
import java.security.cert.X509Certificate;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
public class SamlRelyingPartyRegistrationRepository implements RelyingPartyRegistrationRepository {

    private final Map<String, RelyingPartyRegistration> registrations = new ConcurrentHashMap<>();

    public SamlRelyingPartyRegistrationRepository(IdentityProviderConfigRepository jpaProviderRepository) {
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

    private void loadProviders(IdentityProviderConfigRepository jpaProviderRepository) {
        try {
            List<IdentityProviderConfig> activeProviders = jpaProviderRepository.findAllByActiveTrue();

            if (activeProviders.isEmpty()) {
                return;
            }

            activeProviders.forEach(entity -> {
                try {
                    RelyingPartyRegistration registration = buildRelyingPartyRegistration(entity);
                    registrations.put(entity.getProviderType().name().toLowerCase(), registration);
                    log.info("Loaded SAML provider: {} (type: {})", entity.getName(), entity.getProviderType());
                } catch (Exception e) {
                    log.error("Failed to load SAML provider: {}", entity.getName(), e);
                }
            });

            log.info("Successfully loaded {} SAML provider(s)", registrations.size());
        } catch (Exception e) {
            log.error("Failed to load SAML providers: {}", e.getMessage(), e);
            throw new ApiException(String.format("Failed to load SAML providers: %s", e.getMessage()), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    private RelyingPartyRegistration buildRelyingPartyRegistration(IdentityProviderConfig entity) {
        try {
            String encryptionKey = System.getenv(Constants.ENV_ENCRYPTION_KEY);
            if (encryptionKey == null || encryptionKey.isBlank()) {
                throw new IllegalStateException(
                        "Environment variable " + Constants.ENV_ENCRYPTION_KEY + " not configured");
            }

            String decryptedKey = CipherUtil.decrypt(entity.getSpPrivateKeyPem(), encryptionKey);
            PrivateKey spKey = PemUtils.parsePrivateKey(decryptedKey);
            X509Certificate spCert = PemUtils.parseCertificate(entity.getSpCertificatePem());

            return RelyingPartyRegistrations
                    .fromMetadataLocation(entity.getMetadataUrl())
                    .registrationId(entity.getName())
                    .entityId(entity.getSpEntityId())
                    .assertionConsumerServiceLocation(entity.getSpAcsUrl())
                    .signingX509Credentials(c -> c.add(Saml2X509Credential.signing(spKey, spCert)))
                    .build();
        } catch (Exception e) {
            log.error("Failed to build SAML registration for provider: {}", entity.getName(), e);
            throw new ApiException(String.format("Failed to build SAML registration for provider: %s", entity.getName()), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

}