package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceControlQueryConfig;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface UtmComplianceControlQueryConfigRepository extends JpaRepository<UtmComplianceControlQueryConfig, Long> {

    List<UtmComplianceControlQueryConfig> findByControlConfigId(Long controlConfigId);

}

