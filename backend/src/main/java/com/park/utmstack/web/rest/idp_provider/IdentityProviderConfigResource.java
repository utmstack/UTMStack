package com.park.utmstack.web.rest.idp_provider;


import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderConfigRequestDto;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderConfigResponseDto;
import com.park.utmstack.service.idp_provider.IdentityProviderService;
import com.park.utmstack.web.rest.util.PaginationUtil;
import io.swagger.v3.oas.annotations.Hidden;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.net.URI;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/idp-configs")
@RequiredArgsConstructor
@Hidden
public class IdentityProviderConfigResource {

    private final IdentityProviderService service;

    @PostMapping
    public ResponseEntity<IdentityProviderConfigResponseDto> create(@RequestBody @Valid IdentityProviderConfigRequestDto dto) {
        IdentityProviderConfigResponseDto result = service.create(dto);
        return ResponseEntity
                .created(URI.create("/api/idp-configs/" + result.getId()))
                .body(result);
    }

    @PutMapping("/{id}")
    public ResponseEntity<IdentityProviderConfigResponseDto> update(@PathVariable Long id,
                                                                    @RequestBody IdentityProviderConfigRequestDto dto) {
        IdentityProviderConfigResponseDto result = service.update(id, dto);
        return ResponseEntity.ok(result);
    }


    @GetMapping
    public ResponseEntity<List<IdentityProviderConfigResponseDto>> getAll(Pageable pageable) {

        Page<IdentityProviderConfigResponseDto> page = service.findAll(pageable);

        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-providers");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    @GetMapping("/{id}")
    public ResponseEntity<IdentityProviderConfigResponseDto> getById(@PathVariable Long id) {
        Optional<IdentityProviderConfigResponseDto> dtoOpt = service.findById(id);
        return dtoOpt.map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.notFound().build());
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Long id) {
        service.delete(id);
        return ResponseEntity.noContent().build();
    }
}
