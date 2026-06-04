package com.park.utmstack.domain.correlation.rules;

import com.park.utmstack.service.dto.correlation.AdversaryType;
import lombok.Getter;
import lombok.Setter;

import java.time.Instant;
import java.util.List;


@Setter
@Getter
public class UtmCorrelationRulesFilter {

    private String name;
    private List<Integer> confidentiality;
    private List<Integer> integrity;
    private List<Integer> availability;
    private List<String> category;
    private List<String> technique;
    private Instant initDate;
    private Instant endDate;
    private List<Boolean> active;
    private List<Boolean> systemOwner;
    private List<String> dataTypes;
    private List<AdversaryType> adversary;

    private String search;

}
