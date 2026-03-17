package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlLatestEvaluationDto;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlConfigMapper;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlLatestEvaluationMapper;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;

@Service
public class UtmComplianceControlEvaluationLatestService {

    private final UtmComplianceControlConfigService configService;
    private final ElasticsearchService elasticsearchService;
    private final UtmComplianceControlConfigMapper controlMapper;

    public UtmComplianceControlEvaluationLatestService(UtmComplianceControlConfigService configService,
                                                       ElasticsearchService elasticsearchService,
                                                       UtmComplianceControlConfigMapper controlMapper) {
        this.configService = configService;
        this.elasticsearchService = elasticsearchService;
        this.controlMapper = controlMapper;
    }

    public Page<UtmComplianceControlLatestEvaluationDto> getControlsWithLastEvaluation(
            Long sectionId, Pageable pageable) {

        Page<UtmComplianceControlConfig> controls = configService.findBySection(sectionId, pageable);

        return controls.map(control -> {
            UtmComplianceControlConfigDto controlDto = controlMapper.toDto(control);
            var lastEval = elasticsearchService.getLatestControlEvaluation(control.getId());
            return UtmComplianceControlLatestEvaluationMapper.toDto(controlDto, lastEval);
        });
    }
}


