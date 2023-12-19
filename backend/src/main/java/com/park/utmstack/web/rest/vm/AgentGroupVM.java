package com.park.utmstack.web.rest.vm;

import javax.validation.constraints.NotBlank;

public class AgentGroupVM {
    @NotBlank
    private String groupName;
    private String groupDescription;
    private Long id;

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

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }
}
