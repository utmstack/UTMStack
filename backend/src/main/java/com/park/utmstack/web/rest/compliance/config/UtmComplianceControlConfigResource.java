package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.compliance.config.UtmComplianceControlConfigService;
import com.park.utmstack.service.compliance.config.UtmComplianceQueryConfigService;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigResponseDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryConfigResponseDto;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlConfigMapper;
import com.park.utmstack.service.mapper.compliance.UtmComplianceQueryConfigMapper;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/compliance/control")
public class UtmComplianceControlConfigResource {

    private final UtmComplianceControlConfigService controlService;
    private final UtmComplianceQueryConfigService queryService;
    private final UtmComplianceControlConfigMapper controlMapper;
    private final UtmComplianceQueryConfigMapper queryMapper;

    public UtmComplianceControlConfigResource(
            UtmComplianceControlConfigService controlService,
            UtmComplianceQueryConfigService queryService,
            UtmComplianceControlConfigMapper controlMapper,
            UtmComplianceQueryConfigMapper queryMapper
    ) {
        this.controlService = controlService;
        this.queryService = queryService;
        this.controlMapper = controlMapper;
        this.queryMapper = queryMapper;
    }

    @PostMapping
    public ResponseEntity<UtmComplianceControlConfigResponseDto> createControl(
            @RequestBody UtmComplianceControlConfigRequestDto dto
    ) {
        var entity = controlMapper.toEntity(dto);
        var saved = controlService.create(entity);
        return ResponseEntity.ok(controlMapper.toResponse(saved));
    }

    @GetMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigResponseDto> getControl(@PathVariable Long id) {
        var entity = controlService.findById(id);
        if (entity == null) return ResponseEntity.notFound().build();
        return ResponseEntity.ok(controlMapper.toResponse(entity));
    }

    @PutMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigResponseDto> updateControl(
            @PathVariable Long id,
            @RequestBody UtmComplianceControlConfigRequestDto dto
    ) {
        var existing = controlService.findById(id);
        if (existing == null) return ResponseEntity.notFound().build();

        controlMapper.updateEntity(existing, dto);
        var updated = controlService.update(id, existing);

        return ResponseEntity.ok(controlMapper.toResponse(updated));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteControl(@PathVariable Long id) {
        controlService.delete(id);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/{controlId}/query")
    public ResponseEntity<UtmComplianceQueryConfigResponseDto> addQuery(
            @PathVariable Long controlId,
            @RequestBody UtmComplianceQueryConfigRequestDto dto
    ) {
        dto.setControlConfigId(controlId);

        var entity = queryMapper.toEntity(dto);
        var saved = queryService.create(entity);

        return ResponseEntity.ok(queryMapper.toResponse(saved));
    }

    @DeleteMapping("/query/{queryId}")
    public ResponseEntity<Void> deleteQuery(@PathVariable Long queryId) {
        queryService.delete(queryId);
        return ResponseEntity.noContent().build();
    }
}

