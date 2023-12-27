package com.park.utmstack.service;

import com.park.utmstack.domain.UtmConfigurationSection;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.repository.UtmConfigurationSectionRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmConfigurationSection.
 */
@Service
@Transactional
public class UtmConfigurationSectionService {

    private static final String CLASSNAME = "UtmConfigurationSectionService";

    private final Logger log = LoggerFactory.getLogger(UtmConfigurationSectionService.class);

    private final UtmConfigurationSectionRepository utmConfigurationSectionRepository;

    public UtmConfigurationSectionService(UtmConfigurationSectionRepository utmConfigurationSectionRepository) {
        this.utmConfigurationSectionRepository = utmConfigurationSectionRepository;
    }

    /**
     * Save a utmConfigurationSection.
     *
     * @param utmConfigurationSection the entity to save
     * @return the persisted entity
     */
    public UtmConfigurationSection save(UtmConfigurationSection utmConfigurationSection) {
        log.debug("Request to save UtmConfigurationSection : {}", utmConfigurationSection);
        return utmConfigurationSectionRepository.save(utmConfigurationSection);
    }

    /**
     * Get all the utmConfigurationSections.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmConfigurationSection> findAll(Pageable pageable) {
        log.debug("Request to get all UtmConfigurationSections");
        return utmConfigurationSectionRepository.findAll(pageable);
    }


    /**
     * Get one utmConfigurationSection by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmConfigurationSection> findOne(Long id) {
        log.debug("Request to get UtmConfigurationSection : {}", id);
        return utmConfigurationSectionRepository.findById(id);
    }

    /**
     * Delete the utmConfigurationSection by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmConfigurationSection : {}", id);
        utmConfigurationSectionRepository.deleteById(id);
    }

    /**
     * @param shortName
     * @return
     * @throws Exception
     */
    public List<UtmConfigurationSection> findAllByShortName(ModuleName shortName) throws Exception {
        final String ctx = CLASSNAME + ".findAllByShortName";
        try {
            return utmConfigurationSectionRepository.findAllByModuleNameShort(shortName);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public UtmConfigurationSection findByModuleNameShortAndSection(ModuleName module, String section) throws Exception {
        final String ctx = CLASSNAME + ".findByModuleNameShortAndSection";
        try {
            return utmConfigurationSectionRepository.findByModuleNameShortAndSection(module, section);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
