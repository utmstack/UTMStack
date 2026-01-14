package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;

import java.util.List;

public interface UtmComplianceControlConfigService {

    UtmComplianceControlConfig create(UtmComplianceControlConfig config);

    UtmComplianceControlConfig update(Long id, UtmComplianceControlConfig config);

    void delete(Long id);

    UtmComplianceControlConfig findById(Long id);

    List<UtmComplianceControlConfig> findAll();
}
