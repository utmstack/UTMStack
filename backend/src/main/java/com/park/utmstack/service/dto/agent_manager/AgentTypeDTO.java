package com.park.utmstack.service.dto.agent_manager;


import com.park.utmstack.service.grpc.AgentType;

public class AgentTypeDTO {
    private int id;
    private String typeName;

    public AgentTypeDTO(AgentType agentType) {
        this.id = agentType.getId();
        this.typeName = agentType.getTypeName();
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public String getTypeName() {
        return typeName;
    }

    public void setTypeName(String typeName) {
        this.typeName = typeName;
    }
}


