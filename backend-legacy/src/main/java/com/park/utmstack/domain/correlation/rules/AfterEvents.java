package com.park.utmstack.domain.correlation.rules;

import lombok.Data;

import java.util.List;
@Data
public class AfterEvents {
    private List<SearchRequest> afterEvents;
}
