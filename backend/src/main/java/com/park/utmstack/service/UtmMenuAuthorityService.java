package com.park.utmstack.service;

import com.park.utmstack.domain.UtmMenuAuthority;

import java.util.List;
import java.util.Optional;

/**
 * Service Interface for managing UtmMenuAuthority.
 */
public interface UtmMenuAuthorityService {

    /**
     * Save a utmMenuAuthority.
     *
     * @param utmMenuAuthority the entity to save
     * @return the persisted entity
     */
    UtmMenuAuthority save(UtmMenuAuthority utmMenuAuthority);

    /**
     * @param utmMenuAuthorities
     */
    void saveAll(List<UtmMenuAuthority> utmMenuAuthorities);

    /**
     * Get all the utmMenuAuthorities.
     *
     * @return the list of entities
     */
    List<UtmMenuAuthority> findAll();


    /**
     * Get the "id" utmMenuAuthority.
     *
     * @param id the id of the entity
     * @return the entity
     */
    Optional<UtmMenuAuthority> findOne(Long id);

    /**
     * Delete the "id" utmMenuAuthority.
     *
     * @param id the id of the entity
     */
    void delete(Long id);

    /**
     * Delete all menu authorities by menu identifier
     *
     * @param menuId: Menu identifier
     */
    void deleteByMenuId(Long menuId);

    /**
     * @param menuIds
     */
    void deleteAllByMenuIdIn(List<Long> menuIds);

    /**
     * @param ids
     */
    void deleteAllByIdIn(List<Long> ids);

    /**
     * Find all menu authorities by menu identifier
     *
     * @param menuId : Menu identifier
     * @return
     */
    Optional<List<UtmMenuAuthority>> findAllByMenuIdEquals(Long menuId);

    void deleteMenuAuthoritiesNotIn(Long menuId, List<String> authorities);
}
