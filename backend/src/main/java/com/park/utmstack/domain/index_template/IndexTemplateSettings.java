package com.park.utmstack.domain.index_template;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;

@Getter
@JsonInclude(value = JsonInclude.Include.NON_NULL)
public class IndexTemplateSettings {

    @JsonProperty("opendistro.index_state_management.policy_id")
    private String indexPolicyId;

    @JsonProperty("index.mapping.total_fields.limit")
    private Integer totalFieldsLimit;

    @JsonProperty("number_of_shards")
    private Integer numberOfShards;

    @JsonProperty("number_of_replicas")
    private Integer numberOfReplicas;

    public void setIndexPolicyId(String indexPolicyId) {
        this.indexPolicyId = indexPolicyId;
    }

    public void setTotalFieldsLimit(Integer totalFieldsLimit) {
        this.totalFieldsLimit = totalFieldsLimit;
    }

    public void setNumberOfShards(Integer numberOfShards) {
        this.numberOfShards = numberOfShards;
    }

    public void setNumberOfReplicas(Integer numberOfReplicas) {
        this.numberOfReplicas = numberOfReplicas;
    }
}
