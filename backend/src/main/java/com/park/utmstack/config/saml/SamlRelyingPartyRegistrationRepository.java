package com.park.utmstack.config.saml;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import com.park.utmstack.util.saml.PemUtils;
import org.springframework.security.saml2.core.Saml2X509Credential;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrationRepository;
import org.springframework.security.saml2.provider.service.registration.Saml2MessageBinding;

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
        X509Certificate idpCert = PemUtils.parseCertificate(entity.getCertPem());

        PrivateKey spKey = PemUtils.parsePrivateKey(entity.getSpPrivateKeyPem());
        X509Certificate spCert = PemUtils.parseCertificate(entity.getSpCertificatePem());

        return RelyingPartyRegistration
                .withRegistrationId(entity.getName())
                .entityId("http://localhost:8080/saml/sp") // Issuer del SP (Client ID en Keycloak)
                .assertingPartyDetails(party -> party
                        .entityId("https://localhost:8443/realms/UTMSTACK") // Issuer del IdP
                        .singleSignOnServiceLocation("https://localhost:8443/realms/UTMSTACK/protocol/saml")
                        .singleSignOnServiceBinding(Saml2MessageBinding.POST) // o REDIRECT según tu config
                        .wantAuthnRequestsSigned(true)
                        .verificationX509Credentials(c -> c.add(Saml2X509Credential.verification(idpCert)))
                )
                .assertionConsumerServiceLocation("http://localhost:8080/login/saml2/sso/" + entity.getProviderType().name().toLowerCase())
                .signingX509Credentials(c -> c.add(Saml2X509Credential.signing(spKey, spCert)))
                .build();

    }
}