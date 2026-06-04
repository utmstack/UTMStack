package com.park.utmstack.service.reports;

import com.park.utmstack.domain.reports.UtmReport;
import com.park.utmstack.repository.reports.UtmReportRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmReports.
 */
@Service
@Transactional
public class UtmReportService {

    private final Logger log = LoggerFactory.getLogger(UtmReportService.class);

    private final UtmReportRepository utmReportsRepository;

    public UtmReportService(UtmReportRepository utmReportsRepository) {
        this.utmReportsRepository = utmReportsRepository;
    }

    /**
     * Save a utmReports.
     *
     * @param utmReports the entity to save
     * @return the persisted entity
     */
    public UtmReport save(UtmReport utmReports) {
        log.debug("Request to save UtmReports : {}", utmReports);
        return utmReportsRepository.save(utmReports);
    }

    public void saveAll(List<UtmReport> reports) {
        utmReportsRepository.saveAll(reports);
    }

    /**
     * Get all the utmReports.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmReport> findAll(Pageable pageable) {
        log.debug("Request to get all UtmReports");
        return utmReportsRepository.findAll(pageable);
    }


    /**
     * Get one utmReports by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmReport> findOne(Long id) {
        log.debug("Request to get UtmReports : {}", id);
        return utmReportsRepository.findById(id);
    }

    public void deleteAllByReportSectionIdAndIdNotIn(Long sectionId, List<Long> ids) {
        utmReportsRepository.deleteAllByReportSectionIdAndIdNotIn(sectionId, ids);
    }

    /**
     * Delete the utmReports by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmReports : {}", id);
        utmReportsRepository.deleteById(id);
    }


}
