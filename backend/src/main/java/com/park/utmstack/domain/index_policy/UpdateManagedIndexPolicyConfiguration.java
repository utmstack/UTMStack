package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class UpdateManagedIndexPolicyConfiguration {
    /**
     * Policy to apply
     */
    @SerializedName("policy_id")
    private String policyId;

    /**
     * State to which the index will be modified
     */
    private String state;

    /**
     * List of states that will be modified
     */
    private List<Include> include;

    public UpdateManagedIndexPolicyConfiguration() {
    }

    public UpdateManagedIndexPolicyConfiguration(String policyId, String state, List<Include> include) {
        this.policyId = policyId;
        this.state = state;
        this.include = include;
    }

    public String getPolicyId() {
        return policyId;
    }

    public void setPolicyId(String policyId) {
        this.policyId = policyId;
    }

    public String getState() {
        return state;
    }

    public void setState(String state) {
        this.state = state;
    }

    public List<Include> getInclude() {
        return include;
    }

    public void setInclude(List<Include> include) {
        this.include = include;
    }

    public static class Include {
        private String state;

        public Include(String state) {
            this.state = state;
        }

        public String getState() {
            return state;
        }

        public void setState(String state) {
            this.state = state;
        }
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private String policyId;
        private String state;
        private final List<Include> includes = new ArrayList<>();

        public Builder withPolicyId(String policyId) {
            this.policyId = policyId;
            return this;
        }

        public Builder destinationState(String state) {
            this.state = state;
            return this;
        }

        public Builder sourceStates(String... states) {
            Arrays.stream(states).forEach(state -> this.includes.add(new Include(state)));
            return this;
        }

        public UpdateManagedIndexPolicyConfiguration build() {
            return new UpdateManagedIndexPolicyConfiguration(policyId, state, includes);
        }
    }
}
