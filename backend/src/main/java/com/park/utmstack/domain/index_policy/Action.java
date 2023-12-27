package com.park.utmstack.domain.index_policy;

import com.google.gson.annotations.SerializedName;

/**
 * Actions are the steps that the policy sequentially executes on entering a specific state.
 * They are executed in the order in which they are defined.
 */
public class Action {

    /**
     * @see ActionForceMerge
     */
    @SerializedName(value = "force_merge")
    private ActionForceMerge forceMerge;

    /**
     * @see ActionDelete
     */
    private ActionDelete delete;

    /**
     * @see ActionSnapshot
     */
    private ActionSnapshot snapshot;

    /**
     * @see ActionReadWrite
     */
    @SerializedName(value = "read_write")
    private ActionReadWrite readWrite;

    public Action() {
    }

    public Action(ActionForceMerge forceMerge, ActionDelete delete, ActionSnapshot snapshot, ActionReadWrite readWrite) {
        this.forceMerge = forceMerge;
        this.delete = delete;
        this.snapshot = snapshot;
        this.readWrite = readWrite;
    }

    public ActionForceMerge getForceMerge() {
        return forceMerge;
    }

    public void setForceMerge(ActionForceMerge forceMerge) {
        this.forceMerge = forceMerge;
    }

    public ActionDelete getDelete() {
        return delete;
    }

    public void setDelete(ActionDelete delete) {
        this.delete = delete;
    }

    public ActionSnapshot getSnapshot() {
        return snapshot;
    }

    public void setSnapshot(ActionSnapshot snapshot) {
        this.snapshot = snapshot;
    }

    public ActionReadWrite getReadWrite() {
        return readWrite;
    }

    public void setReadWrite(ActionReadWrite readWrite) {
        this.readWrite = readWrite;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private ActionReadWrite readWrite;
        private ActionForceMerge forceMerge;
        private ActionSnapshot snapshot;
        private ActionDelete delete;

        public Builder readWrite(ActionReadWrite readWrite) {
            this.readWrite = readWrite;
            return this;
        }

        public Builder forceMerge(ActionForceMerge forceMerge) {
            this.forceMerge = forceMerge;
            return this;
        }

        public Builder snapshot(ActionSnapshot snapshot) {
            this.snapshot = snapshot;
            return this;
        }

        public Builder delete(ActionDelete delete) {
            this.delete = delete;
            return this;
        }

        public Action build() {
            return new Action(forceMerge, delete, snapshot, readWrite);
        }
    }
}
