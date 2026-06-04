package com.park.utmstack.service.collectors;

import lombok.Builder;
import lombok.Data;

import java.util.List;

@Data
@Builder
public class BulkCollectorConfigResponseDTO {
    private List<CollectorConfigResultDTO> results;
}

