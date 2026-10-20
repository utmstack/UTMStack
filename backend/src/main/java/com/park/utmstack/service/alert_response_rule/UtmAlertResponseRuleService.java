package com.park.utmstack.service.alert_response_rule;

import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import com.jayway.jsonpath.Criteria;
import com.jayway.jsonpath.Filter;
import com.jayway.jsonpath.Predicate;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseActionTemplate;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRule;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleExecution;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleHistory;
import com.park.utmstack.domain.alert_response_rule.enums.RuleExecutionStatus;
import com.park.utmstack.domain.alert_response_rule.enums.RuleNonExecutionCause;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.shared_types.alert.UtmAlert;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseActionTemplateRepository;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleExecutionRepository;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleHistoryRepository;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleRepository;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.agent_manager.AgentService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmAlertResponseRuleDTO;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.dto.agent_manager.AgentStatusEnum;
import com.park.utmstack.service.grpc.CommandResult;
import com.park.utmstack.service.incident_response.UtmIncidentVariableService;
import com.park.utmstack.service.incident_response.grpc_impl.IncidentResponseCommandService;
import com.park.utmstack.util.UtilJson;
import com.park.utmstack.util.exceptions.UtmNotImplementedException;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
@Transactional
public class UtmAlertResponseRuleService implements ApplicationRunner {

    private static final String CLASSNAME = "UtmAlertResponseRuleService";
    private final Logger log = LoggerFactory.getLogger(UtmAlertResponseRuleService.class);

    private final UtmAlertResponseRuleRepository alertResponseRuleRepository;
    private final UtmAlertResponseRuleHistoryRepository alertResponseRuleHistoryRepository;
    private final UtmNetworkScanRepository networkScanRepository;
    private final ApplicationEventService eventService;
    private final AgentService agentService;
    private final IncidentResponseCommandService incidentResponseCommandService;
    private final UtmAlertResponseRuleExecutionRepository alertResponseRuleExecutionRepository;
    private final UtmIncidentVariableService utmIncidentVariableService;
    private final UtmAlertResponseActionTemplateRepository utmAlertResponseActionTemplateRepository;
    private final UtmAlertResponseActionTemplateService utmAlertResponseActionTemplateService;
    private final JdbcTemplate jdbcTemplate;

    /**
     * issue #2942: an EXECUTING row is finalised by the gRPC callback, which
     * lives only in this JVM's memory. If the backend restarts while a command
     * is in flight the callback is gone and the row would stay EXECUTING
     * forever (and the action may not have run on the agent). On startup,
     * reset every EXECUTING row to PENDING so the scheduler retries it once,
     * the same way it already retries when the agent is offline.
     */
    @Override
    public void run(ApplicationArguments args) {
        final String ctx = CLASSNAME + ".run";
        try {
            int reset = jdbcTemplate.update(
                    "UPDATE utm_alert_response_rule_execution SET execution_status = ?, non_execution_cause = NULL " +
                            "WHERE execution_status = ?",
                    RuleExecutionStatus.PENDING.name(), RuleExecutionStatus.EXECUTING.name());
            if (reset > 0) {
                log.warn(ctx + ": reset {} in-flight (EXECUTING) response executions to PENDING after restart", reset);
            }
        } catch (Exception e) {
            log.error(ctx + ": " + e.getLocalizedMessage(), e);
        }
    }



