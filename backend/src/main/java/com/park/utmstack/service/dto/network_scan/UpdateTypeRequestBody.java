package com.park.utmstack.service.dto.network_scan;

import lombok.Data;
import lombok.Getter;
import lombok.Setter;

import javax.validation.constraints.NotEmpty;
import java.util.List;

@Data
@Getter
@Setter
public class UpdateTypeRequestBody {
    @NotEmpty
    private List<Long> assetsIds;

    private Long assetTypeId;
}