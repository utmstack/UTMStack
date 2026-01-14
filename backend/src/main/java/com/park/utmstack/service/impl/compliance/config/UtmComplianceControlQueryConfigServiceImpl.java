package com.park.utmstack.service.impl.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceControlQueryConfig;
import com.park.utmstack.repository.compliance.UtmComplianceControlQueryConfigRepository;
import com.park.utmstack.service.compliance.config.UtmComplianceControlQueryConfigService;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UtmComplianceControlQueryConfigServiceImpl implements UtmComplianceControlQueryConfigService {

    private final UtmComplianceControlQueryConfigRepository repository;

    public UtmComplianceControlQueryConfigServiceImpl(UtmComplianceControlQueryConfigRepository repository) {
        this.repository = repository;
    }

    @Override
    public UtmComplianceControlQueryConfig create(UtmComplianceControlQueryConfig config) {
        return repository.save(config);
    }

    @Override
    public List<UtmComplianceControlQueryConfig> findByReportConfig(Long reportConfigId) {
        return repository.findByControlConfigId(reportConfigId);
    }

    @Override
    public void delete(Long id) {
        repository.deleteById(id);
    }
}

