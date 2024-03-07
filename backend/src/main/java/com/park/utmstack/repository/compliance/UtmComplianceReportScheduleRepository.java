package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceReportSchedule;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.data.jpa.repository.*;
import org.springframework.lang.Nullable;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

/**
 * Spring Data SQL repository for the UtmComplianceReportSchedule entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmComplianceReportScheduleRepository extends JpaRepository<UtmComplianceReportSchedule, Long>, JpaSpecificationExecutor<UtmComplianceReportSchedule> {

    Optional<UtmComplianceReportSchedule> findFirstByUserIdAndComplianceIdAndScheduleString(Long userId, Long complianceId, String scheduleString);

    Page<UtmComplianceReportSchedule> findAllByUserId(Long userId, Pageable pageable, Specification<UtmComplianceReportSchedule> specification);

    void deleteByUserIdAndComplianceId(Long userId, Long complianceId);
}
