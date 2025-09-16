package com.park.utmstack.aop.utils;

import org.aspectj.lang.ProceedingJoinPoint;

import java.util.Map;

public interface AuditContextExtractor {
    Map<String, Object> extract(ProceedingJoinPoint joinPoint);
}

