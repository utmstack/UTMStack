package com.park.utmstack.loggin;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import net.logstash.logback.argument.StructuredArguments;
import org.slf4j.MDC;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import javax.servlet.*;
import javax.servlet.http.*;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;

/*@Component*/
@Slf4j
public class LoggingContextFilter extends OncePerRequestFilter {

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
            throws ServletException, IOException {
        MDC.clear();

        try {
            String username = Optional.ofNullable(SecurityContextHolder.getContext().getAuthentication())
                    .map(Authentication::getName)
                    .orElse("anonymous");

            Map<String, Object> args = new HashMap<>();
            args.put("username", username);
            args.put("requestId", UUID.randomUUID().toString());
            args.put("path", request.getRequestURI());
            args.put("method", request.getMethod());
            args.put("remoteAddr", request.getRemoteAddr());

            log.debug("Request info: {}", StructuredArguments.keyValue("args", args));

            filterChain.doFilter(request, response);
        } finally {
            MDC.clear();
        }
    }
}/**/

