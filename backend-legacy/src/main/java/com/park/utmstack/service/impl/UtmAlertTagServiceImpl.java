package com.park.utmstack.service.impl;

import com.park.utmstack.domain.UtmAlertTag;
import com.park.utmstack.repository.UtmAlertTagRepository;
import com.park.utmstack.service.UtmAlertTagService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmAlertTag.
 */
@Service
@Transactional
public class UtmAlertTagServiceImpl implements UtmAlertTagService {

    private final Logger log = LoggerFactory.getLogger(UtmAlertTagServiceImpl.class);

    private final UtmAlertTagRepository alertTagRepository;

    public UtmAlertTagServiceImpl(UtmAlertTagRepository utmAlertTagRepository) {
        this.alertTagRepository = utmAlertTagRepository;
    }

    /**
     * Save a utmAlertTag.
     *
     * @param utmAlertTag the entity to save
     * @return the persisted entity
     */
    @Override
    public UtmAlertTag save(UtmAlertTag utmAlertTag) {
        log.debug("Request to save UtmAlertTag : {}", utmAlertTag);
        return alertTagRepository.save(utmAlertTag);
    }

    /**
     * Get all the utmAlertCategories.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Override
    @Transactional(readOnly = true)
    public Page<UtmAlertTag> findAll(Pageable pageable) {
        log.debug("Request to get all UtmAlertCategories");
        return alertTagRepository.findAll(pageable);
    }


    /**
     * Get one utmAlertTag by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Override
    @Transactional(readOnly = true)
    public Optional<UtmAlertTag> findOne(Long id) {
        log.debug("Request to get UtmAlertTag : {}", id);
        return alertTagRepository.findById(id);
    }

    /**
     * Delete the utmAlertTag by id.
     *
     * @param id the id of the entity
     */
    @Override
    public void delete(Long id) {
        log.debug("Request to delete UtmAlertTag : {}", id);
        alertTagRepository.deleteById(id);
    }

    @Override
    @Transactional(readOnly = true)
    public List<UtmAlertTag> findAllByIdIn(List<Long> ids) {
        return alertTagRepository.findAllByIdIn(ids);
    }
}
