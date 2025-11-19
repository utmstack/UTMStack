package com.park.utmstack.domain.shared_types;

import lombok.AllArgsConstructor;
import lombok.Getter;

@AllArgsConstructor
@Getter
public enum ApplicationLayer {
    SERVICE ("SERVICE"),
    CONTROLLER ("CONTROLLER");

    private final String value;
}
