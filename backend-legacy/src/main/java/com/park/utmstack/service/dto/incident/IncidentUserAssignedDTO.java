package com.park.utmstack.service.dto.incident;

public class IncidentUserAssignedDTO {
    private Long id;
    private String login;

    public IncidentUserAssignedDTO() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getLogin() {
        return login;
    }

    public void setLogin(String login) {
        this.login = login;
    }
}
