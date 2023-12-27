package com.utmstack.userauditor.service.type;

import lombok.AllArgsConstructor;
import lombok.Getter;
@AllArgsConstructor
@Getter
public enum SourceType {
    WINDOWS("WINDOWS_AGENT");

    private final String value;
}
