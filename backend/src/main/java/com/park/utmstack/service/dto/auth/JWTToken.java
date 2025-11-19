package com.park.utmstack.service.dto.auth;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class JWTToken {

    private String idToken;
    private boolean authenticated;

    @JsonProperty("id_token")
    String getIdToken() {
        return idToken;
    }
}
