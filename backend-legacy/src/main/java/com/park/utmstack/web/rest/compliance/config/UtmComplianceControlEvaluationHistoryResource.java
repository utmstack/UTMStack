package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlEvaluationHistoryService;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationHistoryResponseDto;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/compliance/control-config")
public class UtmComplianceControlEvaluationHistoryResource {

    private final UtmComplianceControlEvaluationHistoryService evaluationHistoryService;

    public UtmComplianceControlEvaluationHistoryResource(UtmComplianceControlEvaluationHistoryService evaluationHistoryService) {
        this.evaluationHistoryService = evaluationHistoryService;
    }

    @GetMapping("/{id}/evaluations")
    public ResponseEntity<UtmComplianceControlEvaluationHistoryResponseDto> getControlEvaluationHistory(@PathVariable Long id) {
        return ResponseEntity.ok(evaluationHistoryService.getEvaluationsWithRange(id));
    }

}

