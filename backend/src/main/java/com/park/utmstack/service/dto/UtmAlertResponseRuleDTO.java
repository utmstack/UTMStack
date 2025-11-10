package com.park.utmstack.service.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseActionTemplate;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRule;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.validation.constraints.*;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

@Data
@NoArgsConstructor
public class UtmAlertResponseRuleDTO {

    private Long id;

    @NotBlank
    @Size(max = 150)
    private String name;

    @Size(max = 512)
    private String description;

    @NotEmpty
    private List<FilterType> conditions;

    @NotBlank
    private String command;

    @NotNull
    private Boolean active;

    @NotBlank
    private String agentPlatform;

    @Size(max = 500)
    private String defaultAgent;

    private List<String> excludedAgents = new ArrayList<>();

    @JsonProperty(access = JsonProperty.Access.READ_ONLY)
    private String createdBy;

    @JsonProperty(access = JsonProperty.Access.READ_ONLY)
    private Instant createdDate;

    @JsonProperty(access = JsonProperty.Access.READ_ONLY)
    private String lastModifiedBy;

    @JsonProperty(access = JsonProperty.Access.READ_ONLY)
    private Instant lastModifiedDate;

    private List<UtmAlertResponseActionTemplateDTO> actions;

    private Boolean systemOwner;

    public UtmAlertResponseRuleDTO(UtmAlertResponseRule rule) {
        this.id = rule.getId();
        this.name = rule.getRuleName();
        this.description = rule.getRuleDescription();
        this.conditions = new Gson().fromJson(
                rule.getRuleConditions(),
                TypeToken.getParameterized(List.class, FilterType.class).getType()
        );
        this.command = rule.getRuleCmd();
        this.active = rule.getRuleActive();
        this.agentPlatform = rule.getAgentPlatform();
        this.defaultAgent = rule.getDefaultAgent();
        if (StringUtils.hasText(rule.getExcludedAgents())) {
            this.excludedAgents.addAll(Arrays.asList(rule.getExcludedAgents().split(",")));
        }
        this.createdBy = rule.getCreatedBy();
        this.createdDate = rule.getCreatedDate();
        this.lastModifiedBy = rule.getLastModifiedBy();
        this.lastModifiedDate = rule.getLastModifiedDate();
        this.systemOwner = rule.getSystemOwner();

        if (rule.getUtmAlertResponseActionTemplates() != null) {
            this.actions = rule.getUtmAlertResponseActionTemplates()
                    .stream()
                    .map(template -> {
                        UtmAlertResponseActionTemplateDTO dto = new UtmAlertResponseActionTemplateDTO();
                        dto.setId(template.getId());
                        dto.setTitle(template.getTitle());
                        dto.setDescription(template.getDescription());
                        dto.setCommand(template.getCommand());
                        return dto;
                    })
                    .collect(Collectors.toList());
        }

    }

}
