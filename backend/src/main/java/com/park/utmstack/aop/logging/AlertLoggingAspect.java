package com.park.utmstack.aop.logging;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.UtmAlertLog;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.domain.shared_types.AlertType;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.UtmAlertLogService;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.enums.AlertActionEnum;
import com.park.utmstack.util.enums.AlertStatus;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Pointcut;
import org.opensearch.client.opensearch._types.query_dsl.Query;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.*;
import java.util.stream.Collectors;

@Aspect
@Component
public class AlertLoggingAspect {
    private static final String CLASS_NAME = "AlertLoggingAspect";
    private final Logger log = LoggerFactory.getLogger(AlertLoggingAspect.class);
    private final ElasticsearchService elasticsearchService;
    private final UtmAlertLogService alertLogService;

    public AlertLoggingAspect(ElasticsearchService elasticsearchService,
                              UtmAlertLogService alertLogService) {
        this.elasticsearchService = elasticsearchService;
        this.alertLogService = alertLogService;
    }

    @Pointcut("execution(void com.park.utmstack.service.impl.UtmAlertServiceImpl.updateStatus(..))")
    public void manualAlertStatusChangePointcut() {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }

    @Pointcut("execution(void com.park.utmstack.service.impl.UtmAlertServiceImpl.updateTags(..))")
    public void manualAlertTagsChangePointcut() {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }

    @Pointcut("execution(void com.park.utmstack.service.impl.UtmAlertServiceImpl.updateNotes(..))")
    public void manualAlertNotesChangePointcut() {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }

    @Pointcut("execution(void com.park.utmstack.aop.pointcut.AlertPointcut.automaticAlertStatusChangePointcut(..))")
    public void automaticAlertStatusChangePointcut() {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }

    @Pointcut("execution(void com.park.utmstack.aop.pointcut.AlertPointcut.automaticAlertTagsChangePointcut(..))")
    public void automaticAlertTagsChangePointcut() {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }

    @Pointcut("execution(void com.park.utmstack.aop.pointcut.AlertPointcut.convertToIncidentPointcut(..))")
    public void convertToIncidentPointcut() {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }

