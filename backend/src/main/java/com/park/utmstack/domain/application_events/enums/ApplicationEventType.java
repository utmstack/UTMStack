package com.park.utmstack.domain.application_events.enums;

public enum ApplicationEventType {
    AUTH_ATTEMPT,
    AUTH_SUCCESS,
    AUTH_FAILURE,
    TFA_CODE_SENT,
    TFA_VERIFIED,
    AUTH_LOGOUT,
    CONFIG_CHANGED,
    USER_MANAGEMENT,
    ACCESS_DENIED,
    ERROR,
    WARNING,
    INFO
}
