package com.park.utmstack.domain.tfa;

import lombok.Data;

@Data
public class TfaSetupState {
    private  String secret;
    private String lastUsedCode;
    private  long expiresAt;
    private  long setupStartedAt;

    public TfaSetupState(String secret, long expiresAt) {
        this.secret = secret;
        this.expiresAt = expiresAt;
        this.setupStartedAt = System.currentTimeMillis();
    }

    public boolean isExpired() {
        return System.currentTimeMillis() > expiresAt;
    }

    public long getRemainingSeconds() {
        long remaining = expiresAt - System.currentTimeMillis();
        return Math.max(remaining / 1000, 0);
    }
}


