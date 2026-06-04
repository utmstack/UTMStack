package com.park.utmstack.service;

import com.park.utmstack.domain.UtmAlertLog;
import com.park.utmstack.repository.UtmAlertLogRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing UtmAlertLog.
 */
@Service
@Transactional
public class UtmAlertLogService {

    private final Logger log = LoggerFactory.getLogger(UtmAlertLogService.class);

    private final UtmAlertLogRepository utmAlertLogRepository;

    public UtmAlertLogService(UtmAlertLogRepository utmAlertLogRepository) {
        this.utmAlertLogRepository = utmAlertLogRepository;
    }

    /**
     * Save a utmAlertLog.
     *
     * @param utmAlertLog
     *     the entity to save
     *
     * @return the persisted entity
     */
    public UtmAlertLog save(UtmAlertLog utmAlertLog) {
        return utmAlertLogRepository.save(utmAlertLog);
    }

    /**
     * Get all the utmAlertLogs.
     *
     * @param pageable
     *     the pagination information
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmAlertLog> findAll(Pageable pageable) {
        return utmAlertLogRepository.findAll(pageable);
    }


    /**
     * Get one utmAlertLog by id.
     *
     * @param id
     *     the id of the entity
     *
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmAlertLog> findOne(Long id) {
        return utmAlertLogRepository.findById(id);
    }

    /**
     * Delete the utmAlertLog by id.
     *
     * @param id
     *     the id of the entity
     */
    public void delete(Long id) {
        utmAlertLogRepository.deleteById(id);
    }


}
