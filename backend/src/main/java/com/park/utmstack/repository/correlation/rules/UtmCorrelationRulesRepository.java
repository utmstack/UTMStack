package com.park.utmstack.repository.correlation.rules;

import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;

import java.time.Instant;
import java.util.List;

/**
 * Spring Data  repository for the UtmCorrelationRules entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmCorrelationRulesRepository extends JpaRepository<UtmCorrelationRules, Long>, JpaSpecificationExecutor<UtmCorrelationRules> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_correlation_rules_id_seq')")
    Long getNextId();

    @Query(value = "SELECT DISTINCT cr FROM UtmCorrelationRules cr " +
            "LEFT JOIN cr.dataTypes dt " +
            "WHERE " +
            "(:ruleName IS NULL OR (cr.ruleName LIKE :ruleName OR lower(cr.ruleName) LIKE lower(:ruleName))) " +
            "AND ((:ruleConfidentiality) IS NULL OR cr.ruleConfidentiality IN (:ruleConfidentiality)) " +
            "AND ((:ruleIntegrity) IS NULL OR cr.ruleIntegrity IN (:ruleIntegrity)) " +
            "AND ((:ruleAvailability) IS NULL OR cr.ruleAvailability IN (:ruleAvailability)) " +
            "AND ((:ruleCategory) IS NULL OR cr.ruleCategory IN (:ruleCategory)) " +
            "AND ((:ruleTechnique) IS NULL OR cr.ruleTechnique IN (:ruleTechnique)) " +
            "AND ((:ruleActive) IS NULL OR cr.ruleActive IN (:ruleActive)) " +
            "AND ((:systemOwner) IS NULL OR cr.systemOwner IN (:systemOwner)) " +
            "AND ((:ruleSearch) IS NULL OR " +
            "(cr.ruleName LIKE :ruleSearch OR lower(cr.ruleName) LIKE lower(:ruleSearch))) " +
            "AND ((cast(:ruleInitDate as timestamp) is null) or (cast(:ruleEndDate as timestamp) is null) or (cr.ruleLastUpdate BETWEEN :ruleInitDate AND :ruleEndDate)) " +
            "AND ((:dataTypes) IS NULL OR dt.dataType IN (:dataTypes))")
    Page<UtmCorrelationRules> searchByFilters(@Param("ruleName") String ruleName,
                                              @Param("ruleConfidentiality") List<Integer> ruleConfidentiality,
                                              @Param("ruleIntegrity") List<Integer> ruleIntegrity,
                                              @Param("ruleAvailability") List<Integer> ruleAvailability,
                                              @Param("ruleCategory") List<String> ruleCategory,
                                              @Param("ruleTechnique") List<String> ruleTechnique,
                                              @Param("ruleActive") List<Boolean> ruleActive,
                                              @Param("systemOwner") List<Boolean> systemOwner,
                                              @Param("dataTypes") List<String> dataTypes,
                                              @Param("ruleInitDate") Instant ruleInitDate,
                                              @Param("ruleEndDate") Instant ruleEndDate,
                                              @Param("ruleSearch")  String ruleSearch,
                                              Pageable pageable);
}
