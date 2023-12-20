package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

/**
 * Reduces the number of Lucene segments by merging the segments of individual shards.
 * This operation attempts to set the index to a read-only state before starting the merging process
 */
public class ActionForceMerge {
    /**
     * The number of segments to reduce the shard to
     * Required: Yes
     */
    @SerializedName(value = "max_num_segments")
    private Integer maxNumSegments;

    public ActionForceMerge() {
    }

    public ActionForceMerge(Integer maxNumSegments) {
        this.maxNumSegments = maxNumSegments;
    }

    public Integer getMaxNumSegments() {
        return maxNumSegments;
    }

    public void setMaxNumSegments(Integer maxNumSegments) {
        this.maxNumSegments = maxNumSegments;
    }
}
