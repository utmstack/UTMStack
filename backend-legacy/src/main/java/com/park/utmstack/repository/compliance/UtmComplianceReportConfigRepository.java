package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceReportConfig;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the ComplianceTemplate entity.
 */
@Repository
public interface UtmComplianceReportConfigRepository extends JpaRepository<UtmComplianceReportConfig, Long>, JpaSpecificationExecutor<UtmComplianceReportConfig> {

    void deleteAllByConfigSolutionAndStandardSectionIdAndDashboardId(String configSolution, Long sectionId, Long dashboardId);

    Optional<UtmComplianceReportConfig> findByConfigSolutionAndStandardSectionIdAndDashboardId(String configSolution, Long sectionId, Long dashboardId);

    @Modifying
    @Transactional
    @Query("delete from UtmComplianceReportConfig r where r.standardSectionId in (select s.id from UtmComplianceStandardSection s where s.standardId = :standardId) and r.id not in :reportIds")
    void deleteReportsByStandardIdAndIdNotIn(@Param("standardId") Long standardId, @Param("reportIds") List<Long> reportIds);

    @Query(value = "SELECT DISTINCT cfg FROM UtmComplianceReportConfig cfg " +
            "JOIN cfg.section sec " +
            "LEFT JOIN cfg.associatedDashboard d " +
            "LEFT JOIN d.visualizations v " +
            "WHERE (:standardId IS NULL OR sec.standardId = :standardId) " +
            "AND (:solution IS NULL OR lower(cfg.configSolution) LIKE %:solution%) " +
            "AND (:sectionId IS NULL OR sec.id = :sectionId) " +
            "AND (:search IS NULL OR lower(cfg.configReportName) LIKE %:search% OR d.name LIKE %:search%) " +
            "AND (v.pattern.isActive = true)")
    Page<UtmComplianceReportConfig> getReportsByFilters(
            @Param("standardId") Long standardId,
            @Param("solution") String solution,
            @Param("sectionId") Long sectionId,
            @Param("search") String search,
            Pageable pageable);

}
