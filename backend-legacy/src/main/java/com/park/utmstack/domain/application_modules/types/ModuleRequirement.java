package com.park.utmstack.domain.application_modules.types;

import com.park.utmstack.domain.application_modules.enums.ModuleRequirementStatus;

public class ModuleRequirement {
    private String checkName;
    private ModuleRequirementStatus checkStatus;
    private String failMessage;

    public String getCheckName() {
        return checkName;
    }

    public void setCheckName(String checkName) {
        this.checkName = checkName;
    }

    public ModuleRequirementStatus getCheckStatus() {
        return checkStatus;
    }

    public void setCheckStatus(ModuleRequirementStatus checkStatus) {
        this.checkStatus = checkStatus;
    }

    public String getFailMessage() {
        return failMessage;
    }

    public void setFailMessage(String failMessage) {
        this.failMessage = failMessage;
    }
}
