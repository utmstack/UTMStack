package com.park.utmstack.service.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmGroupLogstashPipelineFilters;
import com.park.utmstack.repository.logstash_pipeline.UtmGroupLogstashPipelineFiltersRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing {@link UtmGroupLogstashPipelineFilters}.
 */
@Service
public class UtmGroupLogstashPipelineFiltersService {

    private final Logger log = LoggerFactory.getLogger(UtmGroupLogstashPipelineFiltersService.class);

    private final UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository;

    public UtmGroupLogstashPipelineFiltersService(UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository) {
        this.utmGroupLogstashPipelineFiltersRepository = utmGroupLogstashPipelineFiltersRepository;
    }

    /**
     * Save a utmGroupLogstashPipelineFilters.
     *
     * @param utmGroupLogstashPipelineFilters the entity to save.
     * @return the persisted entity.
     */
    public UtmGroupLogstashPipelineFilters save(UtmGroupLogstashPipelineFilters utmGroupLogstashPipelineFilters) {
        log.debug("Request to save UtmGroupLogstashPipelineFilters : {}", utmGroupLogstashPipelineFilters);
        return utmGroupLogstashPipelineFiltersRepository.save(utmGroupLogstashPipelineFilters);
    }

    /**
     * Update a utmGroupLogstashPipelineFilters.
     *
     * @param utmGroupLogstashPipelineFilters the entity to save.
     * @return the persisted entity.
     */
    public UtmGroupLogstashPipelineFilters update(UtmGroupLogstashPipelineFilters utmGroupLogstashPipelineFilters) {
        log.debug("Request to save UtmGroupLogstashPipelineFilters : {}", utmGroupLogstashPipelineFilters);
        return utmGroupLogstashPipelineFiltersRepository.save(utmGroupLogstashPipelineFilters);
    }

    /**
     * Get all the utmGroupLogstashPipelineFilters.
     *
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public List<UtmGroupLogstashPipelineFilters> findAll() {
        log.debug("Request to get all UtmGroupLogstashPipelineFilters");
        return utmGroupLogstashPipelineFiltersRepository.findAll();
    }

    /**
     * Get one utmGroupLogstashPipelineFilters by id.
     *
     * @param id the id of the entity.
     * @return the entity.
     */
    @Transactional(readOnly = true)
    public Optional<UtmGroupLogstashPipelineFilters> findOne(Long id) {
        log.debug("Request to get UtmGroupLogstashPipelineFilters : {}", id);
        return utmGroupLogstashPipelineFiltersRepository.findById(id);
    }

    /**
     * Delete the utmGroupLogstashPipelineFilters by id.
     *
     * @param id the id of the entity.
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmGroupLogstashPipelineFilters : {}", id);
        utmGroupLogstashPipelineFiltersRepository.deleteById(id);
    }

    /**
     * Delete the utmGroupLogstashPipelineFilters by filter id.
     *
     * @param filterId the id of the filter.
     */
    public void deleteRelations(Integer filterId) {
        log.debug("Request to delete UtmGroupLogstashPipelineFilters : {}", filterId);
        utmGroupLogstashPipelineFiltersRepository.deleteRelationByFilterId(filterId);
    }
}
