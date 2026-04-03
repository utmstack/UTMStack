package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import com.park.utmstack.service.dto.compliance.UtmComplianceStandardSectionDto;
import org.mapstruct.Mapper;

@Mapper(componentModel = "spring", uses = {UtmComplianceStandardMapper.class})
public interface UtmComplianceStandardSectionMapper {
    UtmComplianceStandardSectionDto toDto(UtmComplianceStandardSection entity);
}
