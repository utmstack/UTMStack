package com.park.utmstack.service.dto.api_key;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ApiKeyResponseDTO {

    @Schema(description = "Unique identifier of the API key")
    private Long id;

    @Schema(description = "User-friendly API key name")
    private String name;

    @Schema(description = "Allowed IP address or IP range in CIDR notation (e.g., '192.168.1.100' or '192.168.1.0/24')")
    private List<String> allowedIp;

    @Schema(description = "API key creation timestamp")
    private Instant createdAt;

    @Schema(description = "API key expiration timestamp (if applicable)")
    private Instant expiresAt;

    @Schema(description = "Generated At")
    private Instant generatedAt;
}
