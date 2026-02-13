package com.park.utmstack.service.collectors;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class CollectorConfigResultDTO {
    private int collectorId;
    private boolean success;
    private String errorMessage;
}

