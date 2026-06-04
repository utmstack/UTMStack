package com.park.utmstack.service.application_events;

import com.park.utmstack.domain.application_events.enums.ApplicationEventSource;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_events.types.ApplicationEvent;
import com.park.utmstack.loggin.LogContextBuilder;
import com.park.utmstack.service.elasticsearch.OpensearchClientBuilder;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.logstash.logback.argument.StructuredArguments;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.MDC;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.Map;

@Service
@Slf4j
@RequiredArgsConstructor
public class ApplicationEventService {
    private static final String CLASSNAME = "ApplicationEventService";

    private final OpensearchClientBuilder client;
    private final LogContextBuilder logContextBuilder;


    /**
     * Create an application event. Can be an error, warning or info
     *
     * @param message : Message of the event
     * @param type    : Type of event (ERROR, WARNING, INFO)
     */
    @Async
    public void createEvent(String message, ApplicationEventType type) {
        final String ctx = CLASSNAME + ".createEvent";
        final String V11_LOCAL_INDEX = "v11-backend-logs";
        try {
            ApplicationEvent applicationEvent = ApplicationEvent.builder()
                .message(message)
                .timestamp(Instant.now().toString())
                .source(ApplicationEventSource.PANEL.name())
                .type(type.name())
                .build();
            client.getClient().index(V11_LOCAL_INDEX, applicationEvent);
        } catch (Exception e) {
            log.error( "{}: Error creating application event: {}", ctx, e.getMessage(), e);
        }
    }

    public void createEvent(String message, ApplicationEventType type, Map<String, Object> details) {
        String msg = String.format("%s: %s", MDC.get("context"), message);
        log.info( msg, StructuredArguments.keyValue("args", logContextBuilder.buildArgs(details)));
        this.createEvent(message, this.getType(type));
    }

    private ApplicationEventType getType(ApplicationEventType type) {
        switch (type) {
            case ERROR -> {
                return ApplicationEventType.ERROR;
            }
            case WARNING -> {
                return ApplicationEventType.WARNING;
            }
            default ->  {
                return ApplicationEventType.INFO;
            }
        }
    }
}
