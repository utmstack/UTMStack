package com.park.utmstack.service.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.config.UtmRegexPattern;
import com.park.utmstack.repository.correlation.config.UtmDataTypesRepository;
import com.park.utmstack.repository.correlation.config.UtmRegexPatternRepository;
import io.undertow.util.BadRequestException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing {@link UtmRegexPatternService}.
 */
@Service
@Transactional
public class UtmRegexPatternService {
    private final Logger log = LoggerFactory.getLogger(UtmRegexPatternService.class);
    private static final String CLASSNAME = "UtmRegexPatternRepository";
    private final UtmRegexPatternRepository utmRegexPatternRepository;

    public UtmRegexPatternService(UtmRegexPatternRepository utmRegexPatternRepository) {
        this.utmRegexPatternRepository = utmRegexPatternRepository;
    }

    /**
     * Save a UtmRegexPattern.
     *
     * @param pattern the entity to save.
     * @return the persisted entity.
     */
    private UtmRegexPattern save(UtmRegexPattern pattern) {
        log.debug("Request to save UtmRegexPattern : {}", pattern);
        final String ctx = CLASSNAME + ".save";
        if (pattern.getId() == null) {
            pattern.setId(utmRegexPatternRepository.getNextId());
        }
        try {
            pattern.setLastUpdate();
            return utmRegexPatternRepository.save(pattern);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Add a new UtmRegexPattern
     *
     * @param pattern The regex pattern to add
     * @throws BadRequestException Bad Request if the regex pattern has an id or generic if some error occurs when inserting in DB
     * */
    public void addRegexPattern(UtmRegexPattern pattern) throws BadRequestException {
        final String ctx = CLASSNAME + ".addRegexPattern";
        if (pattern.getId() != null) {
            throw new BadRequestException(ctx + ": A new regex pattern can't have an id.");
        }
        try {
            pattern.setSystemOwner(false);
            this.save(pattern);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding the regex pattern.", ex);
        }
    }

    /**
     * Update a UtmRegexPattern
     *
     * @param pattern The regex pattern to update
     * @throws BadRequestException Bad Request if the regex pattern don't have an id, or is a system regex pattern, or isn't present in database,
     *         or generic error if some error occurs when updating in DB
     * */
    public void updateRegexPattern(UtmRegexPattern pattern) throws BadRequestException {
        final String ctx = CLASSNAME + ".updateRegexPattern";
        if (pattern.getId() == null) {
            throw new BadRequestException(ctx + ": The regex pattern must have an id to update.");
        }
        Optional<UtmRegexPattern> find = utmRegexPatternRepository.findById(pattern.getId());
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The regex pattern you're trying to update is not present in database.");
        }
        if(find.get().getSystemOwner()) {
            throw new BadRequestException(ctx + ": System's regex pattern can't be updated.");
        }
        try {
            this.save(pattern);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while updating the regex pattern.", ex);
        }
    }

    /**
     * Delete the UtmRegexPattern by id.
     *
     * @param id the id of the entity
     * @throws BadRequestException Bad Request if the regex pattern is not present in database or is a system regex pattern
     */
    public void delete(Long id) throws BadRequestException{
        final String ctx = CLASSNAME + ".delete";
        Optional<UtmRegexPattern> find = utmRegexPatternRepository.findById(id);
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The regex pattern you're trying to delete is not present in database.");
        }
        if(find.get().getSystemOwner()) {
            throw new BadRequestException(ctx + ": System's regex pattern can't be removed.");
        }
        try {
            utmRegexPatternRepository.deleteById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get one UtmRegexPattern by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    public Optional<UtmRegexPattern> findOne(Long id) {
        final String ctx = CLASSNAME + ".findOne";
        try {
            return utmRegexPatternRepository.findById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get one UtmRegexPattern by id.
     *
     * @param p the pagination parameters
     * @return the entity
     */
    public Page<UtmRegexPattern> findAll(String search, Pageable p) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return utmRegexPatternRepository.searchByFilters(search != null ? "%"+search+"%" : null, p);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
