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
import java.time.*;
import java.time.format.DateTimeFormatter;
import java.time.temporal.ChronoUnit;
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

    @Value("${app.elasticsearch.searchIntervalMinutes}")
    private int searchIntervalMinutes;

    Map<String, List<EventLog>> userEvents;
    final ElasticsearchService elasticsearchService;

    final SourceScanRepository sourceScanRepository;

    private static final String TARGET_AGG_NAME = "TARGET_USER_NAME";

    private static final String FIELD = "logx.wineventlog.event_data.TargetUserName.keyword";

    private static final int ITEMS_PER_PAGE = 1000;

    private final  DateTimeFormatter shortDateFormatter = DateTimeFormatter.ofPattern("yyyy-MM-dd");
    private final  DateTimeFormatter longDateFormatter = DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm");

    private final ObjectMapper objectMapper;

    final String ctx = "UserService.windowsEventsLogs";

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

        this.userEvents = new HashMap<>();

        SourceScan scan = sourceScanRepository.findLatestBySourceId(userSource.getId())
                .orElse(SourceScan.builder()
                                .source(userSource)
                                .executionDate(LocalDateTime.of(LocalDate.of(startYear, 1, 1), LocalTime.now()))
                                .build());

        LocalDateTime startDate = scan.getExecutionDate();

        try {
            if (!elasticsearchService.indexExist(userSource.getIndexPattern()))
                return this.userEvents;

            LocalDateTime currentDateTime = LocalDateTime.now();

            if (startDate.isBefore(currentDateTime) && ChronoUnit.MINUTES.between(startDate, currentDateTime) >= searchIntervalMinutes) {
                LocalDateTime nextDate = this.executionDate(startDate);

                for (SourceFilter filter : userSource.getFilters()) {
                    this.executeQuery(this.getRange(startDate, nextDate), filter);
                }
                scan.setExecutionDate(nextDate);
                sourceScanRepository.save(scan);
            }

            return this.userEvents;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    List<CompositeBucket> getBuckets(List<String> dateRange, SourceFilter sourceFilter, String after) {

        SearchRequest.Builder srb = getBuilder(sourceFilter, dateRange, after);
        SearchResponse<String> rs = elasticsearchService.search(srb.build(), String.class);
        List<CompositeBucket> buckets = rs.aggregations().get(TARGET_AGG_NAME).composite().buckets().array();

        CompositeAggregate aggregate = rs.aggregations().get(TARGET_AGG_NAME).composite();

        if (!buckets.isEmpty()) {
            after = aggregate.afterKey().get("users").toString().replace("\"", "");
            buckets.addAll(getBuckets(dateRange, sourceFilter, after));
        }
        return buckets;
    }

    void executeQuery(List<String> dateRange, SourceFilter sourceFilter) {

        getBuckets(dateRange, sourceFilter, "").stream()
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


    private static SearchRequest.Builder getBuilder(SourceFilter sourceFilter, List<String> range, String after) {

        List<FilterType> filters = new ArrayList<>();

        filters.add(new FilterType(sourceFilter.getField(), OperatorType.values()[sourceFilter.getOperator()], sourceFilter.getValue()));
        filters.add(new FilterType("@timestamp", OperatorType.IS_BETWEEN, range));

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

        if (currentDateTime.getMonth().equals(startDate.getMonth()) && currentDateTime.getYear() == startDate.getYear() ) {
            if (ChronoUnit.MINUTES.between(startDate, currentDateTime) >= searchIntervalMinutes){
                return LocalDateTime.of(startDate.toLocalDate(), startDate.toLocalTime()).plusMinutes(ChronoUnit.MINUTES.between(startDate, currentDateTime));
            }
        }

        return LocalDateTime.of(startDate.toLocalDate(), LocalTime.of(0,0,0)).plusMonths(searchIntervalMonths);

    }

    private String getKey(CompositeBucket bucket) {
        try {
            return objectMapper.readValue(bucket.key().get("users").toString(), String.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException(e);
        }
    }

    private List<String> getRange(LocalDateTime startDate, LocalDateTime nextDate){
       String from = startDate.format(shortDateFormatter);
       String to = nextDate.format(shortDateFormatter);

       return from.equals(to) ? List.of(getUTC(startDate).format(longDateFormatter), getUTC(nextDate).format(longDateFormatter)) : List.of(from, to);

    }

    private ZonedDateTime getUTC(LocalDateTime dateTime) {

        ZonedDateTime zonaLocal = dateTime.atZone(ZoneId.systemDefault());

       return zonaLocal.withZoneSameInstant(ZoneId.of("UTC"));
    }


}