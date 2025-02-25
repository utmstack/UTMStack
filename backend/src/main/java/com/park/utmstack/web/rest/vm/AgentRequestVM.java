package com.park.utmstack.web.rest.vm;

import com.park.utmstack.service.grpc.AgentRequest;
import javax.validation.constraints.NotEmpty;

/*
* Use this class when you need to update agent's attributes.
* To add new attributes to update, add it to the class. Actually only ip is permitted.
* */
public class AgentRequestVM {
    @NotEmpty
    private String ip;
    @NotEmpty
    private String hostname;


    public AgentRequestVM() {}

    public AgentRequest getAgentRequest() {
        return AgentRequest.newBuilder()
                .setIp(this.ip)
                .setHostname(this.hostname)
                .build();
    }

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public String getHostname() {
        return hostname;
    }

    public void setHostname(String hostname) {
        this.hostname = hostname;
    }
}
