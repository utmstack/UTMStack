package com.park.utmstack.service.network_scan;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.datainput_ingestion.UtmDataInputStatusCheckpoint;
import com.park.utmstack.repository.datainput_ingestion.UtmDataInputStatusCheckpointRepository;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.service.logstash_pipeline.response.statistic.StatisticDocument;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
@Slf4j
@RequiredArgsConstructor
public class SourceActivityProvider {

    private final ElasticsearchService elasticsearchService;
    private final UtmDataInputStatusCheckpointRepository checkpointRepository;

    private static final long CHECKPOINT_ID = 1L;
    private static final int OVERLAP_SECONDS = 5;

    public Map<String, StatisticDocument> fetchLatestSourceActivity() {
        UtmDataInputStatusCheckpoint checkpoint = getOrCreateCheckpoint();

        String fromTimestamp = checkpoint.getLastProcessedTimestamp()
                .minus(OVERLAP_SECONDS, ChronoUnit.SECONDS)
                .toString();

        SearchRequest searchRequest = buildActivityQuery(fromTimestamp);

        try {
            SearchResponse<StatisticDocument> response =
                    elasticsearchService.search(searchRequest, StatisticDocument.class);

            Map<String, StatisticDocument> activityMap = extractLatestHits(response);

            if (!activityMap.isEmpty()) {
                log.debug("Fetched {} active sources from statistics index", activityMap.size());
                updateCheckpoint(checkpoint, activityMap);
            } else {
                log.debug("No new source activity found since checkpoint: {}", checkpoint.getLastProcessedTimestamp());
            }

            return activityMap;
        } catch (Exception e) {
            log.error("Error fetching telemetry from Elasticsearch: {}", e.getMessage(), e);
            return Collections.emptyMap();
        }
    }

    private SearchRequest buildActivityQuery(String fromTimestamp) {
        List<FilterType> filters = List.of(
            new FilterType("type", OperatorType.IS, "enqueue_success"),
            new FilterType("@timestamp", OperatorType.IS_GREATER_THAN, fromTimestamp)
        );

        return SearchRequest.of(s -> s
            .index(Constants.STATISTICS_INDEX_PATTERN)
            .query(SearchUtil.toQuery(filters))
            .aggregations("by_source", agg -> agg
                .terms(t -> t
                    .field("dataSource.keyword")
                    .size(10000)
                )
                .aggregations("by_type", typeAgg -> typeAgg
                    .terms(t -> t
                        .field("dataType.keyword")
                        .size(10000)
                    )
                    // Get the latest document for each dataSource + dataType combination
                    .aggregations("latest_doc", latestAgg -> latestAgg
                        .topHits(th -> th
                            .size(1)
                            .sort(sort -> sort.field(f -> f.field("@timestamp").order(SortOrder.Desc)))
                        )
                    )
                )
            )
            .size(0)
        );
    }

    private Map<String, StatisticDocument> extractLatestHits(SearchResponse<StatisticDocument> response) {
        Map<String, StatisticDocument> results = new HashMap<>();

        try {
            var aggregations = response.aggregations();
            if (aggregations == null || aggregations.get("by_source") == null) {
                log.warn("No aggregation results found in response");
                return results;
            }

            var bySourceAgg = aggregations.get("by_source").sterms();
            if (bySourceAgg == null || bySourceAgg.buckets().array().isEmpty()) {
                log.debug("No data source buckets found in aggregation");
                return results;
            }

            bySourceAgg.buckets().array().forEach(sourceBucket -> {
                String dataSource = sourceBucket.key();

                var byTypeAgg = sourceBucket.aggregations().get("by_type").sterms();
                if (byTypeAgg == null || byTypeAgg.buckets().array().isEmpty()) {
                    log.debug("No data type buckets found for source: {}", dataSource);
                    return;
                }

                byTypeAgg.buckets().array().forEach(typeBucket -> {
                    String dataType = typeBucket.key();

                    var latestDocsAgg = typeBucket.aggregations().get("latest_doc");
                    if (latestDocsAgg != null) {
                        var topHits = latestDocsAgg.topHits();
                        if (topHits != null && !topHits.hits().hits().isEmpty()) {
                            var hit = topHits.hits().hits().get(0);
                            if (hit.source() != null) {
                                StatisticDocument doc = hit.source().to(StatisticDocument.class);
                                if (doc != null) {
                                    String compositeKey = dataSource + "|" + dataType;
                                    results.put(compositeKey, doc);
                                }
                            }
                        }
                    }
                });
            });

            return results;
        } catch (Exception e) {
            log.error("Error extracting latest hits from aggregation response: {}", e.getMessage(), e);
            return results;
        }
    }

    private void updateCheckpoint(UtmDataInputStatusCheckpoint checkpoint, Map<String, StatisticDocument> activityMap) {
        activityMap.values().stream()
                .map(doc -> {
                    try {
                        return Instant.parse(doc.getTimestamp());
                    } catch (Exception e) {
                        log.error("Failed to parse timestamp '{}': {}", doc.getTimestamp(), e.getMessage());
                        return null;
                    }
                })
                .filter(java.util.Objects::nonNull)
                .max(Instant::compareTo)
                .ifPresentOrElse(
                    latest -> {
                        checkpoint.setLastProcessedTimestamp(latest);
                        checkpointRepository.save(checkpoint);
                        log.info("Checkpoint updated to: {} ({} active sources)", latest, activityMap.size());
                    },
                    () -> log.debug("No valid timestamps found to update checkpoint")
                );
    }

    private UtmDataInputStatusCheckpoint getOrCreateCheckpoint() {
        return checkpointRepository.findById(CHECKPOINT_ID)
                .orElseGet(() -> {
                    UtmDataInputStatusCheckpoint cp = new UtmDataInputStatusCheckpoint();
                    cp.setLastProcessedTimestamp(Instant.now().minus(1, ChronoUnit.HOURS));
                    return cp;
                });
    }
}
