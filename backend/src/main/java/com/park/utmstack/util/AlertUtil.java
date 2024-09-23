package com.park.utmstack.util;

import ch.qos.logback.core.util.FixedDelay;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.http.ResponseEntity;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.*;
import java.util.stream.Collectors;

@Component
public class AlertUtil {
    private static final String CLASSNAME = "AlertUtil";

    private final ElasticsearchService elasticsearchService;

    public AlertUtil(ElasticsearchService elasticsearchService) {
        this.elasticsearchService = elasticsearchService;
    }

    public Long countAlertsByStatus(int status) {

        List<FilterType> filters = new ArrayList<>();
        filters.add(new FilterType(Constants.alertStatus, OperatorType.IS, status));

        return this.search(filters);
    }

    public Long countAlertsByStatus(int status, LocalDateTime startDateTime, LocalDateTime endDateTime) {

        List<FilterType> filters = new ArrayList<>();
        filters.add(new FilterType(Constants.alertStatus, OperatorType.IS, status));
        filters.add(new FilterType(Constants.timestamp, OperatorType.IS_BETWEEN, List.of(startDateTime.toString(), endDateTime.toString())));

        return this.search(filters);
    }

    public long search(List<FilterType> filters) {
        final String AGG_NAME = "count_open_alerts";
        final String ctx = CLASSNAME + ".countAlertsByStatus";

        try {
            if (!elasticsearchService.indexExist(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS)))
                return 0L;

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
