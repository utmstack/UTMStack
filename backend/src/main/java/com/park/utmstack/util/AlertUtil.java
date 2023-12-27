package com.park.utmstack.util;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
public class AlertUtil {
    private static final String CLASSNAME = "AlertUtil";

    private final ElasticsearchService elasticsearchService;

    public AlertUtil(ElasticsearchService elasticsearchService) {
        this.elasticsearchService = elasticsearchService;
    }

    public Long countAlertsByStatus(int status) {
        final String ctx = CLASSNAME + ".countAlertsByStatus";
        final String AGG_NAME = "count_open_alerts";
        try {
            if (!elasticsearchService.indexExist(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS)))
                return 0L;

            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertStatus, OperatorType.IS, status));

            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.query(SearchUtil.toQuery(filters))
                .index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS))
                .aggregations(AGG_NAME, a -> a.valueCount(c -> c.field(Constants.alertStatus)))
                .size(0);

            SearchResponse<Object> response = elasticsearchService.search(srb.build(), Object.class);
            return (long) response.aggregations().get(AGG_NAME).valueCount().value();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
