package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface UtmComplianceControlConfigRepository extends JpaRepository<UtmComplianceControlConfig, Long> {

}
