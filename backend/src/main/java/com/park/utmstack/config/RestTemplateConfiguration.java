package com.park.utmstack.config;

import com.park.utmstack.service.web_clients.rest_template.RestTemplateResponseErrorHandler;
import org.apache.http.conn.ssl.SSLConnectionSocketFactory;
import org.apache.http.conn.ssl.TrustStrategy;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.client.HttpComponentsClientHttpRequestFactory;
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
