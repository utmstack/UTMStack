package com.park.utmstack.service.threat_management;


import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.shared_types.alert.Side;
import com.park.utmstack.domain.shared_types.alert.UtmAlert;
import com.park.utmstack.service.dto.threat_management.Adversary;
import com.park.utmstack.service.dto.threat_management.AdversaryAlertsResponseDto;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import lombok.RequiredArgsConstructor;
import org.opensearch.client.json.JsonData;
import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

import static com.park.utmstack.config.Constants.V11_ALERTS_INDEX_PATTERN;

@Service
@RequiredArgsConstructor
public class AdversaryAlertsService {

    private final ElasticsearchService elasticsearchService;

    public List<AdversaryAlertsResponseDto> fetchAdversaryAlerts(List<FilterType> filters){
        SearchRequest request = SearchRequest.of(s -> s
                .index(V11_ALERTS_INDEX_PATTERN)
                .query(SearchUtil.toQuery(filters))
                .size(0)
                .aggregations("adversary", a -> a
                        .terms(t -> t.field("adversary.host.keyword").size(100))
                        .aggregations("adversary_obj", ag -> ag
                                .topHits(th -> th
                                        .size(1)
                                        .sort(srt -> srt.field(f -> f.field("@timestamp")
                                                .order(SortOrder.Desc)))
                                        .source(src -> src.filter(f -> f.includes(List.of("adversary"))))
                                )
                        )
                        .aggregations("alerts", ag -> ag
                                .filter(flt -> flt
                                        .bool(b -> b
                                                .mustNot(mn -> mn.exists(e -> e.field("parentId")))
                                        )
                                )
                                .aggregations("alerts_hits", subAg -> subAg
                                        .topHits(th -> th
                                                .size(100)
                                                .sort(srt -> srt.field(f -> f.field("@timestamp")
                                                        .order(SortOrder.Desc)))
                                        )
                                )
                        )

                        .aggregations("child_alerts", ag -> ag
                                .terms(t -> t.field("parentId.keyword").size(50))
                                .aggregations("child_hits", ch -> ch
                                        .topHits(th -> th
                                                .size(50)
                                                .sort(srt -> srt.field(f -> f.field("@timestamp")
                                                        .order(SortOrder.Desc)))
                                        )
                                )
                        )
                )
        );

        SearchResponse<JsonData> response = elasticsearchService.search(request, JsonData.class);

        List<AdversaryAlertsResponseDto> groups = new ArrayList<>();

        var adversaryBuckets = response.aggregations()
                .get("adversary")
                .sterms()
                .buckets()
                .array();

        for (var bucket : adversaryBuckets) {
            var adversaryHit = bucket.aggregations().get("adversary_obj").topHits().hits().hits().get(0);

            if (adversaryHit.source() != null) {
                Adversary adversaryWrapper = adversaryHit.source().to(Adversary.class);
                Side adversary = adversaryWrapper.getAdversary();

                var topHitsAgg = bucket.aggregations().get("alerts").filter().aggregations()
                        .get("alerts_hits").topHits();
                List<UtmAlert> alerts = topHitsAgg.hits().hits().stream()
                        .filter(hit -> hit.source() != null)
                        .map(hit -> hit.source().to(UtmAlert.class))
                        .filter(alert -> Objects.isNull(alert.getParentId()))
                        .toList();

                Map<String, List<UtmAlert>> childMap = new HashMap<>();
                var childAgg = bucket.aggregations().get("child_alerts").sterms();

                for (var cb : childAgg.buckets().array()) {
                    var childHits = cb.aggregations().get("child_hits").topHits();
                    List<UtmAlert> children = childHits.hits().hits().stream()
                            .filter(hit -> hit.source() != null)
                            .map(hit -> hit.source().to(UtmAlert.class))
                            .toList();
                    childMap.put(cb.key(), children);
                }

                List<AdversaryAlertsResponseDto.AlertWithChildren> alertsWithChildren = alerts.stream()
                        .map(alert -> AdversaryAlertsResponseDto.AlertWithChildren.builder()
                                .alert(alert)
                                .children(childMap.getOrDefault(alert.getId(), Collections.emptyList()))
                                .build())
                        .collect(Collectors.toList());

                groups.add(AdversaryAlertsResponseDto.builder()
                        .adversary(adversary)
                        .alerts(alertsWithChildren)
                        .build());
            }
        }

        return groups;
    }


}

