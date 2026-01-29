package com.park.utmstack.service.dto.visualization.mapper;

import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.service.dto.visualization.UtmVisualizationDto;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.mapstruct.Mapper;
import org.mapstruct.factory.Mappers;

@Mapper(componentModel = "spring")
public interface UtmVisualizationMapper {
    UtmVisualizationMapper INSTANCE = Mappers.getMapper(UtmVisualizationMapper.class);

    UtmVisualizationDto toDto(UtmVisualization entity) throws UtmSerializationException;

    UtmVisualization toEntity(UtmVisualizationDto dto) throws UtmSerializationException;
}
