package com.park.utmstack.loggin.api_key;

import com.park.utmstack.domain.api_keys.ApiKey;
import com.park.utmstack.domain.api_keys.ApiKeyUsageLog;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.api_key.ApiKeyService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;
import org.springframework.web.util.ContentCachingRequestWrapper;
import org.springframework.web.util.ContentCachingResponseWrapper;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.UUID;

import static org.postgresql.PGProperty.APPLICATION_NAME;

@Service
@Slf4j
@RequiredArgsConstructor
public class ApiKeyUsageLoggingService {

    private final ApiKeyService apiKeyService;
    private final ApplicationEventService applicationEventService;
    private static final String LOG_USAGE_FLAG = "LOG_USAGE_DONE";

    public void logUsage(ContentCachingRequestWrapper request,
                         ContentCachingResponseWrapper response,
                         ApiKey apiKey,
                         String ipAddress,
                         String message) {

        if (Boolean.TRUE.equals(request.getAttribute(LOG_USAGE_FLAG))) {
            return;
        }

        try {
            String payload = extractPayload(request);
            String errorText = extractErrorText(response);
            int status = safeStatus(response);

            ApiKeyUsageLog usage = buildUsageLog(apiKey, ipAddress, request, status, errorText, payload);

            apiKeyService.logUsage(usage);

            ApplicationEventType eventType = (status >= 400)
                    ? ApplicationEventType.API_KEY_ACCESS_FAILURE
                    : ApplicationEventType.API_KEY_ACCESS_SUCCESS;

            String eventMessage = (status >= 400)
                    ? "API key access failure"
                    : "API key access";

            applicationEventService.createEvent(eventMessage, eventType, usage.toAuditMap());

        } catch (Exception e) {
            log.error("Error while logging API key usage: {}", e.getMessage(), e);
        } finally {
            request.setAttribute(LOG_USAGE_FLAG, Boolean.TRUE);
        }
    }

    private int safeStatus(HttpServletResponse response) {
        try {
            return response.getStatus();
        } catch (Exception e) {
            return 0;
        }
    }

    private String extractPayload(ContentCachingRequestWrapper request) {
        try {
            if (!"GET".equalsIgnoreCase(request.getMethod()) && !"DELETE".equalsIgnoreCase(request.getMethod())) {
                byte[] content = request.getContentAsByteArray();
                return content.length > 0 ? new String(content, StandardCharsets.UTF_8) : null;
            }
        } catch (Exception ex) {
            log.error("Error extracting payload: {}", ex.getMessage());
        }
        return null;
    }

    private String extractErrorText(ContentCachingResponseWrapper response) {
        int statusCode = response.getStatus();
        if (statusCode >= 400) {
            byte[] content = response.getContentAsByteArray();
            String responseError = content.length > 0 ? new String(content, StandardCharsets.UTF_8) : null;
            String errorHeader = response.getHeader("X-" + APPLICATION_NAME + "-error");
            return StringUtils.hasText(responseError) ? responseError : errorHeader;
        }
        return null;
    }

    private ApiKeyUsageLog buildUsageLog(ApiKey apiKey,
                                         String ipAddress,
                                         HttpServletRequest request,
                                         int status,
                                         String errorText,
                                         String payload) {

        String id = UUID.randomUUID().toString();
        String apiKeyId = apiKey != null && apiKey.getId() != null ? apiKey.getId().toString() : null;
        String apiKeyName = apiKey != null ? apiKey.getName() : null;
        String userId = apiKey != null && apiKey.getUserId() != null ? apiKey.getUserId().toString() : null;
        String timestamp = Instant.now().toString();
        String endpoint = request != null ? request.getRequestURI() : null;
        String queryParams = request != null ? request.getQueryString() : null;
        String userAgent = request != null ? request.getHeader("User-Agent") : null;
        String httpMethod = request != null ? request.getMethod() : null;
        String statusCode = String.valueOf(status);

        String safePayload = null;
        if (payload != null) {
            int PAYLOAD_MAX_LENGTH = 2000;
            safePayload = payload.length() > PAYLOAD_MAX_LENGTH ? payload.substring(0, PAYLOAD_MAX_LENGTH) : payload;
        }

        return ApiKeyUsageLog.builder()
                .id(id)
                .apiKeyId(apiKeyId)
                .apiKeyName(apiKeyName)
                .userId(userId)
                .timestamp(timestamp)
                .endpoint(endpoint)
                .address(ipAddress)
                .errorMessage(errorText)
                .queryParams(queryParams)
                .payload(safePayload)
                .userAgent(userAgent)
                .httpMethod(httpMethod)
                .statusCode(statusCode)
                .build();
    }
}
