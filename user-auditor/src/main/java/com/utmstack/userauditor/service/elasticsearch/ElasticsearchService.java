package com.utmstack.userauditor.service.elasticsearch;

import com.utmstack.userauditor.service.elasticsearch.utils.exception.exceptions.UtmElasticsearchException;
import com.utmstack.userauditor.service.type.FilterType;
import lombok.extern.slf4j.Slf4j;
import org.jetbrains.annotations.NotNull;
import org.opensearch.client.json.JsonData;
import org.opensearch.client.opensearch._types.query_dsl.BoolQuery;
import org.opensearch.client.opensearch._types.query_dsl.MatchAllQuery;
import org.opensearch.client.opensearch._types.query_dsl.Query;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;

import java.util.List;


/**
 * @author Leonardo M. López
 */
@Service
@Slf4j
public class ElasticsearchService {
    private static final String CLASSNAME = "ElasticsearchService";
    private final OpensearchClientBuilder client;

    public ElasticsearchService(OpensearchClientBuilder client) {
        this.client = client;
    }

    public <T> SearchResponse<T> search(SearchRequest request, Class<T> type) {
        final String ctx = CLASSNAME + ".search";
        try {
            return client.getClient().search(request, type);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public <T> SearchResponse<T> search(List<FilterType> filters, Integer top, String indexPattern,
                                        Pageable pageable, Class<T> type) {
        final String ctx = CLASSNAME + ".search";
        try {
            Assert.hasText(indexPattern, "Parameter indexPattern must not be null or empty");
            SearchRequest query = buildQuery(indexPattern, filters, top, pageable);
            return client.getClient().search(query, type);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public boolean indexExist(String index) {
        final String ctx = CLASSNAME + ".indexExist";
        try {
            return client.getClient().indexExist(index);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            return false;
        }
    }

    public <T> SearchResponse<T> searchBySid(String sid, String to, String from, String indexPattern, Pageable pageable, Class<T> type) throws UtmElasticsearchException {

        final String ctx = CLASSNAME + ".buildQuery";
        try {
            BoolQuery.Builder bool = new BoolQuery.Builder();
            bool.filter(Query.of(q -> q.matchAll(MatchAllQuery.of(m -> m))));

            bool.filter(f -> f.range(r -> r.field("@timestamp")
                    .gte(JsonData.of(from))
                    .lte(JsonData.of(to))));

            BoolQuery.Builder shouldList = getBuilder(sid);

            bool.must(f -> f.bool(shouldList.build()));
            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.index(indexPattern);
            SearchUtil.applyPaginationAndSort(srb, pageable, 10000);
            return client.getClient().search(srb.query(Query.of(q -> q.bool(bool.build()))).build(), type);
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }

    }

    @NotNull
    private static BoolQuery.Builder getBuilder(String sid) {
        BoolQuery.Builder shouldList = new BoolQuery.Builder();
        shouldList.minimumShouldMatch("1");
        shouldList.should(f -> f.matchPhrase(m -> m.field(Constants.logxWineventlogEventDataTargetSidKeyword).query(sid)));
        shouldList.should(f -> f.matchPhrase(m -> m.field(Constants.logxWineventlogEventDataTargetUserSidKeyword).query(String.valueOf(sid))));
        shouldList.should(f -> f.matchPhrase(m -> m.field(Constants.logxWineventlogEventDataMemberSidKeyword).query(String.valueOf(sid))));
        return shouldList;
    }

    private SearchRequest buildQuery(String pattern, List<FilterType> filters, Integer top, Pageable pageable) throws UtmElasticsearchException {
        final String ctx = CLASSNAME + ".buildQuery";
        try {
            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.index(pattern);
            SearchUtil.applyPaginationAndSort(srb, pageable, top);
            return srb.query(SearchUtil.toQuery(filters)).build();
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }
    }

}
