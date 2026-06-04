package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the ComplianceTemplate entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmComplianceStandardSectionRepository extends JpaRepository<UtmComplianceStandardSection, Long>, JpaSpecificationExecutor<UtmComplianceStandardSection> {

    @Query(nativeQuery = true, value = "select * from utm_compliance_standard_section css where css.id in (select crc.standard_section_id from utm_compliance_report_config crc) and css.standard_id = :standardId")
    List<UtmComplianceStandardSection> getSectionsWithReportsByStandard(@Param("standardId") Long standardId);

    Optional<UtmComplianceStandardSection> findByStandardSectionNameLike(String sectionName);

    List<UtmComplianceStandardSection> findAllByStandardSectionNameNotIn(List<String> sectionNames);

    List<UtmComplianceStandardSection> findAllByStandardIdAndStandardSectionNameNotIn(Long standardId, List<String> sectionNames);


    @Modifying
    @Query("delete from UtmComplianceStandardSection s where s.standardId = :standardId and s.id not in :sectionIds")
    void deleteSectionsByStandardAndIdNotIn(@Param("standardId") Long standardId, @Param("sectionIds") List<Long> sectionIds);

    Optional<UtmComplianceStandardSection> findFirstByStandardIdOrderByIdDesc(Long standardId);
}
