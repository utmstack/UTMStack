package com.park.utmstack.service.compliance.config;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationsDto;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UtmComplianceControlEvaluationsService {

    private final ElasticsearchService elasticsearchService;

    public UtmComplianceControlEvaluationsService(ElasticsearchService elasticsearchService) {
        this.elasticsearchService = elasticsearchService;
    }

    public List<UtmComplianceControlEvaluationsDto> findByControlId(Long controlId) {
        return elasticsearchService.getControlEvaluations(controlId);
    }

    public UtmComplianceControlEvaluationsDto getLastEvaluationForControl(Long controlId) {
        return elasticsearchService.getLastEvaluation(controlId);
    }
}