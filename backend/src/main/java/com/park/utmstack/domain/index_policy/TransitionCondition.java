package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

public class TransitionCondition {

    /**
     * The minimum age of the index required to transition
     * Example: 30d
     */
    @SerializedName(value = "min_index_age")
    private String minIndexAge;

    public TransitionCondition() {
    }

    public TransitionCondition(String minIndexAge) {
        this.minIndexAge = minIndexAge;
    }

    public String getMinIndexAge() {
        return minIndexAge;
    }

    public void setMinIndexAge(String minIndexAge) {
        this.minIndexAge = minIndexAge;
    }
}
