package com.park.utmstack.domain.index_policy;

import com.fasterxml.jackson.annotation.JsonProperty;

public class IndexPoliciesUpdated {
    @JsonProperty(value = "_id")
    private String id;

    @JsonProperty(value = "_primary_term")
    private Integer primaryTerm;

    @JsonProperty(value = "_seq_no")
    private Integer seqNo;

    private Policy policy;

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

    public Policy getPolicy() {
        return policy;
    }

    public void setPolicy(Policy policy) {
        this.policy = policy;
    }

    public static class Policy {
        private Policy policy;

        public Policy getPolicy() {
            return policy;
        }

        public void setPolicy(Policy policy) {
            this.policy = policy;
        }
    }
}
