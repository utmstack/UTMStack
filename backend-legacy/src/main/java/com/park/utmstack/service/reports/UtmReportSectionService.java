package com.park.utmstack.service.reports;

import com.park.utmstack.domain.reports.UtmReportSection;
import com.park.utmstack.repository.reports.UtmReportSectionRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmReportSection.
 */
@Service
@Transactional
public class UtmReportSectionService {

    private final Logger log = LoggerFactory.getLogger(UtmReportSectionService.class);

    private final UtmReportSectionRepository utmReportSectionRepository;

    public UtmReportSectionService(UtmReportSectionRepository utmReportSectionRepository) {
        this.utmReportSectionRepository = utmReportSectionRepository;
    }

    /**
     * Save a utmReportSection.
     *
     * @param utmReportSection the entity to save
     * @return the persisted entity
     */
    public UtmReportSection save(UtmReportSection utmReportSection) {
        log.debug("Request to save UtmReportSection : {}", utmReportSection);
        return utmReportSectionRepository.save(utmReportSection);
    }

    public void saveAll(List<UtmReportSection> sections) {
        utmReportSectionRepository.saveAll(sections);
    }

    public void saveAndFlush(UtmReportSection section) {
        utmReportSectionRepository.saveAndFlush(section);
    }

    /**
     * Get all the utmReportSections.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmReportSection> findAll(Pageable pageable) {
        log.debug("Request to get all UtmReportSections");
        return utmReportSectionRepository.findAll(pageable);
    }

    @Transactional(readOnly = true)
    public List<UtmReportSection> findAll() {
        return utmReportSectionRepository.findAll();
    }


    /**
     * Get one utmReportSection by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmReportSection> findOne(Long id) {
        log.debug("Request to get UtmReportSection : {}", id);
        return utmReportSectionRepository.findById(id);
    }

    /**
     * Delete the utmReportSection by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmReportSection : {}", id);
        utmReportSectionRepository.deleteById(id);
    }

    public void deleteAllByRepSecSystemTrueAndIdNotIn(List<Long> sectionIds) {
        utmReportSectionRepository.deleteAllByRepSecSystemTrueAndIdNotIn(sectionIds);
    }
}
