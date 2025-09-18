package com.park.utmstack.aop.logging.impl;

import com.park.utmstack.config.Constants;
import com.park.utmstack.loggin.LogContextBuilder;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.logstash.logback.argument.StructuredArguments;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;

@Aspect
@Component
@RequiredArgsConstructor
@Slf4j
public class LoggingMethodAspect {
    private final LogContextBuilder logContextBuilder;

    @Around("@annotation(com.park.utmstack.aop.logging.Loggable)")
    public Object logExecution(ProceedingJoinPoint joinPoint) throws Throwable {
        String traceId = MDC.get(Constants.TRACE_ID_KEY);
        String methodName = joinPoint.getSignature().toShortString();
        long start = System.currentTimeMillis();

        try {
            Object result = joinPoint.proceed();
            long duration = System.currentTimeMillis() - start;
            String msg = String.format("Method %s executed successfully in %sms", methodName, duration);
            log.info( msg, StructuredArguments.keyValue("args", logContextBuilder.buildArgs(methodName, String.valueOf(duration))));
            return result;
        } catch (Exception ex) {
            String msg = String.format("%s Method %s failed: 5s", traceId, methodName);
            log.error(msg, ex, StructuredArguments.keyValue("args", logContextBuilder.buildArgs(ex)));
            throw ex;
        }
    }
}
