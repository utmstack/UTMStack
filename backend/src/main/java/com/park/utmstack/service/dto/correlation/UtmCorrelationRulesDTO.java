package com.park.utmstack.service.dto.correlation;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.rules.AfterEvents;
import com.park.utmstack.domain.correlation.rules.RuleDefinition;
import com.park.utmstack.domain.correlation.rules.SearchRequest;
import lombok.Data;
import lombok.Getter;
import lombok.Setter;
import org.hibernate.validator.constraints.URL;

import javax.validation.constraints.*;
import java.io.Serializable;
import java.util.List;
import java.util.Set;

@Data
public class UtmCorrelationRulesDTO implements Serializable {

    private static final long serialVersionUID = 1L;

    private Long id;

    @NotBlank
    private String name;

    @Min(value = 0)
    @Max(value = 3)
    private Integer confidentiality;

    @Min(value = 0)
    @Max(value = 3)
    private Integer integrity;

    @Min(value = 0)
    @Max(value = 3)
    private Integer availability;

    @NotBlank
    private String category;

    @NotNull
    private AdversaryType adversary;

    @NotBlank
    private String technique;

    private String description;

    private List<@URL(message = "Reference must be a valid URL ") String>references;

    @NotEmpty
    private Set<UtmDataTypes> dataTypes;

    private RuleDefinition definition;

    private Boolean systemOwner;

    private Boolean ruleActive;

    private List<SearchRequest> afterEvents;

    private List<String> deduplicateBy;

}

