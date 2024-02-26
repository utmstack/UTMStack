package com.park.utmstack.service.impl;

import com.park.utmstack.aop.pointcut.AlertPointcut;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.UtmAlertLast;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.domain.shared_types.AlertType;
import com.park.utmstack.domain.shared_types.LogType;
import com.park.utmstack.domain.shared_types.static_dashboard.CardType;
import com.park.utmstack.repository.UtmAlertLastRepository;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.UtmAlertService;
import com.park.utmstack.service.alert_response_rule.UtmAlertResponseRuleService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.application_modules.UtmModuleService;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.service.soc_ai.SocAIService;
import com.park.utmstack.util.enums.AlertStatus;
import com.park.utmstack.util.events.RulesEvaluationEndEvent;
import com.park.utmstack.util.exceptions.DashboardOverviewException;
import com.park.utmstack.util.exceptions.ElasticsearchIndexDocumentUpdateException;
import com.park.utmstack.util.exceptions.UtmElasticsearchException;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import org.apache.commons.text.StringEscapeUtils;
import org.opensearch.client.opensearch._types.query_dsl.Query;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.event.EventListener;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.time.Instant;
import java.time.ZoneOffset;
import java.time.temporal.ChronoUnit;
import java.util.*;
import java.util.stream.Collectors;

@Service
@Transactional
public class UtmAlertServiceImpl implements UtmAlertService {

    private final Logger log = LoggerFactory.getLogger(UtmAlertServiceImpl.class);
    private static final String CLASS_NAME = "UtmAlertServiceImpl";

    private final MailService mailService;
    private final ApplicationEventService eventService;
    private final AlertPointcut alertPointcut;
    private final UtmAlertLastRepository lastAlertRepository;
    private final ElasticsearchService elasticsearchService;
    private final UtmModuleService moduleService;
    private final SocAIService socAIService;
    private final UtmAlertResponseRuleService alertResponseRuleService;

    public UtmAlertServiceImpl(MailService mailService,
                               ApplicationEventService eventService,
                               AlertPointcut alertPointcut,
                               UtmAlertLastRepository lastAlertRepository,
                               ElasticsearchService elasticsearchService,
                               UtmModuleService moduleService, SocAIService socAIService,
                               UtmAlertResponseRuleService alertResponseRuleService) {
        this.mailService = mailService;
        this.eventService = eventService;
        this.alertPointcut = alertPointcut;
        this.lastAlertRepository = lastAlertRepository;
        this.elasticsearchService = elasticsearchService;
        this.moduleService = moduleService;
        this.socAIService = socAIService;
        this.alertResponseRuleService = alertResponseRuleService;
    }

