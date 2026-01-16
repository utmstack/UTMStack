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
        var entity = controlMapper.toEntity(dto);

        if (entity.getQueriesConfigs() != null) {
            entity.getQueriesConfigs().forEach(q -> {
                q.setControlConfigId(entity.getId());
            });
        }

        var saved = controlService.create(entity);

        return ResponseEntity.ok(controlMapper.toResponse(saved));
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
        var existing = controlService.findById(id);
        if (existing == null) {
            return ResponseEntity.notFound().build();
        }

        controlMapper.updateEntity(existing, dto);

        if (existing.getQueriesConfigs() != null) {
            existing.getQueriesConfigs().forEach(q -> {
                q.setControlConfigId(id);
            });
        }

        var updated = controlService.update(id, existing);
        return ResponseEntity.ok(controlMapper.toResponse(updated));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteControl(@PathVariable Long id) {
        controlService.delete(id);
        return ResponseEntity.noContent().build();
    }
}
