package com.park.utmstack.aop.logging.impl;

import lombok.extern.slf4j.Slf4j;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;

@Aspect
@Component
@Slf4j
public class LoggingMethodAspect {
    @Around("@annotation(com.park.utmstack.aop.logging.Loggable)")
    public Object logExecution(ProceedingJoinPoint joinPoint) throws Throwable {
        String traceId = MDC.get("traceId");
        String methodName = joinPoint.getSignature().toShortString();
        long start = System.currentTimeMillis();

        try {
            Object result = joinPoint.proceed();
            long duration = System.currentTimeMillis() - start;
            log.debug("[{}] Method {} executed successfully in {}ms", traceId, methodName, duration);
            return result;
        } catch (Exception ex) {
            log.error("[{}] Method {} failed: {}", traceId, methodName, ex.getMessage(), ex);
            throw ex;
        }
    }
}
