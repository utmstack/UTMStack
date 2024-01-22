package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceReportSchedule;
import org.springframework.data.jpa.repository.*;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

/**
 * Spring Data SQL repository for the UtmComplianceReportSchedule entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmComplianceReportScheduleRepository extends JpaRepository<UtmComplianceReportSchedule, Long> {

    Optional<UtmComplianceReportSchedule> findFirstByUserIdAndComplianceIdAndScheduleString(Long userId, Long complianceId, String scheduleString);

    List<UtmComplianceReportSchedule> findAllByUserId(Long userId);

    void deleteByUserIdAndComplianceId(Long userId, Long complianceId);
}
