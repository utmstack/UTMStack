package com.park.utmstack.service.impl;

import com.park.utmstack.domain.UtmMenuAuthority;
import com.park.utmstack.repository.UtmMenuAuthorityRepository;
import com.park.utmstack.service.UtmMenuAuthorityService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmMenuAuthority.
 */
@Service
@Transactional
public class UtmMenuAuthorityServiceImpl implements UtmMenuAuthorityService {

    private final Logger log = LoggerFactory.getLogger(UtmMenuAuthorityServiceImpl.class);

    private final UtmMenuAuthorityRepository utmMenuAuthorityRepository;

    public UtmMenuAuthorityServiceImpl(UtmMenuAuthorityRepository utmMenuAuthorityRepository) {
        this.utmMenuAuthorityRepository = utmMenuAuthorityRepository;
    }

    /**
     * Save a utmMenuAuthority.
     *
     * @param utmMenuAuthority the entity to save
     * @return the persisted entity
     */
    @Override
    public UtmMenuAuthority save(UtmMenuAuthority utmMenuAuthority) {
        log.debug("Request to save UtmMenuAuthority : {}", utmMenuAuthority);
        return utmMenuAuthorityRepository.save(utmMenuAuthority);
    }

    @Override
    public void saveAll(List<UtmMenuAuthority> utmMenuAuthorities) {
        utmMenuAuthorityRepository.saveAll(utmMenuAuthorities);
    }

    /**
     * Get all the utmMenuAuthorities.
     *
     * @return the list of entities
     */
    @Override
    @Transactional(readOnly = true)
    public List<UtmMenuAuthority> findAll() {
        log.debug("Request to get all UtmMenuAuthorities");
        return utmMenuAuthorityRepository.findAll();
    }


    /**
     * Get one utmMenuAuthority by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Override
    @Transactional(readOnly = true)
    public Optional<UtmMenuAuthority> findOne(Long id) {
        log.debug("Request to get UtmMenuAuthority : {}", id);
        return utmMenuAuthorityRepository.findById(id);
    }

    /**
     * Delete the utmMenuAuthority by id.
     *
     * @param id the id of the entity
     */
    @Override
    public void delete(Long id) {
        log.debug("Request to delete UtmMenuAuthority : {}", id);
        utmMenuAuthorityRepository.deleteById(id);
    }

    @Override
    public void deleteByMenuId(Long menuId) {
        utmMenuAuthorityRepository.deleteAllByMenuIdEquals(menuId);
    }

    @Override
    public void deleteAllByMenuIdIn(List<Long> menuIds) {
        utmMenuAuthorityRepository.deleteAllByMenuIdIn(menuIds);
    }

    @Override
    public void deleteAllByIdIn(List<Long> ids) {
        utmMenuAuthorityRepository.deleteAllByIdIn(ids);
    }

    @Override
    public Optional<List<UtmMenuAuthority>> findAllByMenuIdEquals(Long menuId) {
        return utmMenuAuthorityRepository.findAllByMenuIdEquals(menuId);
    }

    @Override
    public void deleteMenuAuthoritiesNotIn(Long menuId, List<String> authorities) {
        utmMenuAuthorityRepository.deleteMenuAuthoritiesNotIn(menuId, authorities);
    }
}
