package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceQueryConfig;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryConfigRequestDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryConfigResponseDto;
import org.mapstruct.Mapper;
import org.mapstruct.MappingTarget;

@Mapper(componentModel = "spring")
public interface UtmComplianceQueryConfigMapper {
    UtmComplianceQueryConfig toEntity(UtmComplianceQueryConfigRequestDto dto);

    UtmComplianceQueryConfigResponseDto toResponse(UtmComplianceQueryConfig entity);

    void updateEntity(
            @MappingTarget UtmComplianceQueryConfig entity,
            UtmComplianceQueryConfigRequestDto dto
    );
}
