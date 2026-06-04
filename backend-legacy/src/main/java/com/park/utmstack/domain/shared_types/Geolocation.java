package com.park.utmstack.domain.shared_types;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonIgnoreProperties(ignoreUnknown = true)
public class Geolocation {

    @JsonProperty("country")
    private String country;

    @JsonProperty("city")
    private String city;

    @JsonProperty("latitude")
    private Double latitude;

    @JsonProperty("longitude")
    private Double longitude;

    @JsonProperty("asn")
    private Long asn;

    @JsonProperty("aso")
    private String aso;

    @JsonProperty("countryCode")
    private String countryCode;

    @JsonProperty("accuracy")
    private Integer accuracy;
}

