package com.park.utmstack.web.rest.vm;

import javax.validation.constraints.NotNull;

public class GettingStartedInit {
    @NotNull
    public boolean inSaas;

    public boolean isInSaas() {
        return inSaas;
    }

    public void setInSaas(boolean inSaas) {
        this.inSaas = inSaas;
    }
}
