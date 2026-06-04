package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

public class IndexPolicy {

    @SerializedName(value = "_id")
    private String id;

    @SerializedName(value = "_primary_term")
    private Integer primaryTerm;

    @SerializedName(value = "_seq_no")
    private Integer seqNo;

    private Policy policy;

    public IndexPolicy() {
    }

    public IndexPolicy(Policy policy) {
        this.policy = policy;
    }

    public Policy getPolicy() {
        return policy;
    }

    public void setPolicy(Policy policy) {
        this.policy = policy;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public Integer getPrimaryTerm() {
        return primaryTerm;
    }

    public void setPrimaryTerm(Integer primaryTerm) {
        this.primaryTerm = primaryTerm;
    }

    public Integer getSeqNo() {
        return seqNo;
    }

    public void setSeqNo(Integer seqNo) {
        this.seqNo = seqNo;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private Policy policy;

        public Builder withPolicy(Policy policy) {
            this.policy = policy;
            return this;
        }

        public IndexPolicy build() {
            return new IndexPolicy(policy);
        }
    }


}
