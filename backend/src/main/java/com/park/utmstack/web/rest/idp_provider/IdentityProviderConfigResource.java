package com.park.utmstack.web.rest.idp_provider;


import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderConfigRequestDto;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderConfigResponseDto;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderCreateConfigDto;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderCriteria;
import com.park.utmstack.service.idp_provider.IdentityProviderService;
import com.park.utmstack.web.rest.util.PaginationUtil;
import io.swagger.v3.oas.annotations.Hidden;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import javax.validation.Valid;
import java.io.IOException;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/identity-providers")
@RequiredArgsConstructor
@Hidden
public class IdentityProviderConfigResource {

    private final IdentityProviderService service;

    @PostMapping(consumes = MediaType.MULTIPART_FORM_DATA_VALUE)
    public ResponseEntity<IdentityProviderConfigResponseDto> create(
            @RequestParam String name,
            @RequestParam String providerType,
            @RequestParam String metadataUrl,
            @RequestParam Boolean active,
            @RequestPart("spPrivateKeyFile") MultipartFile privateKeyFile,
            @RequestPart("spCertificateFile") MultipartFile certificateFile
    ) throws IOException {
        String privateKeyPem = new String(privateKeyFile.getBytes(), StandardCharsets.UTF_8);
        String certPem = new String(certificateFile.getBytes(), StandardCharsets.UTF_8);

        IdentityProviderCreateConfigDto dto = new IdentityProviderCreateConfigDto();
        dto.setName(name);
        dto.setProviderType(ProviderType.valueOf(providerType));
        dto.setMetadataUrl(metadataUrl);
        dto.setActive(active);
        dto.setSpPrivateKeyPem(privateKeyPem);
        dto.setSpCertificatePem(certPem);

        IdentityProviderConfigResponseDto result = service.create(dto);
        return ResponseEntity
                .created(URI.create("/api/identity-providers/" + result.getId()))
                .body(result);
    }


    @PutMapping("/{id}")
    public ResponseEntity<IdentityProviderConfigResponseDto> update(@PathVariable Long id,
                                                                    @RequestBody @Valid IdentityProviderConfigRequestDto dto) {
        IdentityProviderConfigResponseDto result = service.update(id, dto);
        return ResponseEntity.ok(result);
    }


    @GetMapping
    public ResponseEntity<List<IdentityProviderConfigResponseDto>> getAll(IdentityProviderCriteria criteria, Pageable pageable) {

        Page<IdentityProviderConfigResponseDto> page = service.findAll(criteria, pageable);

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
