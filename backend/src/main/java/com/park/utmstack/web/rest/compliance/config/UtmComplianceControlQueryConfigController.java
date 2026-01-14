package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlConfigService;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlQueryConfigMapper;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;


@RestController
@RequestMapping("/api/compliance/query")
public class UtmComplianceControlQueryConfigController {

    private final UtmComplianceControlConfigService service;
    private final UtmComplianceControlQueryConfigMapper mapper;

    public UtmComplianceControlQueryConfigController(
            UtmComplianceControlConfigService service,
            UtmComplianceControlQueryConfigMapper mapper
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
