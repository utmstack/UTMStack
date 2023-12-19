package com.park.utmstack.service.impl.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboardAuthority;
import com.park.utmstack.repository.chart_builder.UtmDashboardAuthorityRepository;
import com.park.utmstack.service.chart_builder.UtmDashboardAuthorityService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing UtmDashboardAuthority.
 */
@Service
@Transactional
public class UtmDashboardAuthorityServiceImpl implements UtmDashboardAuthorityService {

    private final Logger log = LoggerFactory.getLogger(UtmDashboardAuthorityServiceImpl.class);

    private final UtmDashboardAuthorityRepository utmDashboardAuthorityRepository;

    public UtmDashboardAuthorityServiceImpl(UtmDashboardAuthorityRepository utmDashboardAuthorityRepository) {
        this.utmDashboardAuthorityRepository = utmDashboardAuthorityRepository;
    }

    /**
     * Save a utmDashboardAuthority.
     *
     * @param utmDashboardAuthority the entity to save
     * @return the persisted entity
     */
    @Override
    public UtmDashboardAuthority save(UtmDashboardAuthority utmDashboardAuthority) {
        log.debug("Request to save UtmDashboardAuthority : {}", utmDashboardAuthority);
        return utmDashboardAuthorityRepository.save(utmDashboardAuthority);
    }

    /**
     * Get all the utmDashboardAuthorities.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Override
    @Transactional(readOnly = true)
    public Page<UtmDashboardAuthority> findAll(Pageable pageable) {
        log.debug("Request to get all UtmDashboardAuthorities");
        return utmDashboardAuthorityRepository.findAll(pageable);
    }


    /**
     * Get one utmDashboardAuthority by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Override
    @Transactional(readOnly = true)
    public Optional<UtmDashboardAuthority> findOne(Long id) {
        log.debug("Request to get UtmDashboardAuthority : {}", id);
        return utmDashboardAuthorityRepository.findById(id);
    }

    /**
     * Delete the utmDashboardAuthority by id.
     *
     * @param id the id of the entity
     */
    @Override
    public void delete(Long id) {
        log.debug("Request to delete UtmDashboardAuthority : {}", id);
        utmDashboardAuthorityRepository.deleteById(id);
    }
}
