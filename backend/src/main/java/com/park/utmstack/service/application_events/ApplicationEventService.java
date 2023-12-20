package com.park.utmstack.service.application_events;

import com.park.utmstack.domain.application_events.enums.ApplicationEventSource;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_events.types.ApplicationEvent;
import com.park.utmstack.service.elasticsearch.OpensearchClientBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.time.Instant;

@Service
public class ApplicationEventService {
    private static final String CLASSNAME = "ApplicationEventService";
    private final Logger log = LoggerFactory.getLogger(ApplicationEventService.class);

    private final OpensearchClientBuilder client;

    public ApplicationEventService(OpensearchClientBuilder client) {
        this.client = client;
    }

    /**
     * Create an application event. Can be an error, warning or info
     *
     * @param message : Message of the event
     * @param type    : Type of event (ERROR, WARNING, INFO)
     */
    @Async
    public void createEvent(String message, ApplicationEventType type) {
        final String ctx = CLASSNAME + ".createEvent";
        try {
            ApplicationEvent applicationEvent = ApplicationEvent.builder()
                .message(message).timestamp(Instant.now().toString())
                .source(ApplicationEventSource.PANEL.name()).type(type.name())
                .build();
            client.getClient().index(".utmstack-logs", applicationEvent);
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage());
        }
    }
}
