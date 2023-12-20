package com.park.utmstack.web.rest.vm;

import javax.validation.constraints.NotBlank;

public class CommandVM {
    @NotBlank
    String command;
    @NotBlank
    String originType;
    @NotBlank
    String originId;
    @NotBlank
    String reason;

    public String getCommand() {
        return command;
    }

    public void setCommand(String command) {
        this.command = command;
    }

    public String getOriginType() {
        return originType;
    }

    public void setOriginType(String originType) {
        this.originType = originType;
    }

    public String getOriginId() {
        return originId;
    }

    public void setOriginId(String originId) {
        this.originId = originId;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }
}
