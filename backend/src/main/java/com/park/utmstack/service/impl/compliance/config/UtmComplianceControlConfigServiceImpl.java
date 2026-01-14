package com.park.utmstack.service.impl.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import com.park.utmstack.repository.compliance.UtmComplianceControlConfigRepository;
import com.park.utmstack.service.compliance.config.UtmComplianceControlConfigService;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UtmComplianceControlConfigServiceImpl implements UtmComplianceControlConfigService {

    private final UtmComplianceControlConfigRepository repository;

    public UtmComplianceControlConfigServiceImpl(UtmComplianceControlConfigRepository repository) {
        this.repository = repository;
    }

    @Override
    public UtmComplianceControlConfig create(UtmComplianceControlConfig config) {
        return repository.save(config);
    }

    @Override
    public UtmComplianceControlConfig update(Long id, UtmComplianceControlConfig config) {
        config.setId(id);
        return repository.save(config);
    }

    @Override
    public void delete(Long id) {
        repository.deleteById(id);
    }

    @Override
    public UtmComplianceControlConfig findById(Long id) {
        return repository.findById(id).orElse(null);
    }

    @Override
    public List<UtmComplianceControlConfig> findAll() {
        return repository.findAll();
    }
}
