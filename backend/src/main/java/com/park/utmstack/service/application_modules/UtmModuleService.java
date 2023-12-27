package com.park.utmstack.service.application_modules;

import com.park.utmstack.domain.UtmMenu;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.repository.UtmModuleGroupRepository;
import com.park.utmstack.repository.application_modules.UtmModuleRepository;
import com.park.utmstack.service.UtmMenuService;
import com.park.utmstack.service.index_pattern.UtmIndexPatternService;
import com.park.utmstack.service.logstash_filter.UtmLogstashFilterService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import java.util.List;
import java.util.Objects;
import java.util.Optional;

/**
 * Service Implementation for managing UtmModule.
 */
@Service
@Transactional
public class UtmModuleService {

    private final Logger log = LoggerFactory.getLogger(UtmModuleService.class);
    private static final String CLASSNAME = "UtmModuleService";

    private final UtmModuleRepository moduleRepository;
    private final UtmMenuService menuService;
    private final UtmIndexPatternService indexPatternService;
    private final UtmLogstashFilterService logstashFilterService;
    private final UtmModuleGroupRepository moduleGroupRepository;

    public UtmModuleService(UtmModuleRepository moduleRepository,
                            UtmMenuService menuService,
                            UtmIndexPatternService indexPatternService,
                            UtmLogstashFilterService logstashFilterService,
                            UtmModuleGroupRepository moduleGroupRepository) {
        this.moduleRepository = moduleRepository;
        this.menuService = menuService;
        this.indexPatternService = indexPatternService;
        this.logstashFilterService = logstashFilterService;
        this.moduleGroupRepository = moduleGroupRepository;
    }

    /**
     * Activate or deactivate the module requested
     *
     * @param nameShort        Short name of the module
     * @param activationStatus Activation status
     * @return The current module information
     * @throws Exception In case of any error
     */
    public UtmModule activateDeactivate(Long serverId, ModuleName nameShort, Boolean activationStatus) throws Exception {
        final String ctx = CLASSNAME + ".activateDeactivate";
        try {
            UtmModule module = moduleRepository.findByServerIdAndModuleName(serverId, nameShort);

            if (Objects.isNull(module))
                throw new Exception(String.format("Definition of the module %1$s not found for the server ID %2$s", nameShort.name(), serverId));

            module.setModuleActive(activationStatus);
            module = moduleRepository.save(module);

            List<ModuleName> nonRemovableConf = List.of(ModuleName.SOC_AI);

            if (!activationStatus && !nonRemovableConf.contains(nameShort))
                moduleGroupRepository.deleteAllByModuleId(module.getId());

            enableDisableModuleMenus(nameShort, activationStatus);
            enableDisableModuleIndexPatterns(nameShort, activationStatus);
            enableDisableModuleFilter(nameShort, activationStatus);
            return module;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private void enableDisableModuleMenus(ModuleName nameShort, Boolean activationStatus) throws Exception {
        final String ctx = CLASSNAME + ".enableDisableModuleMenus";
        try {
            List<UtmMenu> menus = menuService.findAllByModuleNameShort(nameShort.name());

            if (CollectionUtils.isEmpty(menus))
                return;

            Integer moduleInstancesActives = moduleRepository.countAllByModuleNameAndModuleActiveIsTrue(nameShort);

            if ((!activationStatus && moduleInstancesActives > 0) || (activationStatus && moduleInstancesActives > 1))
                return;

            menus.forEach(menu -> menu.setMenuActive(activationStatus));
            menuService.saveAll(menus);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private void enableDisableModuleIndexPatterns(ModuleName nameShort, Boolean activationStatus) throws Exception {
        final String ctx = CLASSNAME + ".enableDisableModuleIndexPatterns";
        try {
            List<UtmIndexPattern> patterns = indexPatternService.findAllByPatternModule(nameShort.name());

            if (CollectionUtils.isEmpty(patterns))
                return;

            Integer moduleInstancesActives = moduleRepository.countAllByModuleNameAndModuleActiveIsTrue(nameShort);

            if ((!activationStatus && moduleInstancesActives > 0) || (activationStatus && moduleInstancesActives > 1))
                return;

            patterns.forEach(pattern -> pattern.setActive(activationStatus));
            indexPatternService.saveAll(patterns);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private void enableDisableModuleFilter(ModuleName nameShort, Boolean activationStatus) throws Exception {
        final String ctx = CLASSNAME + ".enableDisableModuleFilter";
        try {
            List<UtmLogstashFilter> filters = logstashFilterService.findAllByModuleName(nameShort.name());

            if (CollectionUtils.isEmpty(filters))
                return;

            Integer moduleInstancesActives = moduleRepository.countAllByModuleNameAndModuleActiveIsTrue(nameShort);

            if ((!activationStatus && moduleInstancesActives > 0) || (activationStatus && moduleInstancesActives > 1))
                return;

            filters.forEach(filter -> filter.setActive(activationStatus));
            logstashFilterService.saveAll(filters);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Gets all distinct categories from de application modules
     *
     * @return A list of string with all module categories
     * @throws Exception In case of any error
     */
    public List<String> getModuleCategories(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getModuleCategories";
        try {
            return moduleRepository.findModuleCategories(serverId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get all the utmModules.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmModule> findAll(Pageable pageable) {
        log.debug("Request to get all UtmModules");
        return moduleRepository.findAll(pageable);
    }


    /**
     * Get one utmModule by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmModule> findOne(Long id) {
        log.debug("Request to get UtmModule : {}", id);
        return moduleRepository.findById(id);
    }

    public UtmModule findByServerIdAndModuleName(Long serverId, ModuleName shortName) throws Exception {
        final String ctx = CLASSNAME + ".findByServerIdAndModuleName";
        try {
            return moduleRepository.findByServerIdAndModuleName(serverId, shortName);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public boolean isModuleActive(ModuleName shortName) throws Exception {
        final String ctx = CLASSNAME + ".isModuleActive";
        try {
            return moduleRepository.countAllByModuleNameAndModuleActiveIsTrue(shortName) > 0;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
