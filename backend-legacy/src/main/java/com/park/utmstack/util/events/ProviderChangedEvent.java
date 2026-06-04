package com.park.utmstack.util.events;

import org.springframework.context.ApplicationEvent;

public class ProviderChangedEvent extends ApplicationEvent {
    public ProviderChangedEvent(Object source) {
        super(source);
    }
}