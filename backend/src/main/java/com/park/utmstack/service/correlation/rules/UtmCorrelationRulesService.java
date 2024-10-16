package com.park.utmstack.service.correlation.rules;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRulesFilter;
import com.park.utmstack.domain.network_scan.Property;
import com.park.utmstack.domain.network_scan.enums.PropertyFilter;
import com.park.utmstack.repository.correlation.config.UtmDataTypesRepository;
import com.park.utmstack.repository.correlation.rules.UtmCorrelationRulesRepository;
import com.park.utmstack.service.dto.correlation.UtmCorrelationRulesDTO;
import com.park.utmstack.service.dto.correlation.UtmCorrelationRulesMapper;
import com.park.utmstack.service.network_scan.UtmNetworkScanService;
import com.park.utmstack.web.rest.vm.UtmCorrelationRulesVM;
import io.undertow.util.BadRequestException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.InvalidDataAccessResourceUsageException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.bind.annotation.RequestParam;

import javax.persistence.EntityNotFoundException;
import java.util.*;

/**
 * Service Implementation for managing {@link UtmCorrelationRulesService}.
 */
@Service
public class UtmCorrelationRulesService {
    private final Logger log = LoggerFactory.getLogger(UtmCorrelationRulesService.class);
    private static final String CLASSNAME = "UtmCorrelationRulesService";

    private final UtmCorrelationRulesRepository utmCorrelationRulesRepository;
    private final UtmNetworkScanService utmNetworkScanService;

    private final UtmCorrelationRulesMapper utmCorrelationRulesMapper;

    private final UtmDataTypesRepository utmDataTypesRepository;

    public UtmCorrelationRulesService(UtmCorrelationRulesRepository utmCorrelationRulesRepository,
                                      UtmNetworkScanService utmNetworkScanService,
                                      UtmCorrelationRulesMapper utmCorrelationRulesMapper, UtmDataTypesRepository utmDataTypesRepository) {
        this.utmCorrelationRulesRepository = utmCorrelationRulesRepository;
        this.utmNetworkScanService = utmNetworkScanService;
        this.utmCorrelationRulesMapper = utmCorrelationRulesMapper;
        this.utmDataTypesRepository = utmDataTypesRepository;
    }

    /**
     * Save a correlation rule.
     *
     * @param rule the entity to save.
     * @return the persisted entity.
     */
    public UtmCorrelationRules save(UtmCorrelationRules rule) {
        log.debug("Request to save UtmCorrelationRules : {}", rule);
        if (rule.getId() == null) {
            rule.setId(utmCorrelationRulesRepository.getNextId());
        }

        rule.setDataTypes(this.saveDataTypes(rule));
        rule.setRuleLastUpdate();
        return utmCorrelationRulesRepository.save(rule);
    }

    /**
     * Update correlation rule definition
     *
     * @param correlationRule The rule to update with its relations
     * @throws Exception Bad Request if the rule don't have an id, or is a system rule, or isn't present in database,
     *         or generic error if some error occurs when updating in DB
     * */
    @Transactional
    public void updateRule(UtmCorrelationRules correlationRule) throws Exception {
        final String ctx = CLASSNAME + ".updateRule";
        Long id = correlationRule.getId();
        if (id == null) {
            throw new BadRequestException(ctx + ": The rule must have an id to update.");
        }

        Optional<UtmCorrelationRules> optionalCorrelationRule = utmCorrelationRulesRepository.findById(id);
        if (optionalCorrelationRule.isEmpty()) {
            throw new EntityNotFoundException("Rule with ID " + id + " not found");
        }
        if (correlationRule.getDataTypes().isEmpty()) {
            throw new BadRequestException(ctx + ": The rule must have at least one data type.");
        }
        if(optionalCorrelationRule.get().getSystemOwner()) {
            throw new BadRequestException(ctx + ": System's rules can't be updated.");
        }
        correlationRule.setDataTypes(this.saveDataTypes(correlationRule));
        correlationRule.setRuleLastUpdate();
        utmCorrelationRulesRepository.save(correlationRule);
    }

