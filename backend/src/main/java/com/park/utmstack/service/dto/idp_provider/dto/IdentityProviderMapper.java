package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.util.CipherUtil;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.MappingTarget;

import java.util.List;

@Mapper(componentModel = "spring", imports = {CipherUtil.class, System.class, Constants.class})
public interface IdentityProviderMapper {

    IdentityProviderConfigResponseDto toDto(IdentityProviderConfig entity);

    @Mapping(target = "clientSecret",
            expression = "java(CipherUtil.encrypt(request.getClientSecret(), System.getenv(Constants.ENV_ENCRYPTION_KEY)))")
    IdentityProviderConfig toEntity(IdentityProviderConfigRequestDto request);

    List<IdentityProviderConfigResponseDto> toDtoList(List<IdentityProviderConfig> entities);

    void updateEntityFromRequest(IdentityProviderConfigRequestDto request, @MappingTarget IdentityProviderConfig entity);
}
