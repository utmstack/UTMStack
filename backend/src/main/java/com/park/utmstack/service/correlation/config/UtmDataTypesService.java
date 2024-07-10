package com.park.utmstack.service.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.repository.correlation.config.UtmDataTypesRepository;
import io.undertow.util.BadRequestException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing {@link UtmDataTypesService}.
 */
@Service
@Transactional
public class UtmDataTypesService {
    private final Logger log = LoggerFactory.getLogger(UtmDataTypesService.class);
    private static final String CLASSNAME = "UtmDataTypesService";
    private final UtmDataTypesRepository utmDataTypesRepository;

    public UtmDataTypesService(UtmDataTypesRepository utmDataTypesRepository) {
        this.utmDataTypesRepository = utmDataTypesRepository;
    }

    /**
     * Save a UtmDataTypes.
     *
     * @param dataType the entity to save.
     * @return the persisted entity.
     */
    private UtmDataTypes save(UtmDataTypes dataType) {
        log.debug("Request to save UtmDataTypes : {}", dataType);
        final String ctx = CLASSNAME + ".save";
        if (dataType.getId() == null) {
            dataType.setId(utmDataTypesRepository.getNextId());
        }
        try {
            dataType.setLastUpdate();
            return utmDataTypesRepository.save(dataType);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Add a new datatype
     *
     * @param dataTypes The datatype to add
     * @throws BadRequestException Bad Request if the datatype has an id or generic if some error occurs when inserting in DB
     * */
    public void addDataType(UtmDataTypes dataTypes) throws BadRequestException {
        final String ctx = CLASSNAME + ".addDataType";
        if (dataTypes.getId() != null) {
            throw new BadRequestException(ctx + ": A new datatype can't have an id.");
        }
        try {
            dataTypes.setSystemOwner(false);
            this.save(dataTypes);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding a datatype.", ex);
        }
    }

    /**
     * Update a datatype
     *
     * @param dataTypes The datatype to update
     * @throws BadRequestException Bad Request if the datatype don't have an id, or is a system datatype, or isn't present in database,
     *         or generic error if some error occurs when updating in DB
     * */
    public void updateDataType(UtmDataTypes dataTypes) throws BadRequestException {
        final String ctx = CLASSNAME + ".updateDataType";
        if (dataTypes.getId() == null) {
            throw new BadRequestException(ctx + ": The datatype must have an id to update.");
        }
        Optional<UtmDataTypes> find = utmDataTypesRepository.findById(dataTypes.getId());
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The datatype you're trying to update is not present in database.");
        }
        if(find.get().getSystemOwner()) {
            throw new BadRequestException(ctx + ": System's datatype can't be updated.");
        }
        try {
            this.save(dataTypes);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding a datatype.", ex);
        }
    }

    /**
     * Delete the UtmDataTypes by id.
     *
     * @param id the id of the entity
     * @throws BadRequestException Bad Request if the datatype is not present in database or is a system datatype
     */
    public void delete(Long id) throws BadRequestException{
        final String ctx = CLASSNAME + ".delete";
        Optional<UtmDataTypes> find = utmDataTypesRepository.findById(id);
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The datatype you're trying to delete is not present in database.");
        }
        if(find.get().getSystemOwner()) {
            throw new BadRequestException(ctx + ": System's datatype can't be removed.");
        }
        try {
            utmDataTypesRepository.deleteById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get one UtmDataTypes by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    public Optional<UtmDataTypes> findOne(Long id) {
        final String ctx = CLASSNAME + ".findOne";
        try {
            return utmDataTypesRepository.findById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get all UtmDataTypes.
     *
     * @param p the pagination parameters
     * @return the list of datatypes
     */
    public Page<UtmDataTypes> findAll(String search, Pageable p) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return utmDataTypesRepository.searchByFilters(search != null ? "%"+search+"%" : null, p);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
