package com.park.utmstack.domain.application_events.types;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ApplicationEvent {

    @JsonProperty("@timestamp")
    private String timestamp;

    private String source;
    private String message;
    private String type;
}
