package com.park.utmstack.service.dto.jwt;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.park.utmstack.domain.tfa.TfaMethod;
import lombok.AllArgsConstructor;
import lombok.Data;

@Data
public class JWTToken {
    private String idToken;
    private boolean authenticated;
    private String tfaMethod;

    public JWTToken(String idToken, boolean authenticated, String tfaMethod) {
        this.idToken = idToken;
        this.authenticated = authenticated;
        this.tfaMethod = tfaMethod;
    }

    public JWTToken(String idToken, boolean authenticated) {
        this.idToken = idToken;
        this.authenticated = authenticated;
        this.tfaMethod = null;
    }

    @JsonProperty("id_token")
    String getIdToken() {
        return idToken;
    }

    void setIdToken(String idToken) {
        this.idToken = idToken;
    }

}
