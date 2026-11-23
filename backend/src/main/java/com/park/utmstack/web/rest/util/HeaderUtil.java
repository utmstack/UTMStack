package com.park.utmstack.web.rest.util;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;

/**
 * Utility class for HTTP headers creation.
 */
public final class HeaderUtil {

    private static final Logger log = LoggerFactory.getLogger(HeaderUtil.class);

    private static final String APPLICATION_NAME = "UtmStack";

    private HeaderUtil() {
    }

    static String sanitizeHeaderValue(String value) {
        if (value == null) {
            return "";
        }
        // Tomcat 11 strictly rejects CR/LF and control characters in HTTP response headers.
        return value.replaceAll("[\\r\\n\\u0000-\\u001F\\u007F]+", " ").trim();
    }

    public static HttpHeaders createAlert(String message, String param) {
        HttpHeaders headers = new HttpHeaders();
        headers.add("X-" + APPLICATION_NAME + "-alert", sanitizeHeaderValue(message));
        headers.add("X-" + APPLICATION_NAME + "-params", sanitizeHeaderValue(param));
        return headers;
    }

    public static HttpHeaders createEntityCreationAlert(String entityName, String param) {
        return createAlert("A new " + entityName + " is created with identifier " + param, param);
    }

    public static HttpHeaders createEntityUpdateAlert(String entityName, String param) {
        return createAlert("A " + entityName + " is updated with identifier " + param, param);
    }

    public static HttpHeaders createEntityDeletionAlert(String entityName, String param) {
        return createAlert("A " + entityName + " is deleted with identifier " + param, param);
    }

    public static HttpHeaders createFailureAlert(String entityName, String errorKey, String defaultMessage) {
       /* log.error("Entity processing failed, {}", defaultMessage);*/
        HttpHeaders headers = new HttpHeaders();
        headers.add("X-" + APPLICATION_NAME + "-error", sanitizeHeaderValue(defaultMessage));
        headers.add("X-" + APPLICATION_NAME + "-params", sanitizeHeaderValue(entityName));
        return headers;
    }

    public static HttpHeaders createFailureAlertWithCode(String message, String params, String code) {
        HttpHeaders headers = new HttpHeaders();
        headers.add("X-" + APPLICATION_NAME + "-error", sanitizeHeaderValue(message));
        headers.add("X-" + APPLICATION_NAME + "-params", sanitizeHeaderValue(params));
        headers.add("X-" + APPLICATION_NAME + "-code", sanitizeHeaderValue(code));
        return headers;
    }
}
