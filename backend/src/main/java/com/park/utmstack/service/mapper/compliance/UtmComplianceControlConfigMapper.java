package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigResponseDto;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.MappingTarget;

@Mapper(componentModel = "spring", uses = UtmComplianceQueryConfigMapper.class)
public interface UtmComplianceControlConfigMapper {

    @Mapping(source = "standardSectionId", target = "standardSectionId")
    UtmComplianceControlConfig toEntity(UtmComplianceControlConfigRequestDto dto);

    UtmComplianceControlConfigResponseDto toResponse(UtmComplianceControlConfig entity);

    @Mapping(source = "standardSectionId", target = "standardSectionId")
    void updateEntity(
            @MappingTarget UtmComplianceControlConfig entity,
            UtmComplianceControlConfigRequestDto dto
    );
}