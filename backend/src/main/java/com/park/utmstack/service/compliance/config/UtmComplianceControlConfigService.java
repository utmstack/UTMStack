package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import com.park.utmstack.domain.compliance.UtmComplianceQueryConfig;
import com.park.utmstack.domain.compliance.enums.EvaluationRule;
import com.park.utmstack.repository.compliance.UtmComplianceControlConfigRepository;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlConfigMapper;
import com.park.utmstack.service.mapper.compliance.UtmComplianceQueryConfigMapper;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;

import javax.transaction.Transactional;
import java.util.List;
import java.util.Map;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

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
    public UtmComplianceControlConfigDto create(UtmComplianceControlConfigDto dto) {

        validateControlConfig(dto);

        UtmComplianceControlConfig entity = mapper.toEntity(dto);

        for (var qdto : dto.getQueriesConfigs()) {
            var q = queryMapper.toEntity(qdto);
            q.setControlConfig(entity);
            entity.getQueriesConfigs().add(q);
        }

        entity = repository.save(entity);

        return mapper.toDto(entity);
    }

    @Transactional
    public UtmComplianceControlConfigDto update(Long id, UtmComplianceControlConfigDto dto) {

        validateControlConfig(dto);

        UtmComplianceControlConfig entity = repository.findByIdWithQueries(id)
                .orElseThrow(() -> new NoSuchElementException("Control not found"));

        mapper.updateEntity(entity, dto);

        Map<Long, UtmComplianceQueryConfig> existing = entity.getQueriesConfigs()
                .stream()
                .collect(Collectors.toMap(UtmComplianceQueryConfig::getId, q -> q));

        entity.getQueriesConfigs().clear();

        for (var qdto : dto.getQueriesConfigs()) {
            if (qdto.getId() != null && existing.containsKey(qdto.getId())) {
                var q = existing.get(qdto.getId());
                queryMapper.updateEntity(q, qdto);
                q.setControlConfig(entity);
                entity.getQueriesConfigs().add(q);
            } else {
                var q = queryMapper.toEntity(qdto);
                q.setControlConfig(entity);
                entity.getQueriesConfigs().add(q);
            }
        }
        return mapper.toDto(entity);
    }

    public void delete(Long id) {
        if (!repository.existsById(id)) {
            throw new NoSuchElementException("Control not found");
        }
        repository.deleteById(id);
    }


    public UtmComplianceControlConfigDto findById(Long id) {
        var entity = repository.findByIdWithQueries(id)
                .orElseThrow(() -> new NoSuchElementException("Control not found"));

        return mapper.toDto(entity);
    }

    private void validateControlConfig(UtmComplianceControlConfigDto dto) {
        if (dto.getQueriesConfigs() == null || dto.getQueriesConfigs().isEmpty()) {
            throw new BadRequestAlertException(
                    "At least one query configuration is required",
                    "utmComplianceControlConfig",
                    "queriesConfigsEmpty"
            );
        }

        for (var q : dto.getQueriesConfigs()) {
            if (q.getEvaluationRule() != EvaluationRule.NO_HITS_ALLOWED && q.getRuleValue() == null) {
                throw new BadRequestAlertException(
                        "ruleValue is required when evaluationRule is not NO_HITS_ALLOWED",
                        "utmComplianceQueryConfig",
                        "ruleValueMissing"
                );
            }
        }
    }

    public Page<UtmComplianceControlConfig> findBySection(Long sectionId, String search, Pageable pageable) {

        Specification<UtmComplianceControlConfig> spec = Specification.where(UtmComplianceControlConfigRepository.bySection(sectionId));

        if (search != null && !search.isBlank()) {
            spec = spec.and(UtmComplianceControlConfigRepository.nameContains(search));
        }

        Page<UtmComplianceControlConfig> page = repository.findAll(spec, pageable);

        List<Long> ids = page.map(UtmComplianceControlConfig::getId).getContent();

        if (ids.isEmpty()) {
            return Page.empty(pageable);
        }

        List<UtmComplianceControlConfig> content = repository.findWithQueriesByIdIn(ids);

        if (content.isEmpty()) {
            return Page.empty(pageable);
        }

        Map<Long, UtmComplianceControlConfig> map = content.stream().collect(Collectors.toMap(UtmComplianceControlConfig::getId, c -> c));
        List<UtmComplianceControlConfig> ordered = ids.stream().map(map::get).collect(Collectors.toList());

        return new PageImpl<>(ordered, pageable, page.getTotalElements());
    }

}
