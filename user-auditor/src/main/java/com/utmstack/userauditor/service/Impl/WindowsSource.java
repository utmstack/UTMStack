package com.utmstack.userauditor.service.Impl;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.utmstack.userauditor.model.*;
import com.utmstack.userauditor.model.winevent.EventLog;
import com.utmstack.userauditor.model.winevent.UserEvent;
import com.utmstack.userauditor.repository.SourceScanRepository;
import com.utmstack.userauditor.service.elasticsearch.ElasticsearchService;
import com.utmstack.userauditor.service.elasticsearch.SearchUtil;
import com.utmstack.userauditor.service.interfaces.Source;
import com.utmstack.userauditor.service.type.FilterType;
import com.utmstack.userauditor.service.type.OperatorType;
import com.utmstack.userauditor.service.type.SourceType;
import org.opensearch.client.json.JsonData;
import org.opensearch.client.opensearch._types.aggregations.CompositeAggregate;
import org.opensearch.client.opensearch._types.aggregations.CompositeBucket;
import org.opensearch.client.opensearch._types.aggregations.TopHitsAggregate;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.opensearch.client.opensearch.core.search.Hit;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.time.Duration;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.LocalTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;


@Component
public class WindowsSource implements Source {

    @Value("${app.elasticsearch.startYear}")
    private int startYear;

    @Value("${app.elasticsearch.searchIntervalMonths}")
    private int searchIntervalMonths;

    Map<String, List<EventLog>> userEvents;
    final ElasticsearchService elasticsearchService;

    final SourceScanRepository sourceScanRepository;

    private static final String TARGET_AGG_NAME = "TARGET_USER_NAME";

    private static final String FIELD = "logx.wineventlog.event_data.TargetUserName.keyword";

    private static final int ITEMS_PER_PAGE = 100;

    private final ObjectMapper objectMapper;

    WindowsSource(ElasticsearchService elasticsearchService, SourceScanRepository sourceScanRepository) {
        this.userEvents = new HashMap<>();
        this.elasticsearchService = elasticsearchService;
        this.sourceScanRepository = sourceScanRepository;
        this.objectMapper = new ObjectMapper();
    }

    @Override
    public SourceType getType() {
        return SourceType.WINDOWS;
    }

    @Override
    public Map<String, List<EventLog>> findUsers(UserSource userSource) throws Exception {

        LocalDateTime currentDateTime = LocalDateTime.now();
        Map<String, List<UserAttribute>> users = new HashMap<>();
        LocalDateTime startDate = LocalDateTime.of(LocalDate.of(startYear, 1, 1), LocalTime.of(0, 0, 0));
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd");
        final String ctx = "UserService.winEventlogs";
        List<SourceScan> scans = sourceScanRepository.findBySource_Id(userSource.getId());

        if (!scans.isEmpty()) {
            startDate = scans.stream()
                    .map(SourceScan::getExecutionDate)
                    .max(LocalDateTime::compareTo)
                    .orElse(startDate);
        }

        try {
            if (!elasticsearchService.indexExist(userSource.getIndexPattern()))
                return this.userEvents;

            if (startDate.toLocalDate().isBefore(currentDateTime.toLocalDate())) {
                for (SourceFilter filter : userSource.getFilters()) {
                    this.executeQuery(startDate.format(formatter), this.executionDate(startDate).format(formatter), filter);
                }
                sourceScanRepository.save(SourceScan.builder()
                        .executionDate(this.executionDate(startDate))
                        .source(userSource)
                        .build());
            }

            return this.userEvents;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    List<CompositeBucket> getBuckets(String from, String to, SourceFilter sourceFilter, String after) {

        SearchRequest.Builder srb = getBuilder(sourceFilter, from, to, after);
        SearchResponse<String> rs = elasticsearchService.search(srb.build(), String.class);
        List<CompositeBucket> buckets = rs.aggregations().get(TARGET_AGG_NAME).composite().buckets().array();

        CompositeAggregate aggregate = rs.aggregations().get(TARGET_AGG_NAME).composite();

        if (!buckets.isEmpty()) {
            after = aggregate.afterKey().get("users").toString().replace("\"", "");
            buckets.addAll(getBuckets(from, to, sourceFilter, after));
        }
        return buckets;
    }

    void executeQuery(String from, String to, SourceFilter sourceFilter) {

        getBuckets(from, to, sourceFilter, "").stream()
                .map(b-> UserEvent.builder()
                        .name(getKey(b))
                        .topEvents(b.aggregations().get("top_events").topHits()).build())
                .filter(user -> !user.getName().isEmpty())
                .forEach(b -> {
                    if (!this.userEvents.containsKey(b.getName())) {
                        this.userEvents.put(b.getName(), getLastEvents(b.getTopEvents()));
                    } else {
                        this.userEvents.get(b.getName()).addAll(getLastEvents(b.getTopEvents()));
                    }
                });
    }


    private static SearchRequest.Builder getBuilder(SourceFilter sourceFilter, String from, String to, String after) {

        List<FilterType> filters = new ArrayList<>();

        filters.add(new FilterType(sourceFilter.getField(), OperatorType.values()[sourceFilter.getOperator()], sourceFilter.getValue()));
        filters.add(new FilterType("@timestamp", OperatorType.IS_BETWEEN, List.of(from, to)));

        SearchRequest.Builder srb = new SearchRequest.Builder();
        srb.size(0);
        srb.index(sourceFilter.getSource().getIndexPattern());
        srb.query(SearchUtil.toQuery(filters));

        srb.aggregations(Map.of(TARGET_AGG_NAME,
                    CompositeAggregationBuilder
                            .buildCompositeTermAggregation(TermsAggregationBuilder.buildTermsAggregation(FIELD), "users", ITEMS_PER_PAGE, after)));
        return srb;
    }

    List<EventLog> getLastEvents(TopHitsAggregate topEvents) {
        //ObjectMapper objectMapper = new ObjectMapper();
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

    private LocalDateTime executionDate(LocalDateTime startDate) {

        LocalDateTime currentDateTime = LocalDateTime.now();
        Duration duration = Duration.between(startDate, currentDateTime);

        long hoursDifference = duration.toHours();
        if (currentDateTime.getMonth().equals(startDate.getMonth()) && currentDateTime.getYear() == startDate.getYear() && hoursDifference >= 24) {
            return LocalDateTime.of(startDate.toLocalDate().plusDays(1), LocalTime.now());
        } else {
            return startDate.plusMonths(searchIntervalMonths);
        }
    }

    private String getKey(CompositeBucket bucket) {
        try {
            return objectMapper.readValue(bucket.key().get("users").toString(), String.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e);
        }
    }

}