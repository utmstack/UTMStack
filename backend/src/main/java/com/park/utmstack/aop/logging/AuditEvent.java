package com.park.utmstack.aop.logging;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface AuditEvent {
    ApplicationEventType attemptType();
    String attemptMessage();

    ApplicationEventType successType();
    String successMessage();
}

