package com.park.utmstack.service.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.List;
import java.util.Optional;

/**
 * Service Interface for managing UtmDashboardVisualization.
 */
public interface UtmDashboardVisualizationService {

    /**
     * Save a utmDashboardVisualization.
     *
     * @param utmDashboardVisualization the entity to save
     * @return the persisted entity
     */
    UtmDashboardVisualization save(UtmDashboardVisualization utmDashboardVisualization);

    void saveAll(List<UtmDashboardVisualization> relations);

    /**
     * Get all the utmDashboardVisualizations.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    Page<UtmDashboardVisualization> findAll(Pageable pageable);


    /**
     * Get the "id" utmDashboardVisualization.
     *
     * @param id the id of the entity
     * @return the entity
     */
    Optional<UtmDashboardVisualization> findOne(Long id);

    /**
     * Delete the "id" utmDashboardVisualization.
     *
     * @param id the id of the entity
     */
    void delete(Long id);

    Optional<List<UtmDashboardVisualization>> findAllByIdDashboard(Long idDashboard);
}
