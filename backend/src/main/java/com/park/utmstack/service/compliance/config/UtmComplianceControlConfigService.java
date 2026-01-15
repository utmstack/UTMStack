package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import com.park.utmstack.repository.compliance.UtmComplianceControlConfigRepository;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UtmComplianceControlConfigService {

    private final UtmComplianceControlConfigRepository repository;

    public UtmComplianceControlConfigService(UtmComplianceControlConfigRepository repository) {
        this.repository = repository;
    }
    
    public UtmComplianceControlConfig create(UtmComplianceControlConfig config) {
        return repository.save(config);
    }
    
    public UtmComplianceControlConfig update(Long id, UtmComplianceControlConfig config) {
        config.setId(id);
        return repository.save(config);
    }
    
    public void delete(Long id) {
        repository.deleteById(id);
    }

    public UtmComplianceControlConfig findById(Long id) {
        return repository.findById(id).orElse(null);
    }
    
    public List<UtmComplianceControlConfig> findAll() {
        return repository.findAll();
    }
}
