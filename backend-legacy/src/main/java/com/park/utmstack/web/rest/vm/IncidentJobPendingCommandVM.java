package com.park.utmstack.web.rest.vm;

public class IncidentJobPendingCommandVM {
    private Long id;
    private String command;

    public IncidentJobPendingCommandVM(Long id, String command) {
        this.id = id;
        this.command = command;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getCommand() {
        return command;
    }

    public void setCommand(String command) {
        this.command = command;
    }
}
