package com.park.utmstack.loggin;

import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.util.RequestContextUtils;
import org.slf4j.MDC;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.Map;
import java.util.Objects;

@Component
public class LogContextBuilder {

    public Map<String, Object> buildArgs(Exception e) {
        return RequestContextUtils.getCurrentRequest()
                .map(request -> buildArgs(e, request))
                .orElse(buildFallbackArgs(e));
    }

    public Map<String, Object> buildArgs() {
        return RequestContextUtils.getCurrentRequest()
                .map(this::buildArgs)
                .orElse(buildFallbackArgs(null));
    }

    public Map<String, Object> buildArgs(HttpServletRequest request) {
        Map<String, Object> args = new HashMap<>();
        args.put("username", SecurityUtils.getCurrentUserLogin().orElse("anonymous"));
        args.put("method", request.getMethod());
        args.put("path", request.getRequestURI());
        args.put("remoteAddr", request.getRemoteAddr());
        args.put("context", MDC.get("context"));
        args.put("traceId", MDC.get("traceId"));
        return args;
    }

    public Map<String, Object> buildArgs(Exception e, HttpServletRequest request) {
        Map<String, Object> args = buildArgs(request);
        if (e != null && e.getCause() != null) {
            args.put("cause", e.getCause().toString());
        }
        return args;
    }

    private Map<String, Object> buildFallbackArgs(Exception e) {
        Map<String, Object> args = new HashMap<>();
        args.put("username", SecurityUtils.getCurrentUserLogin().orElse("anonymous"));
        args.put("context", MDC.get("context"));
        args.put("traceId", MDC.get("traceId"));
        if (e != null && e.getCause() != null) {
            args.put("cause", e.getCause().toString());
        }
        return args;
    }
}
