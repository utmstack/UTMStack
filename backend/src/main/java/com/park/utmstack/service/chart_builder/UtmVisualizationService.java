package com.park.utmstack.service.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmVisualization;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.List;
import java.util.Optional;

/**
 * Service Interface for managing UtmVisualization.
 */
public interface UtmVisualizationService {

    /**
     * Save a utmVisualization.
     *
     * @param utmVisualization the entity to save
     * @return the persisted entity
     */
    UtmVisualization save(UtmVisualization utmVisualization);

    /**
     * Save a list of visualization using batch insert
     *
     * @param visualizations : List of visualization
     */
    void saveAll(Iterable<UtmVisualization> visualizations);

    /**
     * Get all the utmVisualizations.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    Page<UtmVisualization> findAll(Pageable pageable);


    /**
     * Get the "id" utmVisualization.
     *
     * @param id the id of the entity
     * @return the entity
     */
    Optional<UtmVisualization> findOne(Long id);

    /**
     * Delete the "id" utmVisualization.
     *
     * @param id the id of the entity
     */
    void delete(Long id);

    /**
     * Bulk delete
     *
     * @param ids
     */
    void deleteByIdIn(List<Long> ids);

    void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids);

    Optional<UtmVisualization> findByName(String name);

    Long getSystemSequenceNextValue();
}
