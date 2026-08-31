package com.park.utmstack.web.rest.idp_provider;


import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderConfigResponseDto;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderCriteria;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderPublicDto;
import com.park.utmstack.service.idp_provider.IdentityProviderService;
import com.park.utmstack.web.rest.util.PaginationUtil;
import io.swagger.v3.oas.annotations.Hidden;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/utm-providers")
@RequiredArgsConstructor
@Hidden
public class IdentityProviderResource {

    private final IdentityProviderService service;


    /**
     * Anonymous listing used by the login page to render the SSO buttons.
     *
     * <p>Returns {@link IdentityProviderPublicDto} only: exposing the full
     * configuration (metadata URL, SP certificate, timestamps) to
     * unauthenticated callers is an information leak. Administrative reads and
     * writes live on /api/identity-providers, which requires ROLE_ADMIN.
     */
    @GetMapping
    public ResponseEntity<List<IdentityProviderPublicDto>> getAll(IdentityProviderCriteria criteria, Pageable pageable) {

        Page<IdentityProviderConfigResponseDto> page = service.findAll(criteria, pageable);

        List<IdentityProviderPublicDto> body = page.getContent().stream()
                .map(provider -> new IdentityProviderPublicDto(
                        provider.getId(),
                        provider.getName(),
                        provider.getProviderType(),
                        provider.getActive()))
                .collect(Collectors.toList());

        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-providers");
        return ResponseEntity.ok().headers(headers).body(body);
    }


}
