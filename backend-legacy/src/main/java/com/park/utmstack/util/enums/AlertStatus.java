package com.park.utmstack.util.enums;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.google.gson.annotations.SerializedName;

public enum AlertStatus {

    @SerializedName("Automatic review")
    @JsonProperty("Automatic review")
    AUTOMATIC_REVIEW(1, "Automatic review"),
    @SerializedName("Open")
    @JsonProperty("Open")
    OPEN(2, "Open"),
    @SerializedName("In review")
    @JsonProperty("In review")
    IN_REVIEW(3, "In review"),
    @SerializedName("Completed")
    @JsonProperty("Completed")
    COMPLETED(5, "Completed");

    private final int code;
    private final String name;

    AlertStatus(int code, String name) {
        this.code = code;
        this.name = name;
    }

    public int getCode() {
        return code;
    }

    public String getName() {
        return name;
    }

    public static AlertStatus getByCode(Integer code) {
        if (code == null)
            return null;
        switch (code) {
            case 1:
                return AUTOMATIC_REVIEW;
            case 2:
                return OPEN;
            case 3:
                return IN_REVIEW;
            case 5:
                return COMPLETED;
        }
        return null;
    }
}
