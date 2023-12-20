package com.park.utmstack.domain.licence;

public enum LicenceStatus {
    SOLD(1),
    DELIVERED(2),
    ACTIVE(3),
    INACTIVE(4);

    private final int statusValue;

    LicenceStatus(int statusValue) {
        this.statusValue = statusValue;
    }

    public int getStatusValue() {
        return statusValue;
    }
}
