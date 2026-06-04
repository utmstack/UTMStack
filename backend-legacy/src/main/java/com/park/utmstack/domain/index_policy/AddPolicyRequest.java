package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

public class AddPolicyRequest {
    @SerializedName(value = "policy_id")
    private String policyId;

    public AddPolicyRequest(String policyId) {
        this.policyId = policyId;
    }

    public String getPolicyId() {
        return policyId;
    }

    public void setPolicyId(String policyId) {
        this.policyId = policyId;
    }
}
