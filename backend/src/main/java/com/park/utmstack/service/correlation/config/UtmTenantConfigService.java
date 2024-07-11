package com.park.utmstack.service.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmTenantConfig;
import com.park.utmstack.repository.correlation.config.UtmTenantConfigRepository;
import io.undertow.util.BadRequestException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing {@link UtmTenantConfigService}.
 */
@Service
@Transactional
public class UtmTenantConfigService {
    private final Logger log = LoggerFactory.getLogger(UtmTenantConfigService.class);
    private static final String CLASSNAME = "UtmTenantConfigService";
    private final UtmTenantConfigRepository utmTenantConfigRepository;

    public UtmTenantConfigService(UtmTenantConfigRepository utmTenantConfigRepository) {
        this.utmTenantConfigRepository = utmTenantConfigRepository;
    }

    /**
     * Save a UtmTenantConfig.
     *
     * @param tenantConfig the entity to save.
     * @return the persisted entity.
     */
    private UtmTenantConfig save(UtmTenantConfig tenantConfig) {
        log.debug("Request to save UtmTenantConfig : {}", tenantConfig);
        final String ctx = CLASSNAME + ".save";
        try {
            tenantConfig.setLastUpdate();
            return utmTenantConfigRepository.save(tenantConfig);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Add a new tenantConfig
     *
     * @param tenantConfig The tenant configuration to add
     * @throws BadRequestException Bad Request if the tenant configuration has an id or generic if some error occurs when inserting in DB
     * */
    public void addTenantConfig(UtmTenantConfig tenantConfig) throws BadRequestException {
        final String ctx = CLASSNAME + ".addTenantConfig";
        if (tenantConfig.getId() != null) {
            throw new BadRequestException(ctx + ": A new tenant configuration can't have an id.");
        }
        try {
            this.save(tenantConfig);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding the tenant configuration.", ex);
        }
    }

    /**
     * Update a tenantConfig
     *
     * @param tenantConfig The tenant configuration to update
     * @throws BadRequestException Bad Request if the tenant configuration don't have an id, or isn't present in database,
     *         or generic error if some error occurs when updating in DB
     * */
    public void updateTenantConfig(UtmTenantConfig tenantConfig) throws BadRequestException {
        final String ctx = CLASSNAME + ".updateTenantConfig";
        if (tenantConfig.getId() == null) {
            throw new BadRequestException(ctx + ": The tenant configuration must have an id to update.");
        }
        Optional<UtmTenantConfig> find = utmTenantConfigRepository.findById(tenantConfig.getId());
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The tenant configuration you're trying to update is not present in database.");
        }
        try {
            this.save(tenantConfig);
        } catch (Exception ex) {
            throw new RuntimeException(ctx + ": An error occurred while adding the tenant configuration.", ex);
        }
    }

    /**
     * Delete the UtmTenantConfig by id.
     *
     * @param id the id of the entity
     * @throws BadRequestException Bad Request if the tenant configuration is not present in database.
     */
    public void delete(Long id) throws BadRequestException{
        final String ctx = CLASSNAME + ".delete";
        Optional<UtmTenantConfig> find = utmTenantConfigRepository.findById(id);
        if (find.isEmpty()) {
            throw new BadRequestException(ctx + ": The tenantConfig you're trying to delete is not present in database.");
        }
        try {
            utmTenantConfigRepository.deleteById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get one UtmTenantConfig by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    public Optional<UtmTenantConfig> findOne(Long id) {
        final String ctx = CLASSNAME + ".findOne";
        try {
            return utmTenantConfigRepository.findById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get all UtmTenantConfig.
     *
     * @param p the pagination parameters
     * @return the list of tenant configurations
     */
    public Page<UtmTenantConfig> findAll(String search, Pageable p) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return utmTenantConfigRepository.searchByFilters(search != null ? "%"+search+"%" : null, p);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
