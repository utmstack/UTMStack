package com.park.utmstack.service.dto.agent_manager;

import java.util.List;

public class ListAgentsResponseDTO {
    private List<AgentDTO> agents;
    private int total;

    public List<AgentDTO> getAgents() {
        return agents;
    }

    public void setAgents(List<AgentDTO> agents) {
        this.agents = agents;
    }

    public int getTotal() {
        return total;
    }

    public void setTotal(int total) {
        this.total = total;
    }
}
