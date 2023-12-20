package com.park.utmstack.service.logstash_filter;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.repository.logstash_filter.UtmLogstashFilterRepository;
import com.park.utmstack.service.logstash_pipeline.UtmGroupLogstashPipelineFiltersService;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmLogstashFilter.
 */
@Service
@Transactional
public class UtmLogstashFilterService {

    private static final String CLASSNAME = "UtmLogstashFilterService";

    private final UtmLogstashFilterRepository logstashFilterRepository;
    UtmGroupLogstashPipelineFiltersService groupLogstashPipelineFiltersService;

    public UtmLogstashFilterService(UtmLogstashFilterRepository logstashFilterRepository,
                                    UtmGroupLogstashPipelineFiltersService groupLogstashPipelineFiltersService) {
        this.logstashFilterRepository = logstashFilterRepository;
        this.groupLogstashPipelineFiltersService = groupLogstashPipelineFiltersService;
    }

    /**
     * Save a utmLogstashFilter.
     *
     * @param logstashFilter the entity to save
     * @return the persisted entity
     */
    public UtmLogstashFilter save(UtmLogstashFilter logstashFilter) {
        final String ctx = CLASSNAME + ".save";
        try {
            return logstashFilterRepository.save(logstashFilter);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public void saveAll(List<UtmLogstashFilter> filters) {
        final String ctx = CLASSNAME + ".saveAll";
        try {
            logstashFilterRepository.saveAll(filters);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get one utmLogstashFilter by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmLogstashFilter> findOne(Long id) {
        final String ctx = CLASSNAME + ".findOne";
        try {
            return logstashFilterRepository.findById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Delete the utmLogstashFilter by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        final String ctx = CLASSNAME + ".delete";
        try {
            groupLogstashPipelineFiltersService.deleteRelations(id.intValue());
            logstashFilterRepository.deleteById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public List<UtmLogstashFilter> findAll() {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return logstashFilterRepository.findAll();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids) {
        logstashFilterRepository.deleteAllBySystemOwnerIsTrueAndIdNotIn(ids);
    }

    public List<UtmLogstashFilter> findAllByModuleName(String nameShort) {
        return logstashFilterRepository.findAllByModuleName(nameShort);
    }

    /**
     * Get all filters associated to specific pipeline.
     *
     * @param pipelineId the id of the UtmLogstashPipeline to filter
     */
    public List<UtmLogstashFilter> filtersByPipelineId(Long pipelineId) {
        final String ctx = CLASSNAME + ".filtersByPipelineId";
        try {
            return logstashFilterRepository.filtersByPipelineId(pipelineId);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
