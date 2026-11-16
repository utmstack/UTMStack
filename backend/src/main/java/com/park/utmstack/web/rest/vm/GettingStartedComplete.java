package com.park.utmstack.web.rest.vm;

import jakarta.validation.constraints.NotNull;

public class GettingStartedComplete {
    @NotNull
    public String stepShortName;

    public GettingStartedComplete() {
    }

    public String getStepShortName() {
        return stepShortName;
    }

    public void setStepShortName(String stepShortName) {
        this.stepShortName = stepShortName;
    }
}
