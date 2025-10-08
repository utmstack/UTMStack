package com.park.utmstack.aop.logging.impl;

import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.aop.logging.NoLogException;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.shared_types.ApplicationLayer;
import com.park.utmstack.loggin.LogContextBuilder;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.logstash.logback.argument.StructuredArguments;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.reflect.MethodSignature;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

@Aspect
@Component
@Slf4j
@RequiredArgsConstructor
public class AuditAspect {

    private final ApplicationEventService applicationEventService;
    private final LogContextBuilder logContextBuilder;

    @Around("@annotation(auditEvent)")
    public Object logAuditEvent(ProceedingJoinPoint joinPoint, AuditEvent auditEvent) throws Throwable {
        return handleAudit(joinPoint, auditEvent.attemptType(), auditEvent.successType(),
                auditEvent.attemptMessage(), auditEvent.successMessage());
    }

    private Object handleAudit(ProceedingJoinPoint joinPoint,
                               ApplicationEventType attemptType,
                               ApplicationEventType successType,
                               String attemptMessage,
                               String successMessage) throws Throwable {

        MethodSignature signature = (MethodSignature) joinPoint.getSignature();
        String context = signature.getDeclaringType().getSimpleName() + "." + signature.getMethod().getName();
        MDC.put("context", context);

        Map<String, Object> extra = extractAuditData(joinPoint.getArgs());
        extra.put("layer", ApplicationLayer.CONTROLLER.getValue());

        try {
            applicationEventService.createEvent(attemptMessage, attemptType, extra);

            Object result = joinPoint.proceed();

            if (successType != ApplicationEventType.UNDEFINED) {
                applicationEventService.createEvent(successMessage, successType, extra);
            }

            return result;

        } catch (Exception e) {
            if (!e.getClass().isAnnotationPresent(NoLogException.class)) {
                String msg = String.format("%s: %s", context, e.getMessage());
                log.error(msg, e, StructuredArguments.keyValue("args", logContextBuilder.buildArgs(e)));
            }

            throw e;
        }
    }

    private Map<String, Object> extractAuditData(Object[] args) {
        Map<String, Object> extra = new HashMap<>();
        for (Object arg : args) {
            if (arg instanceof AuditableDTO auditable) {
                extra.putAll(auditable.toAuditMap());
            }
        }
        return extra;
    }
}

