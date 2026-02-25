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

    @GetMapping("/get-by-section")
    public ResponseEntity<List<UtmComplianceControlEvaluationDto>> getControlsBySection(
            @RequestParam Long sectionId) {

        var controls = evaluationService.getControlsWithLastEvaluation(sectionId);
        return ResponseEntity.ok(controls);
    }
}

