package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceQueryConfig;
import com.park.utmstack.repository.compliance.UtmComplianceQueryConfigRepository;
import com.park.utmstack.service.dto.compliance.*;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.time.ZoneOffset;
import java.util.*;
import java.util.function.Function;
import java.util.stream.Collectors;

@Service
public class UtmComplianceControlEvaluationHistoryService {

    private final ElasticsearchService elasticsearchService;
    private final UtmComplianceQueryConfigRepository queryConfigRepository;

    public UtmComplianceControlEvaluationHistoryService(ElasticsearchService elasticsearchService,
                                                        UtmComplianceQueryConfigRepository QueryConfigRepository) {
        this.elasticsearchService = elasticsearchService;
        this.queryConfigRepository = QueryConfigRepository;
    }

    public List<UtmComplianceControlEvaluationHistoryDto> findByControlId(Long controlId) {
        return elasticsearchService.getControlEvaluations(controlId);
    }

    public UtmComplianceControlEvaluationHistoryResponseDto getEvaluationsWithRange(Long controlId) {
        var evaluations = findByControlId(controlId);

        if (evaluations.isEmpty()) {
            return new UtmComplianceControlEvaluationHistoryResponseDto(null, null, List.of());
        }

        var queryConfigIds = evaluations.stream()
                .flatMap(ev -> ev.getQueryEvaluations().stream())
                .map(UtmComplianceQueryEvaluationDto::getQueryConfigId)
                .collect(Collectors.toSet());

        var configMap = queryConfigRepository.findAllById(queryConfigIds).stream()
                .collect(Collectors.toMap(UtmComplianceQueryConfig::getId, Function.identity()));

        List<UtmComplianceControlEvaluationGroupedDto> groupedList =
                enrichQueries(evaluations, configMap).stream()
                        .map(evaluation -> {
                            var grouped = groupByIndexPattern(evaluation);
                            return buildGroupedDto(evaluation, grouped);
                        })
                        .toList();

        var timestamps = evaluations.stream()
                .map(UtmComplianceControlEvaluationHistoryDto::getTimestamp)
                .toList();

        return new UtmComplianceControlEvaluationHistoryResponseDto(
                timestamps.stream().min(Instant::compareTo)
                        .get().atZone(ZoneOffset.UTC).toLocalDate(),
                timestamps.stream().max(Instant::compareTo)
                        .get().atZone(ZoneOffset.UTC).toLocalDate(),
                groupedList);
    }


    private List<UtmComplianceControlEvaluationHistoryDto> enrichQueries(
            List<UtmComplianceControlEvaluationHistoryDto> evaluations,
            Map<Long, UtmComplianceQueryConfig> configMap
    ) {
        evaluations.forEach(controlEval ->
                controlEval.getQueryEvaluations().forEach(queryEval -> {
                    var cfg = configMap.get(queryEval.getQueryConfigId());
                    if (cfg != null) {
                        queryEval.setQueryDescription(cfg.getQueryDescription());
                        queryEval.setEvaluationRule(cfg.getEvaluationRule().name());
                        queryEval.setIndexPatternId(cfg.getIndexPattern().getId());
                        queryEval.setIndexPatternName(cfg.getIndexPattern().getPattern());
                    }
                })
        );
        return evaluations;
    }

    private List<UtmComplianceIndexPatternQueriesGroupDto> groupByIndexPattern(
            UtmComplianceControlEvaluationHistoryDto evaluation
    ) {
        return evaluation.getQueryEvaluations().stream()
                .collect(Collectors.groupingBy(UtmComplianceQueryEvaluationDto::getIndexPatternId))
                .entrySet().stream()
                .map(entry -> {
                    var first = entry.getValue().get(0);
                    var dto = new UtmComplianceIndexPatternQueriesGroupDto();
                    dto.setIndexPatternId(entry.getKey());
                    dto.setIndexPatternName(first.getIndexPatternName());
                    dto.setQueries(entry.getValue());
                    return dto;
                })
                .toList();
    }

    private UtmComplianceControlEvaluationGroupedDto buildGroupedDto(
            UtmComplianceControlEvaluationHistoryDto evaluation,
            List<UtmComplianceIndexPatternQueriesGroupDto> groupedEvaluations
    ) {
        var dto = new UtmComplianceControlEvaluationGroupedDto();
        dto.setControlId(evaluation.getControlId());
        dto.setControlName(evaluation.getControlName());
        dto.setStatus(evaluation.getStatus());
        dto.setTimestamp(evaluation.getTimestamp());
        dto.setQueryEvaluations(groupedEvaluations);

        return dto;
    }
}