package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceStandard;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the ComplianceTemplate entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmComplianceStandardRepository extends JpaRepository<UtmComplianceStandard, Long>, JpaSpecificationExecutor<UtmComplianceStandard> {

    Optional<UtmComplianceStandard> findByStandardNameLike(String standardName);

    void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> standardIds);

    Optional<UtmComplianceStandard> findFirstBySystemOwnerIsTrueOrderByIdDesc();
}
