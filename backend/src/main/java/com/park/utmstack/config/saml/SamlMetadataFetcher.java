package com.park.utmstack.config.saml;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.saml2.core.Saml2X509Credential;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrations;

import java.security.PrivateKey;
import java.security.cert.X509Certificate;
import java.time.Duration;
import java.util.concurrent.*;

@Slf4j
public class SamlMetadataFetcher {

    private static final Duration TIMEOUT = Duration.ofSeconds(10);

    private final ExecutorService executor = Executors.newFixedThreadPool(5, r -> {
        Thread t = new Thread(r);
        t.setName("saml-metadata-fetch");
        t.setDaemon(true);
        return t;
    });

    public RelyingPartyRegistration fetch(IdentityProviderConfig entity,
                                          PrivateKey spKey,
                                          X509Certificate spCert) {

        CompletableFuture<RelyingPartyRegistration> future =
                CompletableFuture.supplyAsync(() -> {
                    try {
                        return RelyingPartyRegistrations
                                .fromMetadataLocation(entity.getMetadataUrl())
                                .registrationId(entity.getName())
                                .entityId(entity.getSpEntityId())
                                .assertionConsumerServiceLocation(entity.getSpAcsUrl())
                                .signingX509Credentials(c -> c.add(Saml2X509Credential.signing(spKey, spCert)))
                                .build();
                    } catch (Exception e) {
                        throw new CompletionException(e);
                    }
                }, executor);

        try {
            return future.get(TIMEOUT.getSeconds(), TimeUnit.SECONDS);

        } catch (Exception e) {
            future.cancel(true);
            log.error("Metadata fetch failed for provider '{}': {}", entity.getName(), e.getMessage());
            return null;
        }
    }
}



