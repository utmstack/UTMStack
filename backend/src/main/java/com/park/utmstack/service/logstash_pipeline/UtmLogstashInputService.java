package com.park.utmstack.service.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInput;
import com.park.utmstack.repository.logstash_pipeline.UtmLogstashInputRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing {@link UtmLogstashInput}.
 */
@Service
public class UtmLogstashInputService {

    private final Logger log = LoggerFactory.getLogger(UtmLogstashInputService.class);

    private final UtmLogstashInputRepository utmLogstashInputRepository;

    public UtmLogstashInputService(UtmLogstashInputRepository utmLogstashInputRepository) {
        this.utmLogstashInputRepository = utmLogstashInputRepository;
    }

    /**
     * Save a utmLogstashInput.
     *
     * @param utmLogstashInput the entity to save.
     * @return the persisted entity.
     */
    public UtmLogstashInput save(UtmLogstashInput utmLogstashInput) {
        log.debug("Request to save UtmLogstashInput : {}", utmLogstashInput);
        if (utmLogstashInput.getId()==null) {
            utmLogstashInput.setId(utmLogstashInputRepository.getNextId());
        }
        return utmLogstashInputRepository.save(utmLogstashInput);
    }

    public UtmLogstashInput update(UtmLogstashInput utmLogstashInput) {
        log.debug("Request to update UtmLogstashInput : {}", utmLogstashInput);
        return utmLogstashInputRepository.save(utmLogstashInput);
    }

    /**
     * Get all the utmLogstashInputs.
     *
     * @param pageable the pagination information.
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmLogstashInput> findAll(Pageable pageable) {
        log.debug("Request to get all UtmLogstashInputs");
        return utmLogstashInputRepository.findAll(pageable);
    }

    public List<UtmLogstashInput> getUtmLogstashInputsByPipelineId(Integer pipelineId) {
     return  utmLogstashInputRepository.getUtmLogstashInputsByPipelineId(pipelineId);
    }

    /**
     * Get one utmLogstashInput by id.
     *
     * @param id the id of the entity.
     * @return the entity.
     */
    @Transactional(readOnly = true)
    public Optional<UtmLogstashInput> findOne(Long id) {
        log.debug("Request to get UtmLogstashInput : {}", id);
        return utmLogstashInputRepository.findById(id);
    }

    /**
     * Delete the utmLogstashInput by id.
     *
     * @param id the id of the entity.
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmLogstashInput : {}", id);
        utmLogstashInputRepository.deleteById(id);
    }
    /**
     * Delete list of utmLogstashInput.
     *
     * @param inputs the list of inputs to delete.
     */
    public void deleteAllInputs(List<UtmLogstashInput> inputs) throws Exception{
        log.debug("Request to delete UtmLogstashInput : {}", inputs);
        utmLogstashInputRepository.deleteAll(inputs);
    }
}
