package com.park.utmstack.util;

/**
 * Application-level "bad request" runtime exception, thrown by the correlation-config
 * services when a create/update/delete request is invalid (null id, system row, etc.)
 * and caught by the REST layer to produce a 400.
 *
 * <p>The original implementation used {@code io.undertow.util.BadRequestException},
 * but Boot 4.1.2 has no Undertow on the classpath (it runs on Tomcat). This is a
 * plain {@link RuntimeException} so the existing throw/catch/{@code getLocalizedMessage}
 * call sites in the services and REST controllers compile and behave identically.
 */
public class BadRequestException extends RuntimeException {

    private static final long serialVersionUID = 1L;

    public BadRequestException(String message) {
        super(message);
    }
}
