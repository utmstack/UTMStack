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
import org.opensearch.client.json.JsonData;
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
                updateCheckpoint(checkpoint, activityMap);
            }

            return activityMap;
        } catch (Exception e) {
            log.error("Error consultando telemetría en Elasticsearch: {}", e.getMessage());
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
            .collapse(c -> c
                .field("dataSource.keyword")
                .innerHits(ih -> ih
                    .name("latest")
                    .size(1)
                    .sort(sort -> sort.field(f -> f.field("@timestamp").order(SortOrder.Desc)))
                )
            )
            .size(10000)
        );
    }

    private Map<String, StatisticDocument> extractLatestHits(SearchResponse<StatisticDocument> response) {
        Map<String, StatisticDocument> results = new HashMap<>();

        response.hits().hits().forEach(hit -> {
            if (hit.innerHits() != null && hit.innerHits().containsKey("latest")) {
                var innerHits = hit.innerHits().get("latest").hits().hits();
                if (!innerHits.isEmpty()) {
                    JsonData json = innerHits.get(0).source();
                    if (json != null) {
                        StatisticDocument doc = json.to(StatisticDocument.class);
                        results.put(doc.getDataSource(), doc);
                    }
                }
            }
        });
        return results;
    }

    private void updateCheckpoint(UtmDataInputStatusCheckpoint checkpoint, Map<String, StatisticDocument> activityMap) {
        activityMap.values().stream()
                .map(doc -> Instant.parse(doc.getTimestamp()))
                .max(Instant::compareTo)
                .ifPresent(latest -> {
                    checkpoint.setLastProcessedTimestamp(latest);
                    checkpointRepository.save(checkpoint);
                    log.debug("Checkpoint actualizado a: {}", latest);
                });
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
