package com.park.utmstack.web.rest.agent_manager;

import com.park.utmstack.service.grpc.ListRequest;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.agent_manager.AgentGrpcService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.agent_manager.*;
import com.park.utmstack.service.incident_response.UtmIncidentVariableService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.application_modules.UtmModuleResource;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import javax.validation.constraints.NotNull;
import java.util.List;

@RestController
@RequestMapping("/api/agent-manager")
public class AgentManagerResource {
    private static final String CLASSNAME = "AgentManagerResource";
    private final Logger log = LoggerFactory.getLogger(UtmModuleResource.class);
    private final AgentGrpcService agentGrpcService;

    private final UtmIncidentVariableService utmIncidentVariableService;
    private final ApplicationEventService eventService;

    public AgentManagerResource(AgentGrpcService agentGrpcService,
                                UtmIncidentVariableService utmIncidentVariableService,
                                ApplicationEventService eventService) {
        this.agentGrpcService = agentGrpcService;
        this.utmIncidentVariableService = utmIncidentVariableService;
        this.eventService = eventService;
    }

    @GetMapping("/agents")
    public ResponseEntity<List<AgentDTO>> listAgents(
            @RequestParam(required = false) Integer pageNumber,
            @RequestParam(required = false) Integer pageSize,
            @RequestParam(required = false) String searchQuery,
            @RequestParam(required = false) String sortBy) {

        final String ctx = CLASSNAME + ".listAgents";
        try {
            ListRequest request = ListRequest.newBuilder()
                    .setPageNumber(pageNumber != null ? pageNumber : 0)
                    .setPageSize(pageSize != null ? pageSize : 1000000)
                    .setSearchQuery(searchQuery != null ? searchQuery : "")
                    .setSortBy(sortBy != null ? sortBy : "")
                    .build();
            ListAgentsResponseDTO response = agentGrpcService.listAgents(request);
            List<AgentDTO> agentDTOList = response.getAgents();
            agentDTOList.forEach(agentDTO -> agentDTO.setAgentKey("SECRET"));
            HttpHeaders headers = new HttpHeaders();
            headers.add("X-Total-Count", Long.toString(response.getTotal()));
            return ResponseEntity.ok().headers(headers).body(agentDTOList);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/agents-with-commands")
    public ResponseEntity<List<AgentDTO>> listAgentsWithCommands() {
        final String ctx = CLASSNAME + ".listAgentsWithCommands";
        try {
            ListRequest request = ListRequest.newBuilder()
                    .setPageNumber(1)
                    .setPageSize(1000000)
                    .setSearchQuery("")
                    .setSortBy("")
                    .build();
            ListAgentsResponseDTO response = agentGrpcService.listAgentWithCommands(request);
            List<AgentDTO> agentDTOList = response.getAgents();
            agentDTOList.forEach(agentDTO -> agentDTO.setAgentKey("SECRET"));
            HttpHeaders headers = new HttpHeaders();
            headers.add("X-Total-Count", Long.toString(response.getTotal()));
            return ResponseEntity.ok().headers(headers).body(agentDTOList);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/agent-by-hostname")
    public ResponseEntity<AgentDTO> getAgentByHostname(
            @RequestParam @NotNull String hostname) {
        final String ctx = CLASSNAME + ".getAgentByHostname";
        try {
            AgentDTO response = agentGrpcService.getAgentByHostname(hostname);
            response.setAgentKey("SECRET");
            HttpHeaders headers = new HttpHeaders();
            return ResponseEntity.ok().headers(headers).body(response);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/agent-commands")
    public ResponseEntity<List<AgentCommandDTO>> listAgentCommands(
            @RequestParam(required = false) Integer pageNumber,
            @RequestParam(required = false) Integer pageSize,
            @RequestParam(required = false) String searchQuery,
            @RequestParam(required = false) String sortBy) {

        final String ctx = CLASSNAME + ".listAgentCommands";
        try {
            ListRequest request = ListRequest.newBuilder()
                    .setPageNumber(pageNumber != null ? pageNumber : 0)
                    .setPageSize(pageSize != null ? pageSize : 1000000)
                    .setSearchQuery(searchQuery != null ? searchQuery : "")
                    .setSortBy(sortBy != null ? sortBy : "")
                    .build();
            ListAgentsCommandsResponseDTO response = agentGrpcService.listAgentCommands(request);

            List<AgentCommandDTO> commands = response.getAgentCommands();

            for (AgentCommandDTO command : commands) {
                command.setResult(utmIncidentVariableService
                        .replaceSecretVariableValuesWithPlaceholders(command.getResult()));
            }

            HttpHeaders headers = new HttpHeaders();
            headers.add("X-Total-Count", Long.toString(response.getTotal()));
            return ResponseEntity.ok().headers(headers).body(commands);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }

    }

    @GetMapping("/can-run-command")
    public ResponseEntity<Boolean> canRunCommand(@RequestParam String hostname) {
        final String ctx = CLASSNAME + ".canRunCommand";
        try {
            AgentDTO response = agentGrpcService.getAgentByHostname(hostname);
            return ResponseEntity.ok(response.getStatus() == AgentStatusEnum.ONLINE);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}
