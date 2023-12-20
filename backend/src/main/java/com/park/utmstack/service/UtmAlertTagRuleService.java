package com.park.utmstack.service;

import com.park.utmstack.aop.pointcut.AlertPointcut;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.UtmAlertTag;
import com.park.utmstack.domain.UtmAlertTagRule;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.repository.UtmAlertTagRuleRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.util.AlertUtil;
import com.park.utmstack.util.enums.AlertStatus;
import com.park.utmstack.util.events.RulesEvaluationEndEvent;
import com.park.utmstack.web.rest.vm.AlertTagRuleFilterVM;
import org.hibernate.jpa.TypedParameterValue;
import org.hibernate.type.BooleanType;
import org.hibernate.type.LongType;
import org.hibernate.type.StringType;
import org.opensearch.client.opensearch._types.query_dsl.Query;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmTagRule.
 */
@Service
@Transactional
public class UtmAlertTagRuleService {

    private final Logger log = LoggerFactory.getLogger(UtmAlertTagRuleService.class);
    private static final String CLASSNAME = "UtmAlertTagRuleService";

    private final UtmAlertTagRuleRepository alertTagRuleRepository;
    private final AlertUtil alertUtil;
    private final ApplicationEventPublisher publisher;
    private final ApplicationEventService eventService;
    private final AlertPointcut alertPointcut;
    private final UtmAlertTagService alertTagService;
    private final ElasticsearchService elasticsearchService;

    public UtmAlertTagRuleService(UtmAlertTagRuleRepository alertTagRuleRepository,
                                  AlertUtil alertUtil,
                                  ApplicationEventPublisher publisher,
                                  ApplicationEventService eventService,
                                  AlertPointcut alertPointcut,
                                  UtmAlertTagService alertTagService,
                                  ElasticsearchService elasticsearchService) {
        this.alertTagRuleRepository = alertTagRuleRepository;
        this.alertUtil = alertUtil;
        this.publisher = publisher;
        this.eventService = eventService;
        this.alertPointcut = alertPointcut;
        this.alertTagService = alertTagService;
        this.elasticsearchService = elasticsearchService;
    }

    /**
     * Save a utmTagRule.
     *
     * @param utmAlertTagRule the entity to save
     * @return the persisted entity
     */
    public UtmAlertTagRule save(UtmAlertTagRule utmAlertTagRule) {
        return alertTagRuleRepository.save(utmAlertTagRule);
    }

