package com.park.utmstack.web.rest.idp_provider;


import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import com.park.utmstack.service.dto.idp_provider.dto.*;
import com.park.utmstack.service.idp_provider.IdentityProviderService;
import com.park.utmstack.util.saml.PemUtils;
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
    private final IdentityProviderMapper mapper;

    @PostMapping(consumes = MediaType.MULTIPART_FORM_DATA_VALUE)
    @AuditEvent(
        attemptType = ApplicationEventType.IDP_CONFIG_CREATE_ATTEMPT,
        attemptMessage = "Attempting to create Identity Provider configuration: {name}",
        successType = ApplicationEventType.IDP_CONFIG_CREATE_SUCCESS,
        successMessage = "Identity Provider configuration {name} created successfully"
    )
    public ResponseEntity<IdentityProviderConfigResponseDto> create(@RequestParam String name,
                                                                    @RequestParam String providerType,
                                                                    @RequestParam String metadataUrl,
                                                                    @RequestParam String spEntityId,
                                                                    @RequestParam String spAcsUrl,
                                                                    @RequestParam Boolean active,
                                                                    @RequestPart("spPrivateKeyFile") MultipartFile privateKeyFile,
                                                                    @RequestPart("spCertificateFile") MultipartFile certificateFile) {


        PemUtils.validateFilesForCreate(privateKeyFile, certificateFile);
        IdentityProviderCreateConfigDto dto = mapper.toCreateConfigDto(name, providerType, metadataUrl, active, privateKeyFile, certificateFile, spEntityId , spAcsUrl);

        IdentityProviderConfigResponseDto result = service.create(dto);
        return ResponseEntity
                .created(URI.create("/api/identity-providers/" + result.getId()))
                .body(result);
    }


    @PutMapping("/{id}")
    @AuditEvent(
        attemptType = ApplicationEventType.IDP_CONFIG_UPDATE_ATTEMPT,
        attemptMessage = "Attempting to update Identity Provider configuration with ID: {id}",
        successType = ApplicationEventType.IDP_CONFIG_UPDATE_SUCCESS,
        successMessage = "Identity Provider configuration with ID {id} updated successfully"
    )
    public ResponseEntity<IdentityProviderConfigResponseDto> update(@PathVariable Long id,
                                                                    @RequestParam String name,
                                                                    @RequestParam String providerType,
                                                                    @RequestParam String metadataUrl,
                                                                    @RequestParam String spEntityId,
                                                                    @RequestParam String spAcsUrl,
                                                                    @RequestParam Boolean active,
                                                                    @RequestPart(value = "spPrivateKeyFile", required = false) MultipartFile privateKeyFile,
                                                                    @RequestPart(value = "spCertificateFile", required = false) MultipartFile certificateFile) {

        PemUtils.validateFilesForUpdate(privateKeyFile, certificateFile);
        IdentityProviderCreateConfigDto dto = mapper.toCreateConfigDto(name, providerType, metadataUrl, active, privateKeyFile, certificateFile, spEntityId , spAcsUrl);

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
    @AuditEvent(
        attemptType = ApplicationEventType.IDP_CONFIG_DELETE_ATTEMPT,
        attemptMessage = "Attempting to delete Identity Provider configuration with ID: {id}",
        successType = ApplicationEventType.IDP_CONFIG_DELETE_SUCCESS,
        successMessage = "Identity Provider configuration with ID {id} deleted successfully"
    )
    public ResponseEntity<Void> delete(@PathVariable Long id) {
        service.delete(id);
        return ResponseEntity.noContent().build();
    }
}
