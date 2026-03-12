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
public class UtmComplianceControlEvaluationsService {

    private final ElasticsearchService elasticsearchService;
    private final UtmComplianceQueryConfigRepository queryConfigRepository;

    public UtmComplianceControlEvaluationsService(ElasticsearchService elasticsearchService,
                                                  UtmComplianceQueryConfigRepository QueryConfigRepository) {
        this.elasticsearchService = elasticsearchService;
        this.queryConfigRepository = QueryConfigRepository;
    }

    public List<UtmComplianceControlEvaluationsDto> findByControlId(Long controlId) {
        return elasticsearchService.getControlEvaluations(controlId);
    }

    public UtmComplianceControlEvaluationsDto getLastEvaluationForControl(Long controlId) {
        return elasticsearchService.getLatestControlEvaluation(controlId);
    }

    /*public ControlEvaluationsResponseDto getEvaluationsWithRange(Long controlId) {
        var evaluations = findByControlId(controlId);

        if (evaluations.isEmpty()) {
            return new ControlEvaluationsResponseDto(null, null, evaluations);
        }

        var timestamps = evaluations.stream()
                .map(UtmComplianceControlEvaluationsDto::getTimestamp)
                .toList();

        Instant min = timestamps.stream().min(Instant::compareTo).get();
        Instant max = timestamps.stream().max(Instant::compareTo).get();

        LocalDate start = min.atZone(ZoneOffset.UTC).toLocalDate();
        LocalDate end = max.atZone(ZoneOffset.UTC).toLocalDate();

        return new ControlEvaluationsResponseDto(start, end, evaluations);
    }*/

    /*public ControlEvaluationsResponseDto getEvaluationsWithRange(Long controlId) {

        // 1. Evaluaciones crudas desde OpenSearch - UtmComplianceControlEvaluationsDto
        var evaluations = findByControlId(controlId);

        if (evaluations.isEmpty()) {
            return new ControlEvaluationsResponseDto(
                    null,
                    null,
                    List.of()
            );
        }

        // 2. Extraer todos los queryConfigId de todas las evaluaciones
        var queryConfigIds = evaluations.stream()
                .flatMap(ev -> ev.getQueryEvaluations().stream())
                .map(UtmComplianceQueryEvaluationDto::getQueryConfigId)
                .collect(Collectors.toSet());

        // 3. Lookup masivo en DB - UtmComplianceQueryConfig
        var configs = queryConfigRepository.findAllById(queryConfigIds);
        var configMap = configs.stream()
                .collect(Collectors.toMap(
                        UtmComplianceQueryConfig::getId,
                        Function.identity()
                ));

        // 4. Enriquecer cada query evaluation
        evaluations.forEach(controlEval -> {
            controlEval.getQueryEvaluations().forEach(queryEval -> {
                var cfg = configMap.get(queryEval.getQueryConfigId());
                if (cfg != null) {
                    queryEval.setQueryDescription(cfg.getQueryDescription());
                    queryEval.setEvaluationRule(cfg.getEvaluationRule().name());
                    queryEval.setIndexPatternId(cfg.getIndexPattern().getId());
                    queryEval.setIndexPatternName(cfg.getIndexPattern().getPattern());
                }
            });
        });

        // 5. Aplanar todas las queryEvaluations para agrupar por indexPattern
        var allQueryEvaluations = evaluations.stream()
                .flatMap(ev -> ev.getQueryEvaluations().stream())
                .toList();

        // 6. Agrupar por indexPattern
        var groupedEvaluations = allQueryEvaluations.stream()
                .collect(Collectors.groupingBy(UtmComplianceQueryEvaluationDto::getIndexPatternId))
                .entrySet().stream()
                .map(entry -> {
                    var first = entry.getValue().get(0);
                    var dto = new IndexPatternQueriesGroupDto();
                    dto.setIndexPatternId(entry.getKey());
                    dto.setIndexPatternName(first.getIndexPatternName());
                    dto.setQueries(entry.getValue());
                    return dto;
                })
                .toList();

        // 7. Calcular rango de fechas
        var timestamps = evaluations.stream()
                .map(UtmComplianceControlEvaluationsDto::getTimestamp)
                .toList();

        Instant min = timestamps.stream().min(Instant::compareTo).get();
        Instant max = timestamps.stream().max(Instant::compareTo).get();

        LocalDate startDate = min.atZone(ZoneOffset.UTC).toLocalDate();
        LocalDate endDate = max.atZone(ZoneOffset.UTC).toLocalDate();

        var groupedDto = new UtmComplianceControlEvaluationsGroupedDto();
        groupedDto.setControlId(evaluations.get(0).getControlId());
        groupedDto.setControlName(evaluations.get(0).getControlName());
        groupedDto.setStatus(evaluations.get(0).getStatus());
        groupedDto.setTimestamp(evaluations.get(0).getTimestamp());
        groupedDto.setQueryEvaluations(groupedEvaluations);


        // 8. Devolver DTO final enriquecido y agrupado
        return new ControlEvaluationsResponseDto(
                startDate,
                endDate,
                List.of(groupedDto)
        );

    }*/

