package com.park.utmstack.service.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseActionTemplate;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseActionTemplateRepository;
import com.park.utmstack.service.UtmStackService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@Transactional
public class UtmAlertResponseActionTemplateService {

    private static final String CLASSNAME = "UtmAlertResponseActionTemplateService";
    private final Logger log = LoggerFactory.getLogger(UtmAlertResponseActionTemplateService.class);

    private final UtmAlertResponseActionTemplateRepository utmAlertResponseActionTemplateRepository;

    private final UtmStackService utmStackService;


    public UtmAlertResponseActionTemplateService(UtmAlertResponseActionTemplateRepository utmAlertResponseActionTemplateRepository,
                                                 UtmStackService utmStackService) {
        this.utmAlertResponseActionTemplateRepository = utmAlertResponseActionTemplateRepository;
        this.utmStackService = utmStackService;
    }

    public UtmAlertResponseActionTemplate save(UtmAlertResponseActionTemplate alertResponseActionTemplate) {
        final String ctx = CLASSNAME + ".save";
        try {
            if (utmStackService.isInDevelop()) {
                alertResponseActionTemplate.setId(this.getSystemSequenceNextValue());
                alertResponseActionTemplate.setSystemOwner(true);
            } else {
                alertResponseActionTemplate.setSystemOwner(false);
            }

            return utmAlertResponseActionTemplateRepository.save(alertResponseActionTemplate);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }


    public Long getSystemSequenceNextValue() {
        return utmAlertResponseActionTemplateRepository.findFirstBySystemOwnerIsTrueOrderByIdDesc()
                .map(rule -> rule.getId() + 1)
                .orElse(1L);
    }

}
