package com.park.utmstack.service;

import com.park.utmstack.domain.UtmServerModule;
import com.park.utmstack.repository.UtmServerModuleRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.Collections;
import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmServerModule.
 */
@Service
@Transactional
public class UtmServerModuleService {

    private final Logger log = LoggerFactory.getLogger(UtmServerModuleService.class);

    private final UtmServerModuleRepository utmServerModuleRepository;

    public UtmServerModuleService(UtmServerModuleRepository utmServerModuleRepository) {
        this.utmServerModuleRepository = utmServerModuleRepository;
    }

    /**
     * Save a utmServerModule.
     *
     * @param utmServerModule the entity to save
     * @return the persisted entity
     */
    public UtmServerModule save(UtmServerModule utmServerModule) {
        log.debug("Request to save UtmServerModule : {}", utmServerModule);
        return utmServerModuleRepository.save(utmServerModule);
    }

    /**
     * Save a list of utmServerModule.
     *
     * @param modules the entity to save
     */
    public void saveAll(List<UtmServerModule> modules) {
        log.debug("Request to save UtmServerModule : {}", modules);
        utmServerModuleRepository.saveAll(modules);
    }

    /**
     * Get all the utmServerModules.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmServerModule> findAll(Pageable pageable) {
        log.debug("Request to get all UtmServerModules");
        return utmServerModuleRepository.findAll(pageable);
    }


    /**
     * Get one utmServerModule by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmServerModule> findOne(Long id) {
        log.debug("Request to get UtmServerModule : {}", id);
        return utmServerModuleRepository.findById(id);
    }

    /**
     * Delete the utmServerModule by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmServerModule : {}", id);
        utmServerModuleRepository.deleteById(id);
    }

    /**
     * Only gets those server modules that have integrations
     *
     * @return A list of {@link UtmServerModule}
     */
    public List<UtmServerModule> getModulesWithIntegrations(Long serverId, String prettyName) {
        List<UtmServerModule> modules = utmServerModuleRepository.getModulesWithIntegrations(serverId,
            StringUtils.hasText(prettyName) ? "'%" + prettyName + "%'" : null);
        return !CollectionUtils.isEmpty(modules) ? modules : Collections.emptyList();
    }

    public List<UtmServerModule> findAllByModuleName(String moduleName) {
        return utmServerModuleRepository.findAllByModuleName(moduleName);
    }
}
