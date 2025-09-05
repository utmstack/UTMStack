package com.park.utmstack.config;

import org.slf4j.MDC;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import javax.servlet.*;
import javax.servlet.http.*;
import java.io.IOException;
import java.util.Optional;
import java.util.UUID;

@Component
public class LoggingContextFilter extends OncePerRequestFilter {

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
            throws ServletException, IOException {

        try {
            String username = Optional.ofNullable(SecurityContextHolder.getContext().getAuthentication())
                    .map(Authentication::getName)
                    .orElse("anonymous");

            MDC.put("username", username);
            MDC.put("requestId", UUID.randomUUID().toString());
            MDC.put("path", request.getRequestURI());
            MDC.put("method", request.getMethod());
            MDC.put("remoteAddr", request.getRemoteAddr());

            filterChain.doFilter(request, response);
        } finally {
            MDC.clear();
        }
    }
}

