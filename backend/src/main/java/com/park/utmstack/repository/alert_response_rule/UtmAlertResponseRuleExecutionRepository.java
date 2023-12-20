package com.park.utmstack.repository.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleExecution;
import com.park.utmstack.domain.alert_response_rule.enums.RuleExecutionStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;


@SuppressWarnings("unused")
@Repository
public interface UtmAlertResponseRuleExecutionRepository extends JpaRepository<UtmAlertResponseRuleExecution, Long>, JpaSpecificationExecutor<UtmAlertResponseRuleExecution> {
    List<UtmAlertResponseRuleExecution> findAllByExecutionStatus(RuleExecutionStatus status);

}
