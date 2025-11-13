package com.park.utmstack.domain.alert_response_rule;


import com.google.gson.Gson;
import com.park.utmstack.service.dto.UtmAlertResponseRuleDTO;
import lombok.Getter;
import lombok.Setter;
import org.hibernate.annotations.GenericGenerator;
import org.springframework.data.annotation.CreatedBy;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedBy;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;
import org.springframework.util.CollectionUtils;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Entity
@Table(name = "utm_alert_response_rule")
@Getter
@Setter
@EntityListeners(AuditingEntityListener.class)
public class UtmAlertResponseRule implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;
    @Column(name = "rule_name", length = 150, nullable = false)
    private String ruleName;
    @Column(name = "rule_description", length = 512)
    private String ruleDescription;
    @Column(name = "rule_conditions", nullable = false)
    private String ruleConditions;
    @Column(name = "rule_cmd", nullable = false)
    private String ruleCmd;
    @Column(name = "rule_active", nullable = false)
    private Boolean ruleActive;
    @Column(name = "agent_platform")
    private String agentPlatform;
    @Column(name = "excluded_agents")
    private String excludedAgents;
    @Size(max = 500)
    @Column(name = "default_agent" , length = 500)
    private String defaultAgent;
    @CreatedBy
    @Column(name = "created_by", nullable = false, length = 50, updatable = false)
    private String createdBy;
    @CreatedDate
    @Column(name = "created_date", updatable = false)
    private Instant createdDate;
    @LastModifiedBy
    @Column(name = "last_modified_by", length = 50)
    private String lastModifiedBy;
    @LastModifiedDate
    @Column(name = "last_modified_date")
    private Instant lastModifiedDate;

    @Column(name = "system_owner", nullable = false)
    private Boolean systemOwner;

    @OneToMany(mappedBy = "rule", fetch = FetchType.LAZY, cascade = CascadeType.ALL, orphanRemoval = true)
    List<UtmAlertResponseRuleExecution> utmAlertResponseRuleExecutions;

    @ManyToMany(cascade = {CascadeType.PERSIST, CascadeType.MERGE})
    @JoinTable(
            name = "utm_alert_response_rule_template",
            joinColumns = @JoinColumn(name = "rule_id"),
            inverseJoinColumns = @JoinColumn(name = "template_id")
    )
    private List<UtmAlertResponseActionTemplate> utmAlertResponseActionTemplates = new ArrayList<>();



    public UtmAlertResponseRule() {
    }

    public UtmAlertResponseRule(UtmAlertResponseRuleDTO dto) {
        this.id = dto.getId();
        this.ruleName = dto.getName();
        this.ruleDescription = dto.getDescription();
        this.ruleConditions = new Gson().toJson(dto.getConditions());
        this.ruleCmd = dto.getCommand();
        this.ruleActive = dto.getActive();
        this.agentPlatform = dto.getAgentPlatform();
        this.defaultAgent = dto.getDefaultAgent();
        this.systemOwner = dto.getSystemOwner();
        if (!CollectionUtils.isEmpty(dto.getExcludedAgents()))
            this.excludedAgents = String.join(",", dto.getExcludedAgents());
        else
            this.excludedAgents = null;


        if (dto.getActions() != null) {
            this.utmAlertResponseActionTemplates = dto.getActions()
                    .stream()
                    .map(templateDto -> {
                        UtmAlertResponseActionTemplate template = new UtmAlertResponseActionTemplate();
                        template.setId(templateDto.getId());
                        template.setTitle(templateDto.getTitle());
                        template.setDescription(templateDto.getDescription());
                        template.setCommand(templateDto.getCommand());
                        template.setSystemOwner(false);
                        return template;
                    })
                    .collect(Collectors.toList());
        }
    }

}
