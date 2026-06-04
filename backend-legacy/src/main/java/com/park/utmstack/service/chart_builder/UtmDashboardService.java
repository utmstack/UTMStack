package com.park.utmstack.service.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboard;
import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.List;
import java.util.Optional;

/**
 * Service Interface for managing UtmDashboard.
 */
public interface UtmDashboardService {

    /**
     * Save a utmDashboard.
     *
     * @param utmDashboard the entity to save
     * @return the persisted entity
     */
    UtmDashboard save(UtmDashboard utmDashboard);

    void saveAll(List<UtmDashboard> utmDashboard);

    /**
     * Get all the utmDashboards.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    Page<UtmDashboard> findAll(Pageable pageable);


    /**
     * Get the "id" utmDashboard.
     *
     * @param id the id of the entity
     * @return the entity
     */
    Optional<UtmDashboard> findOne(Long id);

    /**
     * Delete the "id" utmDashboard.
     *
     * @param id the id of the entity
     */
    void delete(Long id);

    void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids);

    void importDashboards(List<UtmDashboardVisualization> dashboards, Boolean override) throws Exception;

    Long getSystemSequenceNextValue();
}
