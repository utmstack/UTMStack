package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.UtmIntegrationConf;
import com.park.utmstack.repository.UtmIntegrationConfRepository;
import com.park.utmstack.util.CipherUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.Optional;

/**
 * Service Implementation for managing UtmIntegrationConf.
 */
@Service
@Transactional
public class UtmIntegrationConfService {
    private static final String CLASSNAME = "UtmIntegrationConfService";

    private final Logger log = LoggerFactory.getLogger(UtmIntegrationConfService.class);

    private final UtmIntegrationConfRepository utmIntegrationConfRepository;

    public UtmIntegrationConfService(UtmIntegrationConfRepository utmIntegrationConfRepository) {
        this.utmIntegrationConfRepository = utmIntegrationConfRepository;
    }

    /**
     * Save a utmIntegrationConf.
     *
     * @param utmIntegrationConf the entity to save
     * @return the persisted entity
     */
    public UtmIntegrationConf save(UtmIntegrationConf utmIntegrationConf) throws Exception {
        final String ctx = CLASSNAME + ".save";
        try {
            String dataType = utmIntegrationConf.getConfDatatype();
            if (StringUtils.hasText(dataType) && dataType.equals("password"))
                utmIntegrationConf.setConfValue(CipherUtil.encrypt(utmIntegrationConf.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
            return utmIntegrationConfRepository.save(utmIntegrationConf);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get all the utmIntegrationConfs.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIntegrationConf> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIntegrationConfs");
        return utmIntegrationConfRepository.findAll(pageable);
    }


    /**
     * Get one utmIntegrationConf by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIntegrationConf> findOne(Long id) {
        log.debug("Request to get UtmIntegrationConf : {}", id);
        return utmIntegrationConfRepository.findById(id);
    }

    /**
     * Delete the utmIntegrationConf by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIntegrationConf : {}", id);
        utmIntegrationConfRepository.deleteById(id);
    }
}
