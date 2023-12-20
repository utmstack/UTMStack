package com.park.utmstack.repository.reports;

import com.park.utmstack.domain.reports.UtmReport;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmReports entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmReportRepository extends JpaRepository<UtmReport, Long>, JpaSpecificationExecutor<UtmReport> {

    void deleteAllByReportSectionIdAndIdNotIn(Long sectionId, List<Long> ids);

}
