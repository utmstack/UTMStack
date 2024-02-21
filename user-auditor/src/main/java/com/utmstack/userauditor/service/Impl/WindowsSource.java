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
import java.time.Duration;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.LocalTime;
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

    static final String SUBJECT_AGG_NAME = "SUBJECT_USER_NAME";
    static final String TARGET_AGG_NAME = "TARGET_USER_NAME";

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
                    this.executeQuery(startDate.format(formatter), this.executionDate(startDate).format(formatter), 10, filter);
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

    void executeQuery(String from, String to, Integer top, SourceFilter sourceFilter) {

        SearchRequest.Builder srb = getBuilder(sourceFilter, from, to);

        SearchResponse<String> rs = elasticsearchService.search(srb.build(), String.class);

        List<BucketAggregation> buckets = TermAggregateParser.parse(rs.aggregations().get(SUBJECT_AGG_NAME));
        buckets.addAll(TermAggregateParser.parse(rs.aggregations().get(TARGET_AGG_NAME)));

        buckets.forEach(b -> {
            if (!this.userEvents.containsKey(this.getKey(b))) {
                this.userEvents.put(this.getKey(b), getLastEvents(b.getSubAggregations().get("top_events").topHits()));
            } else {
                this.userEvents.get(this.getKey(b)).addAll(getLastEvents(b.getSubAggregations().get("top_events").topHits()));
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

        srb.aggregations(SUBJECT_AGG_NAME, agg -> agg
                .terms(t -> t.field("logx.wineventlog.event_data.SubjectUserName.keyword")
                        .order(List.of(Map.of("_count", SortOrder.Desc))))
                .aggregations("top_events", subAgg -> subAgg.topHits(th -> th
                        .size(1)
                        .sort(sort -> sort.field(f -> f.field("@timestamp").order(SortOrder.Desc))))));

        srb.aggregations(TARGET_AGG_NAME, agg -> agg
                .terms(t -> t.field("logx.wineventlog.event_data.TargetUserName.keyword")
                        .order(List.of(Map.of("_count", SortOrder.Desc))))
                .aggregations("top_events", subAgg -> subAgg.topHits(th -> th
                        .size(1)
                        .sort(sort -> sort.field(f -> f.field("@timestamp").order(SortOrder.Desc))))));
        return srb;
    }

    List<EventLog> getLastEvents(TopHitsAggregate topEvents) {
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

    private String getKey(BucketAggregation bucket) {
        if (bucket.getKey().equals("-")) {
            EventLog eventLog = this.getLastEvents(bucket.getSubAggregations().get("top_events").topHits()).get(0);
            return  eventLog.getLogx().wineventlog.eventData.subjectUserName;
        } else {
            return bucket.getKey();
        }
    }

}