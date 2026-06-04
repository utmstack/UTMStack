package com.park.utmstack.service;

import com.park.utmstack.domain.Authority;

import java.util.List;
import java.util.Optional;

public interface AuthorityService {

    /**
     * Save a authority.
     *
     * @param authority the entity to save
     * @return the persisted entity
     */
    Authority save(Authority authority);

    /**
     * Get all the authorities.
     *
     * @return the list of entities
     */
    List<Authority> findAll();

    /**
     * @param name
     * @return
     */
    Optional<Authority> findOne(String name);

    /**
     * Delete the "id" utmRule.
     *
     * @param id the id of the entity
     */
    void delete(String id);
}
