package com.park.utmstack.config;

import com.park.utmstack.service.web_clients.rest_template.RestTemplateResponseErrorHandler;
import org.apache.http.conn.ssl.SSLConnectionSocketFactory;
import org.apache.http.conn.ssl.TrustStrategy;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
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
        factory.setConnectTimeout(10_000);
        factory.setReadTimeout(120_000);
        RestTemplate rest = new RestTemplate(factory);
        rest.setErrorHandler(new DefaultResponseErrorHandler());
        return rest;
    }


    private HttpComponentsClientHttpRequestFactory clientHttpRequestFactory() {
        final String ctx = CLASSNAME + ".clientHttpRequestFactory";
        try {
            TrustStrategy acceptingTrustStrategy = (X509Certificate[] chain, String authType) -> true;
            HostnameVerifier hostnameVerifier = (s, sslSession) -> true;

            SSLContext sslContext = org.apache.http.ssl.SSLContexts.custom()
                .loadTrustMaterial(null, acceptingTrustStrategy)
                .build();

            SSLConnectionSocketFactory csf = new SSLConnectionSocketFactory(sslContext, hostnameVerifier);

            CloseableHttpClient httpClient = HttpClients.custom()
                .setSSLSocketFactory(csf)
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
