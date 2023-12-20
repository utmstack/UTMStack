package com.park.utmstack.util.enums;

import java.util.Optional;

public enum AlertSeverityEnum {
    LOW(1, "Low"),
    MEDIUM(2, "Medium"),
    HIGH(3, "High");

    private int code;
    private String name;

    public int getCode() {
        return code;
    }

    public String getName() {
        return name;
    }

    AlertSeverityEnum(int code, String name) {
        this.code = code;
        this.name = name;
    }

    public static Optional<AlertSeverityEnum> getByCode(Integer code) {
        switch (code) {
            case 1:
                return Optional.of(LOW);
            case 2:
                return Optional.of(MEDIUM);
            case 3:
                return Optional.of(HIGH);
            default:
                return Optional.empty();
        }
    }

    public static Optional<AlertSeverityEnum> getByName(String name) {
        switch (name) {
            case "Low":
                return Optional.of(LOW);
            case "Medium":
                return Optional.of(MEDIUM);
            case "High":
                return Optional.of(HIGH);
            default:
                return Optional.empty();
        }
    }
}
