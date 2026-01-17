package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceQueryConfig;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryConfigDto;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;

@Mapper(componentModel = "spring")
public interface UtmComplianceQueryConfigMapper {

    UtmComplianceQueryConfig toEntity(UtmComplianceQueryConfigDto dto);

    UtmComplianceQueryConfigDto toDto(UtmComplianceQueryConfig entity);
}