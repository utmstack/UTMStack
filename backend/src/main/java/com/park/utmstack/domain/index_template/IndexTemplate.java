package com.park.utmstack.domain.index_template;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

@JsonInclude(JsonInclude.Include.NON_NULL)
public class IndexTemplate {

    @JsonProperty("index_patterns")
    private List<String> indexPatterns;
    private Integer priority;
    private Template template;

    public IndexTemplate() {
    }

    private IndexTemplate(List<String> indexPatterns, Integer priority, Template template) {
        this.indexPatterns = indexPatterns;
        this.priority = priority;
        this.template = template;
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

    public Template getTemplate() {
        return template;
    }

    public void setTemplate(Template template) {
        this.template = template;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private final List<String> indexPatterns = new ArrayList<>();
        private Integer priority;
        private final Template template = new Template();

        public Builder withIndexPattern(String... indexPatterns) {
            this.indexPatterns.addAll(Arrays.asList(indexPatterns));
            return this;
        }

        public Builder withPriority(Integer priority) {
            this.priority = priority;
            return this;
        }

        public Builder withIndexStateManagementPolicyIdSetting(String indexPolicyId) {
            this.template.getSettings().setIndexPolicyId(indexPolicyId);
            return this;
        }

        public Builder withTotalFieldsLimit(Integer limit) {
            this.template.getSettings().setTotalFieldsLimit(limit);
            return this;
        }

        public Builder withNumberOfShards(Integer shards) {
            this.template.getSettings().setNumberOfShards(shards);
            return this;
        }

        public Builder withNumberOfReplicas(Integer replicas) {
            this.template.getSettings().setNumberOfReplicas(replicas);
            return this;
        }

        public IndexTemplate build() {
            return new IndexTemplate(indexPatterns, priority, template);
        }
    }
}
