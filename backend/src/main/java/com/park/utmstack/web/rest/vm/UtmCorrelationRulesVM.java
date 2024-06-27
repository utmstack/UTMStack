package com.park.utmstack.web.rest.vm;

import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import com.park.utmstack.domain.correlation.rules.UtmGroupRulesDataType;

import java.util.List;

public class UtmCorrelationRulesVM {
    private UtmCorrelationRules rule;
    private List<UtmGroupRulesDataType> dataTypeRelations;

    public UtmCorrelationRulesVM(){}
    public UtmCorrelationRulesVM(UtmCorrelationRules rule, List<UtmGroupRulesDataType> dataTypeRelations) {
        this.rule = rule;
        this.dataTypeRelations = dataTypeRelations;
    }

    public UtmCorrelationRules getRule() {
        return rule;
    }

    public void setRule(UtmCorrelationRules rule) {
        this.rule = rule;
    }

    public List<UtmGroupRulesDataType> getDataTypeRelations() {
        return dataTypeRelations;
    }

    public void setDataTypeRelations(List<UtmGroupRulesDataType> dataTypeRelations) {
        this.dataTypeRelations = dataTypeRelations;
    }
}
