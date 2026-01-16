package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlConfigService;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigResponseDto;
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
    public ResponseEntity<UtmComplianceControlConfigResponseDto> createControl(
            @RequestBody UtmComplianceControlConfigRequestDto dto
    ) {
        var created = controlService.create(dto);
        return ResponseEntity.ok(created);
    }

    @GetMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigResponseDto> getControl(@PathVariable Long id) {
        var entity = controlService.findById(id);
        if (entity == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(controlMapper.toResponse(entity));
    }

    @PutMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigResponseDto> updateControl(
            @PathVariable Long id,
            @RequestBody UtmComplianceControlConfigRequestDto dto
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
