package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.mapper.compliance.UtmComplianceQueryConfigMapper;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;


@RestController
@RequestMapping("/api/compliance/query")
public class UtmComplianceControlQueryConfigController {

    private final UtmComplianceControlConfigService service;
    private final UtmComplianceQueryConfigMapper mapper;

    public UtmComplianceControlQueryConfigController(
            UtmComplianceControlConfigService service,
            UtmComplianceQueryConfigMapper mapper
    ) {
        this.service = service;
        this.mapper = mapper;
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Long id) {
        service.delete(id);
        return ResponseEntity.noContent().build();
    }
}
