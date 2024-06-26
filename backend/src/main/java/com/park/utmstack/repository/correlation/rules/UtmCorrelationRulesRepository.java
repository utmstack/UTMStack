package com.park.utmstack.repository.correlation.rules;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;

/**
 * Spring Data  repository for the UtmCorrelationRules entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmCorrelationRulesRepository extends JpaRepository<UtmCorrelationRules, Long>, JpaSpecificationExecutor<UtmCorrelationRules> {

}