    /**
     * Get all the utmTagRules.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmAlertTagRule> findByFilter(AlertTagRuleFilterVM filters, Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".findByFilter";
        try {
            TypedParameterValue id = new TypedParameterValue(new LongType(), filters.getId());
            TypedParameterValue name = new TypedParameterValue(new StringType(), filters.getName());
            TypedParameterValue conditionField = new TypedParameterValue(new StringType(), filters.getConditionField());
            TypedParameterValue conditionValue = new TypedParameterValue(new StringType(), filters.getConditionValue());
            TypedParameterValue tagIds = CollectionUtils.isEmpty(filters.getTagIds()) ? null : new TypedParameterValue(new StringType(), StringUtils.collectionToCommaDelimitedString(filters.getTagIds()));
            TypedParameterValue isActive = new TypedParameterValue(new BooleanType(), filters.getRuleActive());
            TypedParameterValue isDeleted = new TypedParameterValue(new BooleanType(), filters.getRuleDeleted());

            return alertTagRuleRepository.findByFilter(id, name, conditionField,
                conditionValue, tagIds, isActive, isDeleted, pageable);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public List<UtmAlertTagRule> findByIdIn(List<Long> ids) throws Exception {
        final String ctx = CLASSNAME + ".findByIdIn";
        try {
            return alertTagRuleRepository.findAllByIdIn(ids);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }


    /**
     * Get one utmTagRule by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmAlertTagRule> findOne(Long id) {
        return alertTagRuleRepository.findById(id);
    }

    /**
     * Delete the utmTagRule by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        alertTagRuleRepository.deleteById(id);
    }

    @Scheduled(fixedDelay = 30000, initialDelay = 10000)
    public void automaticReview() {
        final String ctx = CLASSNAME + ".automaticReview";
        try {
            // If no new alerts have been received, stop execution
            if (alertUtil.countAlertsByStatus(AlertStatus.AUTOMATIC_REVIEW.getCode()) == 0)
                return;

            // Getting all registered rules
            List<UtmAlertTagRule> tagRules = alertTagRuleRepository.findAll();

            // Getting rules that are actives and are not deleted
            if (!CollectionUtils.isEmpty(tagRules))
                tagRules = tagRules.stream()
                    .filter(rule -> rule.getRuleActive() && !rule.getRuleDeleted())
                    .collect(Collectors.toList());

            // Getting rule evaluation start time
            Instant rulesEvaluationStart = LocalDateTime.now().toInstant(ZoneOffset.UTC)
                .truncatedTo(ChronoUnit.SECONDS);

            // If there is any rule
            if (!CollectionUtils.isEmpty(tagRules))
                applyTagRule(tagRules, rulesEvaluationStart);

            releaseToOpen(rulesEvaluationStart);
            publisher.publishEvent(new RulesEvaluationEndEvent(this));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            log.error(msg);
        }
    }

    private void releaseToOpen(Instant rulesEvaluationStart) throws Exception {
        final String ctx = CLASSNAME + ".releaseToOpen";
        try {
            String ruleScript = "ctx._source.status=%1$s;" +
                "ctx._source.statusLabel='%2$s';" +
                "ctx._source.statusObservation='%3$s';";

            String statusObservation = "This alert has been evaluated by the tag rules engine";

            String script = String.format(ruleScript, AlertStatus.OPEN.getCode(), AlertStatus.OPEN.getName(),
                statusObservation);

            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertStatus, OperatorType.IS, AlertStatus.AUTOMATIC_REVIEW.getCode()));
            filters.add(new FilterType(Constants.timestamp, OperatorType.IS_LESS_THAN_OR_EQUALS, rulesEvaluationStart.toString()));

            Query query = SearchUtil.toQuery(filters);
            String indexPattern = Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS);

            elasticsearchService.updateByQuery(query, indexPattern, script);

            alertPointcut.automaticAlertStatusChangePointcut(query, AlertStatus.OPEN.getCode(),
                statusObservation, indexPattern);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private void applyTagRule(List<UtmAlertTagRule> rules, Instant rulesEvaluationStart) throws Exception {
        final String ctx = CLASSNAME + ".applyTagRule";
        try {
            for (UtmAlertTagRule rule : rules) {
                List<FilterType> filters = rule.getConditions();
                filters.add(new FilterType(Constants.alertStatus, OperatorType.IS, AlertStatus.AUTOMATIC_REVIEW.getCode()));
                filters.add(new FilterType(Constants.timestamp, OperatorType.IS_LESS_THAN_OR_EQUALS, rulesEvaluationStart.toString()));

                Query query = SearchUtil.toQuery(filters);

                String script = "if (!ctx._source.containsKey('tags') || ctx._source.tags == null || ctx._source.tags.empty)\n" +
                    "\tctx._source.tags = new ArrayList();\n" +
                    "ctx._source.tags.addAll([%1$s]);\n" +
                    "ctx._source.tags = ctx._source.tags.stream().distinct().collect(Collectors.toList());\n" +
                    "\n" +
                    "if (!ctx._source.containsKey('TagRulesApplied') || ctx._source.TagRulesApplied == null || ctx._source.TagRulesApplied.empty)\n" +
                    "\tctx._source.TagRulesApplied = new ArrayList();\n" +
                    "ctx._source.TagRulesApplied.add(%2$s);\n" +
                    "ctx._source.TagRulesApplied = ctx._source.TagRulesApplied.stream().distinct().collect(Collectors.toList());" +
                    "\n" +
                    "if (ctx._source.tags.contains('False positive')) {\n" +
                    "\tctx._source.status=%3$s;\n" +
                    "\tctx._source.statusLabel='%4$s';\n" +
                    "\tctx._source.statusObservation='Status changed to completed because alert was tagged as False positive';\n" +
                    "\n}";

                List<UtmAlertTag> tags = alertTagService.findAllByIdIn(rule.getAppliedTagsAsListOfLong());
                String indexPattern = Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS);

                String tagsForInsert = tags.stream().map(tag -> "'" + tag.getTagName() + "'").collect(Collectors.joining(","));

                elasticsearchService.updateByQuery(query, indexPattern, String.format(script, tagsForInsert, rule.getId(),
                    AlertStatus.COMPLETED.getCode(), AlertStatus.COMPLETED.getName()));

                String tagsForLogs = tags.stream().map(UtmAlertTag::getTagName).collect(Collectors.joining(","));
                alertPointcut.automaticAlertTagsChangePointcut(query, tagsForLogs, rule.getRuleName(), indexPattern);
            }
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
