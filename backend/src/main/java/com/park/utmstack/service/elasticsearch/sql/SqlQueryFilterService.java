package com.park.utmstack.service.elasticsearch.sql;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class SqlQueryFilterService {

    /**
     * Applies the given filters to the base SQL query by generating a dynamic WHERE clause.
     * - Handles all FilterType operators.
     * - Treats @timestamp specially (relative and absolute ranges).
     * - Merges the generated WHERE with an existing WHERE if present.
     */
    public String applyFilters(String baseSql, List<FilterType> filters) {
        if (filters == null || filters.isEmpty()) {
            return baseSql;
        }

        List<String> andConditions = new ArrayList<>();
        List<String> orConditions = new ArrayList<>();

        for (FilterType filter : filters) {

            // Special handling for @timestamp: relative/absolute time logic
            if ("@timestamp".equals(filter.getField())) {
                andConditions.add(buildTimestampCondition(filter));
                continue;
            }

            String sqlCondition = toSqlCondition(filter);

            // IS_ONE_OF_TERMS_OR is explicitly an OR-group operator
            if (filter.getOperator() == OperatorType.IS_ONE_OF_TERMS_OR) {
                orConditions.add(sqlCondition);
            } else {
                andConditions.add(sqlCondition);
            }
        }

        String whereClause = combineConditions(andConditions, orConditions);
        return mergeSql(baseSql, whereClause);
    }

    // -------------------------------------------------------------------------
    //  TIMESTAMP HANDLING
    // -------------------------------------------------------------------------

    /**
     * Builds the SQL condition for @timestamp.
     * The value may be:
     * - A List<?> of two elements (for IS_BETWEEN)
     * - A single String (for > or <=)
     */
    private String buildTimestampCondition(FilterType f) {

        Object rawValue = f.getValue();

        switch (f.getOperator()) {

            case IS_BETWEEN:
                if (!(rawValue instanceof List<?> list) || list.size() != 2) {
                    throw new IllegalArgumentException("@timestamp IS_BETWEEN requires a list of two values");
                }

                String from = String.valueOf(list.get(0));
                String to   = String.valueOf(list.get(1));

                return "@timestamp BETWEEN " + toSqlTime(from) + " AND " + toSqlTime(to);

            case IS_GREATER_THAN:
                if (!(rawValue instanceof String singleGt)) {
                    throw new IllegalArgumentException("@timestamp IS_GREATER_THAN requires a single value");
                }

                return "@timestamp > " + toSqlTime(singleGt);

            case IS_LESS_THAN_OR_EQUALS:
                if (!(rawValue instanceof String singleLe)) {
                    throw new IllegalArgumentException("@timestamp IS_LESS_THAN_OR_EQUALS requires a single value");
                }

                return "@timestamp <= " + toSqlTime(singleLe);


            default:
                throw new IllegalArgumentException("Unsupported timestamp operator: " + f.getOperator());
        }
    }


    /**
     * Converts a logical time value into a SQL expression:
     * - "now"      -> NOW()
     * - "now-24h"  -> DATE_SUB(NOW(), INTERVAL 24 HOUR)
     * - "now-15m"  -> DATE_SUB(NOW(), INTERVAL 15 MINUTE)
     * - "now-7d"   -> DATE_SUB(NOW(), INTERVAL 7 DAY)
     * - any other  -> quoted literal (absolute timestamp)
     */
    private String toSqlTime(String value) {
        if ("now".equals(value)) {
            return "NOW()";
        }

        if (value.startsWith("now-")) {
            // Example: now-24h, now-15m, now-7d
            String number = value.substring(4, value.length() - 1);
            char unit = value.toLowerCase().charAt(value.length() - 1);

            String sqlUnit = switch (unit) {
                case 'm' -> "MINUTE";
                case 'h' -> "HOUR";
                case 'd' -> "DAY";
                default -> throw new IllegalArgumentException("Invalid time unit in value: " + value);
            };

            return "DATE_SUB(NOW(), INTERVAL " + number + " " + sqlUnit + ")";
        }

        // Absolute timestamp value
        return "'" + value + "'";
    }

    // -------------------------------------------------------------------------
    //  GENERAL OPERATORS
    // -------------------------------------------------------------------------

    /**
     * Maps a FilterType to a SQL condition string.
     * All non-@timestamp operators are handled here.
     */
    private String toSqlCondition(FilterType f) {

        String field = f.getField();
        Object rawValue = f.getValue();
        List<String> list = asList(rawValue); // safe conversion

        return switch (f.getOperator()) {

            // ---------------------------------------------------------------------
            // Equality
            // ---------------------------------------------------------------------
            case IS -> field + " = '" + list.get(0) + "'";

            case IS_NOT -> field + " <> '" + list.get(0) + "'";

            // ---------------------------------------------------------------------
            // Text contains
            // ---------------------------------------------------------------------
            case CONTAIN -> field + " LIKE '%" + list.get(0) + "%'";

            case DOES_NOT_CONTAIN -> field + " NOT LIKE '%" + list.get(0) + "%'";

            case CONTAIN_ONE_OF ->
                    "(" + list.stream()
                            .map(v -> field + " LIKE '%" + v + "%'")
                            .collect(Collectors.joining(" OR ")) + ")";

            case DOES_NOT_CONTAIN_ONE_OF ->
                    "(" + list.stream()
                            .map(v -> field + " NOT LIKE '%" + v + "%'")
                            .collect(Collectors.joining(" AND ")) + ")";

            // ---------------------------------------------------------------------
            // List membership
            // ---------------------------------------------------------------------
            case IS_ONE_OF ->
                    field + " IN (" + joinQuoted(list) + ")";

            case IS_NOT_ONE_OF ->
                    field + " NOT IN (" + joinQuoted(list) + ")";

            case IS_ONE_OF_TERMS ->
                    field + " IN (" + joinQuoted(list) + ")";

            case IS_ONE_OF_TERMS_OR ->
                    "(" + list.stream()
                            .map(v -> field + " = '" + v + "'")
                            .collect(Collectors.joining(" OR ")) + ")";

            // ---------------------------------------------------------------------
            // Existence
            // ---------------------------------------------------------------------
            case EXIST -> field + " IS NOT NULL";

            case DOES_NOT_EXIST -> field + " IS NULL";

            // ---------------------------------------------------------------------
            // Ranges (non-timestamp fields)
            // ---------------------------------------------------------------------
            case IS_BETWEEN -> {
                if (list.size() != 2) {
                    throw new IllegalArgumentException("IS_BETWEEN requires exactly 2 values");
                }
                yield field + " BETWEEN '" + list.get(0) + "' AND '" + list.get(1) + "'";
            }

            case IS_NOT_BETWEEN -> {
                if (list.size() != 2) {
                    throw new IllegalArgumentException("IS_NOT_BETWEEN requires exactly 2 values");
                }
                yield field + " NOT BETWEEN '" + list.get(0) + "' AND '" + list.get(1) + "'";
            }

            case IS_GREATER_THAN ->
                    field + " > '" + list.get(0) + "'";

            case IS_LESS_THAN_OR_EQUALS ->
                    field + " <= '" + list.get(0) + "'";

            // ---------------------------------------------------------------------
            // Starts / ends with
            // ---------------------------------------------------------------------
            case START_WITH ->
                    field + " LIKE '" + list.get(0) + "%'";

            case NOT_START_WITH ->
                    field + " NOT LIKE '" + list.get(0) + "%'";

            case ENDS_WITH ->
                    field + " LIKE '%" + list.get(0) + "'";

            case NOT_ENDS_WITH ->
                    field + " NOT LIKE '%" + list.get(0) + "'";

            // ---------------------------------------------------------------------
            // Value in multiple fields
            // ---------------------------------------------------------------------
            case IS_IN_FIELDS ->
                    "'" + list.get(0) + "' IN (" + String.join(", ", list) + ")";

            case IS_NOT_IN_FIELDS ->
                    "'" + list.get(0) + "' NOT IN (" + String.join(", ", list) + ")";

            // ---------------------------------------------------------------------
            default -> throw new IllegalArgumentException("Unsupported operator: " + f.getOperator());
        };
    }

    /**
     * Joins a list of values into a comma-separated list of quoted literals.
     * Example: ["a","b"] -> 'a', 'b'
     */
    private String joinQuoted(List<String> values) {
        return values.stream()
                .map(v -> "'" + v + "'")
                .collect(Collectors.joining(", "));
    }

    // -------------------------------------------------------------------------
    //  AND / OR COMBINATION
    // -------------------------------------------------------------------------

    /**
     * Combines AND and OR condition lists into a single SQL expression.
     * - AND conditions are grouped in parentheses.
     * - OR conditions are grouped in parentheses.
     * - If both exist: (AND...) AND (OR...)
     */
    private String combineConditions(List<String> ands, List<String> ors) {

        String andPart = ands.isEmpty()
                ? ""
                : "(" + String.join(" AND ", ands) + ")";

        String orPart = ors.isEmpty()
                ? ""
                : "(" + String.join(" OR ", ors) + ")";

        if (!andPart.isEmpty() && !orPart.isEmpty()) {
            return andPart + " AND " + orPart;
        }

        return andPart + orPart;
    }

    // -------------------------------------------------------------------------
    //  MERGING WITH BASE SQL
    // -------------------------------------------------------------------------

    /**
     * Merges the generated WHERE clause into the base SQL.
     * - If base SQL already has WHERE, appends "AND <whereClause>".
     * - Otherwise, appends "WHERE <whereClause>".
     */
    private String mergeSql(String sql, String whereClause) {
        if (whereClause == null || whereClause.isBlank()) {
            return sql;
        }

        String normalized = normalizeForSearch(sql);

        // Case 1: SQL already contains WHERE → append AND
        int whereIndex = normalized.indexOf(" where ");
        if (whereIndex != -1) {
            // Find where the WHERE clause ends (before GROUP BY / ORDER BY / LIMIT)
            int endOfWhere = findEndOfWhereClause(normalized, whereIndex + 7);
            return sql.substring(0, endOfWhere)
                    + " AND " + whereClause + " "
                    + sql.substring(endOfWhere);
        }

        // Case 2: Insert WHERE after FROM <target>
        int fromIndex = normalized.indexOf(" from ");
        if (fromIndex != -1) {
            int insertPos = findEndOfFromTarget(sql, normalized, fromIndex + 6);
            return sql.substring(0, insertPos)
                    + " WHERE " + whereClause + " "
                    + sql.substring(insertPos);
        }

        // Case 3: fallback
        return sql + " WHERE " + whereClause;
    }

    private int findEndOfFromTarget(String originalSql, String normalizedSql, int start) {

        List<String> keywords = List.of(" group by ", " order by ", " limit ", " having ", " where ");

        int nextKeywordPos = normalizedSql.length();

        for (String kw : keywords) {
            int idx = normalizedSql.indexOf(kw, start);
            if (idx != -1 && idx < nextKeywordPos) {
                nextKeywordPos = idx;
            }
        }

        // Convert normalized index back to original index
        String before = normalizedSql.substring(0, nextKeywordPos);
        String lastToken = before.trim().substring(before.trim().lastIndexOf(" ") + 1);

        int originalIndex = originalSql.toLowerCase().indexOf(lastToken);

        return originalIndex == -1 ? originalSql.length() : originalIndex + lastToken.length();
    }

    private int findEndOfWhereClause(String normalized, int start) {
        List<String> keywords = List.of(" group by ", " order by ", " limit ", " having ");

        int nextKeywordPos = normalized.length();

        for (String kw : keywords) {
            int idx = normalized.indexOf(kw, start);
            if (idx != -1 && idx < nextKeywordPos) {
                nextKeywordPos = idx;
            }
        }

        return nextKeywordPos;
    }

    private List<String> asList(Object value) {
        if (value == null) {
            return List.of();
        }
        if (value instanceof List<?> list) {
            return list.stream().map(String::valueOf).toList();
        }
        return List.of(String.valueOf(value)); // single value → list of one
    }

    private String normalizeForSearch(String sql) {
        return sql
                .replace("\n", " ")
                .replace("\r", " ")
                .replace("\t", " ")
                .replaceAll(" +", " ")
                .toLowerCase();
    }
}

