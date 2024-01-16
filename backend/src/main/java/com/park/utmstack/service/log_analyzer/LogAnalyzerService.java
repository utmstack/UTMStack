package com.park.utmstack.service.log_analyzer;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.log_analyzer.LogAnalyzerQuery;
import com.park.utmstack.repository.log_analyzer.LogAnalyzerQueryRepository;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.util.exceptions.UtmElasticsearchException;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch._types.aggregations.Aggregate;
import org.opensearch.client.opensearch._types.aggregations.CalendarInterval;
import org.opensearch.client.opensearch._types.aggregations.DateHistogramBucket;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
@Transactional
public class LogAnalyzerService {
    private final Logger log = LoggerFactory.getLogger(LogAnalyzerService.class);
    private static final String CLASSNAME = "LogAnalyzerService";
    private final LogAnalyzerQueryRepository logAnalyzerQueryRepository;
    private final ElasticsearchService elasticsearchService;

    public LogAnalyzerService(LogAnalyzerQueryRepository logAnalyzerQueryRepository,
                              ElasticsearchService elasticsearchService) {
        this.logAnalyzerQueryRepository = logAnalyzerQueryRepository;
        this.elasticsearchService = elasticsearchService;
    }

    /**
     * Save a logAnalyzerQuery.
     *
     * @param logAnalyzerQuery the entity to save
     * @return the persisted entity
     */
    public LogAnalyzerQuery save(LogAnalyzerQuery logAnalyzerQuery) {
        log.debug("Request to save LogAnalyzerQuery : {}", logAnalyzerQuery);
        return logAnalyzerQueryRepository.save(logAnalyzerQuery);
    }

    /**
     * Get all the logAnalyzerQueries.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<LogAnalyzerQuery> findAll(Pageable pageable) {
        log.debug("Request to get all LogAnalyzerQueries");
        return logAnalyzerQueryRepository.findAll(pageable);
    }


    /**
     * Get one logAnalyzerQuery by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<LogAnalyzerQuery> findOne(Long id) {
        log.debug("Request to get LogAnalyzerQuery : {}", id);
        return logAnalyzerQueryRepository.findById(id);
    }

    /**
     * Delete the logAnalyzerQuery by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete LogAnalyzerQuery : {}", id);
        logAnalyzerQueryRepository.deleteById(id);
    }

    public TopValuesResult topXValues(List<FilterType> filters, int top, String field, String indexPattern,
                                      Pageable pageable) {
        final String ctx = CLASSNAME + ".topXValues";
        final String AGG_NAME_TOP_VALUES = "topXvalueOf";
        final String AGG_NAME_TOTAL = "totalField";
        try {
            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.size(Constants.LOG_ANALYZER_TOTAL_RESULTS).query(SearchUtil.toQuery(filters))
                .index(indexPattern);
            SearchUtil.applySort(srb, pageable.getSort());

            TopValuesResult retValue = new TopValuesResult();
            retValue.setTotal(0);

            HitsMetadata<Object> hits = elasticsearchService.search(srb.build(), Object.class).hits();

            if (hits.total().value() <= 0)
                return retValue;

            List<String> ids = hits.hits().stream().map(Hit::id).collect(Collectors.toList());

            filters.clear();
            filters.add(new FilterType(Constants._id, OperatorType.IS_ONE_OF_TERMS, ids));

            srb = new SearchRequest.Builder().query(SearchUtil.toQuery(filters))
                .size(0).index(indexPattern)
                .aggregations(AGG_NAME_TOP_VALUES, a -> a.terms(t -> t.field(field).size(top)))
                .aggregations(AGG_NAME_TOTAL, c -> c.valueCount(v -> v.field(field)));

            Map<String, Aggregate> aggs = elasticsearchService.search(srb.build(), Object.class).aggregations();

            List<BucketAggregation> topValues = TermAggregateParser.parse(aggs.get(AGG_NAME_TOP_VALUES));

            if (CollectionUtils.isEmpty(topValues))
                return retValue;

            long total = (long) aggs.get(AGG_NAME_TOTAL).valueCount().value();

            if (total <= 0)
                return retValue;

            retValue.setTotal((int) total);

            topValues.forEach(agg -> {
                TopValues value = new TopValues();
                value.setCount(agg.getDocCount());
                value.setValue(agg.getKey());
                value.setPercent((agg.getDocCount().floatValue() / total) * 100);
                retValue.addTop(value);
            });

            return retValue;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public ChartView getChartViewData(String indexPattern, List<FilterType> filters, CalendarInterval interval,
                                      Integer top, String field, String fieldDataType)
        throws UtmElasticsearchException {
        final String ctx = CLASSNAME + ".getChartViewData";
        final String CHART_VIEW_AGG = "amountOfDocument";
        try {
            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.query(SearchUtil.toQuery(filters)).index(indexPattern).size(0);

            if (fieldDataType.equalsIgnoreCase("date")) {
                srb.aggregations(CHART_VIEW_AGG, a -> a.dateHistogram(h -> h
                    .calendarInterval(interval != null ? interval : CalendarInterval.Month)
                    .format("yyyy-MM-dd HH:mm:ss").field(field).minDocCount(1)));
            } else {
                srb.aggregations(CHART_VIEW_AGG, a -> a.terms(t -> t.field(field).size(top)
                    .order(List.of(Map.of("_count", SortOrder.Desc)))));
            }

            Aggregate agg = elasticsearchService.search(srb.build(), Object.class).aggregations().get(CHART_VIEW_AGG);
            ChartView retValue = new ChartView();

            if (fieldDataType.equalsIgnoreCase("date")) {
                List<DateHistogramBucket> buckets = agg.dateHistogram().buckets().array();
                if (CollectionUtils.isEmpty(buckets))
                    return retValue;
                buckets.forEach(bucket -> retValue.addCategory(bucket.keyAsString()).addValue(bucket.docCount()));
            } else {
                List<BucketAggregation> buckets = TermAggregateParser.parse(agg);
                if (CollectionUtils.isEmpty(buckets))
                    return retValue;
                buckets.forEach(b -> retValue.addCategory(b.getKey()).addValue(b.getDocCount()));
            }

            return retValue;
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }
    }

    public static class TopValuesResult {
        private int total;
        private List<TopValues> top = new ArrayList<>();

        public int getTotal() {
            return total;
        }

        public void setTotal(int total) {
            this.total = total;
        }

        public List<TopValues> getTop() {
            return top;
        }

        public void setTop(List<TopValues> top) {
            this.top = top;
        }

        public void addTop(TopValues top) {
            this.top.add(top);
        }
    }

    public static class TopValues {
        private String value;
        private long count;
        private float percent;

        public String getValue() {
            return value;
        }

        public void setValue(String value) {
            this.value = value;
        }

        public long getCount() {
            return count;
        }

        public void setCount(long count) {
            this.count = count;
        }

        public float getPercent() {
            return percent;
        }

        public void setPercent(float percent) {
            this.percent = percent;
        }
    }

    public static class ChartView {
        private List<String> categories = new ArrayList<>();
        private List<Long> values = new ArrayList<>();

        public List<String> getCategories() {
            return categories;
        }

        public void setCategories(List<String> categories) {
            this.categories = categories;
        }

        public ChartView addCategory(String category) {
            categories.add(category);
            return this;
        }

        public List<Long> getValues() {
            return values;
        }

        public void setValues(List<Long> values) {
            this.values = values;
        }

        public ChartView addValue(Long value) {
            values.add(value);
            return this;
        }
    }
}
