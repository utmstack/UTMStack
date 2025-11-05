package com.park.utmstack.service.mapper;

import com.park.utmstack.domain.api_keys.ApiKey;
import com.park.utmstack.service.dto.api_key.ApiKeyResponseDTO;
import org.mapstruct.Mapper;

import java.util.Arrays;
import java.util.Collections;
import java.util.Optional;
import java.util.stream.Collectors;

@Mapper(componentModel = "spring")
public class ApiKeyMapper {

   public ApiKeyResponseDTO toDto(ApiKey apiKey){
        return ApiKeyResponseDTO.builder()
            .id(apiKey.getId())
            .name(apiKey.getName())
            .createdAt(apiKey.getCreatedAt())
            .expiresAt(apiKey.getExpiresAt())
            .allowedIp(
                Optional.ofNullable(apiKey.getAllowedIp())
                    .map(s -> Arrays.stream(s.split(","))
                        .map(String::trim)
                        .filter(str -> !str.isEmpty())
                        .collect(Collectors.toList()))
                    .orElse(Collections.emptyList())
            )
            .build();
    }
}
