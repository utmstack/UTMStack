package com.park.utmstack.config.saml;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.util.exceptions.SamlMetadataUrlInvalidException;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.saml2.core.Saml2X509Credential;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrations;

import java.net.InetAddress;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.UnknownHostException;
import java.security.PrivateKey;
import java.security.cert.X509Certificate;
import java.time.Duration;
import java.util.concurrent.*;
import java.util.function.Function;

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
                        // Closes the direct-internal-IP fetch path (loopback/site-local/link-local/
                        // any-local/multicast, plus non-http(s) schemes). Two documented residuals
                        // remain because Spring's RelyingPartyRegistrations.fromMetadataLocation uses
                        // its own static DefaultResourceLoader and we cannot inject a no-redirect,
                        // address-pinned connection into it:
                        //   (1) redirect-following — a public metadata_url that 302-redirects to an
                        //       internal target will be followed; defeating this requires replacing
                        //       fromMetadataLocation with a hand-rolled fetch that pins the resolved
                        //       InetAddress and disables redirects (tracked follow-up, needs a live
                        //       IdP to E2E-verify it does not break SAML login).
                        //   (2) DNS rebinding (TOCTOU) — we resolve once here, Spring resolves again
                        //       at fetch; a rebinding host could swap records between the two.
                        assertFetchableMetadataUrl(entity.getMetadataUrl());
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

    /**
     * Validates that a persisted IdP metadata URL is fetchable: it must use the
     * http or https scheme, carry a host, and resolve to a public address
     * (rejects loopback, site-local, link-local, any-local and multicast).
     * Resolution uses real DNS via {@link InetAddress#getByName}.
     *
     * @throws SamlMetadataUrlInvalidException if the URL is malformed, uses a
     *                                        disallowed scheme, has no host, or
     *                                        resolves to a non-public address
     */
    public static void assertFetchableMetadataUrl(String url) {
        assertFetchableMetadataUrl(url, host -> {
            try {
                return InetAddress.getByName(host);
            } catch (UnknownHostException e) {
                throw new SamlMetadataUrlInvalidException("Unknown host: " + host);
            }
        });
    }

    /**
     * Testable overload of {@link #assertFetchableMetadataUrl(String)} that accepts
     * an injected resolver so unit tests can assert host rejection without performing
     * real DNS lookups.
     */
    public static void assertFetchableMetadataUrl(String url, Function<String, InetAddress> resolver) {
        if (url == null || url.trim().isEmpty()) {
            throw new SamlMetadataUrlInvalidException("Metadata URL is required");
        }

        URL parsed;
        try {
            parsed = new URL(url);
        } catch (MalformedURLException e) {
            throw new SamlMetadataUrlInvalidException(
                    "Invalid metadata URL format: " + e.getMessage());
        }

        String scheme = parsed.getProtocol() == null ? "" : parsed.getProtocol().toLowerCase();
        if (!scheme.equals("http") && !scheme.equals("https")) {
            throw new SamlMetadataUrlInvalidException(
                    "Metadata URL must use the http or https scheme");
        }

        String host = parsed.getHost();
        if (host == null || host.trim().isEmpty()) {
            throw new SamlMetadataUrlInvalidException(
                    "Metadata URL must include a host");
        }

        InetAddress address = resolver.apply(host);
        if (!isFetchableHost(address)) {
            throw new SamlMetadataUrlInvalidException(
                    "Metadata URL host must be publicly resolvable");
        }
    }

    /**
     * Address-class check shared by the fetch-time validator and the API path:
     * returns {@code true} when the address is a legitimate public target —
     * i.e. NOT loopback, NOT site-local (RFC1918), NOT link-local
     * (169.254/16, e.g. the cloud metadata endpoint), NOT any-local, and NOT
     * multicast. Unit-testable without DNS.
     *
     * <p>Note: {@link InetAddress#isSiteLocalAddress()} is IPv4-only, so IPv6
     * site-local {@code fec0::/64} is NOT rejected; that range is obsolete per
     * RFC 3879 and unroutable, so this is accepted.
     */
    public static boolean isFetchableHost(InetAddress address) {
        return !address.isLoopbackAddress()
                && !address.isSiteLocalAddress()
                && !address.isLinkLocalAddress()
                && !address.isAnyLocalAddress()
                && !address.isMulticastAddress();
    }
}



