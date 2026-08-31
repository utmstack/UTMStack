package com.park.utmstack.web.rest.vm;

import javax.validation.constraints.NotNull;

/**
 * Body of POST /api/check-credentials. The password used to travel in the query
 * string, where it landed in nginx access logs and browser history.
 */
public class CheckCredentialsVM {

    @NotNull
    private String password;

    @NotNull
    private String checkUUID;

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }

    public String getCheckUUID() {
        return checkUUID;
    }

    public void setCheckUUID(String checkUUID) {
        this.checkUUID = checkUUID;
    }
}
