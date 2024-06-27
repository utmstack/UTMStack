package com.park.utmstack.service.correlation.rules;

import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import com.park.utmstack.domain.correlation.rules.UtmGroupRulesDataType;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;
import com.park.utmstack.repository.correlation.rules.UtmCorrelationRulesRepository;
import com.park.utmstack.repository.correlation.rules.UtmGroupRulesDataTypeRepository;
import com.park.utmstack.web.rest.vm.UtmCorrelationRulesVM;
import io.undertow.util.BadRequestException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing {@link UtmCorrelationRulesService}.
 */
@Service
public class UtmCorrelationRulesService {
    private final Logger log = LoggerFactory.getLogger(UtmCorrelationRulesService.class);
    private static final String CLASSNAME = "UtmCorrelationRulesService";

    private final UtmCorrelationRulesRepository utmCorrelationRulesRepository;
    private final UtmGroupRulesDataTypeRepository utmGroupRulesDataTypeRepository;

    public UtmCorrelationRulesService(UtmCorrelationRulesRepository utmCorrelationRulesRepository, UtmGroupRulesDataTypeRepository utmGroupRulesDataTypeRepository) {
        this.utmCorrelationRulesRepository = utmCorrelationRulesRepository;
        this.utmGroupRulesDataTypeRepository = utmGroupRulesDataTypeRepository;
    }

    /**
     * Save a utmLogstashPipeline.
     *
     * @param rule the entity to save.
     * @return the persisted entity.
     */
    public UtmCorrelationRules save(UtmCorrelationRules rule) {
        log.debug("Request to save UtmCorrelationRules : {}", rule);
        if (rule.getId() == null) {
            rule.setId(utmCorrelationRulesRepository.getNextId());
        }
        return utmCorrelationRulesRepository.save(rule);
    }

    /**
     * Add a correlation rule definition
     *
     * @param rulesVM VM with rule and its relations
     * @throws Exception Bad Request if the rule has an id or generic if some error occurs when inserting in DB
     * */
    @Transactional
    public void addRule(UtmCorrelationRulesVM rulesVM) throws Exception {
        final String ctx = CLASSNAME + ".addRule";
        if (rulesVM.getRule().getId() != null) {
            throw new BadRequestException(ctx + ": A new rule can't have an id.");
        }
        try {
            UtmCorrelationRules rule = rulesVM.getRule();
            rule.setSystemOwner(false);
            List<UtmGroupRulesDataType> dataTypes = rulesVM.getDataTypeRelations();

            // Saving relations with datatypes
            utmGroupRulesDataTypeRepository.saveAll(dataTypes);
            this.save(rule);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding a rule.", ex);
        }
    }

    /**
     * Update correlation rule definition
     *
     * @param rulesVM The rule to update with its relations
     * @throws Exception Bad Request if the rule don't have an id, or is a system rule, or isn't present in database,
     *         or generic error if some error occurs when updating in DB
     * */
    @Transactional
    public void updateRule(UtmCorrelationRulesVM rulesVM) throws Exception {
        final String ctx = CLASSNAME + ".updateRule";
        if (rulesVM.getRule().getId() == null) {
            throw new BadRequestException(ctx + ": The rule must have an id to update.");
        }
        Optional<UtmCorrelationRules> find = utmCorrelationRulesRepository.findById(rulesVM.getRule().getId());
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The rule you're trying to update is not present in database.");
        }
        if(find.get().getSystemOwner()) {
            throw new BadRequestException(ctx + ": System's rules can't be updated.");
        }
        try {
            UtmCorrelationRules rule = rulesVM.getRule();

            List<UtmGroupRulesDataType> dataTypesCurrent = utmGroupRulesDataTypeRepository.findByRuleId(rule.getId());
            List<UtmGroupRulesDataType> dataTypesUpdated = rulesVM.getDataTypeRelations();

            // Removing deleted relations
            /*List<UtmGroupRulesDataType> dataTypesDeleted = new ArrayList<>();
            dataTypesDeleted.addAll(
                    dataTypesCurrent.stream().filter(f->dataTypesUpdated.stream()
                            .filter(d->d.getId()==f.getId()).findFirst().isPresent()).collect(Collectors.toList())*/
            //);
            utmGroupRulesDataTypeRepository.deleteAll(dataTypesCurrent.stream().filter(f-> dataTypesUpdated.stream()
                    .noneMatch(d-> Objects.equals(d.getId(), f.getId()))).collect(Collectors.toList()));
            // Saving relations with datatypes
            utmGroupRulesDataTypeRepository.saveAll(dataTypesUpdated);
            this.save(rule);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding a rule.", ex);
        }
    }

    /**
     * Remove correlation rule from database
     *
     * @param id The id of the rule to remove
     * @throws Exception if the rule can't be removed from database
     * */
    @Transactional
    public void deleteRule (Long id) throws Exception {
        utmCorrelationRulesRepository.deleteById(id);
    }


}
