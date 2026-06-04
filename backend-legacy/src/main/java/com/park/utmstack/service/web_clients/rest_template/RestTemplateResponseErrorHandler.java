package com.park.utmstack.service.web_clients.rest_template;

import org.springframework.http.HttpStatus;
import org.springframework.http.client.ClientHttpResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.client.DefaultResponseErrorHandler;
import org.springframework.web.client.RestClientResponseException;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.util.Arrays;
import java.util.stream.Collectors;

@Component
public class RestTemplateResponseErrorHandler extends DefaultResponseErrorHandler {

    @Override
    public void handleError(ClientHttpResponse response) throws IOException {
        if (Arrays.asList(HttpStatus.Series.SERVER_ERROR, HttpStatus.Series.CLIENT_ERROR).contains(response.getStatusCode().series())) {
            BufferedReader reader = new BufferedReader(new InputStreamReader(response.getBody()));
            String httpBodyResponse = reader.lines().collect(Collectors.joining(""));
            throw new RestClientResponseException(String.format("Error with status: %1$s and message: %2$s", response.getRawStatusCode(), httpBodyResponse), response.getRawStatusCode(), response.getStatusText(), null, null, null);
        }
    }
}
