package com.park.utmstack.service.application_events;

import com.park.utmstack.domain.application_events.enums.ApplicationEventSource;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_events.types.ApplicationEvent;
import com.park.utmstack.loggin.LogContextBuilder;
import com.park.utmstack.service.elasticsearch.OpensearchClientBuilder;
import lombok.RequiredArgsConstructor;
import net.logstash.logback.argument.StructuredArguments;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.MDC;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.Map;

@Service
@RequiredArgsConstructor
public class ApplicationEventService {
    private static final String CLASSNAME = "ApplicationEventService";
    private final Logger log = LoggerFactory.getLogger(ApplicationEventService.class);

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
        try {
            ApplicationEvent applicationEvent = ApplicationEvent.builder()
                .message(message).timestamp(Instant.now().toString())
                .source(ApplicationEventSource.PANEL.name()).type(type.name())
                .build();
            client.getClient().index(".utmstack-logs", applicationEvent);
        } catch (Exception e) {
            log.error(ctx + ": {}", e.getMessage());
        }
    }

    public void createEvent(String message, ApplicationEventType type, Map<String, Object> details) {
        String msg = String.format("%s: %s", MDC.get("context"), message);
        log.info( msg, StructuredArguments.keyValue("args", logContextBuilder.buildArgs(details)));
    }
}
