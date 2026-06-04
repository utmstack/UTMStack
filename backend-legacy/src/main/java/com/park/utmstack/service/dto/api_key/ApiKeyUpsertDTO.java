package com.park.utmstack.service.dto.api_key;

import com.park.utmstack.validation.api_key.ValidIPOrCIDR;
import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotNull;
import java.time.Instant;
import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ApiKeyUpsertDTO {
    @NotNull
    @Schema(description = "API Key name", requiredMode = Schema.RequiredMode.REQUIRED)
    private String name;

    @Schema(description = "Allowed IP address or IP range in CIDR notation (e.g., '192.168.1.100' or '192.168.1.0/24'). If null, no IP restrictions are applied.")
    private List<@ValidIPOrCIDR String> allowedIp;

    @Schema(description = "Expiration timestamp of the API key")
    private Instant expiresAt;
}
