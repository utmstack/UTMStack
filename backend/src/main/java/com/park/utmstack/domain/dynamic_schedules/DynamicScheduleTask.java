package com.park.utmstack.domain.dynamic_schedules;

import java.util.Optional;

public class DynamicScheduleTask {
    protected String id;

    protected DynamicScheduleTask() {
    }

    protected DynamicScheduleTask(String id) {
        this.id = id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getId() {
        return id;
    }

    protected Optional<Runnable> getAction() {
        return Optional.empty();
    }
}
