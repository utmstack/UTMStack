package com.park.utmstack.domain.index_policy;

import javax.validation.constraints.NotNull;

public class PolicySettings {
    @NotNull
    private Boolean snapshotActive;
    @NotNull
    private String deleteAfter;

    public Boolean isSnapshotActive() {
        return snapshotActive;
    }

    public void setSnapshotActive(Boolean snapshotActive) {
        this.snapshotActive = snapshotActive;
    }

    public String getDeleteAfter() {
        return deleteAfter;
    }

    public void setDeleteAfter(String deleteAfter) {
        this.deleteAfter = deleteAfter;
    }
}
