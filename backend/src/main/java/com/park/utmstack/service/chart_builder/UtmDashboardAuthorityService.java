package com.park.utmstack.service.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboardAuthority;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.Optional;

/**
 * Service Interface for managing UtmDashboardAuthority.
 */
public interface UtmDashboardAuthorityService {

    /**
     * Save a utmDashboardAuthority.
     *
     * @param utmDashboardAuthority the entity to save
     * @return the persisted entity
     */
    UtmDashboardAuthority save(UtmDashboardAuthority utmDashboardAuthority);

    /**
     * Get all the utmDashboardAuthorities.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    Page<UtmDashboardAuthority> findAll(Pageable pageable);


    /**
     * Get the "id" utmDashboardAuthority.
     *
     * @param id the id of the entity
     * @return the entity
     */
    Optional<UtmDashboardAuthority> findOne(Long id);

    /**
     * Delete the "id" utmDashboardAuthority.
     *
     * @param id the id of the entity
     */
    void delete(Long id);
}
