package com.park.utmstack.domain.api_keys;

import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.*;

import java.util.HashMap;
import java.util.Map;

@Builder
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
public class ApiKeyUsageLog implements AuditableDTO {

    private String id;
    private String apiKeyId;
    private String apiKeyName;
    private String userId;
    private String timestamp;
    private String endpoint;
    private String address;
    private String errorMessage;
    private String queryParams;
    private String payload;
    private String userAgent;
    private String httpMethod;
    private String statusCode;

    @Override
    public Map<String, Object> toAuditMap() {
        Map<String, Object> map = new HashMap<>();

        map.put("id", id);
        map.put("api_key_id", apiKeyId);
        map.put("api_key_name", apiKeyName);
        map.put("user_id", userId);
        map.put("timestamp", timestamp != null ? timestamp : null);
        map.put("endpoint", endpoint);
        map.put("address", address);
        map.put("error_message", errorMessage);
        map.put("query_params", queryParams);
        map.put("user_agent", userAgent);
        map.put("http_method", httpMethod);
        map.put("status_code", statusCode);

        return map;
    }
}
