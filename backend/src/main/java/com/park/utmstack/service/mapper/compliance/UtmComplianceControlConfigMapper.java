package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;
import org.mapstruct.MappingTarget;

@Mapper(componentModel = "spring", uses = {
        UtmComplianceQueryConfigMapper.class,
        UtmComplianceStandardSectionMapper.class,
        UtmComplianceStandardMapper.class})
public interface UtmComplianceControlConfigMapper {

    @Mapping(target = "queriesConfigs", ignore = true)
    UtmComplianceControlConfig toEntity(UtmComplianceControlConfigDto dto);

    UtmComplianceControlConfigDto toDto(UtmComplianceControlConfig entity);

    @Mapping(target = "id", ignore = true)
    @Mapping(target = "queriesConfigs", ignore = true)
    void updateEntity(@MappingTarget UtmComplianceControlConfig entity, UtmComplianceControlConfigDto dto);
}