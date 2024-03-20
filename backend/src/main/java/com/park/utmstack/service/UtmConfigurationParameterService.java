package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.UtmConfigurationParameter;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.repository.UtmConfigurationParameterRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.exceptions.UtmMailException;
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

    private final UtmConfigurationParameterRepository configParamRepository;
    private final UserService userService;
    private final MailService mailService;
    private final ApplicationEventService applicationEventService;

    public UtmConfigurationParameterService(UtmConfigurationParameterRepository configParamRepository,
                                            UserService userService,
                                            MailService mailService,
                                            ApplicationEventService applicationEventService) {
        this.configParamRepository = configParamRepository;
        this.userService = userService;
        this.mailService = mailService;
        this.applicationEventService = applicationEventService;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void init() {
        final String ctx = CLASSNAME + ".init";
        try {
            List<UtmConfigurationParameter> params = configParamRepository.findAll();
            if (CollectionUtils.isEmpty(params))
                return;
            for (UtmConfigurationParameter p : params) {
                String value = p.getConfParamValue();
                if (StringUtils.hasText(value) && p.getConfParamDatatype().equalsIgnoreCase("password"))
                    try {
                        value = CipherUtil.decrypt(value, System.getenv(Constants.ENV_ENCRYPTION_KEY));
                    } catch (Exception e) {
                        String msg = String.format("%1$s: Fail to decrypt the value of the configuration parameter %2$s, error is: %3$s",
                                ctx, p.getConfParamLarge(), e.getLocalizedMessage());
                        log.error(msg);
                        applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
                        continue;
                    }
                Constants.CFG.put(p.getConfParamShort(), value);
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    public void saveAll(List<UtmConfigurationParameter> params) throws UtmMailException {
        final String ctx = CLASSNAME + ".saveAll";
        try {
            // If the configuration to save is: Enable Two-Factor Authentication then we need to check
            // if the email configuration is OK
            params.stream().filter(p -> p.getConfParamShort().equals(Constants.PROP_TFA_ENABLE)
                            && Boolean.parseBoolean(p.getConfParamValue()))
                    .findFirst().ifPresent(tfa -> validateMailConfOnMFAActivation());

            Map<String, String> cfg = new HashMap<>();
            for (UtmConfigurationParameter p : params) {
                cfg.put(p.getConfParamShort(), p.getConfParamValue());
                if (StringUtils.hasText(p.getConfParamValue()) && p.getConfParamDatatype().equalsIgnoreCase("password"))
                    p.setConfParamValue(CipherUtil.encrypt(p.getConfParamValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
            }
            configParamRepository.saveAll(params);
            Constants.CFG.putAll(cfg);
        } catch (UtmMailException e) {
            throw new UtmMailException(ctx + ": " + e.getMessage());
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
        return configParamRepository.findAll(pageable);
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
        return configParamRepository.findById(id);
    }

    /**
     * Delete the utmConfigurationParameter by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmConfigurationParameter : {}", id);
        configParamRepository.deleteById(id);
    }

    public Map<String, String> getValueMapForDateSetting() throws Exception {
        final String ctx = CLASSNAME + ".getValueMapForDateSetting";
        try {
            return configParamRepository
                    .findAllBySectionId(DATE_FORMAT_SETTING_ID).stream()
                    .collect(Collectors.toMap(UtmConfigurationParameter::getConfParamShort,
                            UtmConfigurationParameter::getConfParamValue));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private void validateMailConfOnMFAActivation() throws UtmMailException {
        final String ctx = CLASSNAME + ".validateMailConfOnMFAActivation";
        try {
            mailService.sendCheckEmail(List.of(userService.getCurrentUserLogin().getEmail()));
        } catch (Exception e) {
            throw new UtmMailException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}
