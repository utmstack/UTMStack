package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlConfigService;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlConfigMapper;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/compliance/control")
public class UtmComplianceControlConfigResource {

    private final UtmComplianceControlConfigService controlService;
    private final UtmComplianceControlConfigMapper controlMapper;

    public UtmComplianceControlConfigResource(
            UtmComplianceControlConfigService controlService,
            UtmComplianceControlConfigMapper controlMapper
    ) {
        this.controlService = controlService;
        this.controlMapper = controlMapper;
    }

    @PostMapping
    public ResponseEntity<UtmComplianceControlConfigDto> createControl(
            @RequestBody UtmComplianceControlConfigDto dto
    ) {
        var created = controlService.create(dto);
        return ResponseEntity.ok(created);
    }

    @GetMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigDto> getControl(@PathVariable Long id) {
        var entity = controlService.findById(id);
        if (entity == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(controlMapper.toDto(entity));
    }

    @PutMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigDto> updateControl(
            @PathVariable Long id,
            @RequestBody UtmComplianceControlConfigDto dto
    ) {
        var updated = controlService.update(id, dto);
        return ResponseEntity.ok(updated);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteControl(@PathVariable Long id) {
        controlService.delete(id);
        return ResponseEntity.noContent().build();
    }
}
