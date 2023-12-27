package com.park.utmstack.service.index_pattern;

import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.List;
import java.util.Optional;

/**
 * Service Interface for managing UtmIndexPattern.
 */
public interface UtmIndexPatternService {

    /**
     * Save a utmIndexPattern.
     *
     * @param utmIndexPattern the entity to save
     * @return the persisted entity
     */
    UtmIndexPattern save(UtmIndexPattern utmIndexPattern);

    void saveAll(List<UtmIndexPattern> patterns);

    /**
     * Get all the utmIndexPatterns.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    Page<UtmIndexPattern> findAll(Pageable pageable);

    List<UtmIndexPattern> findAll();

    /**
     * Get the "id" utmIndexPattern.
     *
     * @param id the id of the entity
     * @return the entity
     */
    Optional<UtmIndexPattern> findOne(Long id);

    /**
     * Delete the "id" utmIndexPattern.
     *
     * @param id the id of the entity
     */
    void delete(Long id);

    void deleteAllByPatternSystemIsTrueAndIdNotIn(List<Long> ids);

    Long getSystemSequenceNextValue();

    List<UtmIndexPattern> findAllByPatternModule(String nameShort);
}
