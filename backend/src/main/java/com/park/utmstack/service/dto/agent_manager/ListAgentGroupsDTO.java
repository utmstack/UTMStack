package com.park.utmstack.service.dto.agent_manager;

import java.util.List;

public class ListAgentGroupsDTO {
    private List<AgentGroupDTO> groups;
    private int total;

    public List<AgentGroupDTO> getGroups() {
        return groups;
    }

    public void setGroups(List<AgentGroupDTO> groups) {
        this.groups = groups;
    }

    public int getTotal() {
        return total;
    }

    public void setTotal(int total) {
        this.total = total;
    }
}
