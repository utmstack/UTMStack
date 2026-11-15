package com.park.utmstack.config;

import com.park.utmstack.service.web_clients.rest_template.RestTemplateResponseErrorHandler;
import org.apache.hc.client5.http.impl.classic.CloseableHttpClient;
import org.apache.hc.client5.http.impl.classic.HttpClients;
import org.apache.hc.client5.http.impl.io.PoolingHttpClientConnectionManagerBuilder;
import org.apache.hc.client5.http.ssl.SSLConnectionSocketFactory;
import org.apache.hc.core5.ssl.SSLContextBuilder;
import org.apache.hc.core5.ssl.TrustStrategy;
import org.jetbrains.annotations.NotNull;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.client.ClientHttpResponse;
import org.springframework.http.client.HttpComponentsClientHttpRequestFactory;
import org.springframework.web.client.DefaultResponseErrorHandler;
import org.springframework.web.client.RestTemplate;

import javax.net.ssl.HostnameVerifier;
import javax.net.ssl.SSLContext;
import java.security.cert.X509Certificate;
import java.util.Objects;

/**
 * Spring Boot 3 / Spring Framework 6: the default {@code HttpComponentsClientHttpRequestFactory}
 * is backed by Apache HttpClient 5 (not the 4.x the Boot 2 version used), so the SSL-trusting
 * client factory is rebuilt with the {@code org.apache.hc.client5.*} / {@code org.apache.hc.core5.*}
 * APIs. The trust-all behavior is intentionally preserved (these are internal service-mesh
 * endpoints using self-signed certs).
 */
@Configuration
public class RestTemplateConfiguration {
    public static final String CLASSNAME = "RestTemplateConfiguration";
    private final Logger log = LoggerFactory.getLogger(RestTemplateConfiguration.class);

    @Bean
    public RestTemplate restTemplate() {
        RestTemplate rest = new RestTemplate();
        rest.setErrorHandler(new RestTemplateResponseErrorHandler());
        return rest;
    }

    @Bean
    public RestTemplate restTemplateWithSsl() {
        RestTemplate rest = new RestTemplate();
        rest.setRequestFactory(Objects.requireNonNull(clientHttpRequestFactory()));
        rest.setErrorHandler(new RestTemplateResponseErrorHandler());
        return rest;
    }

    @Bean(name = "rawRestTemplate")
    public RestTemplate rawRestTemplate() {
        RestTemplate rest = new RestTemplate();
        rest.setErrorHandler(new DefaultResponseErrorHandler() {
            @Override
            public boolean hasError(@NotNull ClientHttpResponse response) {
                return false;
            }
        });
        return rest;
    }

    /**
     * RestTemplate for the web-pdf microservice. PDF generation is the slowest
     * backend call (headless browser render + print of a report), and it must be
     * bounded: without a read timeout a hung or OOM-ing printPage call makes the
     * UI spinner spin forever (issue #2934). Connect is quick (local service);
     * the generous read timeout leaves room for legitimate large reports.
     */
    @Bean(name = "pdfRestTemplate")
    public RestTemplate pdfRestTemplate() {
        HttpComponentsClientHttpRequestFactory factory = new HttpComponentsClientHttpRequestFactory();
        // Spring 7 removed setConnectTimeout(int); setConnectionRequestTimeout covers the
        // connection-acquisition wait (the same 10 s we set before).
        factory.setConnectionRequestTimeout(10_000);
        factory.setReadTimeout(120_000);
        RestTemplate rest = new RestTemplate(factory);
        rest.setErrorHandler(new DefaultResponseErrorHandler());
        return rest;
    }


    private HttpComponentsClientHttpRequestFactory clientHttpRequestFactory() {
        final String ctx = CLASSNAME + ".clientHttpRequestFactory";
        try {
            // Trust-all strategy (internal service-mesh TLS with self-signed certs) —
            // unchanged from the pre-migration httpclient4 config.
            TrustStrategy acceptingTrustStrategy = (X509Certificate[] chain, String authType) -> true;
            HostnameVerifier hostnameVerifier = (s, session) -> true;

            SSLContext sslContext = SSLContextBuilder.create()
                .loadTrustMaterial(null, acceptingTrustStrategy)
                .build();

            SSLConnectionSocketFactory sslSocketFactory = new SSLConnectionSocketFactory(sslContext, hostnameVerifier);

            CloseableHttpClient httpClient = HttpClients.custom()
                .setConnectionManager(PoolingHttpClientConnectionManagerBuilder.create()
                    .setSSLSocketFactory(sslSocketFactory)
                    .build())
                .build();

            HttpComponentsClientHttpRequestFactory requestFactory = new HttpComponentsClientHttpRequestFactory();
            requestFactory.setHttpClient(httpClient);

            return requestFactory;
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            e.printStackTrace();
            return null;
        }
    }
}
