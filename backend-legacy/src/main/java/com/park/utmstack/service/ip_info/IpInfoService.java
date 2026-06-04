package com.park.utmstack.service.ip_info;

import com.park.utmstack.config.ApplicationProperties;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.ip_info.GeoIp;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.util.exceptions.UtmIpInfoException;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Service
public class IpInfoService {
    private static final String CLASSNAME = "IpInfoService";

    private final ApplicationProperties applicationProperties;
    private final ElasticsearchService elasticsearchService;

    public IpInfoService(ApplicationProperties applicationProperties,
                         ElasticsearchService elasticsearchService) {
        this.applicationProperties = applicationProperties;
        this.elasticsearchService = elasticsearchService;
    }

    /**
     * Get information about an Ip
     *
     * @param ip The ip to get the related information
     * @return A ${@link GeoIp} object with the ip information
     */
    public GeoIp getIpInfo(String ip) throws UtmIpInfoException {
        final String ctx = CLASSNAME + "getIpV4Info";
        try {
            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType("network", OperatorType.IS, ip));

            SearchRequest sr = SearchRequest.of(s -> s.index(applicationProperties.getChartBuilder().getIpInfoIndexName())
                .query(SearchUtil.toQuery(filters)).size(1));

            HitsMetadata<GeoIp> hits = elasticsearchService.search(sr, GeoIp.class).hits();

            if (hits.total().value() <= 0)
                return null;

            return hits.hits().get(0).source();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}
