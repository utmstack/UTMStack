package com.park.utmstack.service.elasticsearch;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.UtmSpaceNotificationControl;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.UtmSpaceNotificationControlService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.chart_builder.IndexPropertyType;
import com.park.utmstack.util.exceptions.OpenSearchIndexNotFoundException;
import com.park.utmstack.util.exceptions.UtmElasticsearchException;
import com.utmstack.opensearch_connector.enums.IndexSortableProperty;
import com.utmstack.opensearch_connector.enums.TermOrder;
import com.utmstack.opensearch_connector.exceptions.OpenSearchException;
import com.utmstack.opensearch_connector.types.ElasticCluster;
import com.utmstack.opensearch_connector.types.IndexSort;
import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch._types.query_dsl.Query;
import org.opensearch.client.opensearch.cat.indices.IndicesRecord;
import org.opensearch.client.opensearch.core.IndexResponse;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.support.PagedListHolder;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.data.support.PageableExecutionUtils;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.time.temporal.ChronoUnit;
import java.util.*;
import java.util.stream.Collectors;

/**
 * @author Leonardo M. López
 */
@Service
public class ElasticsearchService {
    private static final String CLASSNAME = "ElasticsearchService";
    private final Logger log = LoggerFactory.getLogger(ElasticsearchService.class);
    private final ApplicationEventService eventService;
    private final UserRepository userRepository;
    private final MailService mailService;
    private final UtmSpaceNotificationControlService spaceNotificationControlService;
    private final OpensearchClientBuilder client;

    public ElasticsearchService(ApplicationEventService eventService, UserRepository userRepository,
                                MailService mailService,
                                UtmSpaceNotificationControlService spaceNotificationControlService,
                                OpensearchClientBuilder client) {
        this.eventService = eventService;
        this.userRepository = userRepository;
        this.mailService = mailService;
        this.spaceNotificationControlService = spaceNotificationControlService;
        this.client = client;
    }

