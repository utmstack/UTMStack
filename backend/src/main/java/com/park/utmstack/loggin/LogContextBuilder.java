package com.park.utmstack.loggin;

import com.park.utmstack.domain.User;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.UserService;
import lombok.RequiredArgsConstructor;
import net.logstash.logback.argument.StructuredArgument;
import org.slf4j.MDC;
import net.logstash.logback.argument.StructuredArguments;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.Map;
import java.util.Objects;

@Component
@RequiredArgsConstructor
public class LogContextBuilder {

    private final UserService userService;

    public Map<String, Object> buildArgs(Exception e, HttpServletRequest request) {
        Map<String, Object> args = new HashMap<>();

        String userName = SecurityUtils.getCurrentUserLogin().orElse("anonymous");

        if (Objects.nonNull(e.getCause())) {
            args.put("cause", e.getCause().toString());
        }

        args.put("username", userName);
        args.put("method", request.getMethod());
        args.put("path", request.getRequestURI());
        args.put("remoteAddr", request.getRemoteAddr());
        args.put("context", MDC.get("context"));

      return args;
    }
}