    public ControlEvaluationsResponseDto getEvaluationsWithRange(Long controlId) {
        //TODO: Elena Ordenar por fecha descendente
        var evaluations = findByControlId(controlId);

        if (evaluations.isEmpty()) {
            return new ControlEvaluationsResponseDto(null, null, List.of());
        }

        var queryConfigIds = evaluations.stream()
                .flatMap(ev -> ev.getQueryEvaluations().stream())
                .map(UtmComplianceQueryEvaluationDto::getQueryConfigId)
                .collect(Collectors.toSet());

        var configMap = queryConfigRepository.findAllById(queryConfigIds).stream()
                .collect(Collectors.toMap(UtmComplianceQueryConfig::getId, Function.identity()));

        //evaluations = enrichQueries(evaluations, configMap);

        List<UtmComplianceControlEvaluationsGroupedDto> groupedList =
                enrichQueries(evaluations, configMap).stream()
                        .map(evaluation -> {
                            var grouped = groupByIndexPattern(evaluation);
                            return buildGroupedDto(evaluation, grouped);
                        })
                        .toList();

        var timestamps = evaluations.stream()
                .map(UtmComplianceControlEvaluationsDto::getTimestamp)
                .toList();

        return new ControlEvaluationsResponseDto(
                timestamps.stream().min(Instant::compareTo)
                        .get().atZone(ZoneOffset.UTC).toLocalDate(),
                timestamps.stream().max(Instant::compareTo)
                        .get().atZone(ZoneOffset.UTC).toLocalDate(),
                groupedList);
    }


    private List<UtmComplianceControlEvaluationsDto> enrichQueries(
            List<UtmComplianceControlEvaluationsDto> evaluations,
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

    private List<IndexPatternQueriesGroupDto> groupByIndexPattern(
            UtmComplianceControlEvaluationsDto evaluation
    ) {
        return evaluation.getQueryEvaluations().stream()
                .collect(Collectors.groupingBy(UtmComplianceQueryEvaluationDto::getIndexPatternId))
                .entrySet().stream()
                .map(entry -> {
                    var first = entry.getValue().get(0);
                    var dto = new IndexPatternQueriesGroupDto();
                    dto.setIndexPatternId(entry.getKey());
                    dto.setIndexPatternName(first.getIndexPatternName());
                    dto.setQueries(entry.getValue());
                    return dto;
                })
                .toList();
    }

    private UtmComplianceControlEvaluationsGroupedDto buildGroupedDto(
            UtmComplianceControlEvaluationsDto evaluation,
            List<IndexPatternQueriesGroupDto> groupedEvaluations
    ) {
        var dto = new UtmComplianceControlEvaluationsGroupedDto();
        dto.setControlId(evaluation.getControlId());
        dto.setControlName(evaluation.getControlName());
        dto.setStatus(evaluation.getStatus());
        dto.setTimestamp(evaluation.getTimestamp());
        dto.setQueryEvaluations(groupedEvaluations);

        return dto;
    }
}