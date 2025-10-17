package com.park.utmstack.domain.api_keys;


import lombok.*;
import java.time.Instant;
import java.util.UUID;

@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class ApiKeyUsageLog {

    private UUID id;

    private UUID apiKeyId;

    private String apiKeyName;

    private Long userId;

    private Instant timestamp;

    private String endpoint;

    private String address;

    private String errorMessage;

    private String queryParams;

    private String payload;

    private String userAgent;

    private String httpMethod;

    private Integer statusCode;
}
