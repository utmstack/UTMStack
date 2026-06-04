package com.park.utmstack.loggin;

import com.park.utmstack.config.Constants;
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

    public Map<String, Object> buildArgs(String methodName, String duration) {
        Map<String, Object> args = new HashMap<>();
        args.put(Constants.USERNAME_KEY, SecurityUtils.getCurrentUserLogin().orElse("anonymous"));
        args.put(Constants.CONTEXT_KEY, methodName);
        args.put(Constants.DURATION_KEY, duration);
        args.put(Constants.TRACE_ID_KEY, MDC.get(Constants.TRACE_ID_KEY));
        return args;
    }

    public Map<String, Object> buildArgs(HttpServletRequest request) {
        Map<String, Object> args = new HashMap<>();
        args.put(Constants.USERNAME_KEY, SecurityUtils.getCurrentUserLogin().orElse("anonymous"));
        args.put(Constants.METHOD_KEY, request.getMethod());
        args.put(Constants.PATH_KEY, request.getRequestURI());
        args.put(Constants.REMOTE_ADDR_KEY, request.getRemoteAddr());
        args.put(Constants.CONTEXT_KEY, MDC.get(Constants.CONTEXT_KEY));
        args.put(Constants.TRACE_ID_KEY, MDC.get(Constants.TRACE_ID_KEY));
        return args;
    }

    public Map<String, Object> buildArgs(Exception e, HttpServletRequest request) {
        Map<String, Object> args = buildArgs(request);
        if (e != null && e.getCause() != null) {
            args.put(Constants.CAUSE_KEY, e.getCause().toString());
        }
        return args;
    }

    public Map<String, Object> buildArgs(Map<String, Object> extra) {
        Map<String, Object> base = buildArgs();
        return mergeArgs(base, extra);
    }

    private Map<String, Object> buildFallbackArgs(Exception e) {
        Map<String, Object> args = new HashMap<>();
        args.put(Constants.USERNAME_KEY, SecurityUtils.getCurrentUserLogin().orElse("anonymous"));
        args.put(Constants.CONTEXT_KEY, MDC.get(Constants.CONTEXT_KEY));
        args.put(Constants.TRACE_ID_KEY, MDC.get(Constants.TRACE_ID_KEY));
        if (e != null && e.getCause() != null) {
            args.put(Constants.CAUSE_KEY, e.getCause().toString());
        }
        return args;
    }

    private Map<String, Object> mergeArgs(Map<String, Object> base, Map<String, Object> extra) {
        if (extra != null) {
            base.putAll(extra);
        }
        return base;
    }
}
