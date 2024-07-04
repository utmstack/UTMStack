package com.park.utmstack.domain.correlation.rules;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotEmpty;
import java.util.List;

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

    public List<RuleVariable> getRuleVariables() {
        return ruleVariables;
    }

    public void setRuleVariables(List<RuleVariable> ruleVariables) {
        this.ruleVariables = ruleVariables;
    }

    public String getRuleExpression() {
        return ruleExpression;
    }

    public void setRuleExpression(String ruleExpression) {
        this.ruleExpression = ruleExpression;
    }
}