    public UtmAlertResponseRule save(UtmAlertResponseRule alertResponseRule, boolean isCreate) {
        final String ctx = CLASSNAME + ".save";
        try {
            if (!isCreate) {
                String alertRuleId = String.valueOf(alertResponseRule.getId());
                UtmAlertResponseRule current = alertResponseRuleRepository.findById(alertResponseRule.getId())
                        .orElseThrow(() -> new RuntimeException(String.format("Incident response rule with ID: %1$s not found", alertRuleId)));
                this.mergeInto(current, alertResponseRule);
                alertResponseRuleHistoryRepository.save(new UtmAlertResponseRuleHistory(new UtmAlertResponseRuleDTO(alertResponseRule)));
                alertResponseRule = current;
            }

            return alertResponseRuleRepository.save(alertResponseRule);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    @Transactional(readOnly = true)
    public Optional<UtmAlertResponseRule> findOne(Long id) {
        final String ctx = CLASSNAME + ".findOne";
        try {
            return alertResponseRuleRepository.findById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    public void delete(Long id) {
        final String ctx = CLASSNAME + ".delete";
        try {
            alertResponseRuleRepository.deleteById(id);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    public Map<String, List<String>> resolveFilterValues() {
        final String ctx = CLASSNAME + ".resolveFilterValues";
        try {
            return Map.of(
                    "agentPlatform", alertResponseRuleRepository.findAgentPlatformValues(),
                    "users", alertResponseRuleRepository.findUserValues()
            );
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    @Async
    public void evaluateRules(List<UtmAlert> alerts) {
        final String ctx = CLASSNAME + ".evaluateRules";
        try {
            if (CollectionUtils.isEmpty(alerts))
                return;

            List<UtmAlertResponseRule> rules = alertResponseRuleRepository.findAllByRuleActiveIsTrue();
            if (CollectionUtils.isEmpty(rules))
                return;

            // Excluding alerts tagged as false positive
            alerts = alerts.stream().filter(a -> (CollectionUtils.isEmpty(a.getTags()) || !a.getTags().contains("False positive")))
                    .collect(Collectors.toList());

            // Do nothing if there is no valid alerts to check
            if (CollectionUtils.isEmpty(alerts))
                return;

            String alertJsonArray = new Gson().toJson(alerts);
            for (UtmAlertResponseRule rule : rules) {
                List<String> agentNames = networkScanRepository.findAgentNamesByPlatform(rule.getAgentPlatform());

                if (CollectionUtils.isEmpty(agentNames))
                    continue;

                // Matching agents (these are the alerts made from logs coming from an agent)
                //------------------------------------------------------------------------------------------
                createResponseRuleExecution(rule,alertJsonArray,agentNames,true);

                // Then the alerts that match the filters but aren't from an agent, gets executed using the default agent if there is one
                //-----------------------------------------------------------------------------------------------------------------------
                if (StringUtils.hasText(rule.getDefaultAgent())) {
                    createResponseRuleExecution(rule,alertJsonArray,agentNames,false);
                }
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    private void createResponseRuleExecution (UtmAlertResponseRule rule, String alertJsonArray, List<String> agentNames, boolean isAgent) throws Exception {
        final String ctx = CLASSNAME + ".createResponseRuleExecution";
        List<FilterType> conditions = new ArrayList<>();
        try {
            // Common conditions
            if (StringUtils.hasText(rule.getRuleConditions()))
                conditions.addAll(new Gson().fromJson(rule.getRuleConditions(), TypeToken.getParameterized(List.class, FilterType.class).getType()));

            if (StringUtils.hasText(rule.getExcludedAgents()))
                conditions.add(new FilterType(Constants.alertDataSourceKeyword, OperatorType.IS_NOT_ONE_OF, List.of(rule.getExcludedAgents().split(","))));

            // Specific condition for agent and non agents
            if (isAgent) {
                conditions.add(new FilterType(Constants.alertDataSourceKeyword, OperatorType.IS_ONE_OF, agentNames));
            } else {
                conditions.add(new FilterType(Constants.alertDataSourceKeyword, OperatorType.IS_NOT_ONE_OF, agentNames));
            }

            // Processing the alerts and generating the rule executions
            Filter filter = buildFilters(conditions);
            List<?> matches = UtilJson.read("$[?]", alertJsonArray, filter);

            if (!CollectionUtils.isEmpty(matches)) {

                for (Object match : matches) {
                    String matchAsJson = new Gson().toJson(match);

                    UtmAlertResponseRuleExecution exe = new UtmAlertResponseRuleExecution();
                    // Execution agent takes the rule's default agent if the alert was generated by logs from non agent datasource
                    if (isAgent) {
                        exe.setAgent(UtilJson.read("$.dataSource", matchAsJson));
                    } else {
                        exe.setAgent(rule.getDefaultAgent());
                    }

                    exe.setAlertId(UtilJson.read("$.id", matchAsJson));
                    exe.setRuleId(rule.getId());
                    exe.setCommand(buildCommand(rule.getRuleCmd(), matchAsJson));
                    exe.setExecutionStatus(RuleExecutionStatus.PENDING);
                    alertResponseRuleExecutionRepository.save(exe);
                }
            }

        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            throw new Exception(msg);
        }
    }

    private Filter buildFilters(List<FilterType> filters) {
        final String ctx = CLASSNAME + ".buildFilters";
        try {
            Iterator<FilterType> it = filters.iterator();
            List<Predicate> predicates = new ArrayList<>();
            while (it.hasNext()) {
                FilterType filter = it.next();
                switch (filter.getOperator()) {
                    case IS:
                        predicates.add(Criteria.where(filter.getField().replace(".keyword", "")).eq(filter.getValue()));
                        break;
                    case IS_NOT_ONE_OF:
                        Assert.isInstanceOf(List.class, filter.getValue(), "To use the IS_NOT_ONE_OF operator the value must be a list");
                        predicates.add(Criteria.where(filter.getField().replace(".keyword", "")).nin((List<?>) filter.getValue()));
                        break;
                    case IS_ONE_OF:
                        Assert.isInstanceOf(List.class, filter.getValue(), "To use the IS_ONE_OF operator the value must be a list");
                        predicates.add(Criteria.where(filter.getField().replace(".keyword", "")).in((List<?>) filter.getValue()));
                        break;
                    default:
                        throw new UtmNotImplementedException(String.format("Operator: %1$s is not implemented", filter.getOperator()));
                }
            }
            return Filter.filter(predicates);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private String buildCommand(String rawCommand, String alertJson) {
        final String ctx = CLASSNAME + ".buildCommand";
        try {
            String regex = "\\$\\((.*?)\\)";
            StringBuilder output = new StringBuilder();
            int lastIndex = 0;

            Pattern pattern = Pattern.compile(regex);
            Matcher matcher = pattern.matcher(rawCommand);

            while (matcher.find()) {
                String fieldName = matcher.group(1);
                Object value = UtilJson.read(String.format("$.%s", fieldName), alertJson);
                output.append(rawCommand, lastIndex, matcher.start());
                output.append(value);
                lastIndex = matcher.end();
            }
            output.append(rawCommand.substring(lastIndex));
            return output.toString();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    @Scheduled(fixedDelay = 30, timeUnit = TimeUnit.SECONDS)
    public void executeRuleCommands() {
        final String ctx = CLASSNAME + ".executeRuleCommands";
        try {
            List<UtmAlertResponseRuleExecution> cmds = alertResponseRuleExecutionRepository.findAllRuleByExecutionStatusAndRule_RuleActiveTrue(RuleExecutionStatus.PENDING);
            if (CollectionUtils.isEmpty(cmds))
                return;

            for (UtmAlertResponseRuleExecution cmd : cmds) {
                Optional<AgentDTO> opt = agentService.getAgentByHostName(cmd.getAgent());
                if (opt.isEmpty()) {
                    cmd.setExecutionStatus(RuleExecutionStatus.FAILED);
                    cmd.setNonExecutionCause(RuleNonExecutionCause.AGENT_NOT_FOUND);
                    alertResponseRuleExecutionRepository.save(cmd);
                } else {
                    AgentDTO agent = opt.get();
                    if (agent.getStatus().equals(AgentStatusEnum.ONLINE)) {
                        String reason = "The incident response automation executed this command because it was accomplished the conditions of the rule with ID: " + cmd.getRuleId();
                        String shell = cmd.getRule() != null ? cmd.getRule().getRuleShell() : null;
                        // Mark the execution in-flight in its OWN committed transaction BEFORE
                        // dispatching (issue #2942). The scheduler only re-selects PENDING rows,
                        // so this makes the record invisible to the next tick and prevents a
                        // duplicate dispatch if the command takes longer than the 30s fixedDelay.
                        // The gRPC result callback arrives later, on another thread, so it must
                        // not rely on the outer @Transactional context still being open.
                        jdbcTemplate.update(
                                "UPDATE utm_alert_response_rule_execution SET execution_status = ? WHERE id = ?",
                                RuleExecutionStatus.EXECUTING.name(), cmd.getId());
                        incidentResponseCommandService.sendCommand(String.valueOf(agent.getId()), utmIncidentVariableService.replaceVariablesInCommand(cmd.getCommand()), "INCIDENT_RESPONSE_AUTOMATION",
                                cmd.getRuleId().toString(), reason, Constants.SYSTEM_ACCOUNT, shell,
                                new SoarExecutionObserver(ctx, cmd, alertResponseRuleExecutionRepository));
                    } else {
                        if (cmd.getExecutionRetries() < Constants.IRA_EXECUTION_RETRIES) {
                            cmd.setExecutionStatus(RuleExecutionStatus.PENDING);
                            cmd.setExecutionRetries(cmd.getExecutionRetries() + 1);
                            cmd.setNonExecutionCause(RuleNonExecutionCause.AGENT_OFFLINE);
                        } else {
                            cmd.setExecutionStatus(RuleExecutionStatus.FAILED);
                        }
                        alertResponseRuleExecutionRepository.save(cmd);
                    }
                }
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
        }
    }

    public void mergeInto(UtmAlertResponseRule target, UtmAlertResponseRule source) {
        target.setRuleName(source.getRuleName());
        target.setRuleDescription(source.getRuleDescription());
        target.setRuleCmd(source.getRuleCmd());
        target.setRuleActive(source.getRuleActive());
        target.setAgentPlatform(source.getAgentPlatform());
        target.setDefaultAgent(source.getDefaultAgent());
        target.setExcludedAgents(source.getExcludedAgents());
        target.setRuleConditions(source.getRuleConditions());

        target.getUtmAlertResponseActionTemplates().clear();

        if (!source.getUtmAlertResponseActionTemplates().isEmpty()) {
            List<UtmAlertResponseActionTemplate> managedTemplates = source.getUtmAlertResponseActionTemplates()
                    .stream()
                    .map(t -> {
                        if (t.getId() != null) {
                            UtmAlertResponseActionTemplate existing = utmAlertResponseActionTemplateRepository.findById(t.getId())
                                    .orElseThrow(() -> new RuntimeException("Template not found: " + t.getId()));
                            existing.setTitle(t.getTitle());
                            existing.setDescription(t.getDescription());
                            existing.setCommand(t.getCommand());
                            return existing;
                        } else {
                            UtmAlertResponseActionTemplate newT = new UtmAlertResponseActionTemplate();
                            newT.setTitle(t.getTitle());
                            newT.setDescription(t.getDescription());
                            newT.setCommand(t.getCommand());
                            return utmAlertResponseActionTemplateService.save(newT);
                        }
                    })
                    .collect(Collectors.toList());

            target.getUtmAlertResponseActionTemplates().addAll(managedTemplates);
        }
    }

    public Long getSystemSequenceNextValue() {
        return alertResponseRuleRepository.findFirstBySystemOwnerIsTrueOrderByIdDesc()
                .map(rule -> rule.getId() + 1)
                .orElse(1L);
    }
}
