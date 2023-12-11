package com.utmstack.userauditor.service.Impl;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import com.utmstack.userauditor.model.*;
import com.utmstack.userauditor.model.winevent.EventLog;
import com.utmstack.userauditor.model.SourceFilter;
import com.utmstack.userauditor.model.SourceScan;
import com.utmstack.userauditor.repository.SourceScanRepository;
import com.utmstack.userauditor.service.elasticsearch.ElasticsearchService;
import com.utmstack.userauditor.service.elasticsearch.SearchUtil;
import com.utmstack.userauditor.service.interfaces.Source;
import com.utmstack.userauditor.service.type.FilterType;
import com.utmstack.userauditor.service.type.OperatorType;
import com.utmstack.userauditor.service.type.SourceType;
import org.jetbrains.annotations.NotNull;
import org.opensearch.client.json.JsonData;
import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch._types.aggregations.TopHitsAggregate;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.opensearch.client.opensearch.core.search.Hit;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.*;

@Component
public class WindowsSource implements Source {

    @Value("${app.elasticsearch.startYear}")
    private int startYear;

    @Value("${app.elasticsearch.searchIntervalMonths}")
    private int searchIntervalMonths;

    Map<String, List<EventLog>> userEvents;
    final ElasticsearchService elasticsearchService;

    final SourceScanRepository sourceScanRepository;

    WindowsSource(ElasticsearchService elasticsearchService, SourceScanRepository sourceScanRepository) {
        this.userEvents = new HashMap<>();
        this.elasticsearchService = elasticsearchService;
        this.sourceScanRepository = sourceScanRepository;
    }

    @Override
    public SourceType getType() {
        return SourceType.WINDOWS;
    }

    @Override
    public Map<String, List<EventLog>> findUsers(UserSource userSource) throws Exception {

        LocalDate currentDate = LocalDate.now();
        Map<String, List<UserAttribute>> users = new HashMap<>();
        LocalDate startDate = LocalDate.of( startYear, 1, 1);
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd");
        final String ctx = "UserService.userlogs";
        List<SourceScan> scans = sourceScanRepository.findBySource_Id(userSource.getId());

        if (!scans.isEmpty()) {
            startDate = scans.stream()
                    .map(SourceScan::getExecutionDate)
                    .max(LocalDate::compareTo)
                    .orElse(startDate);
        }

        try {
            if (!elasticsearchService.indexExist(userSource.getIndexPattern()))
                return this.userEvents;

            if (currentDate.isAfter(startDate)) {
                for (SourceFilter filter : userSource.getFilters()) {
                    this.executeQuery(startDate.format(formatter), startDate.plusMonths(searchIntervalMonths).format(formatter), 10, filter);
                }
                sourceScanRepository.save(SourceScan.builder()
                        .executionDate(startDate.plusMonths(searchIntervalMonths))
                        .source(userSource)
                        .build());
            }

            return this.userEvents;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    void executeQuery(String from, String to, Integer top, SourceFilter sourceFilter) {

        // Pageable pageable = PageRequest.of(3, 10000, Sort.by(Sort.Order.asc("@timestamp")));

        final String AGG_NAME = "recent_events_by_user";

        SearchRequest.Builder srb = getBuilder(sourceFilter, from, to);

        SearchResponse<String> rs = elasticsearchService.search(srb.build(), String.class);

        List<BucketAggregation> buckets = TermAggregateParser.parse(rs.aggregations().get(AGG_NAME));
        buckets.forEach(b -> {
            if (!b.getKey().equals("-")){
                if (!this.userEvents.containsKey(b.getKey())) {
                    this.userEvents.put(b.getKey(), userAttributes(b.getSubAggregations().get("top_events").topHits()));
                } else {
                    this.userEvents.get(b.getKey()).addAll(userAttributes(b.getSubAggregations().get("top_events").topHits()));
                }
            }
        });
    }

    @NotNull
    private static SearchRequest.Builder getBuilder(SourceFilter sourceFilter, String from, String to) {

        List<FilterType> filters = new ArrayList<>();

        filters.add(new FilterType(sourceFilter.getField(), OperatorType.values()[sourceFilter.getOperator()], sourceFilter.getValue()));
        filters.add(new FilterType("@timestamp", OperatorType.IS_BETWEEN, List.of(from, to)));

        SearchRequest.Builder srb = new SearchRequest.Builder();
        srb.size(0);
        srb.index(sourceFilter.getSource().getIndexPattern());
        srb.query(SearchUtil.toQuery(filters));
        srb.aggregations("recent_events_by_user", agg -> agg
                .terms(t -> t.field("logx.wineventlog.event_data.SubjectUserName.keyword").order(List.of(Map.of("_count", SortOrder.Desc))))
                .aggregations("top_events", subAgg -> subAgg.topHits(th -> th
                        .size(1)
                        .sort(sort -> sort.field(f -> f.field("@timestamp").order(SortOrder.Desc))))));
        return srb;
    }

    List<EventLog> userAttributes(TopHitsAggregate topEvents) {
        ObjectMapper objectMapper = new ObjectMapper();
        List<EventLog> eventLogs = new ArrayList<>();
        List<Hit<JsonData>> searchHits = topEvents.hits().hits();

        searchHits.forEach(hit -> {
            JsonData jsonData = hit.source();
            try {
                if (jsonData != null) {
                    eventLogs.add(objectMapper.readValue(jsonData.toString(), EventLog.class));
                }
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        });

        return eventLogs;
    }

}