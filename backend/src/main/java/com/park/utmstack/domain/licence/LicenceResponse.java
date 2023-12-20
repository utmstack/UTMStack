package com.park.utmstack.domain.licence;

import com.fasterxml.jackson.annotation.JsonInclude;

@JsonInclude(value = JsonInclude.Include.NON_NULL)
public class LicenceResponse {
    private Boolean success;
    private Licence data;

    public Boolean getSuccess() {
        return success;
    }

    public void setSuccess(Boolean success) {
        this.success = success;
    }

    public Licence getData() {
        return data;
    }

    public void setData(Licence data) {
        this.data = data;
    }
}
