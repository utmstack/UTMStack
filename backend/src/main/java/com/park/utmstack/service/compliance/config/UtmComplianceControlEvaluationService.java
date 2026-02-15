package com.park.utmstack.service.compliance.config;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationDto;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UtmComplianceControlEvaluationService {

    private final ElasticsearchService elasticsearchService;

    public UtmComplianceControlEvaluationService(ElasticsearchService elasticsearchService) {
        this.elasticsearchService = elasticsearchService;
    }

    public List<UtmComplianceControlEvaluationDto> findByControlId(Long controlId) {
        return elasticsearchService.getControlEvaluations(controlId);
    }
}


