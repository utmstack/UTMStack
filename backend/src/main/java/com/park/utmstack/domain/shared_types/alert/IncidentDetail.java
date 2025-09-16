package com.park.utmstack.domain.shared_types.alert;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.Data;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonIgnoreProperties(ignoreUnknown = true)
public class IncidentDetail {
    private String createdBy;
    private String observation;
    private String creationDate;
    private String source;
}