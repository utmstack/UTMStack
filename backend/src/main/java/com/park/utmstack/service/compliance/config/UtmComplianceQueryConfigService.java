package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceQueryConfig;
import com.park.utmstack.repository.compliance.UtmComplianceQueryConfigRepository;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UtmComplianceQueryConfigService {

    private final UtmComplianceQueryConfigRepository repository;

    public UtmComplianceQueryConfigService(UtmComplianceQueryConfigRepository repository) {
        this.repository = repository;
    }
    
    public UtmComplianceQueryConfig create(UtmComplianceQueryConfig config) {
        return repository.save(config);
    }

    public List<UtmComplianceQueryConfig> findByReportConfig(Long reportConfigId) {
        return repository.findByControlConfigId(reportConfigId);
    }

    public void delete(Long id) {
        repository.deleteById(id);
    }
}

