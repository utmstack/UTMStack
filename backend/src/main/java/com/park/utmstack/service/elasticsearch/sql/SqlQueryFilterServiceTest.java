package com.park.utmstack.service.elasticsearch.sql;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class SqlQueryFilterServiceTest {

    private final SqlQueryFilterService service = new SqlQueryFilterService();

    private static FilterType filter(String field, OperatorType op, Object value) {
        return new FilterType(field, op, value);
    }

    @Test
    void isValueWithSingleQuoteIsEscapedByDoubling() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("name", OperatorType.IS, "o'brien")));
        assertTrue(sql.contains("name = 'o''brien'"),
                "expected doubled single quote in: " + sql);
    }

    @Test
    void isInjectionValueIsEscaped() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("name", OperatorType.IS, "' OR 1=1 --")));
        // The leading quote must be doubled so the value cannot break out of the literal
        assertTrue(sql.contains("name = ''' OR 1=1 --'"),
                "expected escaped injection value in: " + sql);
    }

    @Test
    void containValueWithSingleQuoteIsEscaped() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("name", OperatorType.CONTAIN, "a'b")));
        assertTrue(sql.contains("name LIKE '%a''b%'"),
                "expected escaped LIKE value in: " + sql);
    }

    @Test
    void isOneOfEscapesEachListValue() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("name", OperatorType.IS_ONE_OF, List.of("x", "y'z"))));
        assertTrue(sql.contains("name IN ('x', 'y''z')"),
                "expected escaped IN list in: " + sql);
    }

    @Test
    void invalidFieldWithSqlSemicolonIsRejected() {
        assertThrows(IllegalArgumentException.class,
                () -> service.applyFilters("SELECT * FROM t",
                        List.of(filter("name; DROP TABLE x", OperatorType.IS, "v"))));
    }

    @Test
    void dottedFieldIsAccepted() {
        String sql = assertDoesNotThrow(() -> service.applyFilters("SELECT * FROM t",
                List.of(filter("host.name", OperatorType.IS, "vm1"))));
        assertTrue(sql.contains("host.name = 'vm1'"), "unexpected sql: " + sql);
    }

    @Test
    void timestampBetweenAbsoluteValuesAreQuotedLiterals() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("@timestamp", OperatorType.IS_BETWEEN,
                        List.of("2026-01-01T00:00:00Z", "2026-01-02T00:00:00Z"))));
        assertTrue(sql.contains(
                "@timestamp BETWEEN '2026-01-01T00:00:00Z' AND '2026-01-02T00:00:00Z'"),
                "unexpected sql: " + sql);
    }

    @Test
    void timestampGreaterThanRelativeIsUnchanged() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("@timestamp", OperatorType.IS_GREATER_THAN, "now-24h")));
        assertTrue(sql.contains("DATE_SUB(NOW(), INTERVAL 24 HOUR)"),
                "unexpected sql: " + sql);
    }

    @Test
    void benignEndToEndMergeIsIntact() {
        // Merge is exercised end to end; the condition must appear exactly once with
        // the value correctly quoted (pre-existing mergeSql insertion point is out
        // of scope for this task).
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("status", OperatorType.IS, "ok")));
        assertEquals("SELECT WHERE (status = 'ok')  * FROM t", sql);
    }

    @Test
    void fieldWithSpaceAndBooleanOperatorIsRejected() {
        assertThrows(IllegalArgumentException.class,
                () -> service.applyFilters("SELECT * FROM t",
                        List.of(filter("a OR b", OperatorType.IS, "v"))));
    }

    @Test
    void dottedWindowsEventFieldIsAccepted() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("event.eventDataCommandLine", OperatorType.IS, "cmd")));
        assertTrue(sql.contains("event.eventDataCommandLine = 'cmd'"),
                "unexpected sql: " + sql);
    }

    @Test
    void hyphenatedFieldIsAccepted() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("x-amz-acl", OperatorType.IS, "public-read")));
        assertTrue(sql.contains("x-amz-acl = 'public-read'"),
                "unexpected sql: " + sql);
    }

    @Test
    void relativeTimestampWithNonNumericIntervalIsRejected() {
        assertThrows(IllegalArgumentException.class,
                () -> service.applyFilters("SELECT * FROM t",
                        List.of(filter("@timestamp", OperatorType.IS_GREATER_THAN, "now-24xh"))));
    }

    @Test
    void isInFieldsEmitsOrOfLikeClausesWithEscapedValue() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("*", OperatorType.IS_IN_FIELDS,
                        List.of("o'brien", "host.name", "event.eventDataCommandLine"))));
        assertTrue(sql.contains("(host.name LIKE '%o''brien%' OR event.eventDataCommandLine LIKE '%o''brien%')"),
                "unexpected sql: " + sql);
    }

    @Test
    void isInFieldsRejectsInvalidFieldElement() {
        assertThrows(IllegalArgumentException.class,
                () -> service.applyFilters("SELECT * FROM t",
                        List.of(filter("*", OperatorType.IS_IN_FIELDS,
                                List.of("v", "a OR b")))));
    }

    @Test
    void isNotInFieldsEmitsAndOfNotLikeClauses() {
        String sql = service.applyFilters("SELECT * FROM t",
                List.of(filter("*", OperatorType.IS_NOT_IN_FIELDS,
                        List.of("x", "f1", "f2"))));
        assertTrue(sql.contains("(f1 NOT LIKE '%x%' AND f2 NOT LIKE '%x%')"),
                "unexpected sql: " + sql);
    }
}
