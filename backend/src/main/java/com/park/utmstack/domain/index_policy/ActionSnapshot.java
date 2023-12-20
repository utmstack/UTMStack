package com.park.utmstack.domain.index_policy;

/**
 * Backup your cluster’s indices and state.
 */
public class ActionSnapshot {
    /**
     * The repository name
     * Required: Yes
     */
    private String repository;

    /**
     * The name of the snapshot
     * Required: Yes
     */
    private String snapshot;

    public ActionSnapshot() {
    }

    public ActionSnapshot(String repository, String snapshot) {
        this.repository = repository;
        this.snapshot = snapshot;
    }

    public String getRepository() {
        return repository;
    }

    public void setRepository(String repository) {
        this.repository = repository;
    }

    public String getSnapshot() {
        return snapshot;
    }

    public void setSnapshot(String snapshot) {
        this.snapshot = snapshot;
    }
}
