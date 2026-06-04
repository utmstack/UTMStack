package com.park.utmstack.domain.correlation.rules;

import lombok.Getter;
import lombok.Setter;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotEmpty;
import java.util.List;

@Setter
@Getter
public class RuleDefinition {
    @NotEmpty
    private List<RuleVariable> ruleVariables;
    @NotBlank
    private String ruleExpression;

    public RuleDefinition(List<RuleVariable> ruleVariables, String ruleExpression) {
        this.ruleVariables = ruleVariables;
        this.ruleExpression = ruleExpression;
    }
    public RuleDefinition(){}

}
