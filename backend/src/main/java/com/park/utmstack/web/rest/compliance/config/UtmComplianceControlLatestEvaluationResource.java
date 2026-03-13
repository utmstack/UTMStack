package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlEvaluationLatestService;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlLatestEvaluationDto;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/compliance/control-config")
public class UtmComplianceControlLatestEvaluationResource {

    private final UtmComplianceControlEvaluationLatestService latestEvaluationService;

    public UtmComplianceControlLatestEvaluationResource(UtmComplianceControlEvaluationLatestService latestEvaluationService) {
        this.latestEvaluationService = latestEvaluationService;
    }

    @GetMapping("/get-by-section")
    public ResponseEntity<List<UtmComplianceControlLatestEvaluationDto>> getControlsLatestEvaluationBySection(
            @RequestParam Long sectionId) {

        var controls = latestEvaluationService.getControlsWithLastEvaluation(sectionId);
        return ResponseEntity.ok(controls);
    }
}

