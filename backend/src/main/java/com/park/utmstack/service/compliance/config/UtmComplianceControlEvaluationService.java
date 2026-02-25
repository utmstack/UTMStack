package com.park.utmstack.service.compliance.config;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationDto;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlEvaluationMapper;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UtmComplianceControlEvaluationService {

    //private final ElasticsearchService elasticsearchService;
    private final UtmComplianceControlConfigService configService;
    private final UtmComplianceControlEvaluationsService evaluationsService;

    public UtmComplianceControlEvaluationService(ElasticsearchService elasticsearchService,
                                                 UtmComplianceControlConfigService configService,
                                                 UtmComplianceControlEvaluationsService evaluationsService) {
        //this.elasticsearchService = elasticsearchService;
        this.configService = configService;
        this.evaluationsService = evaluationsService;
    }

     public List<UtmComplianceControlEvaluationDto> getControlsWithLastEvaluation(Long sectionId) {

            List<UtmComplianceControlConfigDto> controls =
                    configService.getControlsBySection(sectionId);

            return controls.stream()
                    .map(control -> {
                        var lastEval = evaluationsService.getLastEvaluationForControl(control.getId());
                        return UtmComplianceControlEvaluationMapper.toDto(control, lastEval);
                    })
                    .toList();
        }
    }


