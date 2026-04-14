package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceQueryConfig;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryConfigDto;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.MappingTarget;

@Mapper(componentModel = "spring")
public interface UtmComplianceQueryConfigMapper {
    @Mapping(target = "controlConfig", ignore = true)
    UtmComplianceQueryConfig toEntity(UtmComplianceQueryConfigDto dto);

    @Mapping(target = "controlConfigId", source = "controlConfig.id")
    UtmComplianceQueryConfigDto toDto(UtmComplianceQueryConfig entity);

    @Mapping(target = "id", ignore = true)
    @Mapping(target = "controlConfig", ignore = true)
    void updateEntity(@MappingTarget UtmComplianceQueryConfig entity, UtmComplianceQueryConfigDto dto);
}