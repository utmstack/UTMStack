package com.park.utmstack.service.impl;

import com.park.utmstack.domain.UtmSchedule;
import com.park.utmstack.repository.UtmScheduleRepository;
import com.park.utmstack.service.UtmScheduleService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing UtmSchedule.
 */
@Service
@Transactional
public class UtmScheduleServiceImpl implements UtmScheduleService {

    private final Logger log = LoggerFactory.getLogger(UtmScheduleServiceImpl.class);

    private final UtmScheduleRepository utmScheduleRepository;

    public UtmScheduleServiceImpl(UtmScheduleRepository utmScheduleRepository) {
        this.utmScheduleRepository = utmScheduleRepository;
    }

    /**
     * Save a utmSchedule.
     *
     * @param utmSchedule the entity to save
     * @return the persisted entity
     */
    @Override
    public UtmSchedule save(UtmSchedule utmSchedule) {
        log.debug("Request to save UtmSchedule : {}", utmSchedule);
        return utmScheduleRepository.save(utmSchedule);
    }

    /**
     * Get all the utmSchedules.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Override
    @Transactional(readOnly = true)
    public Page<UtmSchedule> findAll(Pageable pageable) {
        log.debug("Request to get all UtmSchedules");
        return utmScheduleRepository.findAll(pageable);
    }


    /**
     * Get one utmSchedule by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Override
    @Transactional(readOnly = true)
    public Optional<UtmSchedule> findOne(Long id) {
        log.debug("Request to get UtmSchedule : {}", id);
        return utmScheduleRepository.findById(id);
    }

    /**
     * Delete the utmSchedule by id.
     *
     * @param id the id of the entity
     */
    @Override
    public void delete(Long id) {
        log.debug("Request to delete UtmSchedule : {}", id);
        utmScheduleRepository.deleteById(id);
    }
}
