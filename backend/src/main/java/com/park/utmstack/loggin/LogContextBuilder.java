package com.park.utmstack.loggin;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.MDC;

import java.util.*;
import java.util.stream.Collectors;

public class LogContextBuilder {

    private static final ObjectMapper mapper = new ObjectMapper();

    public static void init(String traceCode, Map<String, Object> args, Throwable exception) {
        MDC.put("code", traceCode != null ? traceCode : UUID.randomUUID().toString().replace("-", ""));
        MDC.put("args", serializeArgs(args));
        MDC.put("trace", serializeStackTrace(exception));
    }

    public static void clear() {
        MDC.remove("code");
        MDC.remove("args");
        MDC.remove("trace");
    }

    private static String serializeArgs(Map<String, Object> args) {
        try {
            return args != null ? mapper.writeValueAsString(args) : "{}";
        } catch (Exception e) {
            return "{\"error\":\"Failed to serialize args\"}";
        }
    }

    private static String serializeStackTrace(Throwable ex) {
        if (ex == null) return "[]";
        List<String> traceList = Arrays.stream(ex.getStackTrace())
                .map(ste -> ste.getClassName() + "." + ste.getMethodName() + " " + ste.getLineNumber())
                .collect(Collectors.toList());
        try {
            return mapper.writeValueAsString(traceList);
        } catch (Exception e) {
            return "[\"Failed to serialize stack trace\"]";
        }
    }
}

