package com.park.utmstack.service.dto.application_modules;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Component
public class UtmModuleMapper {

    private final Logger logger = LoggerFactory.getLogger(UtmModuleMapper.class);

    public ModuleDTO toDto(UtmModule entity, boolean includeDataType) {
        ModuleDTO dto = new ModuleDTO();
        dto.setId(entity.getId());
        dto.setServerId(entity.getServerId());
        dto.setPrettyName(entity.getPrettyName());
        dto.setModuleName(entity.getModuleName());
        dto.setModuleDescription(entity.getModuleDescription());
        dto.setModuleActive(entity.getModuleActive());
        dto.setModuleIcon(entity.getModuleIcon());
        dto.setModuleCategory(entity.getModuleCategory());
        dto.setLiteVersion(entity.getLiteVersion());
        dto.setNeedsRestart(entity.getNeedsRestart());
        dto.setIsActivatable(entity.getActivatable());
        dto.setModuleGroups(entity.getModuleGroups());

        if (includeDataType) {
            entity.getFilters().stream().findFirst().ifPresent(filter ->
                    dto.setDataType(filter.getDatatype()));
        }

        return dto;
    }

    public List<ModuleDTO> toListDTO(List<UtmModule> modules) {
        return modules.stream()
                .map((m) -> this.toDto(m, false))
                .collect(Collectors.toList());
    }
}