    @Around("manualAlertStatusChangePointcut()")
    public Object logManualAlertStatusChange(ProceedingJoinPoint joinPoint) throws Throwable {
        final String ctx = CLASS_NAME + ".logManualAlertStatusChange";

        try {
            Object[] args = joinPoint.getArgs();
            List<?> alertIds = (List<?>) args[0];
            Integer status = (Integer) args[1];
            String statusObservation = (String) args[2];

            List<AlertType> alerts = getAlerts(alertIds);

            joinPoint.proceed();

            if (CollectionUtils.isEmpty(alerts))
                return null;

            for (AlertType alert : alerts) {
                UtmAlertLog alertLog = new UtmAlertLog();
                alertLog.setAlertId(alert.getId());
                alertLog.setLogUser(SecurityUtils.getCurrentUserLogin().orElse("system"));
                alertLog.setLogDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                alertLog.setLogAction(AlertActionEnum.UPDATE_STATUS.name());
                alertLog.setLogOldValue(UtilSerializer.jsonSerialize(alert));

                Map<String, Object> newValue = new LinkedHashMap<>();
                newValue.put("status", status);
                newValue.put("statusLabel", AlertStatus.getByCode(status).getName());
                newValue.put("statusObservation", statusObservation);

                alertLog.setLogNewValue(UtilSerializer.jsonSerialize(newValue));

                String basicTemplate = "The user %1$s changed alert status from %2$s to %3$s";
                String templateWithObservation = "The user %1$s changed alert status from %2$s to %3$s with the observation: %4$s";

                if (!StringUtils.hasText(statusObservation))
                    alertLog.setLogMessage(String.format(basicTemplate, alertLog.getLogUser(), alert.getStatusLabel().getName(),
                        AlertStatus.getByCode(status).getName()));
                else
                    alertLog.setLogMessage(String.format(templateWithObservation, alertLog.getLogUser(),
                        alert.getStatusLabel().getName(), AlertStatus.getByCode(status).getName(), statusObservation));

                alertLogService.save(alertLog);
            }
            return null;
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage(), e);
            joinPoint.proceed();
            return null;
        }
    }

    @Around("automaticAlertStatusChangePointcut()")
    public Object logAutomaticAlertStatusChange(ProceedingJoinPoint joinPoint) throws Throwable {
        final String ctx = CLASS_NAME + ".logAutomaticAlertStatusChange";
        try {
            Object[] args = joinPoint.getArgs();
            Query query = (Query) args[0];
            Integer status = (Integer) args[1];
            String statusObservation = (String) args[2];
            String indexPattern = (String) args[3];

            SearchRequest.Builder srb = new SearchRequest.Builder();
            srb.query(query).size(Constants.LOG_ANALYZER_TOTAL_RESULTS).index(indexPattern);

            SearchRequest sr = SearchRequest.of(s -> s.index(indexPattern)
                .size(Constants.LOG_ANALYZER_TOTAL_RESULTS)
                .query(query));

            HitsMetadata<AlertType> response = elasticsearchService.search(sr, AlertType.class).hits();

            joinPoint.proceed();

            if (response.total().value() <= 0)
                return null;

            List<AlertType> alerts = response.hits().stream().map(Hit::source).collect(Collectors.toList());

            if (CollectionUtils.isEmpty(alerts))
                return null;

            for (AlertType alert : alerts) {
                UtmAlertLog alertLog = new UtmAlertLog();
                alertLog.setAlertId(alert.getId());
                alertLog.setLogUser("system");
                alertLog.setLogDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                alertLog.setLogAction(AlertActionEnum.UPDATE_STATUS.name());
                alertLog.setLogOldValue(UtilSerializer.jsonSerialize(alert));

                Map<String, Object> newValue = new LinkedHashMap<>();
                newValue.put("status", status);
                newValue.put("statusLabel", AlertStatus.getByCode(status).getName());
                newValue.put("statusObservation", statusObservation);

                alertLog.setLogNewValue(UtilSerializer.jsonSerialize(newValue));

                String msgTemplate = "The system changed alert status from %1$s to %2$s with the observation: %3$s";

                alertLog.setLogMessage(String.format(msgTemplate, alert.getStatusLabel().getName(),
                    AlertStatus.getByCode(status).getName(), statusObservation));

                alertLogService.save(alertLog);
            }
            return null;
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage(), e);
            joinPoint.proceed();
            return null;
        }
    }

    @Around("manualAlertTagsChangePointcut()")
    public Object logManualAlertTagsChange(ProceedingJoinPoint joinPoint) throws Throwable {
        final String ctx = CLASS_NAME + ".logManualAlertTagsChange";

        try {
            Object[] args = joinPoint.getArgs();
            List alertIds = (List) args[0];
            List tags = (List) args[1];

            List<AlertType> alerts = getAlerts(alertIds);

            joinPoint.proceed();

            if (CollectionUtils.isEmpty(alerts))
                return null;

            String user = SecurityUtils.getCurrentUserLogin().orElse("system");

            for (AlertType alert : alerts) {
                UtmAlertLog alertLog = new UtmAlertLog();
                alertLog.setAlertId(alert.getId());
                alertLog.setLogUser(user);
                alertLog.setLogDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                alertLog.setLogAction(AlertActionEnum.UPDATE_TAGS.name());
                alertLog.setLogOldValue(UtilSerializer.jsonSerialize(alert));

                Map<String, Object> newValue = new LinkedHashMap<>();
                if (!CollectionUtils.isEmpty(tags)) {
                    String tagsToString = String.join(",", tags);
                    newValue.put("tags", tagsToString);
                    alertLog.setLogMessage(String.format("The user %1$s tagged the alert with: %2$s", user, tagsToString));
                } else {
                    newValue.put("tags", "");
                    alertLog.setLogMessage(String.format("The %1$s user removed all tags of the alert", user));
                }
                alertLog.setLogNewValue(UtilSerializer.jsonSerialize(newValue));
                alertLogService.save(alertLog);
            }
            return null;
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage(), e);
            joinPoint.proceed();
            return null;
        }
    }

    @Around("automaticAlertTagsChangePointcut()")
    public Object logAutomaticAlertTagsChange(ProceedingJoinPoint joinPoint) throws Throwable {
        final String ctx = CLASS_NAME + ".logAutomaticAlertTagsChange";
        try {
            Object[] args = joinPoint.getArgs();
            Query query = (Query) args[0];
            String tags = (String) args[1];
            String ruleName = (String) args[2];
            String indexPattern = (String) args[3];

            SearchRequest request = SearchRequest.of(s -> s.query(query)
                .size(Constants.LOG_ANALYZER_TOTAL_RESULTS).index(indexPattern));

            HitsMetadata<AlertType> hits = elasticsearchService.search(request, AlertType.class).hits();

            joinPoint.proceed();

            if (hits.total().value() <= 0)
                return null;

            List<AlertType> alerts = hits.hits().stream().map(Hit::source).collect(Collectors.toList());

            if (CollectionUtils.isEmpty(alerts))
                return null;

            String user = SecurityUtils.getCurrentUserLogin().orElse("system");

            for (AlertType alert : alerts) {
                UtmAlertLog alertLog = new UtmAlertLog();
                alertLog.setAlertId(alert.getId());
                alertLog.setLogUser(user);
                alertLog.setLogDate(Instant.now());
                alertLog.setLogAction(AlertActionEnum.UPDATE_TAGS.name());
                alertLog.setLogOldValue(UtilSerializer.jsonSerialize(alert));

                Map<String, Object> newValue = new LinkedHashMap<>();
                newValue.put("tags", tags);
                alertLog.setLogNewValue(UtilSerializer.jsonSerialize(newValue));
                alertLog.setLogMessage(String.format(
                    "The system tagged the alert automatically with: %2$s due a match with rule: %3$s", user, tags, ruleName));

                alertLogService.save(alertLog);
            }
            return null;
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage(), e);
            joinPoint.proceed();
            return null;
        }
    }

    @Around("manualAlertNotesChangePointcut()")
    public Object logManualAlertNotesChange(ProceedingJoinPoint joinPoint) throws Throwable {
        final String ctx = CLASS_NAME + ".logManualAlertNotesChange";

        try {
            Object[] args = joinPoint.getArgs();
            String alertId = (String) args[0];
            String notes = (String) args[1];

            List<AlertType> alerts = getAlerts(Collections.singletonList(alertId));

            joinPoint.proceed();

            if (CollectionUtils.isEmpty(alerts))
                return null;

            String user = SecurityUtils.getCurrentUserLogin().orElse("system");

            for (AlertType alert : alerts) {
                UtmAlertLog alertLog = new UtmAlertLog();
                alertLog.setAlertId(alert.getId());
                alertLog.setLogUser(user);
                alertLog.setLogDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                alertLog.setLogAction(AlertActionEnum.UPDATE_NOTES.name());
                alertLog.setLogOldValue(UtilSerializer.jsonSerialize(alert));

                Map<String, Object> newValue = new LinkedHashMap<>();
                if (StringUtils.hasText(notes)) {
                    newValue.put("notes", notes);
                    alertLog.setLogMessage(String.format("The %1$s user updated the alert notes with: <br>%2$s", user, notes));
                } else {
                    newValue.put("notes", "");
                    alertLog.setLogMessage(String.format("The %1$s user removed all notes of the alert", user));
                }
                alertLog.setLogNewValue(UtilSerializer.jsonSerialize(newValue));
                alertLogService.save(alertLog);
            }
            return null;
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage(), e);
            joinPoint.proceed();
            return null;
        }
    }

    @Around("convertToIncidentPointcut()")
    public Object logConvertToIncident(ProceedingJoinPoint joinPoint) throws Throwable {
        final String ctx = CLASS_NAME + ".logConvertToIncident";

        try {
            Object[] args = joinPoint.getArgs();
            Query query = (Query) args[0];
            String incidentName = (String) args[1];
            Integer incidentId = (Integer) args[2];
            Instant incidentCreationDate = (Instant) args[3];
            String incidentCreatedBy = (String) args[4];
            String incidentSource = (String) args[5];
            String indexPattern = (String) args[6];

            SearchRequest request = SearchRequest.of(s -> s.size(Constants.LOG_ANALYZER_TOTAL_RESULTS)
                .query(query).index(indexPattern));

            HitsMetadata<AlertType> hits = elasticsearchService.search(request, AlertType.class).hits();

            joinPoint.proceed();

            if (hits.total().value() <= 0)
                return null;

            List<AlertType> alerts = hits.hits().stream().map(Hit::source).collect(Collectors.toList());

            if (CollectionUtils.isEmpty(alerts))
                return null;

            for (AlertType alert : alerts) {
                UtmAlertLog alertLog = new UtmAlertLog();
                alertLog.setAlertId(alert.getId());
                alertLog.setLogUser(incidentCreatedBy);
                alertLog.setLogDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                alertLog.setLogAction(AlertActionEnum.MARK_AS_INCIDENT.name());
                alertLog.setLogOldValue(UtilSerializer.jsonSerialize(alert));

                Map<String, Object> newValue = new LinkedHashMap<>();
                newValue.put("isIncident", true);
                newValue.put("incidentDetail.incidentName", incidentName);
                newValue.put("incidentDetail.incidentId", incidentId);
                newValue.put("incidentDetail.createdBy", incidentCreatedBy);
                newValue.put("incidentDetail.creationDate", incidentCreationDate.toString());
                newValue.put("incidentDetail.source", incidentSource);
                alertLog.setLogNewValue(UtilSerializer.jsonSerialize(newValue));
                alertLog.setLogMessage(String.format("This alert was added to incident %2$s by user %1$s",
                    incidentCreatedBy, (incidentName + "(" + incidentId.toString() + ")")));

                alertLogService.save(alertLog);
            }
            return null;
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage(), e);
            joinPoint.proceed();
            return null;
        }
    }

    /**
     * @param ids
     * @return
     * @throws Exception
     */
    private List<AlertType> getAlerts(List<?> ids) throws Exception {
        final String ctx = CLASS_NAME + ".getAlerts";
        try {
            List<FilterType> filters = new ArrayList<>();
            filters.add(new FilterType(Constants.alertIdKeyword, OperatorType.IS_ONE_OF_TERMS, ids));
            SearchRequest request = SearchRequest.of(s -> s.query(SearchUtil.toQuery(filters))
                .index(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS)));
            HitsMetadata<AlertType> hits = elasticsearchService.search(request, AlertType.class).hits();
            if (hits.total().value() <= 0)
                return Collections.emptyList();
            return hits.hits().stream().map(Hit::source).collect(Collectors.toList());
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
