package com.park.utmstack.service.agent_manager;

import com.park.utmstack.service.grpc.DeleteRequest;
import com.park.utmstack.service.grpc.ListRequest;
import com.park.utmstack.config.Constants;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.dto.agent_manager.AgentCommandDTO;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.dto.agent_manager.ListAgentsCommandsResponseDTO;
import com.park.utmstack.service.dto.agent_manager.ListAgentsResponseDTO;
import com.park.utmstack.service.grpc.*;
import com.park.utmstack.web.rest.errors.AgentNotfoundException;
import io.grpc.*;
import io.grpc.Status;
import io.grpc.stub.MetadataUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;
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
        return mapToListAgentsResponseDTO(blockingStub.listAgents(request));
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
            // Load agent list one time, will be used to search the agent that execute the command, by agent_id
            ListRequest req = ListRequest.newBuilder()
                    .setPageNumber(1)
                    .setPageSize(1000000)
                    .setSearchQuery("")
                    .setSortBy("")
                    .build();

            ListAgentsResponse agentResp = blockingStub.listAgents(req);
            List<Agent> agentList = agentResp.getRowsList();

            List<AgentCommandDTO> agentCommandDTOs = response.getRowsList().stream()
                    .map(ac -> {
                        try {
                            return this.protoToDTOAgentCommand(ac, agentList);
                        } catch (Exception e) {
                            throw new RuntimeException(e);
                        }
                    })
                    .collect(Collectors.toList());

            dto.setAgentCommands(agentCommandDTOs);
            dto.setTotal(response.getTotal());

            return dto;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public AgentCommandDTO protoToDTOAgentCommand(AgentCommand agentCommand, List<Agent> agentList) throws Exception {
        // Look for the agent with id = agentCommand.getAgentId() to get the agent that executed the command
        Optional<Agent> agent = agentList.stream().filter(a -> agentCommand.getAgentId() == a.getId()).findFirst();
        if (agentList.isEmpty() || agent.isEmpty()) {
            throw new AgentNotfoundException();
        }
        return new AgentCommandDTO(agentCommand, agent.get());
    }


    public AgentDTO protoToDTOAgent(Agent agent) {
        return new AgentDTO(agent);
    }


    public AgentDTO getAgentByHostname(String hostname) throws AgentNotfoundException {
        final String ctx = CLASSNAME + ".getAgentByHostname";
        try {
            ListRequest req = ListRequest.newBuilder()
                    .setPageNumber(1)
                    .setPageSize(1000000)
                    .setSearchQuery("hostname.Is=" + hostname)
                    .setSortBy("")
                    .build();
            ListAgentsResponseDTO response = listAgents(req);
            List<AgentDTO> agentDTOList = response.getAgents();
            if (agentDTOList.isEmpty()) {
                throw new AgentNotfoundException();
            }

            return agentDTOList.get(0);
        } catch (AgentNotfoundException e) {
            throw e;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    @SuppressWarnings("ResultOfMethodCallIgnored")
    public void deleteAgent(String hostname) throws AgentNotfoundException {
        final String ctx = CLASSNAME + ".deleteAgent";
        try {
            String currentUser = SecurityUtils.getCurrentUserLogin().orElseThrow(() -> new RuntimeException("No current user login"));
            AgentDTO agent = null;
            try {
                agent = getAgentByHostname(hostname);
                if (agent == null) {
                    log.error(String.format("%1$s: Agent %2$s could not be deleted because no information was obtained from the agent", ctx, hostname));
                    return;
                }
            } catch (StatusRuntimeException e) {
                if (e.getStatus().getCode() == Status.Code.NOT_FOUND) {
                    throw new AgentNotfoundException();
                }
            } catch (AgentNotfoundException e) {
                throw e;
            }

            DeleteRequest request = DeleteRequest.newBuilder().setDeletedBy(currentUser).build();

            Metadata customHeaders = new Metadata();
            customHeaders.put(Metadata.Key.of("key", Metadata.ASCII_STRING_MARSHALLER), agent.getAgentKey());
            customHeaders.put(Metadata.Key.of("id", Metadata.ASCII_STRING_MARSHALLER), String.valueOf(agent.getId()));
            customHeaders.put(Metadata.Key.of("type", Metadata.ASCII_STRING_MARSHALLER), Constants.AGENT_HEADER);

            Channel intercept = ClientInterceptors.intercept(grpcManagedChannel, MetadataUtils.newAttachHeadersInterceptor(customHeaders));
            AgentServiceGrpc.AgentServiceBlockingStub newStub = AgentServiceGrpc.newBlockingStub(intercept);
            newStub.deleteAgent(request);

        } catch (AgentNotfoundException e) {
            throw e;
        } catch (Exception e) {
            String msg = e.getLocalizedMessage();
            if(msg.contains("UNAVAILABLE")) {
                msg = ctx + ": Agent couldn't be deleted, agent manager is not available.";
            } else {
                msg = ctx + ": " + e.getMessage();
            }
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }
}
