package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.UtmConfigurationParameter;
import com.park.utmstack.repository.UtmConfigurationParameterRepository;
import com.park.utmstack.util.CipherUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

import static com.park.utmstack.config.Constants.DATE_FORMAT_SETTING_ID;

/**
 * Service Implementation for managing UtmConfigurationParameter.
 */
@Service
@Transactional
public class UtmConfigurationParameterService {

    private static final String CLASSNAME = "UtmConfigurationParameterService";
    private final Logger log = LoggerFactory.getLogger(UtmConfigurationParameterService.class);

    private final UtmConfigurationParameterRepository utmConfigurationParameterRepository;

    public UtmConfigurationParameterService(UtmConfigurationParameterRepository utmConfigurationParameterRepository) {
        this.utmConfigurationParameterRepository = utmConfigurationParameterRepository;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void init() {
        final String ctx = CLASSNAME + ".init";
        try {
            List<UtmConfigurationParameter> params = utmConfigurationParameterRepository.findAll();
            if (CollectionUtils.isEmpty(params))
                return;
            params.forEach(p -> {
                String value = p.getConfParamValue();
                if (StringUtils.hasText(value) && p.getConfParamDatatype().equalsIgnoreCase("password"))
                    value = CipherUtil.decrypt(value, System.getenv(Constants.ENV_ENCRYPTION_KEY));
                Constants.CFG.put(p.getConfParamShort(), value);
            });
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    public void saveAll(List<UtmConfigurationParameter> parameters) throws Exception {
        final String ctx = CLASSNAME + ".saveAll";
        try {
            Map<String, String> cfg = new HashMap<>();
            parameters.forEach(p -> {
                cfg.put(p.getConfParamShort(), p.getConfParamValue());
                if (StringUtils.hasText(p.getConfParamValue()) && p.getConfParamDatatype().equalsIgnoreCase("password"))
                    p.setConfParamValue(CipherUtil.encrypt(p.getConfParamValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
            });
            utmConfigurationParameterRepository.saveAll(parameters);
            Constants.CFG.putAll(cfg);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get all the utmConfigurationParameters.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmConfigurationParameter> findAll(Pageable pageable) {
        log.debug("Request to get all UtmConfigurationParameters");
        return utmConfigurationParameterRepository.findAll(pageable);
    }


    /**
     * Get one utmConfigurationParameter by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmConfigurationParameter> findOne(Long id) {
        log.debug("Request to get UtmConfigurationParameter : {}", id);
        return utmConfigurationParameterRepository.findById(id);
    }

    /**
     * Delete the utmConfigurationParameter by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmConfigurationParameter : {}", id);
        utmConfigurationParameterRepository.deleteById(id);
    }

    public Map<String, String> getValueMapForDateSetting() throws Exception {
        final String ctx = CLASSNAME + ".getValueMapForDateSetting";
        try {
            return utmConfigurationParameterRepository
                .findAllBySectionId(DATE_FORMAT_SETTING_ID).stream()
                .collect(Collectors.toMap(UtmConfigurationParameter::getConfParamShort,
                    UtmConfigurationParameter::getConfParamValue));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
