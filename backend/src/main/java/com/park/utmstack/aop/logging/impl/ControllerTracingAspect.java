package com.park.utmstack.aop.logging.impl;

import com.park.utmstack.aop.logging.NoLogException;
import com.park.utmstack.loggin.LogContextBuilder;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.logstash.logback.argument.StructuredArguments;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.reflect.MethodSignature;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;

@Aspect
@Slf4j
@Component
@RequiredArgsConstructor
public class ControllerTracingAspect {

    private final LogContextBuilder logContextBuilder;

    @Around("within(@org.springframework.web.bind.annotation.RestController *)")
    public Object enrichMDC(ProceedingJoinPoint joinPoint) throws Throwable {
        MethodSignature signature = (MethodSignature) joinPoint.getSignature();
        String context = signature.getDeclaringType().getSimpleName() + "." + signature.getMethod().getName();
        MDC.put("context", context);
        try {
            return joinPoint.proceed();
        } catch (Exception e) {
            if (!e.getClass().isAnnotationPresent(NoLogException.class)) {
                String msg = String.format("%s: %s", context, e.getMessage());
                log.error(msg, e, StructuredArguments.keyValue("args", logContextBuilder.buildArgs(e)));
            }
            throw e;
        }
    }
}

