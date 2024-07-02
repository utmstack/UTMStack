package com.park.utmstack.service.correlation.rules.dto;

import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.springframework.stereotype.Component;

import java.time.Clock;
import java.time.Instant;

@Component
public class UtmCorrelationRulesMapper {

    public UtmCorrelationRulesDTO toDto(UtmCorrelationRules entity) throws UtmSerializationException {
        UtmCorrelationRulesDTO dto = new UtmCorrelationRulesDTO();
        dto.setId(entity.getId());
        dto.setName(entity.getRuleName());
        dto.setConfidentiality(entity.getRuleConfidentiality());
        dto.setIntegrity(entity.getRuleIntegrity());
        dto.setAvailability(entity.getRuleAvailability());
        dto.setCategory(entity.getRuleCategory());
        dto.setTechnique(entity.getRuleTechnique());
        dto.setDescription(entity.getRuleDescription());
        dto.setReferences(entity.getRuleReferences());
        dto.setVariables(entity.getRuleVariables());
        dto.setExpression(entity.getRuleExpression());
        dto.setDataTypes(entity.getDataTypes());
        return dto;
    }

    public UtmCorrelationRules toEntity(UtmCorrelationRulesDTO dto) throws UtmSerializationException {
        UtmCorrelationRules entity = new UtmCorrelationRules();
        entity.setId(dto.getId());
        entity.setRuleName(dto.getName());
        entity.setRuleConfidentiality(dto.getConfidentiality());
        entity.setRuleIntegrity(dto.getIntegrity());
        entity.setRuleAvailability(dto.getAvailability());
        entity.setRuleCategory(dto.getCategory());
        entity.setRuleTechnique(dto.getTechnique());
        entity.setRuleDescription(dto.getDescription());
        entity.setRuleReferences(dto.getReferences());
        entity.setRuleExpression(dto.getExpression());
        entity.setDataTypes(dto.getDataTypes());
        entity.setRuleActive(false);
        entity.setSystemOwner(false);
        entity.setRuleLastUpdate(Instant.now(Clock.systemUTC()));
        entity.setRuleVariables(dto.getVariables());

        return entity;
    }
}
