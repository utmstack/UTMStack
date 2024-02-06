package com.park.utmstack.service.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmPortsConfiguration;
import com.park.utmstack.repository.logstash_pipeline.UtmPortsConfigurationRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing {@link UtmPortsConfiguration}.
 */
@Service
@Transactional
public class UtmPortsConfigurationService {

    private final Logger log = LoggerFactory.getLogger(UtmPortsConfigurationService.class);

    private final UtmPortsConfigurationRepository utmPortsConfigurationRepository;

    public UtmPortsConfigurationService(UtmPortsConfigurationRepository utmPortsConfigurationRepository) {
        this.utmPortsConfigurationRepository = utmPortsConfigurationRepository;
    }

    /**
     * Get all the utmPortsConfigurations.
     *
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public List<UtmPortsConfiguration> findAll() {
        log.debug("Request to get all UtmPortsConfigurations");
        return utmPortsConfigurationRepository.findAll();
    }

    /**
     * Get one utmPortsConfiguration by id.
     *
     * @param id the id of the entity.
     * @return the entity.
     */
    @Transactional(readOnly = true)
    public Optional<UtmPortsConfiguration> findOne(Long id) {
        log.debug("Request to get UtmPortsConfiguration : {}", id);
        return utmPortsConfigurationRepository.findById(id);
    }
}
