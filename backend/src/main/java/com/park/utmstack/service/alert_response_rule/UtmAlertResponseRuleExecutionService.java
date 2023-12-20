package com.park.utmstack.service.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleExecution;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleExecutionRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

@Service
@Transactional
public class UtmAlertResponseRuleExecutionService {

    private static final String CLASSNAME = "UtmAlertResponseRuleExecutionService";
    private final Logger log = LoggerFactory.getLogger(UtmAlertResponseRuleExecutionService.class);

    private final UtmAlertResponseRuleExecutionRepository ruleExecutionRepository;

    public UtmAlertResponseRuleExecutionService(UtmAlertResponseRuleExecutionRepository ruleExecutionRepository) {
        this.ruleExecutionRepository = ruleExecutionRepository;
    }

    public UtmAlertResponseRuleExecution save(UtmAlertResponseRuleExecution execution) {
        final String ctx = CLASSNAME + ".save";
        try {
            return ruleExecutionRepository.save(execution);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(e);
        }
    }

    @Transactional(readOnly = true)
    public Page<UtmAlertResponseRuleExecution> findAll(Pageable pageable) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return ruleExecutionRepository.findAll(pageable);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    @Transactional(readOnly = true)
    public Optional<UtmAlertResponseRuleExecution> findOne(Long id) {
        final String ctx = CLASSNAME + ".findOne";
        try {
            return ruleExecutionRepository.findById(id);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    public void delete(Long id) {
        final String ctx = CLASSNAME + ".delete";
        try {
            ruleExecutionRepository.deleteById(id);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }
}
