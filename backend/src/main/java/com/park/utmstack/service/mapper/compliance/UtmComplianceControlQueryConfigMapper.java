package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceControlQueryConfig;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlQueryConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlQueryConfigResponseDto;
import org.mapstruct.Mapper;
import org.mapstruct.MappingTarget;

@Mapper(componentModel = "spring")
public interface UtmComplianceControlQueryConfigMapper {
    UtmComplianceControlQueryConfig toEntity(UtmComplianceControlQueryConfigRequestDto dto);

    UtmComplianceControlQueryConfigResponseDto toResponse(UtmComplianceControlQueryConfig entity);

    void updateEntity(@MappingTarget UtmComplianceControlQueryConfig entity, UtmComplianceControlQueryConfigRequestDto dto);
}
