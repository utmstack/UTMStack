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
        // Post the event-bag rename, an IS on a bag field dual-reads both
        // generations: a bool(should) of term(event.<leaf>) OR
        // match_phrase(log.<leaf>). The top-level filter clause is that bool,
        // and no bare term / match_phrase sits at the top level.
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS, "event.eventCode", "4624")));
        BoolQuery b = bool(q);
        assertTrue(clauses(b).stream().anyMatch(Query::isBool),
            "IS on a bag field should emit a dual-read bool, got: " + q);
        assertTrue(clauses(b).stream().noneMatch(Query::isTerm),
            "dual-read must nest the term inside a bool, not emit a bare term: " + q);
        assertTrue(clauses(b).stream().noneMatch(Query::isMatchPhrase),
            "dual-read must nest the match_phrase inside a bool: " + q);
    }

    @Test
    public void isOneOfOnBagFieldWithEmptyListMatchesNothing() {
        // A regression guard: an empty IS_ONE_OF on a bag field must NOT be a
        // match-all. (A bool with minimum_should_match=1 and zero should
        // clauses is a match-all in OpenSearch, which would silently return
        // every document.) The guard emits mustNot(matchAll).
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS_ONE_OF, "event.eventCode", List.of())));
        BoolQuery b = bool(q);
        assertTrue(clauses(b).stream().anyMatch(c -> c.isBool()
            && c.bool().mustNot().stream().anyMatch(Query::isMatchAll)),
            "empty IS_ONE_OF on a bag field should match nothing, got: " + q);
    }

    @Test
    public void bagFieldWithKeywordSuffixStripsSuffix() {
        // A filter written against the legacy "event.<leaf>.keyword" form must
        // still resolve to the same flat_object field "event.<leaf>" (which has
        // no .keyword sub-field) on the new generation.
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS, "event.eventCode.keyword", "4624")));
        BoolQuery b = bool(q);
        assertTrue(clauses(b).stream().anyMatch(c -> c.isBool()
            && c.bool().should().stream().anyMatch(s -> s.isTerm()
                && s.term().field().equals("event.eventCode"))),
            ".keyword suffix should be stripped to the flat_object field, got: " + q);
    }

    @Test
    public void isOneOfOnFlattenedFieldUsesTerms() {
        // Same dual-read for the list operator: a bool(should) carrying
        // term(event.<leaf>) / match_phrase(log.<leaf>) per value.
        Query q = SearchUtil.toQuery(List.of(f(OperatorType.IS_ONE_OF, "event.eventCode", List.of("4624", "4670"))));
        BoolQuery b = bool(q);
        assertTrue(clauses(b).stream().anyMatch(Query::isBool),
            "IS_ONE_OF on a bag field should emit a dual-read bool, got: " + q);
        assertTrue(clauses(b).stream().noneMatch(Query::isTerms),
            "dual-read must nest the terms inside a bool, not emit a bare terms: " + q);
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
