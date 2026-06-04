package com.park.utmstack.domain.index_template;

public class CreateTemplateResponse {
    private boolean acknowledged;

    public boolean isAcknowledged() {
        return acknowledged;
    }

    public void setAcknowledged(boolean acknowledged) {
        this.acknowledged = acknowledged;
    }
}
