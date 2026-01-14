package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceControlQueryConfig;

import java.util.List;

public interface UtmComplianceControlQueryConfigService {

    UtmComplianceControlQueryConfig create(UtmComplianceControlQueryConfig config);

    List<UtmComplianceControlQueryConfig> findByReportConfig(Long reportConfigId);

    void delete(Long id);
}

