package com.park.utmstack.web.rest.vm;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.park.utmstack.domain.UtmAlertTag;
import com.park.utmstack.domain.UtmAlertTagRule;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import org.springframework.util.CollectionUtils;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotEmpty;
import java.time.Instant;
import java.util.List;
import java.util.stream.Collectors;

public class AlertTagRuleVM {
    private Long id;
    @NotBlank
    private String name;
    @NotBlank
    private String description;
    @NotEmpty
    private List<FilterType> conditions;
    @NotEmpty
    private List<AlertTagVM> tags;
    private Boolean active;
    private Boolean deleted;
    private String createdBy;
    private Instant createdDate;
    private String lastModifiedBy;
    private Instant lastModifiedDate;

    public AlertTagRuleVM() {
    }

    public AlertTagRuleVM(UtmAlertTagRule rule, List<UtmAlertTag> tags) throws JsonProcessingException {
        this.id = rule.getId();
        this.name = rule.getRuleName();
        this.description = rule.getRuleDescription();
        this.active = rule.getRuleActive();
        this.deleted = rule.getRuleDeleted();
        this.createdBy = rule.getCreatedBy();
        this.createdDate = rule.getCreatedDate();
        this.lastModifiedBy = rule.getLastModifiedBy();
        this.lastModifiedDate = rule.getLastModifiedDate();
        this.tags = !CollectionUtils.isEmpty(tags) ? tags.stream().map(AlertTagVM::new).collect(Collectors.toList()) : null;

        ObjectMapper om = new ObjectMapper();
        this.conditions = om.readValue(rule.getRuleConditions(), new TypeReference<>() {
        });
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public List<FilterType> getConditions() {
        return conditions;
    }

    public void setConditions(List<FilterType> conditions) {
        this.conditions = conditions;
    }

    public List<AlertTagVM> getTags() {
        return tags;
    }

    public void setTags(List<AlertTagVM> tags) {
        this.tags = tags;
    }

    public Boolean getActive() {
        return active;
    }

    public void setActive(Boolean active) {
        this.active = active;
    }

    public Boolean getDeleted() {
        return deleted;
    }

    public void setDeleted(Boolean deleted) {
        this.deleted = deleted;
    }

    public String getCreatedBy() {
        return createdBy;
    }

    public void setCreatedBy(String createdBy) {
        this.createdBy = createdBy;
    }

    public Instant getCreatedDate() {
        return createdDate;
    }

    public void setCreatedDate(Instant createdDate) {
        this.createdDate = createdDate;
    }

    public String getLastModifiedBy() {
        return lastModifiedBy;
    }

    public void setLastModifiedBy(String lastModifiedBy) {
        this.lastModifiedBy = lastModifiedBy;
    }

    public Instant getLastModifiedDate() {
        return lastModifiedDate;
    }

    public void setLastModifiedDate(Instant lastModifiedDate) {
        this.lastModifiedDate = lastModifiedDate;
    }
}
