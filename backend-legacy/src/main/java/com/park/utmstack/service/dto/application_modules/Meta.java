package com.park.utmstack.service.dto.application_modules;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
public class Meta {
    @JsonProperty("trace_id") private String traceId;
}
