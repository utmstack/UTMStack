package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigResponseDto;
import org.mapstruct.Mapper;
import org.mapstruct.MappingTarget;

@Mapper(componentModel = "spring")
public interface UtmComplianceControlConfigMapper {
    UtmComplianceControlConfig toEntity(UtmComplianceControlConfigRequestDto dto);

    UtmComplianceControlConfigResponseDto toResponse(UtmComplianceControlConfig entity);

    void updateEntity(
            @MappingTarget UtmComplianceControlConfig entity,
            UtmComplianceControlConfigRequestDto dto
    );
}