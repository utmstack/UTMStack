package com.park.utmstack.service.dto.agent_manager;


import com.park.utmstack.service.grpc.AgentGroup;

public class AgentGroupDTO {
    private int id;
    private String groupName;
    private String groupDescription;

    public AgentGroupDTO(AgentGroup group) {
        this.groupName = group.getGroupName();
        this.id = group.getId();
        this.groupDescription = group.getGroupDescription();
    }

    public String getGroupName() {
        return groupName;
    }

    public void setGroupName(String groupName) {
        this.groupName = groupName;
    }

    public String getGroupDescription() {
        return groupDescription;
    }

    public void setGroupDescription(String groupDescription) {
        this.groupDescription = groupDescription;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }
}


