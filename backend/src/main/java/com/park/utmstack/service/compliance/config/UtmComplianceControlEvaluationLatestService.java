package com.park.utmstack.service.compliance.config;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlLatestEvaluationDto;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlLatestEvaluationMapper;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UtmComplianceControlEvaluationLatestService {

    private final UtmComplianceControlConfigService configService;
    private final ElasticsearchService elasticsearchService;

    public UtmComplianceControlEvaluationLatestService(UtmComplianceControlConfigService configService,
                                                       ElasticsearchService elasticsearchService) {
        this.configService = configService;
        this.elasticsearchService = elasticsearchService;
    }

     public List<UtmComplianceControlLatestEvaluationDto> getControlsWithLastEvaluation(Long sectionId) {

            List<UtmComplianceControlConfigDto> controls = configService.getControlsBySection(sectionId);

            return controls.stream()
                    .map(control -> {
                        var lastEval = this.elasticsearchService.getLatestControlEvaluation(control.getId());
                        return UtmComplianceControlLatestEvaluationMapper.toDto(control, lastEval);
                    })
                    .toList();
    }
}


