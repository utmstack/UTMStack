package com.park.utmstack.service.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInputConfiguration;
import com.park.utmstack.domain.logstash_pipeline.UtmPortsConfiguration;
import com.park.utmstack.domain.logstash_pipeline.enums.InputConfigTypes;
import com.park.utmstack.repository.logstash_pipeline.UtmLogstashInputConfigurationRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;

/**
 * Service Implementation for managing {@link UtmLogstashInputConfiguration}.
 */
@Service
public class UtmLogstashInputConfigurationService {

    private final Logger log = LoggerFactory.getLogger(UtmLogstashInputConfigurationService.class);
    private static final String CLASSNAME = "UtmLogstashInputConfigurationService";

    private final UtmLogstashInputConfigurationRepository utmLogstashInputConfigurationRepository;

    private final UtmPortsConfigurationService utmPortsConfigurationService;

    public UtmLogstashInputConfigurationService(UtmLogstashInputConfigurationRepository utmLogstashInputConfigurationRepository, UtmPortsConfigurationService utmPortsConfigurationService) {
        this.utmLogstashInputConfigurationRepository = utmLogstashInputConfigurationRepository;
        this.utmPortsConfigurationService = utmPortsConfigurationService;
    }

    /**
     * Save a utmLogstashInputConfiguration.
     *
     * @param utmLogstashInputConfiguration the entity to save.
     * @return the persisted entity.
     */
    public UtmLogstashInputConfiguration save(UtmLogstashInputConfiguration utmLogstashInputConfiguration) {
        log.debug("Request to save UtmLogstashInputConfiguration : {}", utmLogstashInputConfiguration);
        if (utmLogstashInputConfiguration.getId() == null) {
            utmLogstashInputConfiguration.setId(utmLogstashInputConfigurationRepository.getNextId());
        }
        return utmLogstashInputConfigurationRepository.save(utmLogstashInputConfiguration);
    }

    /**
     * Update a utmLogstashInputConfiguration.
     *
     * @param utmLogstashInputConfiguration the entity to save.
     * @return the persisted entity.
     */
    public UtmLogstashInputConfiguration update(UtmLogstashInputConfiguration utmLogstashInputConfiguration) {
        log.debug("Request to save UtmLogstashInputConfiguration : {}", utmLogstashInputConfiguration);
        return utmLogstashInputConfigurationRepository.save(utmLogstashInputConfiguration);
    }

    /**
     * Get all the utmLogstashInputConfigurations.
     *
     * @param pageable the pagination information.
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmLogstashInputConfiguration> findAll(Pageable pageable) {
        log.debug("Request to get all UtmLogstashInputConfigurations");
        return utmLogstashInputConfigurationRepository.findAll(pageable);
    }

    /**
     * Get one utmLogstashInputConfiguration by id.
     *
     * @param id the id of the entity.
     * @return the entity.
     */
    @Transactional(readOnly = true)
    public Optional<UtmLogstashInputConfiguration> findOne(Long id) {
        log.debug("Request to get UtmLogstashInputConfiguration : {}", id);
        return utmLogstashInputConfigurationRepository.findById(id);
    }

    /**
     * Delete the utmLogstashInputConfiguration by id.
     *
     * @param id the id of the entity.
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmLogstashInputConfiguration : {}", id);
        utmLogstashInputConfigurationRepository.deleteById(id);
    }

    /**
     * Get available ports by type
     * @param type must be one of (tcp)
     * */
    public List<String> getAvailableConfigsByType(String type) {
        // Replace _ssl to avoid conflicts with the type parameter in UtmLogstashInputResource methods
        // Now you can use same type for all resources in UtmLogstashInputResource
        String finalType = type.replace("_ssl","");
        Map<String, String> ports = getAllConfiguredPorts(finalType);
        List<String> inputAvailable = new ArrayList<>();
        List<UtmLogstashInputConfiguration> allInputs = utmLogstashInputConfigurationRepository.allConfigsByType(Arrays.asList(InputConfigTypes.PORT.getValue()));
        ports.forEach((e, v) -> {
            String proto = e.split("-")[0];
            Optional<UtmLogstashInputConfiguration> lc = allInputs.stream().filter(
                search -> search.getConfKey().startsWith(proto) && search.getConfValue().compareTo(v) == 0
            ).findFirst();
            if (lc == null || !lc.isPresent()) {
                inputAvailable.add(v);
            }
        });
        return inputAvailable;
    }

    /**
     * Method to generate all variants (ports in range or simple port) using the values of UtmPortsConfiguration table
     * You can pass a type to get only the variants for that type (protocol in the table)
     * TYPE must be one of (tcp)
     */
    public Map<String, String> getAllConfiguredPorts(String type) {
        final String ctx = CLASSNAME + ".getAllConfiguredPorts";
        List<UtmPortsConfiguration> configs = utmPortsConfigurationService.findAll();
        Map<String, String> available = new LinkedHashMap<>();
        configs.stream().forEach((conf) -> {
            String protocol = conf.getProtocol();
            if (protocol.compareToIgnoreCase(type) == 0) {
                if (conf.getPortOrRange().contains("-")) {
                    available.putAll(generateRange(protocol, conf.getPortOrRange()));
                } else {
                    try {
                        int port = Integer.parseInt(conf.getPortOrRange());
                        available.put(protocol + "-" + port, conf.getPortOrRange());
                    } catch (Exception e) {
                        log.warn(ctx + ": This port can't be generated -> " + e.getMessage());
                    }
                }
            }
        });

        return available;
    }

    /**
     * Method to generate all values between range in the way number-number
     */
    public Map<String, String> generateRange(String protocol, String range) {
        final String ctx = CLASSNAME + ".generateRange";
        Map<String, String> generated = new LinkedHashMap<>();
        String[] rangeValues = range.split("-");
        try {
            int r1 = Integer.parseInt(rangeValues[0].trim());
            int r2 = Integer.parseInt(rangeValues[1].trim());

            int start = Math.min(r1, r2);
            int end = Math.max(r1, r2);

            for (int i = start; i <= end; i++) {
                generated.put(protocol + "-" + i, "" + i);
            }
        } catch (Exception e) {
            log.warn(ctx + ": This range can't be generated -> " + e.getMessage());
            return generated;
        }
        return generated;
    }
    public Long getNextId(){
        return utmLogstashInputConfigurationRepository.getNextId();
    }
    public List<UtmLogstashInputConfiguration> getUtmLogstashInputConfigurationsByInputId(Integer inputId){
        return utmLogstashInputConfigurationRepository.getUtmLogstashInputConfigurationsByInputId(inputId);
    }
}
