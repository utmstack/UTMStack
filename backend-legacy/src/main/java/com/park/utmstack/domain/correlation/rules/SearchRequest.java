package com.park.utmstack.domain.correlation.rules;

import lombok.Data;

import java.util.List;

@Data
public class SearchRequest {
    private String indexPattern;
    private List<Expression> with;
    private List<SearchRequest> or;
    private String within;
    private Integer count;
}
