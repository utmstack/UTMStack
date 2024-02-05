package com.park.utmstack.service.elasticsearch;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.util.CustomStringEscapeUtil;
import com.park.utmstack.util.UtilPagination;
import org.apache.commons.lang3.ObjectUtils;
import org.opensearch.client.json.JsonData;
import org.opensearch.client.opensearch._types.FieldValue;
import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch._types.query_dsl.*;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.StringFilter;

import java.time.Instant;
import java.util.Arrays;
import java.util.List;
import java.util.Objects;
import java.util.stream.Collectors;

public class SearchUtil {

    private static final String CLASSNAME = "SearchUtil";

    public static Query toQuery(List<FilterType> filters) {
        final String ctx = CLASSNAME + ".toQuery";
        try {
            BoolQuery.Builder bool = new BoolQuery.Builder();
            bool.filter(Query.of(q -> q.matchAll(MatchAllQuery.of(m -> m))));

            if (!CollectionUtils.isEmpty(filters)) {
                for (FilterType filter : filters) {
                    switch (filter.getOperator()) {
                        case IS:
                            buildIsOperator(bool, filter);
                            break;
                        case IS_NOT:
                            buildIsNotOperator(bool, filter);
                            break;
                        case CONTAIN:
                            buildContainOperator(bool, filter);
                            break;
                        case DOES_NOT_CONTAIN:
                            buildDoesNotContainOperator(bool, filter);
                            break;
                        case IS_ONE_OF:
                            buildIsOneOfOperator(bool, filter);
                            break;
                        case IS_NOT_ONE_OF:
                            buildIsNotOneOfOperator(bool, filter);
                            break;
                        case EXIST:
                            buildExistOperator(bool, filter);
                            break;
                        case DOES_NOT_EXIST:
                            buildDoesNotExistOperator(bool, filter);
                            break;
                        case IS_BETWEEN:
                            buildIsBetweenOperator(bool, filter);
                            break;
                        case IS_NOT_BETWEEN:
                            buildIsNotBetweenOperator(bool, filter);
                            break;
                        case IS_IN_FIELDS:
                            buildIsInFields(bool, filter);
                            break;
                        case IS_NOT_IN_FIELDS:
                            buildIsNotInFields(bool, filter);
                            break;
                        case ENDS_WITH:
                            buildEndsWith(bool, filter);
                            break;
                        case NOT_ENDS_WITH:
                            buildNotEndsWith(bool, filter);
                            break;
                        case START_WITH:
                            buildStartWith(bool, filter);
                            break;
                        case NOT_START_WITH:
                            buildNotStartWith(bool, filter);
                            break;
                        case IS_ONE_OF_TERMS:
                            buildIsOneOfTermsOperator(bool, filter);
                            break;
                        case IS_GREATER_THAN:
                            buildIsGreaterThan(bool, filter);
                            break;
                        case IS_LESS_THAN_OR_EQUALS:
                            buildIsLessThanOrEquals(bool, filter);
                            break;
                    }
                }
            }
            return Query.of(q -> q.bool(bool.build()));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsOperator";
        try {
            filter.validate();
            bool.filter(f -> f.matchPhrase(m -> m.field(filter.getField()).query(String.valueOf(filter.getValue()))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsNotOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsNotOperator";
        try {
            filter.validate();
            bool.mustNot(n -> n.matchPhrase(m -> m.field(filter.getField()).query(String.valueOf(filter.getValue()))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildContainOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildContainOperator";
        try {
            filter.validate();
            bool.filter(f -> f.queryString(s -> s.fields(filter.getField())
                .query("*" + filter.getValue() + "*")
                .defaultOperator(Operator.And)
                .lenient(true)
                .type(TextQueryType.BestFields)));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildDoesNotContainOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildDoesNotContainOperator";
        try {
            filter.validate();
            bool.mustNot(n -> n.queryString(s -> s.fields(filter.getField())
                .query("*" + filter.getValue() + "*")
                .defaultOperator(Operator.And)
                .lenient(true)
                .type(TextQueryType.BestFields)));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsOneOfOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsOneOfOperator";
        try {
            filter.validate();
            BoolQuery.Builder shouldList = new BoolQuery.Builder();
            shouldList.minimumShouldMatch("1");
            for (Object val : (List<?>) filter.getValue())
                shouldList.should(f -> f.matchPhrase(m -> m.field(filter.getField()).query(String.valueOf(val))));
            bool.filter(f -> f.bool(shouldList.build()));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsNotOneOfOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsNotOneOfOperator";
        try {
            filter.validate();
            BoolQuery.Builder shouldList = new BoolQuery.Builder();
            shouldList.minimumShouldMatch("1");
            for (Object val : (List<?>) filter.getValue())
                shouldList.should(f -> f.matchPhrase(m -> m.field(filter.getField()).query(String.valueOf(val))));
            bool.mustNot(f -> f.bool(shouldList.build()));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildExistOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildExistOperator";
        try {
            filter.validate();
            bool.filter(f -> f.exists(e -> e.field(filter.getField())));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildDoesNotExistOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildDoesNotExistOperator";
        try {
            filter.validate();
            bool.mustNot(m -> m.exists(e -> e.field(filter.getField())));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsBetweenOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsBetweenOperator";
        try {
            filter.validate();
            List<?> values = (List<?>) filter.getValue();
            bool.filter(f -> f.range(r -> r.field(filter.getField())
                .gte(JsonData.of(values.get(0)))
                .lte(JsonData.of(values.get(1)))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsNotBetweenOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsNotBetweenOperator";
        try {
            filter.validate();
            List<?> values = (List<?>) filter.getValue();
            bool.mustNot(m -> m.range(r -> r.field(filter.getField())
                .gte(JsonData.of(values.get(0)))
                .lte(JsonData.of(values.get(1)))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsInFields(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsInFields";
        try {
            filter.validate();
            String value = CustomStringEscapeUtil.openSearchQueryStringEscap(String.valueOf(filter.getValue()));
            bool.filter(f -> f.queryString(q -> q
                .defaultField("*")
                .query("*" + value + "*")));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsNotInFields(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsNotInFields";
        try {
            filter.validate();
            String value = CustomStringEscapeUtil.openSearchQueryStringEscap(String.valueOf(filter.getValue()));
            bool.filter(f -> f.bool(b -> b.mustNot(n -> n.queryString(q -> q
                .defaultField("*")
                .query("*" + value + "*")))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildEndsWith(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildEndsWith";
        try {
            filter.validate();
            bool.filter(f -> f.wildcard(w -> w.field(filter.getField()).value(String.format("*%1$s", filter.getValue()))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildNotEndsWith(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildNotEndsWith";
        try {
            filter.validate();
            bool.filter(f -> f.bool(b -> b.mustNot(n -> n.wildcard(w -> w
                .field(filter.getField())
                .value(String.format("*%1$s", filter.getValue()))))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildStartWith(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildStartWith";
        try {
            filter.validate();
            bool.filter(f -> f.wildcard(w -> w.field(filter.getField()).value(String.format("%1$s*", filter.getValue()))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildNotStartWith(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildNotStartWith";
        try {
            filter.validate();
            bool.filter(f -> f.bool(b -> b.mustNot(n -> n.wildcard(w -> w
                .field(filter.getField())
                .value(String.format("%1$s*", filter.getValue()))))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsOneOfTermsOperator(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsOneOfTermsOperator";
        try {
            filter.validate();
            List<FieldValue> termsValues = ((List<?>) filter.getValue()).stream().map(str -> FieldValue.of((String) str))
                .collect(Collectors.toList());
            TermsQuery terms = TermsQuery.of(t -> t.field(filter.getField()).terms(q -> q.value(termsValues)));
            bool.filter(f -> f.terms(terms));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsGreaterThan(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsGreaterThan";
        try {
            bool.filter(f -> f.range(RangeQuery.of(r -> r.field(filter.getField())
                .gt(JsonData.of(filter.getValue())).format(Constants.INDEX_TIMESTAMP_FORMAT))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private static void buildIsLessThanOrEquals(BoolQuery.Builder bool, FilterType filter) {
        final String ctx = CLASSNAME + ".buildIsLessThanOrEquals";
        try {
            bool.filter(f -> f.range(RangeQuery.of(r -> r.field(filter.getField())
                .lte(JsonData.of(filter.getValue())).format(Constants.INDEX_TIMESTAMP_FORMAT))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Apply a sort to a elasticsearch query
     *
     * @param srb  : A {@link SearchRequest.Builder} object with the search request definition
     * @param sort : A {@link Sort} object with sorting information
     */
    public static void applySort(SearchRequest.Builder srb, Sort sort) {
        final String ctx = CLASSNAME + ".applySort";
        try {
            if (srb == null)
                return;
            if (sort == null || !sort.isSorted())
                srb.sort(s -> s.field(f -> f.field("_source").order(SortOrder.Desc)));
            else
                sort.forEach(order -> srb.sort(s -> s.field(f -> f.field(order.getProperty())
                    .order(order.isAscending() ? SortOrder.Asc : SortOrder.Desc))));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Apply pagination and sort
     *
     * @param srb      A {@link SearchRequest.Builder} object with the search request definition
     * @param pageable A {@link Pageable} object with pagination and sort information
     * @param top      A top f results to get
     */
    public static void applyPaginationAndSort(SearchRequest.Builder srb, Pageable pageable, Integer top) {
        final String ctx = CLASSNAME + ".applyPaginationAndSort";
        try {
            if (Objects.isNull(srb) || Objects.isNull(pageable))
                return;

            // Applying pagination
            if (pageable.isPaged()) {
                int first = UtilPagination.getFirstForElasticsearch(pageable.getPageSize(), pageable.getPageNumber());
                srb.from(first).size((first + pageable.getPageSize()) > top ? (top - first) : pageable.getPageSize());
            } else {
                srb.size(top);
            }

            // Applying sort
            Sort sort = pageable.getSort();
            if (sort.isSorted()) {
                sort.forEach(order -> srb.sort(s -> s.field(f -> f.field(order.getProperty())
                    .order(order.isAscending() ? SortOrder.Asc : SortOrder.Desc))));
            } else {
                srb.sort(s -> s.field(f -> f.field("_score").order(SortOrder.Desc)));
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public static FilterType fromStringFilter(StringFilter filter, String fieldName) {
        final String ctx = CLASSNAME + ".fromStringFilter";
        if (StringUtils.hasText(filter.getEquals()))
            return new FilterType(fieldName, OperatorType.IS, filter.getEquals());
        else if (StringUtils.hasText(filter.getContains()))
            return new FilterType(fieldName, OperatorType.CONTAIN, filter.getContains());
        else if (!CollectionUtils.isEmpty(filter.getIn()))
            return new FilterType(fieldName, OperatorType.IS_ONE_OF_TERMS, filter.getIn());
        else
            throw new RuntimeException(String.format("%1$s: The operator you are trying to execute for field %2$s is not implemented", ctx, fieldName));
    }

    public static FilterType fromInstantFilter(InstantFilter filter, String fieldName) {
        final String ctx = CLASSNAME + ".fromInstantFilter";
        if (!Objects.isNull(filter.getEquals()))
            return new FilterType(fieldName, OperatorType.IS, filter.getEquals());

        Instant from = ObjectUtils.firstNonNull(filter.getGreaterThan(), filter.getGreaterThanOrEqual());
        Instant to = ObjectUtils.firstNonNull(filter.getLessThan(), filter.getLessThanOrEqual(), Instant.now());

        if (Objects.isNull(from))
            throw new RuntimeException(ctx + ": We can't perform a range operation without the initial part of the range");

        return new FilterType(fieldName, OperatorType.IS_BETWEEN, Arrays.asList(from, to));
    }
}
