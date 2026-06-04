package com.park.utmstack.domain.tfa;

import lombok.Data;

@Data
public class TfaSetupState {
    private  String secret;
    private String lastUsedCode;
    private  long expiresAt;
    private  long setupStartedAt;

    private long lastChallengeAt;
    private static final long COOLDOWN_MS = 28_000;

    public TfaSetupState(String secret, long expiresAt) {
        this.secret = secret;
        this.expiresAt = expiresAt;
        this.setupStartedAt = System.currentTimeMillis();
        this.lastChallengeAt = 0;
    }

    public boolean isExpired() {
        return System.currentTimeMillis() > expiresAt;
    }

    public long getRemainingSeconds() {
        long remaining = expiresAt - System.currentTimeMillis();
        return Math.max(remaining / 1000, 0);
    }

    public boolean canRequestChallenge() {
        return System.currentTimeMillis() - lastChallengeAt >= COOLDOWN_MS;
    }

    public long getCooldownRemainingSeconds() {
        long remaining = (lastChallengeAt + COOLDOWN_MS) - System.currentTimeMillis();
        return Math.max(remaining / 1000, 0);
    }

    public void markChallengeRequested() {
        this.lastChallengeAt = System.currentTimeMillis();
    }
}


