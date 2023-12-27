package com.park.utmstack.service.dto.agent_manager;

import java.util.List;

public class ListAgentsCommandsResponseDTO {
    private List<AgentCommandDTO> agentCommands;
    private int total;

    public List<AgentCommandDTO> getAgentCommands() {
        return agentCommands;
    }

    public void setAgentCommands(List<AgentCommandDTO> agentCommands) {
        this.agentCommands = agentCommands;
    }

    public int getTotal() {
        return total;
    }

    public void setTotal(int total) {
        this.total = total;
    }
}
