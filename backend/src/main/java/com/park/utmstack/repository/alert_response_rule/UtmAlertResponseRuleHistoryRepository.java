package com.park.utmstack.repository.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleHistory;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmAlertResponseRuleHistory entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmAlertResponseRuleHistoryRepository extends JpaRepository<UtmAlertResponseRuleHistory, Long>, JpaSpecificationExecutor<UtmAlertResponseRuleHistory> {

}
