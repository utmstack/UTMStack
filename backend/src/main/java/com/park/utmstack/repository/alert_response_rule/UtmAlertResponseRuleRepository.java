package com.park.utmstack.repository.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRule;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmAlertSoarRule entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmAlertResponseRuleRepository extends JpaRepository<UtmAlertResponseRule, Long>, JpaSpecificationExecutor<UtmAlertResponseRule> {

    @Query("select distinct r.agentPlatform from UtmAlertResponseRule r")
    List<String> findAgentPlatformValues();

    @Query(nativeQuery = true, value = "SELECT r1.created_by AS users FROM utm_alert_response_rule r1 UNION SELECT r2.last_modified_by FROM utm_alert_response_rule r2 where r2.last_modified_by is not null")
    List<String> findUserValues();

    List<UtmAlertResponseRule> findAllByRuleActiveIsTrue();

}
