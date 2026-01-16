package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import com.park.utmstack.repository.compliance.UtmComplianceControlConfigRepository;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigResponseDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryConfigRequestDto;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlConfigMapper;
import com.park.utmstack.service.mapper.compliance.UtmComplianceQueryConfigMapper;
import org.springframework.stereotype.Service;

import javax.transaction.Transactional;
import java.util.ArrayList;
import java.util.List;

@Service
public class UtmComplianceControlConfigService {

    private final UtmComplianceControlConfigRepository repository;
    private final UtmComplianceControlConfigMapper mapper;
    private final UtmComplianceQueryConfigMapper queryMapper;

    public UtmComplianceControlConfigService(UtmComplianceControlConfigRepository repository,
                                             UtmComplianceControlConfigMapper mapper,
                                             UtmComplianceQueryConfigMapper queryMapper) {
        this.repository = repository;
        this.mapper = mapper;
        this.queryMapper = queryMapper;
    }

    @Transactional
    public UtmComplianceControlConfigResponseDto create(UtmComplianceControlConfigRequestDto dto) {
        UtmComplianceControlConfig entity = mapper.toEntity(dto);
        entity.setQueriesConfigs(new ArrayList<>());

        entity = repository.save(entity);

        for (var qdto : dto.getQueriesConfigs()) {
            var q = queryMapper.toEntity(qdto);
            q.setControlConfigId(entity.getId());
            entity.getQueriesConfigs().add(q);
        }

        entity = repository.save(entity);

        return mapper.toResponse(entity);
    }


    @Transactional
    public UtmComplianceControlConfigResponseDto update(Long id, UtmComplianceControlConfigRequestDto dto) {

        UtmComplianceControlConfig entity = repository.findById(id)
                .orElseThrow(() -> new RuntimeException("Control not found"));

        mapper.updateEntity(entity, dto);
        entity.getQueriesConfigs().clear();

        for (var qdto : dto.getQueriesConfigs()) {
            var q = queryMapper.toEntity(qdto);
            q.setControlConfigId(id);
            entity.getQueriesConfigs().add(q);
        }

        entity = repository.save(entity);

        return mapper.toResponse(entity);
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
