package com.park.utmstack.config.saml;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.saml.PemUtils;
import org.springframework.security.saml2.core.Saml2X509Credential;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrationRepository;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrations;

import java.security.PrivateKey;
import java.security.cert.X509Certificate;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

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
        jpaProviderRepository.findAllByActiveTrue().forEach(entity -> {
            RelyingPartyRegistration registration = buildRelyingPartyRegistration(entity);
            registrations.put(entity.getProviderType().name().toLowerCase(), registration);
        });
    }

    private RelyingPartyRegistration buildRelyingPartyRegistration(IdentityProviderConfig entity) {

        PrivateKey spKey = PemUtils.parsePrivateKey(CipherUtil.decrypt(
                entity.getSpPrivateKeyPem(),
                System.getenv(Constants.ENV_ENCRYPTION_KEY)
        ));
        X509Certificate spCert = PemUtils.parseCertificate(entity.getSpCertificatePem());

        // Build registration directly from IdP metadata URL
        return RelyingPartyRegistrations
                .fromMetadataLocation(entity.getMetadataUrl())
                .registrationId(entity.getName())
                .entityId("http://localhost:8080/saml/sp")
                .assertionConsumerServiceLocation(
                        "http://localhost:8080/login/saml2/sso/" + entity.getProviderType().name().toLowerCase()
                )
                .signingX509Credentials(c -> c.add(Saml2X509Credential.signing(spKey, spCert)))
                .build();
    }

}