package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

import java.util.ArrayList;
import java.util.List;

public class UpdateManagedIndexPolicyResponse {
    @SerializedName("updated_indices")
    private int updatedIndices = 0;
    private boolean failures;
    @SerializedName("failed_indices")
    private List<FailedIndices> failedIndices = new ArrayList<>();

    public Integer getUpdatedIndices() {
        return updatedIndices;
    }

    public void setUpdatedIndices(Integer updatedIndices) {
        this.updatedIndices = updatedIndices;
    }

    public Boolean getFailures() {
        return failures;
    }

    public void setFailures(Boolean failures) {
        this.failures = failures;
    }

    public List<FailedIndices> getFailedIndices() {
        return failedIndices;
    }

    public void setFailedIndices(List<FailedIndices> failedIndices) {
        this.failedIndices = failedIndices;
    }
}
