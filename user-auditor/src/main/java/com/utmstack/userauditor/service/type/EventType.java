package com.utmstack.userauditor.service.type;

import lombok.AllArgsConstructor;
import lombok.Getter;

@AllArgsConstructor
@Getter
public enum EventType {
    USER_CREATED(4720),
    USER_DELETED(4726),
    USER_LAST_LOGON(4624);

    private final int eventId;
}
