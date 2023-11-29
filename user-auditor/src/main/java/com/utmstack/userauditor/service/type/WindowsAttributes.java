package com.utmstack.userauditor.service.type;

import lombok.AllArgsConstructor;
import lombok.Getter;

@AllArgsConstructor
@Getter
public enum WindowsAttributes {

    SAMAccountName("SAMAccountName"),
    Disabled("Disabled"),
    ObjectSID("ObjectSID"),
    ObjectType("ObjectType"),
    IsAdmin("IsAdmin"),
    CreatedAt("CreatedAt"),
    LastLogon("LastLogon"),
    Location("Location"),
    AccountExpire("AccountExpire");

    private final String value;
}
