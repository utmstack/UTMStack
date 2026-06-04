package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

public class FailedIndices {
    @SerializedName("index_name")
    private String indexName;
    @SerializedName("index_uuid")
    private String indexUuid;
    private String reason;

    public String getIndexName() {
        return indexName;
    }

    public void setIndexName(String indexName) {
        this.indexName = indexName;
    }

    public String getIndexUuid() {
        return indexUuid;
    }

    public void setIndexUuid(String indexUuid) {
        this.indexUuid = indexUuid;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }
}
