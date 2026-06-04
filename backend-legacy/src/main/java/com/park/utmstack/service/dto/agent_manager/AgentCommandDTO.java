package com.park.utmstack.service.dto.agent_manager;

import com.park.utmstack.service.grpc.Agent;
import com.park.utmstack.service.grpc.AgentCommand;

import java.time.Instant;

public class AgentCommandDTO {
    private String createdAt;
    private String updatedAt;
    private int agentId;
    private String command;
    private String commandStatus;
    private String result;
    private String executedBy;
    private String cmdId;

    private AgentDTO agent;

    private String reason;
    private String originType;
    private String originId;

    public AgentCommandDTO(AgentCommand agentCommand, Agent originalAgent) {
        this.createdAt = Instant.ofEpochSecond(agentCommand.getCreatedAt().getSeconds()).toString();
        this.updatedAt = agentCommand.getUpdatedAt().toString();
        this.agentId = agentCommand.getAgentId();
        this.command = agentCommand.getCommand();
        this.commandStatus = agentCommand.getCommandStatus().toString();
        this.result = agentCommand.getResult();
        this.executedBy = agentCommand.getExecutedBy();
        this.cmdId = agentCommand.getCmdId();
        this.agent = new AgentDTO(originalAgent);
        this.agent.setAgentKey("SECRET");
        this.reason = agentCommand.getReason();
        this.originId = agentCommand.getOriginId();
        this.originType = agentCommand.getOriginType();
    }

    public String getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(String createdAt) {
        this.createdAt = createdAt;
    }

    public String getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(String updatedAt) {
        this.updatedAt = updatedAt;
    }

    public int getAgentId() {
        return agentId;
    }

    public void setAgentId(int agentId) {
        this.agentId = agentId;
    }

    public String getCommand() {
        return command;
    }

    public void setCommand(String command) {
        this.command = command;
    }

    public String getCommandStatus() {
        return commandStatus;
    }

    public void setCommandStatus(String commandStatus) {
        this.commandStatus = commandStatus;
    }

    public String getResult() {
        return result;
    }

    public void setResult(String result) {
        this.result = result;
    }

    public String getExecutedBy() {
        return executedBy;
    }

    public void setExecutedBy(String executedBy) {
        this.executedBy = executedBy;
    }

    public String getCmdId() {
        return cmdId;
    }

    public void setCmdId(String cmdId) {
        this.cmdId = cmdId;
    }

    public AgentDTO getAgent() {
        return agent;
    }

    public void setAgent(AgentDTO agent) {
        this.agent = agent;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }

    public String getOriginType() {
        return originType;
    }

    public void setOriginType(String originType) {
        this.originType = originType;
    }

    public String getOriginId() {
        return originId;
    }

    public void setOriginId(String originId) {
        this.originId = originId;
    }
}

