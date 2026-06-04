package com.park.utmstack.service;

import com.park.utmstack.domain.UtmSchedule;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.Optional;

/**
 * Service Interface for managing UtmSchedule.
 */
public interface UtmScheduleService {

    /**
     * Save a utmSchedule.
     *
     * @param utmSchedule the entity to save
     * @return the persisted entity
     */
    UtmSchedule save(UtmSchedule utmSchedule);

    /**
     * Get all the utmSchedules.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    Page<UtmSchedule> findAll(Pageable pageable);


    /**
     * Get the "id" utmSchedule.
     *
     * @param id the id of the entity
     * @return the entity
     */
    Optional<UtmSchedule> findOne(Long id);

    /**
     * Delete the "id" utmSchedule.
     *
     * @param id the id of the entity
     */
    void delete(Long id);
}
