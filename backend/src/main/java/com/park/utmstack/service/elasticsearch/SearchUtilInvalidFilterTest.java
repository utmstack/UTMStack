package com.park.utmstack.service.elasticsearch;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.util.exceptions.ApiException;
import org.junit.jupiter.api.Test;
import org.springframework.http.HttpStatus;

import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

/**
 * #2968: structurally invalid filters (IS_ONE_OF / IS_NOT_ONE_OF whose value
 * is not a list) must surface as a 400 ApiException, not a 500 RuntimeException.
 */
public class SearchUtilInvalidFilterTest {

    private static FilterType f(OperatorType op, String field, Object value) {
        return new FilterType(field, op, value);
    }

    @Test
    public void isOneOfWithObjectValueThrowsApiException400() {
        // Row-to-filters used to emit the raw alert row fields: an object
        // value (e.g. adversary = {host: ...}) reaches validate() as a
        // LinkedHashMap. Must be a 400, not a 500.
        ApiException ex = assertThrows(ApiException.class,
            () -> SearchUtil.toQuery(List.of(f(OperatorType.IS_ONE_OF, "adversary", Map.of("host", "10.0.0.1")))));
        assertEquals(HttpStatus.BAD_REQUEST, ex.getStatus());
    }

    @Test
    public void isOneOfWithNullValueThrowsApiException400() {
        // target was null on the alert row -> value null after sanitizing
        // failed. Must be a 400, not a 500.
        ApiException ex = assertThrows(ApiException.class,
            () -> SearchUtil.toQuery(List.of(f(OperatorType.IS_ONE_OF, "target", null))));
        assertEquals(HttpStatus.BAD_REQUEST, ex.getStatus());
    }

    @Test
    public void isOneOfWithScalarValueThrowsApiException400() {
        // echoes was a bare Integer (0) -> not a list -> 400.
        ApiException ex = assertThrows(ApiException.class,
            () -> SearchUtil.toQuery(List.of(f(OperatorType.IS_ONE_OF, "echoes", 0))));
        assertEquals(HttpStatus.BAD_REQUEST, ex.getStatus());
    }

    @Test
    public void isNotOneOfWithObjectValueThrowsApiException400() {
        ApiException ex = assertThrows(ApiException.class,
            () -> SearchUtil.toQuery(List.of(f(OperatorType.IS_NOT_ONE_OF, "impact", Map.of("confidentiality", 1)))));
        assertEquals(HttpStatus.BAD_REQUEST, ex.getStatus());
    }

    @Test
    public void isOneOfWithProperListStillBuildsQuery() {
        // Sanity: a correctly-shaped filter keeps working.
        assertDoesNotThrow(() -> SearchUtil.toQuery(List.of(f(OperatorType.IS_ONE_OF, "name", List.of("rule")))));
    }
}
