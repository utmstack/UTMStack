package com.park.utmstack.service.dto.network_scan;

import lombok.Data;

import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;
import java.util.List;

@Data
public class UpdateGroupDTO {

    @NotEmpty(message = "assetsIds cannot be empty")
    private List<Long> assetsIds;

    @NotNull(message = "assetGroupId is required")
    private Long assetGroupId;
}