    @EventListener(RulesEvaluationEndEvent.class)
    public void checkForNewAlerts() {
        final String ctx = CLASS_NAME + ".checkForNewAlerts";
        try {
            UtmAlertLast initialDate = lastAlertRepository.findById(1L)
                    .orElse(new UtmAlertLast(Instant.now().atZone(ZoneOffset.UTC).toInstant().minus(1, ChronoUnit.HOURS)));

            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertStatus, OperatorType.IS, AlertStatus.OPEN.getCode()));
            filters.add(new FilterType(Constants.timestamp, OperatorType.IS_GREATER_THAN, initialDate.getLastAlertTimestamp().toString()));

            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.query(SearchUtil.toQuery(filters))
                    .index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS))
                    .size(1000);
            SearchUtil.applySort(srb, Sort.by(Sort.Order.asc(Constants.timestamp)));

            SearchRequest build = srb.build();

            HitsMetadata<AlertType> hitsMetadata = elasticsearchService.search(build, AlertType.class).hits();

            if (hitsMetadata.total().value() <= 0)
                return;

            List<AlertType> alerts = hitsMetadata.hits().stream().map(Hit::source).collect(Collectors.toList());

            if (CollectionUtils.isEmpty(alerts))
                return;

            initialDate.setLastAlertTimestamp(alerts.get(alerts.size() - 1).getTimestampAsInstant());
            lastAlertRepository.save(initialDate);

            for (AlertType alert : alerts) {
                List<LogType> relatedLogs;
                try {
                    relatedLogs = getRelatedAlerts(alert.getLogs());
                } catch (Exception e) {
                    log.error(ctx + ": " + e.getMessage());
                    continue;
                }

                String emails = Constants.CFG.get(Constants.PROP_ALERT_ADDRESS_TO_NOTIFY_ALERTS);

                if (!StringUtils.hasText(emails)) {
                    String msg = String.format("%1$s: Could not sent email notification for alert with id: %2$s and name %3$s because there is no addresses configured", ctx, alert.getId(), alert.getName());
                    log.warn(msg);
                    eventService.createEvent(msg, ApplicationEventType.WARNING);
                    continue;
                }

                String[] addressToNotify = emails.replace(" ", "").split(",");
                mailService.sendAlertEmail(Arrays.asList(addressToNotify), alert, relatedLogs);
            }

            alertResponseRuleService.evaluateRules(alerts);

            if (moduleService.isModuleActive(ModuleName.SOC_AI))
                socAIService.requestSocAiProcess(alerts.stream().map(AlertType::getId).collect(Collectors.toList()));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    private List<LogType> getRelatedAlerts(List<String> logs) throws UtmElasticsearchException {
        final String ctx = CLASS_NAME + ".getRelatedAlerts";

        try {
            if (CollectionUtils.isEmpty(logs))
                return Collections.emptyList();

            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants._id, OperatorType.IS_ONE_OF_TERMS, logs));
            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.query(SearchUtil.toQuery(filters)).size(100).index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS));
            SearchUtil.applySort(srb, Sort.by(Sort.Order.desc(Constants.timestamp)));

            HitsMetadata<LogType> hits = elasticsearchService.search(srb.build(), LogType.class).hits();

            if (hits.total().value() <= 0)
                return Collections.emptyList();

            return hits.hits().stream().map(Hit::source).collect(Collectors.toList());
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }
    }

    @Override
    public void updateStatus(List<String> alertIds, int status, String statusObservation) throws
            ElasticsearchIndexDocumentUpdateException {
        final String ctx = CLASS_NAME + ".updateStatus";
        try {
            String ruleScript = "ctx._source.status=%1$s;" +
                    "ctx._source.statusLabel='%2$s';" +
                    "ctx._source.statusObservation=\"%3$s\";";

            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertIdKeyword, OperatorType.IS_ONE_OF_TERMS, alertIds));

            String script = String.format(ruleScript, status,
                    AlertStatus.getByCode(status).getName(), StringEscapeUtils.escapeJava(statusObservation));

            elasticsearchService.updateByQuery(SearchUtil.toQuery(filters),
                    Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS), script);
        } catch (Exception e) {
            throw new ElasticsearchIndexDocumentUpdateException(ctx + ": " + e.getMessage());
        }
    }

    @Override
    public void updateTags(List<String> alertIds, List<String> tags, Boolean createRule) throws
            ElasticsearchIndexDocumentUpdateException {
        final String ctx = CLASS_NAME + ".updateTags";
        try {
            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertIdKeyword, OperatorType.IS_ONE_OF_TERMS, alertIds));

            String tagScript = "ctx._source.tags=%1$s;";
            String script = String.format(tagScript, (Object) null);
            if (!CollectionUtils.isEmpty(tags))
                script = String.format(tagScript, "['" + String.join("','", tags) + "']");

            elasticsearchService.updateByQuery(SearchUtil.toQuery(filters), Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS),
                    script);
        } catch (Exception e) {
            throw new ElasticsearchIndexDocumentUpdateException(ctx + ": " + e.getMessage());
        }
    }

    public List<CardType> countAlertsByStatus(List<FilterType> filters) throws DashboardOverviewException {
        final String ctx = CLASS_NAME + ".countAlertsByStatus";
        final String AGG_NAME = "alert_today_and_last_week";
        try {
            Map<Integer, CardType> alertStatus = new HashMap<>();
            alertStatus.put(2, new CardType(AlertStatus.OPEN.getName(), 0));
            alertStatus.put(3, new CardType(AlertStatus.IN_REVIEW.getName(), 0));
            alertStatus.put(5, new CardType(AlertStatus.COMPLETED.getName(), 0));

            if (!elasticsearchService.indexExist(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS)))
                return new ArrayList<>(alertStatus.values());

            SearchRequest srb = SearchRequest.of(r -> r.size(0)
                    .index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS))
                    .query(SearchUtil.toQuery(filters))
                    .aggregations(AGG_NAME, a -> a.terms(t -> t.field(Constants.alertStatus).size(5))));

            SearchResponse<Object> response = elasticsearchService.search(srb, Object.class);

            List<BucketAggregation> parse = TermAggregateParser.parse(response.aggregations().get(AGG_NAME));

            if (CollectionUtils.isEmpty(parse))
                return new ArrayList<>(alertStatus.values());

            parse.forEach(agg -> alertStatus.get(Integer.parseInt(agg.getKey())).setValue(agg.getDocCount().intValue()));
            return new ArrayList<>(alertStatus.values());
        } catch (Exception e) {
            throw new DashboardOverviewException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Update field notes of a alert
     *
     * @param alertId : Identifier of alert to update
     * @param notes   : New value for the field notes
     */
    @Override
    public void updateNotes(String alertId, String notes) throws ElasticsearchIndexDocumentUpdateException {
        final String ctx = CLASS_NAME + ".updateNotes";
        try {
            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertIdKeyword, OperatorType.IS, alertId));

            String script = String.format("ctx._source.notes=\"%1$s\";", StringEscapeUtils.escapeJava(notes));

            elasticsearchService.updateByQuery(SearchUtil.toQuery(filters),
                    Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS), script);
        } catch (Exception e) {
            throw new ElasticsearchIndexDocumentUpdateException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * This method set field isIncident to true to the events that are passed
     *
     * @param alertIds     : List of alert ids to mark as incident
     * @param incidentName : Name of the incident
     * @param incidentId   : Incident id
     * @throws ElasticsearchIndexDocumentUpdateException In case of any error
     */
    public void convertToIncident(List<String> alertIds, String incidentName, Integer incidentId, String incidentSource) throws ElasticsearchIndexDocumentUpdateException {
        final String ctx = CLASS_NAME + ".convertToIncident";
        try {
            Assert.notEmpty(alertIds, "Parameter eventIds is null or empty");
            Assert.hasText(incidentName, "Parameter incidentObservation is null or empty");
            Assert.notNull(incidentId, "Parameter incidentId is null");

            Instant incidentCreationDate = Instant.now();
            String incidentCreatedBy = SecurityUtils.getCurrentUserLogin().orElse("system");

            String script = String.format("ctx._source.isIncident=true;" +
                            "if(ctx._source.incidentDetail == null) {" +
                            "ctx._source.incidentDetail = new HashMap();}" +
                            "ctx._source.incidentDetail.incidentName=\"%1$s\";" +
                            "ctx._source.incidentDetail.incidentId=\"%2$s\";" +
                            "ctx._source.incidentDetail.creationDate=\"%3$s\";" +
                            "ctx._source.incidentDetail.createdBy=\"%4$s\";" +
                            "ctx._source.incidentDetail.source=\"%5$s\";",
                    incidentName, incidentId, incidentCreationDate, incidentCreatedBy, incidentSource);

            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertIdKeyword, OperatorType.IS_ONE_OF_TERMS, alertIds));
            filters.add(new FilterType(Constants.alertIsIncident, OperatorType.IS, false));

            String indexPattern = Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS);
            Query query = SearchUtil.toQuery(filters);
            elasticsearchService.updateByQuery(query, indexPattern, script);
            alertPointcut.convertToIncidentPointcut(query, incidentName, incidentId, incidentCreationDate, incidentCreatedBy, incidentSource, indexPattern);
        } catch (Exception e) {
            throw new ElasticsearchIndexDocumentUpdateException(ctx + ": " + e.getMessage());
        }
    }

    public List<AlertType> getAlertsByIds(List<String> alertIds) throws UtmElasticsearchException {
        final String ctx = CLASS_NAME + ".getAlertsByIds";
        try {
            if (CollectionUtils.isEmpty(alertIds))
                return new ArrayList<>();

            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertIdKeyword, OperatorType.IS_ONE_OF_TERMS, alertIds));

            Query query = SearchUtil.toQuery(filters);
            SearchRequest request = SearchRequest.of(r -> r.query(query)
                    .index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS))
                    .size(Constants.LOG_ANALYZER_TOTAL_RESULTS));

            HitsMetadata<AlertType> hits = elasticsearchService.search(request, AlertType.class).hits();

            if (hits.total().value() <= 0)
                return new ArrayList<>();

            List<AlertType> alerts = hits.hits().stream().map(Hit::source).collect(Collectors.toList());

            if (CollectionUtils.isEmpty(alerts))
                return new ArrayList<>();

            return alerts;
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }
    }
}
