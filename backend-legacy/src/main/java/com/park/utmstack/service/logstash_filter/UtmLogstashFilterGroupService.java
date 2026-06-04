package com.park.utmstack.service.logstash_filter;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilterGroup;
import com.park.utmstack.repository.logstash_filter.UtmLogstashFilterGroupRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmLogstashFilterGroup.
 */
@Service
@Transactional
public class UtmLogstashFilterGroupService {

    private final Logger log = LoggerFactory.getLogger(UtmLogstashFilterGroupService.class);

    private final UtmLogstashFilterGroupRepository utmLogstashFilterGroupRepository;

    public UtmLogstashFilterGroupService(UtmLogstashFilterGroupRepository utmLogstashFilterGroupRepository) {
        this.utmLogstashFilterGroupRepository = utmLogstashFilterGroupRepository;
    }

    /**
     * Save a utmLogstashFilterGroup.
     *
     * @param utmLogstashFilterGroup the entity to save
     * @return the persisted entity
     */
    public UtmLogstashFilterGroup save(UtmLogstashFilterGroup utmLogstashFilterGroup) {
        log.debug("Request to save UtmLogstashFilterGroup : {}", utmLogstashFilterGroup);
        return utmLogstashFilterGroupRepository.save(utmLogstashFilterGroup);
    }

    /**
     * Save a list of utmLogstashFilterGroup.
     *
     * @param groups List of the groups to save
     */
    public void saveAll(List<UtmLogstashFilterGroup> groups) {
        utmLogstashFilterGroupRepository.saveAll(groups);
    }

    /**
     * Get all the utmLogstashFilterGroups.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmLogstashFilterGroup> findAll(Pageable pageable) {
        log.debug("Request to get all UtmLogstashFilterGroups");
        return utmLogstashFilterGroupRepository.findAll(pageable);
    }

    @Transactional(readOnly = true)
    public List<UtmLogstashFilterGroup> findAll() {
        log.debug("Request to get all UtmLogstashFilterGroups");
        return utmLogstashFilterGroupRepository.findAll();
    }


    /**
     * Get one utmLogstashFilterGroup by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmLogstashFilterGroup> findOne(Long id) {
        log.debug("Request to get UtmLogstashFilterGroup : {}", id);
        return utmLogstashFilterGroupRepository.findById(id);
    }

    /**
     * Delete the utmLogstashFilterGroup by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmLogstashFilterGroup : {}", id);
        utmLogstashFilterGroupRepository.deleteById(id);
    }

    public void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids) {
        utmLogstashFilterGroupRepository.deleteAllBySystemOwnerIsTrueAndIdNotIn(ids);
    }
}
