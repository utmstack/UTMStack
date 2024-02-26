package com.park.utmstack.web.rest.elasticsearch;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.shared_types.CsvExportingParams;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.util.UtilCsv;
import com.park.utmstack.util.UtilPagination;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.util.chart_builder.IndexPropertyType;
import com.park.utmstack.util.chart_builder.IndexType;
import com.park.utmstack.util.exceptions.OpenSearchIndexNotFoundException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import com.utmstack.opensearch_connector.types.ElasticCluster;
import org.opensearch.client.opensearch.cat.indices.IndicesRecord;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletResponse;
import javax.validation.Valid;
import javax.validation.constraints.NotNull;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.stream.Collectors;


@RestController
@RequestMapping("/api/elasticsearch")
public class ElasticsearchResource {

    private final Logger log = LoggerFactory.getLogger(ElasticsearchResource.class);
    private static final String CLASSNAME = "ElasticsearchResource";

    private final ElasticsearchService elasticsearchService;
    private final ApplicationEventService applicationEventService;

    public ElasticsearchResource(ElasticsearchService elasticsearchService,
                                 ApplicationEventService applicationEventService) {
        this.elasticsearchService = elasticsearchService;
        this.applicationEventService = applicationEventService;
    }

    @GetMapping("/property/values")
    public ResponseEntity<List<String>> getFieldValues(@RequestParam String keyword,
                                                       @RequestParam String indexPattern) {
        final String ctx = CLASSNAME + ".getFieldValues";
        try {
            return ResponseEntity.ok(elasticsearchService.getFieldValues(keyword, indexPattern));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @PostMapping("/property/values-with-count")
    public ResponseEntity<Map<String, Long>> getFieldValuesWithCount(@Valid @RequestBody PropertyValuesWithCountRequest rq) {
        final String ctx = CLASSNAME + ".getFieldValuesWithCount";
        try {
            return ResponseEntity.ok(elasticsearchService.getFieldValuesWithCount(rq.getField(), rq.getIndex(),
                    rq.getFilters(), rq.getTop(), rq.isOrderByCount(), rq.isSortAsc()));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * Gets all properties of an index by his pattern
     *
     * @param indexPattern: Pattern defined for an index
     * @return A list with the properties of the index passed
     */
    @GetMapping("/index/properties")
    public ResponseEntity<List<IndexPropertyType>> getIndexProperties(@RequestParam String indexPattern) {
        final String ctx = CLASSNAME + ".getIndexProperties";
        try {
            return ResponseEntity.ok(elasticsearchService.getIndexProperties(indexPattern));
        } catch (OpenSearchIndexNotFoundException e) {
            String msg = ctx + ": " + e.getMessage();
            log.info(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.INFO);
            return UtilResponse.buildNotFoundResponse(msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildInternalServerErrorResponse(msg);
        }
    }

    @GetMapping("/index/all")
    public ResponseEntity<List<IndexType>> getAllIndexes(@RequestParam(defaultValue = "false") boolean includeSystemIndex,
                                                         @RequestParam(required = false) String pattern, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllIndexes";
        try {
            Page<IndicesRecord> page = elasticsearchService.getAllIndexes(includeSystemIndex, pattern, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/index/all");
            return ResponseEntity.ok().headers(headers).body(page.getContent().stream().map(IndexType::new).collect(Collectors.toList()));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * Bulk delete for indexes
     *
     * @param indexes : List of the names pf all indexes to be removed
     * @return
     */
    @PostMapping("/index/delete-index")
    public ResponseEntity<Void> deleteIndex(@RequestBody List<String> indexes) {
        final String ctx = CLASSNAME + ".deleteIndex";
        try {
            elasticsearchService.deleteIndex(indexes);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PostMapping("/search")
    public ResponseEntity<List<Map>> search(@RequestBody(required = false) List<FilterType> filters,
                                            @RequestParam Integer top, @RequestParam String indexPattern,
                                            Pageable pageable) {
        final String ctx = CLASSNAME + ".search";
        try {
            SearchResponse<Map> searchResponse = elasticsearchService.search(filters, top, indexPattern,
                    pageable, Map.class);

            if (Objects.isNull(searchResponse) || Objects.isNull(searchResponse.hits()) || searchResponse.hits().total().value() == 0)
                return ResponseEntity.ok(Collections.emptyList());

            HitsMetadata<Map> hits = searchResponse.hits();
            HttpHeaders headers = UtilPagination.generatePaginationHttpHeaders(Math.min(hits.total().value(), top),
                    pageable.getPageNumber(), pageable.getPageSize(), "/api/elasticsearch/search");

            return ResponseEntity.ok().headers(headers).body(hits.hits().stream()
                    .map(Hit::source).collect(Collectors.toList()));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @PostMapping("/search/csv")
    public ResponseEntity<Void> searchToCsv(@RequestBody @Valid CsvExportingParams params, HttpServletResponse response) {
        final String ctx = CLASSNAME + ".searchToCsv";
        try {
            SearchResponse<Map> searchResponse = elasticsearchService.search(params.getFilters(), params.getTop(),
                    params.getIndexPattern(), Pageable.unpaged(), Map.class);

            if (Objects.isNull(searchResponse) || Objects.isNull(searchResponse.hits()) || searchResponse.hits().total().value() == 0)
                return ResponseEntity.ok().build();

            List<Map> hits = searchResponse.hits().hits().stream().map(Hit::source).collect(Collectors.toList());
            UtilCsv.prepareToDownload(response, params.getColumns(), hits);

            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @PostMapping("/generic-search")
    public ResponseEntity<List<Map>> genericSearch(@Valid @RequestBody GenericSearchBody body, Pageable pageable) {
        final String ctx = CLASSNAME + ".genericSearch";
        try {
            SearchResponse<Map> searchResponse = elasticsearchService.search(body.getFilters(), body.getTop(),
                    body.getIndex(), pageable, Map.class);

            if (Objects.isNull(searchResponse) || Objects.isNull(searchResponse.hits()) || searchResponse.hits().total().value() == 0)
                return ResponseEntity.ok().build();

            HitsMetadata<Map> hits = searchResponse.hits();

            HttpHeaders headers = UtilPagination.generatePaginationHttpHeaders(Math.min(hits.total().value(), body.getTop()),
                    pageable.getPageNumber(), pageable.getPageSize(), "/api/elasticsearch/generic-search");

            return ResponseEntity.ok().headers(headers).body(hits.hits().stream()
                    .map(Hit::source).collect(Collectors.toList()));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/cluster/status")
    public ResponseEntity<ElasticCluster> getClusterStatus() {
        final String ctx = CLASSNAME + ".getClusterStatus";
        try {
            return elasticsearchService.getClusterStatus().map(ResponseEntity::ok)
                    .orElseGet(() -> ResponseEntity.noContent().build());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    public static class PropertyValuesWithCountRequest {
        private List<FilterType> filters;
        private String field;
        private Integer top;
        @NotNull
        private String index;
        private boolean orderByCount;
        private boolean sortAsc;

        public List<FilterType> getFilters() {
            return filters;
        }

        public void setFilters(List<FilterType> filters) {
            this.filters = filters;
        }

        public String getField() {
            return field;
        }

        public void setField(String field) {
            this.field = field;
        }

        public Integer getTop() {
            return top;
        }

        public void setTop(Integer top) {
            this.top = top;
        }

        public String getIndex() {
            return index;
        }

        public void setIndex(String index) {
            this.index = index;
        }

        public boolean isOrderByCount() {
            return orderByCount;
        }

        public void setOrderByCount(boolean orderByCount) {
            this.orderByCount = orderByCount;
        }

        public boolean isSortAsc() {
            return sortAsc;
        }

        public void setSortAsc(boolean sortAsc) {
            this.sortAsc = sortAsc;
        }
    }

    public static class GenericSearchBody {
        @NotNull
        private String index;

        private List<FilterType> filters;

        @NotNull
        private Integer top;

        public String getIndex() {
            return index;
        }

        public void setIndex(String index) {
            this.index = index;
        }

        public List<FilterType> getFilters() {
            return filters;
        }

        public void setFilters(List<FilterType> filters) {
            this.filters = filters;
        }

        public Integer getTop() {
            return top;
        }

        public void setTop(Integer top) {
            this.top = top;
        }
    }
}
