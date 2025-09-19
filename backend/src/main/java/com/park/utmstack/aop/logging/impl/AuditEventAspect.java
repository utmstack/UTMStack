package com.park.utmstack.aop.logging.impl;

import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.aop.utils.AuditContextExtractor;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.loggin.LogContextBuilder;
import com.park.utmstack.service.application_events.ApplicationEventService;
import lombok.RequiredArgsConstructor;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;
import java.util.Objects;

@Aspect
@Component
@RequiredArgsConstructor
public class AuditEventAspect {

    private final ApplicationEventService applicationEventService;
    private final LogContextBuilder logContextBuilder;
    private final List<AuditContextExtractor> extractors;

    @Around("@annotation(auditEvent)")
    public Object logAuditEvent(ProceedingJoinPoint joinPoint, AuditEvent auditEvent) throws Throwable {
        Map<String, Object> args = logContextBuilder.buildArgs();
        for (AuditContextExtractor extractor : extractors) {
            args.putAll(extractor.extract(joinPoint));
        }

        applicationEventService.createEvent(auditEvent.attemptMessage(), auditEvent.attemptType(), args);

        Object result = joinPoint.proceed();

        if (!auditEvent.successType().equals(ApplicationEventType.UNDEFINED)) {
            applicationEventService.createEvent(auditEvent.successMessage(), auditEvent.successType(), args);
        }

        return result;
    }

}