    /**
     * Activate or deactivate correlation rule
     *
     * @param ruleId The rule's id to activate or deactivate
     * @throws Exception Bad Request if the rule don't have an id, or isn't present in database,
     *         or generic error if some error occurs when updating in DB
     * */
    @Transactional
    public void setRuleActivation(Long ruleId, boolean setActive) throws Exception {
        final String ctx = CLASSNAME + ".setRuleActivation";
        if (ruleId == null) {
            throw new BadRequestException(ctx + ": The rule must have an id to activate or deactivate.");
        }
        Optional<UtmCorrelationRules> find = utmCorrelationRulesRepository.findById(ruleId);
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The rule you're trying to activate or deactivate is not present in database.");
        }
        try {
            UtmCorrelationRules rule = find.get();
            rule.setRuleActive(setActive);
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
        final String ctx = CLASSNAME + ".deleteRule";
        Optional<UtmCorrelationRules> find = utmCorrelationRulesRepository.findById(id);
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The rule you're trying to delete is not present in database.");
        }
        if(find.get().getSystemOwner()) {
            throw new BadRequestException(ctx + ": System's rules can't be removed.");
        }
        utmCorrelationRulesRepository.deleteById(id);
    }

    /**
     * Gets UtmCorrelationRules result by filters
     *
     * @param f Object with all filters to be applied
     * @param p For paginate the result
     * @return A list of {@link UtmCorrelationRulesVM} with results
     * @throws RuntimeException In case of any error
     */
    @Transactional
    public Page<UtmCorrelationRulesDTO> searchByFilters(UtmCorrelationRulesFilter f, Pageable p) throws RuntimeException {
        final String ctx = CLASSNAME + ".searchByFilters";
        try {
            return filter(f, p);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private Page<UtmCorrelationRulesDTO> filter(UtmCorrelationRulesFilter f, Pageable p) throws Exception {
        final String ctx = CLASSNAME + ".filter";
        try {
            Page<UtmCorrelationRules> page = utmCorrelationRulesRepository.searchByFilters(
                    f.getName() == null ? null : "%" + f.getName() + "%",f.getConfidentiality(),f.getIntegrity(),f.getAvailability(),
                    f.getCategory(),f.getTechnique(),f.getActive(),f.getSystemOwner(),f.getDataTypes(),
                    f.getInitDate(),f.getEndDate(), f.getSearch() == null ? null :"%" + f.getSearch() + "%", p );

            List<UtmCorrelationRulesDTO> rulesList = this.utmCorrelationRulesMapper.toListDTO(page.getContent());
            return new PageImpl<>(rulesList, p, page.getTotalElements());
        } catch (InvalidDataAccessResourceUsageException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            throw new Exception(msg);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Search all no repeated values for a property field
     *
     * @param prop     Property field to get the values
     * @param pageable For paginate the result
     * @return A list with values of the property field
     * @throws Exception In case of any error
     */
    public List<?> searchPropertyValues(@RequestParam Property prop, String value, Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".searchPropertyValues";
        try {
            return utmNetworkScanService.searchPropertyValues(prop, value, false, pageable);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get one UtmCorrelationRulesVM by rule id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmCorrelationRules> findOne(Long id) {
        return this.utmCorrelationRulesRepository.findById(id);
    }

    private Set<UtmDataTypes> saveDataTypes(UtmCorrelationRules rule) {
        Set<UtmDataTypes> existingDataTypes = new HashSet<>();
        Set<UtmDataTypes> newDataTypes = new HashSet<>();

        for (UtmDataTypes dataType : rule.getDataTypes()) {
            dataType.setLastUpdate();
            if (dataType.getId() == null || !utmDataTypesRepository.existsById(dataType.getId())) {
                dataType.setSystemOwner(false);
                newDataTypes.add(dataType);
            } else {
                dataType.setSystemOwner(utmDataTypesRepository.findById(dataType.getId()).get().getSystemOwner());
                existingDataTypes.add(dataType);
            }
        }

        if (!newDataTypes.isEmpty()) {
            newDataTypes = new HashSet<>(utmDataTypesRepository.saveAll(newDataTypes));
        }

        existingDataTypes.addAll(newDataTypes);

        return existingDataTypes;
    }

}
