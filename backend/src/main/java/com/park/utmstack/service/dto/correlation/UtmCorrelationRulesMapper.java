package com.park.utmstack.service.dto.correlation;

import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.time.Clock;
import java.time.Instant;
import java.util.List;
import java.util.stream.Collectors;

@Component
public class UtmCorrelationRulesMapper {

    private final Logger logger = LoggerFactory.getLogger(UtmCorrelationRulesMapper.class);

    public UtmCorrelationRulesDTO toDto(UtmCorrelationRules entity) {
        try {
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
            dto.setDefinition(entity.getRuleDefinition());
            dto.setDataTypes(entity.getDataTypes());
            dto.setSystemOwner(entity.getSystemOwner());
            dto.setRuleActive(entity.getRuleActive());
            return dto;
        } catch (UtmSerializationException e) {
            logger.error("Error serializing rule references", e);
            throw new RuntimeException("Error serializing rule references", e);
        }
    }

    public UtmCorrelationRules toEntity(UtmCorrelationRulesDTO dto) {
        try {
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
            entity.setRuleDefinition(dto.getDefinition());
            entity.setDataTypes(dto.getDataTypes());
            entity.setRuleActive(dto.getRuleActive() == null || dto.getRuleActive());
            entity.setSystemOwner(false);
            entity.setRuleLastUpdate(Instant.now(Clock.systemUTC()));

            return entity;
        } catch (UtmSerializationException e) {
            logger.error("Error serializing rule references", e);
            throw new RuntimeException("Error serializing rule references", e);
        }
    }

    public List<UtmCorrelationRulesDTO> toListDTO(List<UtmCorrelationRules> rules) {
        return rules.stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }
}
