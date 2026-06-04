package com.park.utmstack.service;

import com.park.utmstack.domain.UtmServerModule;
import com.park.utmstack.repository.UtmServerModuleRepository;
import com.park.utmstack.util.exceptions.ApiException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
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

    @Transactional
    public void markForRestart(String moduleName) {
        log.info("Marking module '{}' for restart due to inactivity.", moduleName);

        try {
            List<UtmServerModule> modules = utmServerModuleRepository.findAllByModuleName(moduleName);

            if (modules.isEmpty()) {
                log.warn("No modules found with name '{}'. Skipping restart trigger.", moduleName);
                return;
            }

            List<UtmServerModule> modulesToUpdate = modules.stream()
                    .filter(m -> !m.isNeedsRestart())
                    .peek(m -> m.setNeedsRestart(true))
                    .toList();

            if (!modulesToUpdate.isEmpty()) {
                utmServerModuleRepository.saveAll(modulesToUpdate);
                log.info("Successfully marked {} instances of '{}' for restart.",
                        modulesToUpdate.size(), moduleName);
            } else {
                log.debug("Module '{}' was already marked for restart.", moduleName);
            }

        } catch (Exception e) {
            log.error("Failed to mark module '{}' for restart: {}", moduleName, e.getMessage());
            throw new ApiException("Error triggering module restart for: " + moduleName, HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }
}
