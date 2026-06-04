package com.park.utmstack.service.dynamic_schedules;

import org.springframework.context.event.ContextRefreshedEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

@Component
public class Schedulers {
    private final DynamicScheduleTaskService dynamicScheduleService;

    public Schedulers(DynamicScheduleTaskService dynamicScheduleService) {
        this.dynamicScheduleService = dynamicScheduleService;
    }

    @EventListener({ContextRefreshedEvent.class})
    public void contextRefreshedEvent() {
    }
}
