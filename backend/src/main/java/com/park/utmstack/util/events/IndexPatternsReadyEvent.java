package com.park.utmstack.util.events;

import org.springframework.context.ApplicationEvent;

public class IndexPatternsReadyEvent extends ApplicationEvent {
    /**
     * Create a new ApplicationEvent.
     *
     * @param source the object on which the event initially occurred (never {@code null})
     */
    public IndexPatternsReadyEvent(Object source) {
        super(source);
    }
}
