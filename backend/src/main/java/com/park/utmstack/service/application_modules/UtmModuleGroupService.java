package com.park.utmstack.service.application_modules;

import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.repository.UtmModuleGroupRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmConfigurationGroup.
 */
@Service
@Transactional
public class UtmModuleGroupService {

    private static final String CLASSNAME = "UtmModuleGroupService";
    private final Logger log = LoggerFactory.getLogger(UtmModuleGroupService.class);

    private final UtmModuleGroupRepository moduleGroupRepository;

    public UtmModuleGroupService(UtmModuleGroupRepository moduleGroupRepository) {
        this.moduleGroupRepository = moduleGroupRepository;
    }

    /**
     * Save a utmConfigurationGroup.
     *
     * @param utmModuleGroup the entity to save
     * @return the persisted entity
     */
    public UtmModuleGroup save(UtmModuleGroup utmModuleGroup) {
        log.debug("Request to save UtmConfigurationGroup : {}", utmModuleGroup);
        return moduleGroupRepository.save(utmModuleGroup);
    }

    /**
     * Get all the utmConfigurationGroups.
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public List<UtmModuleGroup> findAll() {
        log.debug("Request to get all UtmConfigurationGroups");
        return moduleGroupRepository.findAll();
    }


    /**
     * Get one utmConfigurationGroup by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmModuleGroup> findOne(Long id) {
        log.debug("Request to get UtmConfigurationGroup : {}", id);
        return moduleGroupRepository.findById(id);
    }

    /**
     * Delete the utmConfigurationGroup by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmConfigurationGroup : {}", id);
        moduleGroupRepository.deleteById(id);
    }

    public void deleteAllByModuleId(Long id) {
        moduleGroupRepository.deleteAllByModuleId(id);
    }

    public List<UtmModuleGroup> findAllByModuleId(Long moduleId) throws Exception {
        final String ctx = CLASSNAME + ".findAllByModuleName";
        try {
            return moduleGroupRepository.findAllByModuleId(moduleId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
