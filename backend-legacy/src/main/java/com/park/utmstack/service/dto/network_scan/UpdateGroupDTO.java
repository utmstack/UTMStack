package com.park.utmstack.service.dto.network_scan;

import lombok.Data;

import javax.validation.constraints.NotEmpty;
import javax.validation.constraints.NotNull;
import java.util.List;

@Data
public class UpdateGroupDTO {

    @NotEmpty(message = "assetsIds cannot be empty")
    private List<Long> assetsIds;

    @NotNull(message = "assetGroupId is required")
    private Long assetGroupId;
}

