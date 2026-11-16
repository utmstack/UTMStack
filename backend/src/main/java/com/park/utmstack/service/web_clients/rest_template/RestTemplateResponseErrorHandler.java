package com.park.utmstack.service.web_clients.rest_template;

import org.springframework.http.client.ClientHttpResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.client.DefaultResponseErrorHandler;
import org.springframework.web.client.RestClientResponseException;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import org.springframework.http.HttpMethod;
import java.util.stream.Collectors;

@Component
public class RestTemplateResponseErrorHandler extends DefaultResponseErrorHandler {

    // Spring 7: ResponseErrorHandler.handleError now takes (URI, HttpMethod, ClientHttpResponse).
    @Override
    public void handleError(URI uri, HttpMethod httpMethod, ClientHttpResponse response) throws IOException {
        // Spring 7: ClientHttpResponse.getStatusCode() returns HttpStatusCode (int via value()).
        if (response.getStatusCode().isError()) {
            BufferedReader reader = new BufferedReader(new InputStreamReader(response.getBody()));
            String httpBodyResponse = reader.lines().collect(Collectors.joining(""));
            throw new RestClientResponseException(String.format("Error with status: %1$s and message: %2$s", response.getStatusCode().value(), httpBodyResponse), response.getStatusCode(), response.getStatusText(), null, null, StandardCharsets.UTF_8);
        }
    }
}
