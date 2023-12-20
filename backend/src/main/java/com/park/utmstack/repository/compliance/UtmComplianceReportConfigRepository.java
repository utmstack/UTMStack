package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceReportConfig;
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
@SuppressWarnings("unused")
@Repository
public interface UtmComplianceReportConfigRepository extends JpaRepository<UtmComplianceReportConfig, Long>, JpaSpecificationExecutor<UtmComplianceReportConfig> {

    void deleteAllByConfigSolutionAndStandardSectionIdAndDashboardId(String configSolution, Long sectionId, Long dashboardId);

    Optional<UtmComplianceReportConfig> findByConfigSolutionAndStandardSectionIdAndDashboardId(String configSolution, Long sectionId, Long dashboardId);

    @Modifying
    @Transactional
    @Query("delete from UtmComplianceReportConfig r where r.standardSectionId in (select s.id from UtmComplianceStandardSection s where s.standardId = :standardId) and r.id not in :reportIds")
    void deleteReportsByStandardIdAndIdNotIn(@Param("standardId") Long standardId, @Param("reportIds") List<Long> reportIds);
}
