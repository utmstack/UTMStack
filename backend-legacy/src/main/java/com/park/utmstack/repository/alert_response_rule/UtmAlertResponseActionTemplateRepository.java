package com.park.utmstack.repository.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseActionTemplate;
import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.Optional;


/**
 * Spring Data  repository for the UtmAlertSoarRule entity.
 */

@Repository
public interface UtmAlertResponseActionTemplateRepository extends JpaRepository<UtmAlertResponseActionTemplate, Long>, JpaSpecificationExecutor<UtmAlertResponseActionTemplate> {

    Optional<UtmAlertResponseActionTemplate> findFirstBySystemOwnerIsTrueOrderByIdDesc();

}
