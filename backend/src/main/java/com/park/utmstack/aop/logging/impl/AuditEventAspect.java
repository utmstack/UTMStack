package com.park.utmstack.aop.logging.impl;

import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.aop.utils.AuditContextExtractor;
import com.park.utmstack.loggin.LogContextBuilder;
import com.park.utmstack.service.application_events.ApplicationEventService;
import lombok.RequiredArgsConstructor;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;

@Aspect
@Component
@RequiredArgsConstructor
public class AuditEventAspect {

    private final ApplicationEventService applicationEventService;
    private final LogContextBuilder logContextBuilder;
    private final List<AuditContextExtractor> extractors;

    @Around("@annotation(auditEvent)")
    public Object logAuditEvent(ProceedingJoinPoint joinPoint, AuditEvent auditEvent) throws Throwable {
        Object result = joinPoint.proceed();

        Map<String, Object> args = logContextBuilder.buildArgs();
        for (AuditContextExtractor extractor : extractors) {
            args.putAll(extractor.extract(joinPoint));
        }
        applicationEventService.createEvent(auditEvent.message(), auditEvent.value(), args);

        return result;
    }
}

