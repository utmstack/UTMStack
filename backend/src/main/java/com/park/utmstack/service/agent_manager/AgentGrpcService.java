package com.park.utmstack.service.agent_manager;

import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.dto.agent_manager.AgentCommandDTO;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.dto.agent_manager.ListAgentsCommandsResponseDTO;
import com.park.utmstack.service.dto.agent_manager.ListAgentsResponseDTO;
import com.park.utmstack.service.grpc.*;
import io.grpc.*;
import io.grpc.stub.MetadataUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
public class AgentGrpcService {
    private static final String CLASSNAME = "AgentGrpcService";
    private final Logger log = LoggerFactory.getLogger(AgentService.class);

    private final AgentServiceGrpc.AgentServiceBlockingStub blockingStub;
    private final ManagedChannel grpcManagedChannel;

    public AgentGrpcService(ManagedChannel grpcManagedChannel) {
        this.grpcManagedChannel = grpcManagedChannel;
        this.blockingStub = AgentServiceGrpc.newBlockingStub(grpcManagedChannel);
    }

    public ListAgentsResponseDTO listAgents(ListRequest request) throws Exception {
        return mapToListAgentsResponseDTO(blockingStub.listAgents(request));
    }

    public ListAgentsCommandsResponseDTO listAgentCommands(ListRequest request) throws Exception {
        return mapToListAgentsCommandsResponseDTO(blockingStub.listAgentCommands(request));
    }

    public ListAgentsResponseDTO listAgentWithCommands(ListRequest request) throws Exception {
        return mapToListAgentsResponseDTO(blockingStub.listAgentsWithCommands(request));
    }


    public ListAgentsResponseDTO mapToListAgentsResponseDTO(ListAgentsResponse response) throws Exception {
        final String ctx = CLASSNAME + ".mapToListAgentsResponseDTO";
        try {
            ListAgentsResponseDTO dto = new ListAgentsResponseDTO();

            List<AgentDTO> agentDTOs = response.getRowsList().stream()
                .map(this::protoToDTOAgent)
                .collect(Collectors.toList());

            dto.setAgents(agentDTOs);
            dto.setTotal(response.getTotal());

            return dto;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public ListAgentsCommandsResponseDTO mapToListAgentsCommandsResponseDTO(ListAgentsCommandsResponse response) throws Exception {
        final String ctx = CLASSNAME + ".mapToListAgentsCommandsResponseDTO";
        try {
            ListAgentsCommandsResponseDTO dto = new ListAgentsCommandsResponseDTO();

            List<AgentCommandDTO> agentCommandDTOs = response.getRowsList().stream()
                .map(this::protoToDTOAgentCommand)
                .collect(Collectors.toList());

            dto.setAgentCommands(agentCommandDTOs);
            dto.setTotal(response.getTotal());

            return dto;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public AgentCommandDTO protoToDTOAgentCommand(AgentCommand agentCommand) {
        return new AgentCommandDTO(agentCommand);
    }


    public AgentDTO protoToDTOAgent(Agent agent) {
        return new AgentDTO(agent);
    }


    public AgentDTO updateAgentType(AgentTypeUpdate newType) throws Exception {
        final String ctx = CLASSNAME + ".updateAgentType";
        try {
            Agent agent = blockingStub.updateAgentType(newType);
            return protoToDTOAgent(agent);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public AgentDTO updateAgentGroup(AgentGroupUpdate newGroup) throws Exception {
        final String ctx = CLASSNAME + ".updateAgentGroup";
        try {
            Agent agent = blockingStub.updateAgentGroup(newGroup);
            return protoToDTOAgent(agent);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public AgentDTO getAgentByHostname(Hostname hostname) {
        final String ctx = CLASSNAME + ".getAgentByHostname";
        try {
            Agent agent = blockingStub.getAgentByHostname(hostname);
            return protoToDTOAgent(agent);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    @SuppressWarnings("ResultOfMethodCallIgnored")
    public void deleteAgent(String hostname) {
        final String ctx = CLASSNAME + ".deleteAgent";
        try {
            String currentUser = SecurityUtils.getCurrentUserLogin().orElseThrow(() -> new RuntimeException("No current user login"));
            Agent agent = null;
            try {
                agent = blockingStub.getAgentByHostname(Hostname.newBuilder().setHostname(hostname).build());
                if (agent == null) {
                    log.error(String.format("%1$s: Agent %2$s could not be deleted because no information was obtained from the agent", ctx, hostname));
                    return;
                }
            } catch (StatusRuntimeException e) {
                if (e.getStatus().getCode() == Status.Code.NOT_FOUND) {
                    log.error(String.format("%1$s: Agent %2$s could not be deleted because was not found", ctx, hostname));
                    return;
                }
            }

            AgentDelete request = AgentDelete.newBuilder().setAgentKey(agent.getAgentKey())
                .setDeletedBy(currentUser).build();

            Metadata customHeaders = new Metadata();
            customHeaders.put(Metadata.Key.of("agent-key", Metadata.ASCII_STRING_MARSHALLER), agent.getAgentKey());
            customHeaders.put(Metadata.Key.of("agent-id", Metadata.ASCII_STRING_MARSHALLER), String.valueOf(agent.getId()));

            Channel intercept = ClientInterceptors.intercept(grpcManagedChannel, MetadataUtils.newAttachHeadersInterceptor(customHeaders));
            AgentServiceGrpc.AgentServiceBlockingStub newStub = AgentServiceGrpc.newBlockingStub(intercept);
            newStub.deleteAgent(request);

        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }
}
