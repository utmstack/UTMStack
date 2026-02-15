package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlEvaluationService;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationDto;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/compliance/control-config")
public class UtmComplianceControlEvaluationResource {

    private final UtmComplianceControlEvaluationService evaluationService;

    public UtmComplianceControlEvaluationResource(UtmComplianceControlEvaluationService evaluationService) {
        this.evaluationService = evaluationService;
    }

    @GetMapping("/{id}/evaluations")
    public ResponseEntity<List<UtmComplianceControlEvaluationDto>> getControlEvaluations(@PathVariable Long id) {
        var evaluations = evaluationService.findByControlId(id);
        return ResponseEntity.ok(evaluations);
    }
}