    /**
     * Gets all values from an index keyword field
     *
     * @param keyword:      Keyword field name
     * @param indexPattern: Index pattern
     * @return List of field value
     */
    public List<String> getFieldValues(String keyword, String indexPattern) {
        final String ctx = CLASSNAME + ".getFieldValues";
        try {
            return new ArrayList<>(client.getClient().getFieldValues(keyword, indexPattern,
                null, 10000, TermOrder.Count, SortOrder.Desc).keySet());
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Gets all values for a field and count the documents for each value
     *
     * @param filters : Filters to apply
     * @param field   : Field to get values
     * @param top     : Top of result to get as result
     * @param index   : Index to get the field values
     * @return A map with field value as key and amount of documents as value
     */
    public Map<String, Long> getFieldValuesWithCount(String field, String index, List<FilterType> filters, Integer top,
                                                     boolean orderByCount, boolean sortAsc) {
        final String ctx = CLASSNAME + ".getFieldValuesWithCount";
        try {
            return client.getClient().getFieldValues(field, index, SearchUtil.toQuery(filters), top,
                orderByCount ? TermOrder.Count : TermOrder.Key, sortAsc ? SortOrder.Asc : SortOrder.Desc);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Check if some index exist
     *
     * @param index Index where the indexing will be performed, you can use a pattern too
     * @return True if index exist, false otherwise
     */
    public boolean indexExist(String index) {
        final String ctx = CLASSNAME + ".indexExist";
        try {
            return client.getClient().indexExist(index);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return false;
        }
    }

    public <T> IndexResponse index(String index, T document) {
        final String ctx = CLASSNAME + ".index";
        try {
            return client.getClient().index(index, document);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Gets all fields of an index
     *
     * @param indexPattern: Index pattern for get fields
     * @return A list of IndexProperty with a name and type of field
     */
    public List<IndexPropertyType> getIndexProperties(String indexPattern) {
        final String ctx = CLASSNAME + ".getIndexProperties";

        if (!indexExist(indexPattern))
            throw new OpenSearchIndexNotFoundException(ctx  + ": Index [" + indexPattern + "] not found");

        try {
            Map<String, String> properties = client.getClient().getIndexProperties(indexPattern);
            if (CollectionUtils.isEmpty(properties))
                return Collections.emptyList();
            return properties.entrySet()
                .stream().map(e -> new IndexPropertyType(e.getKey(), e.getValue())).collect(Collectors.toList());
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Make a query to elasticsearch to get all indexes. Depending on includeSystemIndex param it includes in the result the
     * elasticsearch system indexes
     *
     * @param includeSystemIndex: Decide if include elasticsearch system indexes to the result
     * @param pattern:            Just return indexes that his name match with pattern
     * @return A list of IndexType object.
     * @throws UtmElasticsearchException In case of any error
     */
    public Page<IndicesRecord> getAllIndexes(boolean includeSystemIndex, String pattern, Pageable pageable) throws
        UtmElasticsearchException {
        final String ctx = CLASSNAME + ".getAllIndexes";
        try {
            Assert.notNull(pageable, "Argument pageable can't be null");

            List<IndicesRecord> indices = client.getClient().getIndices(pattern, from(pageable.getSort()));

            if (CollectionUtils.isEmpty(indices))
                return PageableExecutionUtils.getPage(indices, pageable, indices::size);

            if (!includeSystemIndex)
                indices = indices.stream().filter(index -> !index.index().startsWith("."))
                    .collect(Collectors.toList());

            PagedListHolder<IndicesRecord> pageDefinition = new PagedListHolder<>();
            pageDefinition.setSource(indices);
            pageDefinition.setPageSize(pageable.getPageSize());
            pageDefinition.setPage(pageable.getPageNumber());
            return PageableExecutionUtils.getPage(pageDefinition.getPageList(), pageable, indices::size);
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }
    }

    private IndexSort from(Sort sort) {
        final String ctx = CLASSNAME + ".from";
        try {
            if (Objects.isNull(sort) || sort.isUnsorted())
                return IndexSort.unSorted();
            IndexSort.Builder sortBuilder = IndexSort.builder();
            sort.forEach(order -> sortBuilder.with(IndexSortableProperty.fromJsonValue(order.getProperty()),
                order.getDirection().isAscending() ? SortOrder.Asc : SortOrder.Desc));
            return sortBuilder.build();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    public Optional<ElasticCluster> getClusterStatus() throws UtmElasticsearchException {
        final String ctx = CLASSNAME + ".getClusterStatus";
        try {
            return client.getClient().getClusterNodesInfo();
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }
    }

    @Scheduled(fixedDelay = 60000, initialDelay = 60000)
    public void preventSystemCrashBySpace() {
        final String ctx = CLASSNAME + ".preventSystemCrashBySpace";

        try {
            Optional<ElasticCluster> opt = getClusterStatus();

            if (opt.isEmpty())
                return;

            ElasticCluster clusterStatus = opt.get();

            float diskPercent = clusterStatus.getResume().getDiskUsedPercent();

            if (diskPercent < 70)
                return;

            if (diskPercent >= 85) {
                deleteOldestIndices();
            } else if (diskPercent >= 70) {
                List<User> admins = userRepository.findAllAdmins();
                if (CollectionUtils.isEmpty(admins))
                    return;

                UtmSpaceNotificationControl notificationControl = spaceNotificationControlService.findById(1L)
                    .orElse(new UtmSpaceNotificationControl());
                if (Objects.isNull(notificationControl.getId()))
                    notificationControl.setId(1L);

                Instant now = LocalDateTime.now().toInstant(ZoneOffset.UTC);

                if (Objects.isNull(notificationControl.getNextNotification()) ||
                    now.isAfter(notificationControl.getNextNotification())) {
                    mailService.sendLowSpaceEmail(admins, clusterStatus);
                    notificationControl.setNextNotification(now.plus(24, ChronoUnit.HOURS));
                    spaceNotificationControlService.save(notificationControl);
                }
            }
        } catch (Exception e) {
            String msg = String.format("%1$s: %2$s", ctx, e.getMessage());
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    /**
     *
     */
    private void deleteOldestIndices() {
        final String ctx = CLASSNAME + ".deleteOldestIndices";
        try {
            List<IndicesRecord> indices = client.getClient().getIndices(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS), IndexSort.builder()
                .with(IndexSortableProperty.CreationDate, SortOrder.Asc).build());

            // If no index that match with log-* was found then te function is terminated
            if (CollectionUtils.isEmpty(indices))
                return;

            // Indices are returned from oldest to newest ordered by creation.date asc
            for (IndicesRecord index : indices) {
                try {
                    // Delete oldest indices
                    deleteIndex(Collections.singletonList(index.index()));
                    eventService.createEvent(String.format("Index %1$s was deleted to avoid system crash by space:\n" +
                            "Creation Date: %2$s\n" +
                            "Docs Count: %3$s\n" +
                            "Size: %4$s",
                        index.index(), index.creationDateString(), index.docsCount(), index.storeSize()), ApplicationEventType.INFO);
                } catch (Exception e) {
                    String msg = String.format("%1$s: Fail to delete index: %2$s with message: %3$s", ctx, index.index(), e.getMessage());
                    eventService.createEvent(msg, ApplicationEventType.WARNING);
                }

                Optional<ElasticCluster> opt = getClusterStatus();

                if (opt.isEmpty() || opt.get().getResume().getDiskUsedPercent() < 70)
                    break;
            }
        } catch (Exception e) {
            String msg = String.format("%1$s: %2$s", ctx, e.getMessage());
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    /**
     * Bulk delete for indexes
     *
     * @param indices : List of the names pf all indexes to be removed
     * @throws Exception In case of any error
     */
    public void deleteIndex(List<String> indices) throws Exception {
        final String ctx = CLASSNAME + ".deleteIndex";
        try {
            if (CollectionUtils.isEmpty(indices))
                return;
            client.getClient().deleteIndex(indices);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
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

    public <T> SearchResponse<T> search(SearchRequest request, Class<T> type) {
        final String ctx = CLASSNAME + ".search";
        try {
            return client.getClient().search(request, type);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public void updateByQuery(Query query, String index, String script) {
        final String ctx = CLASSNAME + ".updateByQuery";
        try {
            client.getClient().updateByQuery(query, index, script);
        } catch (OpenSearchException e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Build a query based on filters provided
     *
     * @param filters : Filters to apply
     * @return A SearchSourceBuilder with the query to execute
     */
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
