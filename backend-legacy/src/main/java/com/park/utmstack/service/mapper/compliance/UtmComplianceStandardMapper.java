package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceStandard;
import com.park.utmstack.service.dto.compliance.UtmComplianceStandardDto;
import org.mapstruct.Mapper;

@Mapper(componentModel = "spring")
public interface UtmComplianceStandardMapper {
    UtmComplianceStandardDto toDto(UtmComplianceStandard entity);
}