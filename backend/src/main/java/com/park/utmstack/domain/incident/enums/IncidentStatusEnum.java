package com.park.utmstack.domain.incident.enums;

/**
 * The IncidentStatusEnum enumeration.
 */
public enum IncidentStatusEnum {
    OPEN(2),
    IN_REVIEW(3),
    COMPLETED(5),
    MERGED(0);
    private final int value;

    IncidentStatusEnum(final int newValue) {
        value = newValue;
    }

    public int getValue() {
        return value;
    }
}
