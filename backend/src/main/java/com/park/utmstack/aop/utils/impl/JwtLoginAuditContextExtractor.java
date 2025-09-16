package com.park.utmstack.aop.utils.impl;

import com.park.utmstack.aop.utils.AuditContextExtractor;
import com.park.utmstack.web.rest.vm.LoginVM;
import org.aspectj.lang.ProceedingJoinPoint;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

@Component
public class JwtLoginAuditContextExtractor implements AuditContextExtractor {

    @Override
    public Map<String, Object> extract(ProceedingJoinPoint joinPoint) {
        Map<String, Object> context = new HashMap<>();

        for (Object arg : joinPoint.getArgs()) {
            if (arg instanceof LoginVM loginVM) {
                context.put("loginAttempt", loginVM.getUsername());
            }
        }

        return context;
    }
}

