package com.park.utmstack.service;

import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Service;
import org.springframework.web.reactive.function.client.WebClient;

@Service
public class WebClientService {
    private static final String CLASSNAME = "WebClientService";

    /**
     * Create a springboot reactive non blocking client
     *
     * @param url Base url of the client
     * @return A ${@link WebClient} object
     * @throws Exception In case of any error
     */
    public WebClient webClient(String url) throws Exception {
        final String ctx = CLASSNAME + ".webClient";
        try {
            return WebClient.builder().baseUrl(url)
                .defaultHeader(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_JSON_VALUE)
                .build();
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
