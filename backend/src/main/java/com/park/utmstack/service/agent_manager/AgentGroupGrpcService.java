package com.park.utmstack.service.agent_manager;

import com.park.utmstack.service.dto.agent_manager.AgentGroupDTO;
import com.park.utmstack.service.dto.agent_manager.ListAgentGroupsDTO;
import com.park.utmstack.service.grpc.*;
import io.grpc.ManagedChannel;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
public class AgentGroupGrpcService {
    private static final String CLASSNAME = "AgentGroupGrpcService";
    private final Logger log = LoggerFactory.getLogger(AgentService.class);

    private final AgentGroupServiceGrpc.AgentGroupServiceBlockingStub blockingStub;

    public AgentGroupGrpcService(ManagedChannel grpcManagedChannel) {
        this.blockingStub = AgentGroupServiceGrpc.newBlockingStub(grpcManagedChannel);
    }

    public ListAgentGroupsDTO listGroups(ListRequest request) throws Exception {
        final String ctx = CLASSNAME + ".createGroup";
        try {
            return mapToListAgentsGroupDTO(blockingStub.listGroups(request));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public AgentGroupDTO createGroup(AgentGroup request) throws Exception {
        final String ctx = CLASSNAME + ".createGroup";
        try {
            return protoToDTOGroup(blockingStub.createGroup(request));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public AgentGroupDTO editGroup(AgentGroup request) throws Exception {
        final String ctx = CLASSNAME + ".editGroup";
        try {
            return protoToDTOGroup(blockingStub.editGroup(request));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public Long deleteGroup(Long groupId) throws Exception {
        final String ctx = CLASSNAME + ".deleteGroup";
        try {
            Id id = Id.newBuilder().setId(groupId).build();
            return blockingStub.deleteGroup(id).getId();
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public ListAgentGroupsDTO mapToListAgentsGroupDTO(ListAgentsGroupResponse response) throws Exception {
        final String ctx = CLASSNAME + ".mapToListAgentsGroupDTO";
        try {
            ListAgentGroupsDTO dto = new ListAgentGroupsDTO();

            List<AgentGroupDTO> groups = response.getRowsList().stream()
                .map(this::protoToDTOGroup)
                .collect(Collectors.toList());

            dto.setGroups(groups);
            dto.setTotal(response.getTotal());

            return dto;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public AgentGroupDTO protoToDTOGroup(AgentGroup group) {
        return new AgentGroupDTO(group);
    }


}
