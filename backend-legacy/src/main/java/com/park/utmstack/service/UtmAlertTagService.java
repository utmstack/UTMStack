package com.park.utmstack.service;

import com.park.utmstack.domain.UtmAlertTag;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.List;
import java.util.Optional;

/**
 * Service Interface for managing UtmAlertTag.
 */
public interface UtmAlertTagService {

    /**
     * Save a utmAlertTag.
     *
     * @param utmAlertTag
     *     the entity to save
     *
     * @return the persisted entity
     */
    UtmAlertTag save(UtmAlertTag utmAlertTag);

    /**
     * Get all the utmAlertCategories.
     *
     * @param pageable
     *     the pagination information
     *
     * @return the list of entities
     */
    Page<UtmAlertTag> findAll(Pageable pageable);


    /**
     * Get the "id" utmAlertTag.
     *
     * @param id
     *     the id of the entity
     *
     * @return the entity
     */
    Optional<UtmAlertTag> findOne(Long id);

    /**
     * Delete the "id" utmAlertTag.
     *
     * @param id
     *     the id of the entity
     */
    void delete(Long id);

    List<UtmAlertTag> findAllByIdIn(List<Long> ids);
}
