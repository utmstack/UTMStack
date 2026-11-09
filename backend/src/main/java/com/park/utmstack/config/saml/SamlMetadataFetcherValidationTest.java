package com.park.utmstack.config.saml;

import com.park.utmstack.util.exceptions.SamlMetadataUrlInvalidException;
import org.junit.jupiter.api.Test;

import java.net.InetAddress;
import java.net.UnknownHostException;
import java.util.function.Function;

import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.junit.jupiter.api.Assertions.assertThrows;

/**
 * CVE-2026-82040 (fetch-time leg) — a persisted IdP row pointing at
 * loopback/link-local/private ranges must never make the backend issue an
 * outbound metadata request. All host resolution is injected (no real DNS):
 * literal-IP cases use InetAddress.getByName, which parses without a lookup.
 */
class SamlMetadataFetcherValidationTest {

    /** Parses literal IPs without DNS; returns fixed addresses for named hosts. */
    private static final Function<String, InetAddress> RESOLVER = host -> {
        switch (host) {
            case "idp.example.com":
                return ip("8.8.8.8");
            case "example.com":
                return ip("93.184.216.34");
            case "internal.example.com":
                return ip("10.0.0.5");
            default:
                try {
                    return InetAddress.getByName(host);
                } catch (UnknownHostException e) {
                    throw new IllegalStateException("unexpected DNS lookup in test: " + host, e);
                }
        }
    };

    private static InetAddress ip(String s) {
        try {
            return InetAddress.getByName(s);
        } catch (UnknownHostException e) {
            throw new IllegalStateException(e);
        }
    }

    // --- REJECT: internal address classes

    @Test
    void rejectsLoopback() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "http://127.0.0.1:9999/m.xml", RESOLVER));
    }

    @Test
    void rejectsSiteLocal() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "http://10.0.0.5/m.xml", RESOLVER));
    }

    @Test
    void rejectsLinkLocalMetadataEndpoint() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "http://169.254.169.254/latest/meta-data", RESOLVER));
    }

    @Test
    void rejectsIpv6Loopback() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "http://[::1]/m.xml", RESOLVER));
    }

    @Test
    void rejectsHostnameResolvingToSiteLocal() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "https://internal.example.com/m.xml", RESOLVER));
    }

    @Test
    void rejectsIpv6LinkLocal() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "http://[fe80::1]/m.xml", RESOLVER));
    }

    @Test
    void rejectsMulticastAndAnyLocal() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "http://224.0.0.1/m.xml", RESOLVER));
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "http://0.0.0.0/m.xml", RESOLVER));
    }

    // --- REJECT: schemes and malformed URLs

    @Test
    void rejectsFileScheme() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "file:///etc/passwd", RESOLVER));
    }

    @Test
    void rejectsFtpScheme() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "ftp://example.com/m.xml", RESOLVER));
    }

    @Test
    void rejectsMalformedUrl() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "not a url", RESOLVER));
    }

    @Test
    void rejectsNullOrEmptyUrl() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl("", RESOLVER));
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(null, RESOLVER));
    }

    @Test
    void rejectsUrlWithNoHost() {
        assertThrows(SamlMetadataUrlInvalidException.class,
                () -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                        "http:///m.xml", RESOLVER));
    }

    // --- ACCEPT: public addresses over http/https

    @Test
    void acceptsPublicHttpsHost() {
        assertDoesNotThrow(() -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                "https://idp.example.com/m.xml", RESOLVER));
    }

    @Test
    void acceptsPublicHttpHost() {
        assertDoesNotThrow(() -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                "http://example.com/m.xml", RESOLVER));
    }

    @Test
    void schemeCheckIsCaseInsensitive() {
        assertDoesNotThrow(() -> SamlMetadataFetcher.assertFetchableMetadataUrl(
                "HTTPS://idp.example.com/m.xml", RESOLVER));
    }
}
