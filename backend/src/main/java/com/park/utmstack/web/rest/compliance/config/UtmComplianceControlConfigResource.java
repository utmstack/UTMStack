package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigResponseDto;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlConfigMapper;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/compliance/control")
public class UtmComplianceControlConfigResource {

    private final UtmComplianceControlConfigService service;
    private final UtmComplianceControlConfigMapper mapper;

    public UtmComplianceControlConfigResource(
            UtmComplianceControlConfigService service,
            UtmComplianceControlConfigMapper mapper
    ) {
        this.service = service;
        this.mapper = mapper;
    }

    @PostMapping
    public ResponseEntity<UtmComplianceControlConfigResponseDto> create(@RequestBody UtmComplianceControlConfigRequestDto dto) {
        var entity = mapper.toEntity(dto);
        var saved = service.create(entity);
        return ResponseEntity.ok(mapper.toResponse(saved));
    }

    @GetMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigResponseDto> getById(@PathVariable Long id) {
        var entity = service.findById(id);
        if (entity == null) return ResponseEntity.notFound().build();
        return ResponseEntity.ok(mapper.toResponse(entity));
    }

    @PutMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigResponseDto> update(
            @PathVariable Long id,
            @RequestBody UtmComplianceControlConfigRequestDto dto
    ) {
        var existing = service.findById(id);
        if (existing == null) return ResponseEntity.notFound().build();

        mapper.updateEntity(existing, dto);
        var updated = service.update(id, existing);

        return ResponseEntity.ok(mapper.toResponse(updated));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Long id) {
        service.delete(id);
        return ResponseEntity.noContent().build();
    }
}
