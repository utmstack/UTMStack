package com.park.utmstack.service.elasticsearch;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.util.exceptions.ApiException;
import org.junit.jupiter.api.Test;
import org.opensearch.client.opensearch._types.query_dsl.BoolQuery;
import org.opensearch.client.opensearch._types.query_dsl.Query;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.springframework.data.domain.Sort;
import org.springframework.http.HttpStatus;

import java.util.ArrayList;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

public class SearchUtilFlattenedTest {

    private static FilterType f(OperatorType op, String field, Object value) {
        return new FilterType(field, op, value);
    }

    private static BoolQuery bool(Query q) {
        return q.bool();
    }

    /**
     * SearchUtil puts positive clauses in the filter() list of the bool query
     * (negative ones in mustNot()), so scan both to assert clause presence.
     *
     * NOTE: the OpenSearch Java client's tagged-union accessors (term(), prefix(),
     * ...) throw IllegalStateException on a mismatching variant, so always guard
     * with isTerm()/isPrefix()/... before calling the accessor.
     */
    private static List<Query> clauses(BoolQuery b) {
        List<Query> all = new ArrayList<>(b.must());
        all.addAll(b.filter());
        return all;
    }

    @Test
    public void isOnFlattenedFieldUsesTerm() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS, "event.eventCode", "4624")));
        BoolQuery b = bool(q);
        assertTrue(clauses(b).stream().anyMatch(c -> c.isTerm()
            && c.term().field().equals("event.eventCode")
            && c.term().value().stringValue().equals("4624")));
        assertTrue(clauses(b).stream().noneMatch(c -> c.isMatchPhrase()));
    }

    @Test
    public void isOneOfOnFlattenedFieldUsesTerms() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS_ONE_OF, "event.eventCode", List.of("4624", "4670"))));
        BoolQuery b = bool(q);
        assertTrue(clauses(b).stream().anyMatch(c -> c.isTerms()
            && c.terms().field().equals("event.eventCode")
            && c.terms().terms().value().stream().anyMatch(v -> v.stringValue().equals("4624"))));
    }

    @Test
    public void startWithOnFlattenedFieldUsesPrefix() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.START_WITH, "event.message", "SYS")));
        BoolQuery b = bool(q);
        assertTrue(clauses(b).stream().anyMatch(c -> c.isPrefix()
            && c.prefix().field().equals("event.message")));
        assertTrue(clauses(b).stream().noneMatch(c -> c.isWildcard()));
    }

    @Test
    public void textFieldsStillUseMatchPhrase() {
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS, "action", "deny")));
        BoolQuery b = bool(q);
        assertTrue(clauses(b).stream().anyMatch(c -> c.isMatchPhrase()
            && c.matchPhrase().field().equals("action")));
    }

    @Test
    public void sortByFlattenedFieldThrowsApiExceptionBadRequest() {
        SearchRequest.Builder srb = new SearchRequest.Builder();
        Sort sort = Sort.by(Sort.Order.asc("event.eventCode"));
        ApiException ex = assertThrows(ApiException.class, () -> SearchUtil.applySort(srb, sort));
        assertEquals(HttpStatus.BAD_REQUEST, ex.getStatus());
    }

    @Test
    public void sortByNonFlattenedFieldDoesNotThrow() {
        SearchRequest.Builder srb = new SearchRequest.Builder();
        Sort sort = Sort.by(Sort.Order.asc("action"));
        assertDoesNotThrow(() -> SearchUtil.applySort(srb, sort));
    }

    @Test
    public void dateFormattedRangeOnFlattenedFieldThrowsBadRequest() {
        ApiException ex = assertThrows(ApiException.class,
            () -> SearchUtil.toQuery(List.of(f(OperatorType.IS_GREATER_THAN, "event.timestamp", "2026-01-01"))));
        assertEquals(HttpStatus.BAD_REQUEST, ex.getStatus());
    }
}
