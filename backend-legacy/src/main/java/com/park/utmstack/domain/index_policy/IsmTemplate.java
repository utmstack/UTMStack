package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

import java.util.ArrayList;
import java.util.List;

/**
 * Specify an ISM template pattern that matches the index to apply the policy.
 */
public class IsmTemplate {
    @SerializedName(value = "index_patterns")
    private List<String> indexPatterns;
    private Integer priority;

    public IsmTemplate() {
    }

    public List<String> getIndexPatterns() {
        return indexPatterns;
    }

    public void setIndexPatterns(List<String> indexPatterns) {
        this.indexPatterns = indexPatterns;
    }

    public Integer getPriority() {
        return priority;
    }

    public void setPriority(Integer priority) {
        this.priority = priority;
    }

    public IsmTemplate(List<String> indexPatterns, Integer priority) {
        this.indexPatterns = indexPatterns;
        this.priority = priority;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private final List<String> indexPatterns = new ArrayList<>();
        private Integer priority;

        public Builder withIndexPattern(String indexPattern) {
            this.indexPatterns.add(indexPattern);
            return this;
        }

        public Builder withPriority(Integer priority) {
            this.priority = priority;
            return this;
        }

        public IsmTemplate build() {
            return new IsmTemplate(indexPatterns, priority);
        }
    }
}
