package com.park.utmstack.web.rest.idp_provider;


import com.park.utmstack.domain.UtmDataInputStatus;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderConfigResponseDto;
import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderCriteria;
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

@RestController
@RequestMapping("/api/utm-providers")
@RequiredArgsConstructor
@Hidden
public class IdentityProviderResource {

    private final IdentityProviderService service;


    @GetMapping
    public ResponseEntity<List<IdentityProviderConfigResponseDto>> getAll(IdentityProviderCriteria criteria, Pageable pageable) {

        Page<IdentityProviderConfigResponseDto> page = service.findAll(criteria, pageable);

        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-providers");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }


}
