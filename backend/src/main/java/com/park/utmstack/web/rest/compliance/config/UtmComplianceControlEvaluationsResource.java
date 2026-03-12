package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlEvaluationsService;
import com.park.utmstack.service.dto.compliance.ControlEvaluationsResponseDto;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/compliance/control-config")
public class UtmComplianceControlEvaluationsResource {

    private final UtmComplianceControlEvaluationsService evaluationsService;

    public UtmComplianceControlEvaluationsResource(UtmComplianceControlEvaluationsService evaluationService) {
        this.evaluationsService = evaluationService;
    }

    @GetMapping("/{id}/evaluations")
    public ResponseEntity<ControlEvaluationsResponseDto> getControlEvaluations(@PathVariable Long id) {
        return ResponseEntity.ok(evaluationsService.getEvaluationsWithRange(id));
    }

}

