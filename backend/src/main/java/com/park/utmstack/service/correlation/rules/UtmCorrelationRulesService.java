package com.park.utmstack.service.correlation.rules;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRulesFilter;
import com.park.utmstack.domain.correlation.rules.UtmGroupRulesDataType;
import com.park.utmstack.domain.network_scan.enums.PropertyFilter;
import com.park.utmstack.repository.correlation.rules.UtmCorrelationRulesRepository;
import com.park.utmstack.repository.correlation.rules.UtmGroupRulesDataTypeRepository;
import com.park.utmstack.service.network_scan.UtmNetworkScanService;
import com.park.utmstack.web.rest.vm.UtmCorrelationRulesVM;
import io.undertow.util.BadRequestException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.support.PagedListHolder;
import org.springframework.dao.InvalidDataAccessResourceUsageException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.support.PageableExecutionUtils;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
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
    private final UtmNetworkScanService utmNetworkScanService;

    public UtmCorrelationRulesService(UtmCorrelationRulesRepository utmCorrelationRulesRepository, UtmGroupRulesDataTypeRepository utmGroupRulesDataTypeRepository, UtmNetworkScanService utmNetworkScanService) {
        this.utmCorrelationRulesRepository = utmCorrelationRulesRepository;
        this.utmGroupRulesDataTypeRepository = utmGroupRulesDataTypeRepository;
        this.utmNetworkScanService = utmNetworkScanService;
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
        rule.setRuleLastUpdate();
        return utmCorrelationRulesRepository.save(rule);
    }

    /**
     * Add a correlation rule definition
     *
     * @param rulesVM VM with rule and its relations
     * @throws BadRequestException Bad Request if the rule has an id or generic if some error occurs when inserting in DB
     * */
    @Transactional
    public void addRule(UtmCorrelationRulesVM rulesVM) throws BadRequestException {
        final String ctx = CLASSNAME + ".addRule";
        if (rulesVM.getRule().getId() != null) {
            throw new BadRequestException(ctx + ": A new rule can't have an id.");
        }
        try {
            UtmCorrelationRules rule = rulesVM.getRule();
            rule.setSystemOwner(false);

            // Saving relations with datatypes
            Long ruleId = this.save(rule).getId();
            List<UtmGroupRulesDataType> dataTypes = rulesVM.getDataTypeRelations();
            dataTypes.forEach(d-> {
                d.setRuleId(ruleId);
                d.setLastUpdate();
            });
            utmGroupRulesDataTypeRepository.saveAll(dataTypes);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding a rule.", ex);
        }
    }

    /**
     * Update correlation rule definition
     *
     * @param rulesVM The rule to update with its relations
     * @throws BadRequestException Bad Request if the rule don't have an id, or is a system rule, or isn't present in database,
     *         or generic error if some error occurs when updating in DB
     * */
    @Transactional
    public void updateRule(UtmCorrelationRulesVM rulesVM) throws BadRequestException {
        final String ctx = CLASSNAME + ".updateRule";
        if (rulesVM.getRule().getId() == null) {
            throw new BadRequestException(ctx + ": The rule must have an id to update.");
        }
        Optional<UtmCorrelationRules> find = utmCorrelationRulesRepository.findById(rulesVM.getRule().getId());
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The rule you're trying to update is not present in database.");
        }
        if (rulesVM.getDataTypeRelations().isEmpty()) {
            throw new BadRequestException(ctx + ": The rule must have at least one data type.");
        }
        if(find.get().getSystemOwner()) {
            throw new BadRequestException(ctx + ": System's rules can't be updated.");
        }
        try {
            UtmCorrelationRules rule = rulesVM.getRule();

            List<UtmGroupRulesDataType> dataTypesCurrent = utmGroupRulesDataTypeRepository.findByRuleId(rule.getId());
            List<UtmGroupRulesDataType> dataTypesUpdated = rulesVM.getDataTypeRelations();

            // Removing deleted relations
            utmGroupRulesDataTypeRepository.deleteAll(dataTypesCurrent.stream().filter(f-> dataTypesUpdated.stream()
                    .noneMatch(d-> Objects.equals(d.getId(), f.getId()))).collect(Collectors.toList()));
            // Saving relations with datatypes
            dataTypesUpdated.forEach(d-> {
                d.setRuleId(rule.getId());
                d.setLastUpdate();
            });
            utmGroupRulesDataTypeRepository.saveAll(dataTypesUpdated);
            this.save(rule);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding a rule.", ex);
        }
    }

    /**
     * Activate or deactivate correlation rule
     *
     * @param rulesVM The rule to activate or deactivate
     * @throws BadRequestException Bad Request if the rule don't have an id, or isn't present in database,
     *         or generic error if some error occurs when updating in DB
     * */
    @Transactional
    public void setRuleActivation(UtmCorrelationRulesVM rulesVM, boolean setActive) throws BadRequestException {
        final String ctx = CLASSNAME + ".updateRule";
        if (rulesVM.getRule().getId() == null) {
            throw new BadRequestException(ctx + ": The rule must have an id to activate or deactivate.");
        }
        Optional<UtmCorrelationRules> find = utmCorrelationRulesRepository.findById(rulesVM.getRule().getId());
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
    public Page<UtmCorrelationRulesVM> searchByFilters(UtmCorrelationRulesFilter f, Pageable p) throws RuntimeException {
        final String ctx = CLASSNAME + ".searchByFilters";
        try {
            return filter(f, p);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private Page<UtmCorrelationRulesVM> filter(UtmCorrelationRulesFilter f, Pageable p) throws Exception {
        final String ctx = CLASSNAME + ".filter";
        try {
            List<UtmCorrelationRules> rulesList = utmCorrelationRulesRepository.searchByFilters(
                    f.getRuleName() == null ? null : "%" + f.getRuleName() + "%",f.getRuleConfidentiality(),f.getRuleIntegrity(),f.getRuleAvailability(),
                    f.getRuleCategory(),f.getRuleTechnique(),f.getRuleActive(),f.getSystemOwner(),f.getDataTypes(),
                    f.getRuleInitDate(),f.getRuleEndDate());

            List<UtmCorrelationRulesVM> rulesVMList = new ArrayList<>();
            if (!rulesList.isEmpty()) {
                rulesList.forEach(l -> {
                    UtmCorrelationRulesVM vm = new UtmCorrelationRulesVM();
                    vm.setRule(l);
                    vm.setDataTypeRelations(utmGroupRulesDataTypeRepository.findByRuleId(l.getId()));
                    rulesVMList.add(vm);
                });
            }
            PagedListHolder<UtmCorrelationRulesVM> pageDefinition = new PagedListHolder<>();
            pageDefinition.setSource(rulesVMList);
            pageDefinition.setPageSize(p.getPageSize());
            pageDefinition.setPage(p.getPageNumber());
            return PageableExecutionUtils.getPage(pageDefinition.getPageList(), p, rulesList::size);
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
    public List<?> searchPropertyValues(PropertyFilter prop, String value, Pageable pageable) throws Exception {
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
    public Optional<UtmCorrelationRulesVM> findOne(Long id) {
        final String ctx = CLASSNAME + ".findOne";
        try {
            UtmCorrelationRulesVM vm = new UtmCorrelationRulesVM();
            Optional<UtmCorrelationRules> optVm = utmCorrelationRulesRepository.findById(id);
            if (optVm.isPresent()) {
                vm.setRule(optVm.get());
                vm.setDataTypeRelations(utmGroupRulesDataTypeRepository.findByRuleId(optVm.get().getId()));
                return Optional.of(vm);
            }

            return Optional.empty();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /*public Page<UtmDataTypes> findAll(Pageable p) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return utmCorrelationRulesRepository.findAll(p);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }*/


}
