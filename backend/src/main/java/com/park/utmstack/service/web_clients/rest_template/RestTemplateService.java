package com.park.utmstack.service.web_clients.rest_template;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestClientResponseException;
import org.springframework.web.client.RestTemplate;

import java.util.List;
import java.util.Objects;

@Service
public class RestTemplateService {
    private static final String CLASSNAME = "RestTemplateService";

    private final RestTemplate rest;
    private final RestTemplate restTemplateWithSsl;

    private final HttpHeaders headers = new HttpHeaders();

    public RestTemplateService(RestTemplate restTemplate, RestTemplate restTemplateWithSsl) {
        this.rest = restTemplate;
        this.restTemplateWithSsl = restTemplateWithSsl;
        headers.add("Content-Type", "application/json");
        headers.add("Accept", "*/*");
    }

    public <T> ResponseEntity<T> get(String url, Class<T> type) throws Exception {
        final String ctx = CLASSNAME + ".get";
        try {
            HttpEntity<String> requestEntity = new HttpEntity<>("", headers);
            return rest.exchange(url, HttpMethod.GET, requestEntity, type);
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public <T> ResponseEntity<T> get(String url, Class<T> type, HttpHeaders headers) throws Exception {
        final String ctx = CLASSNAME + ".get";
        try {
            HttpEntity<String> requestEntity = new HttpEntity<>(headers);
            return rest.exchange(url, HttpMethod.GET, requestEntity, type);
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public ResponseEntity<String> getFromRaw(String url) throws Exception {
        final String ctx = CLASSNAME + ".get";
        try {
            return rest.getForEntity(url, String.class);
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public <T> ResponseEntity<T> getFromRaw(String url, Class<T> type) throws Exception {
        final String ctx = CLASSNAME + ".get";
        try {
            ResponseEntity<String> rs = rest.getForEntity(url, String.class);
            ObjectMapper om = new ObjectMapper();
            om.registerModule(new JavaTimeModule());
            return ResponseEntity.ok(om.readValue(rs.getBody(), type));
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public <T> ResponseEntity<List<T>> getFromRaw(String url, TypeReference<List<T>> p) throws Exception {
        final String ctx = CLASSNAME + ".getFromRaw";
        try {
            ResponseEntity<String> rs = rest.getForEntity(url, String.class);
            ObjectMapper om = new ObjectMapper();
            om.registerModule(new JavaTimeModule());
            return ResponseEntity.ok(om.readValue(rs.getBody(), p));
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public ResponseEntity<String> head(String url) throws Exception {
        final String ctx = CLASSNAME + ".head";
        try {
            HttpEntity<String> requestEntity = new HttpEntity<>("", headers);
            return rest.exchange(url, HttpMethod.HEAD, requestEntity, String.class);
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        } catch (ResourceAccessException e) {
            String msg = ctx + ": " + e.getMessage();
            return ResponseEntity.status(HttpStatus.NOT_FOUND).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public <T, U> ResponseEntity<U> post(String url, T body, Class<U> type) throws Exception {
        final String ctx = CLASSNAME + ".post";
        try {
            HttpEntity<T> requestEntity = new HttpEntity<>(body, headers);
            return rest.exchange(url, HttpMethod.POST, requestEntity, type);
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public <T, U> ResponseEntity<U> put(String url, T body, Class<U> type) {
        final String ctx = CLASSNAME + ".put";
        try {
            HttpEntity<T> requestEntity = new HttpEntity<>(body, headers);
            return rest.exchange(url, HttpMethod.PUT, requestEntity, type);
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage() + ": " + e.getResponseBodyAsString();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public <T> ResponseEntity<List<T>> get(String url, ParameterizedTypeReference<List<T>> p) throws Exception {
        final String ctx = CLASSNAME + ".get";
        try {
            HttpEntity<String> requestEntity = new HttpEntity<>(headers);
            return rest.exchange(url, HttpMethod.GET, requestEntity, p);
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage() + ": " + e.getResponseBodyAsString();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public <T> ResponseEntity<List<T>> get(String url, ParameterizedTypeReference<List<T>> p, HttpHeaders headers, boolean withSslSupport) throws Exception {
        final String ctx = CLASSNAME + ".get";
        try {
            HttpEntity<String> requestEntity = new HttpEntity<>(headers);
            RestTemplate rt = withSslSupport ? restTemplateWithSsl : rest;
            return rt.exchange(url, HttpMethod.GET, requestEntity, p);
        } catch (RestClientResponseException e) {
            String msg = ctx + ": " + e.getMessage() + ": " + e.getResponseBodyAsString();
            return ResponseEntity.status(e.getRawStatusCode()).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public String extractErrorMessage(ResponseEntity<?> rs) {
        if (Objects.isNull(rs))
            return "";
        return Objects.requireNonNull(rs.getHeaders().get("X-UtmStack-error")).get(0);
    }

    public RestTemplateService addHeader(String key, List<String> value) {
        headers.put(key, value);
        return this;
    }

    public RestTemplateService acceptingSsl() {


        return this;

    }
}
