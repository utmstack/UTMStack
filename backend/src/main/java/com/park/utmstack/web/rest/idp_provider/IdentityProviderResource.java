package com.park.utmstack.web.rest.idp_provider;


import com.park.utmstack.service.dto.idp_provider.dto.IdentityProviderConfigResponseDto;
import com.park.utmstack.service.idp_provider.IdentityProviderService;
import io.swagger.v3.oas.annotations.Hidden;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/utm-providers")
@RequiredArgsConstructor
@Hidden
public class IdentityProviderResource {

    private final IdentityProviderService service;


    @GetMapping
    public ResponseEntity<Page<IdentityProviderConfigResponseDto>> getAll(Pageable pageable) {

        Page<IdentityProviderConfigResponseDto> result = service.findAll(pageable);
        return ResponseEntity.ok(result);
    }


}
