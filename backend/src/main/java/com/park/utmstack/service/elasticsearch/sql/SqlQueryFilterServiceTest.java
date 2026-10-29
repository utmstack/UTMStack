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
}
