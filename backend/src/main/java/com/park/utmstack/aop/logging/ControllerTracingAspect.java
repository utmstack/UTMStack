package com.park.utmstack.aop.logging;

import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.reflect.MethodSignature;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;

@Aspect
@Component
public class ControllerTracingAspect {

    @Around("within(@org.springframework.web.bind.annotation.RestController *)")
    public Object enrichMDC(ProceedingJoinPoint joinPoint) throws Throwable {
        MethodSignature signature = (MethodSignature) joinPoint.getSignature();
        String context = signature.getDeclaringType().getSimpleName() + "." + signature.getMethod().getName();
        MDC.put("context", context);
       /* String username = Optional.ofNullable(SecurityContextHolder.getContext().getAuthentication())
                .map(Authentication::getName)
                .orElse("anonymous");
        MDC.put("username", username);*/

        return joinPoint.proceed();
    }
}

