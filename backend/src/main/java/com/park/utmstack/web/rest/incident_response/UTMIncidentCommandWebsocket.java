package com.park.utmstack.web.rest.incident_response;

import com.google.gson.Gson;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.agent_manager.AgentGrpcService;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.dto.agent_manager.AgentStatusEnum;
import com.park.utmstack.service.grpc.CommandResult;
import com.park.utmstack.service.grpc.Hostname;
import com.park.utmstack.service.incident_response.UtmIncidentVariableService;
import com.park.utmstack.service.incident_response.grpc_impl.IncidentResponseCommandService;
import com.park.utmstack.util.exceptions.InvalidEchoCommandException;
import com.park.utmstack.web.rest.errors.AgentNotfoundException;
import com.park.utmstack.web.rest.errors.AgentOfflineException;
import com.park.utmstack.web.rest.errors.InternalServerErrorException;
import com.park.utmstack.web.rest.vm.CommandVM;
import io.grpc.stub.StreamObserver;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.messaging.handler.annotation.DestinationVariable;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.bind.annotation.RestController;

import javax.validation.constraints.NotNull;

@RestController
public class UTMIncidentCommandWebsocket {
    private static final String CLASSNAME = "UTMIncidentCommandWebsocket";
    private final Logger log = LoggerFactory.getLogger(UTMIncidentCommandWebsocket.class);

    private final SimpMessagingTemplate messagingTemplate;
    private final IncidentResponseCommandService incidentResponseCommandService;

    private final AgentGrpcService agentGrpcService;

    private final UtmIncidentVariableService utmIncidentVariableService;

    public UTMIncidentCommandWebsocket(SimpMessagingTemplate messagingTemplate,
                                       IncidentResponseCommandService incidentResponseCommandService,
                                       AgentGrpcService agentGrpcService, UtmIncidentVariableService utmIncidentVariableService) {
        this.messagingTemplate = messagingTemplate;
        this.incidentResponseCommandService = incidentResponseCommandService;
        this.agentGrpcService = agentGrpcService;
        this.utmIncidentVariableService = utmIncidentVariableService;
    }

    @MessageMapping("/command/{hostname}")
    public void processCommand(@NotNull String command, @DestinationVariable @NotNull String hostname) {
        final String ctx = CLASSNAME + ".processCommand";
        try {
            String executedBy = SecurityUtils.getCurrentUserLogin()
                    .orElseThrow(() -> new InternalServerErrorException("Current user login not found"));
            String destination = String.format("/topic/%1$s", hostname);
            Hostname req = Hostname.newBuilder()
                    .setHostname(hostname)
                    .build();
            try {
                AgentDTO agentDTO = agentGrpcService.getAgentByHostname(req);
                if (agentDTO.getStatus() == AgentStatusEnum.OFFLINE) {
                    throw new AgentOfflineException();
                }

                CommandVM commandVM = new Gson().fromJson(command, CommandVM.class);
                String commandVar = utmIncidentVariableService.replaceVariablesInCommand(commandVM.getCommand());

                incidentResponseCommandService.sendCommand(agentDTO.getAgentKey(), commandVar, commandVM.getOriginType(),
                        commandVM.getOriginId(), commandVM.getReason(), executedBy, new StreamObserver<>() {
                            @Override
                            public void onNext(CommandResult value) {
                                String output = utmIncidentVariableService.replaceSecretVariableValuesWithPlaceholders(value.getResult());
                                messagingTemplate.convertAndSendToUser(executedBy, destination, output);
                            }

                            @Override
                            public void onError(Throwable t) {
                                String msg = ctx + ": " + t.getLocalizedMessage();
                                log.error(msg);
                                messagingTemplate.convertAndSendToUser(executedBy, destination, t.getLocalizedMessage());
                            }

                            @Override
                            public void onCompleted() {
                            }
                        });

            } catch (Exception exception) {
                throw new AgentNotfoundException();
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
        }
    }
}
