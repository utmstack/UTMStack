package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlEvaluationsService;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationsDto;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/compliance/control-config")
public class UtmComplianceControlEvaluationsResource {

    private final UtmComplianceControlEvaluationsService evaluationsService;

    public UtmComplianceControlEvaluationsResource(UtmComplianceControlEvaluationsService evaluationService) {
        this.evaluationsService = evaluationService;
    }

    @GetMapping("/{id}/evaluations")
    public ResponseEntity<List<UtmComplianceControlEvaluationsDto>> getControlEvaluations(@PathVariable Long id) {
        var evaluations = evaluationsService.findByControlId(id);
        return ResponseEntity.ok(evaluations);
    }
}

