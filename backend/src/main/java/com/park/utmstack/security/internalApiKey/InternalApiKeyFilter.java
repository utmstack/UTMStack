package com.park.utmstack.security.internalApiKey;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.lang.NonNull;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource;
import org.springframework.util.StringUtils;
import org.springframework.web.filter.OncePerRequestFilter;

import javax.servlet.FilterChain;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;

public class InternalApiKeyFilter extends OncePerRequestFilter {
    private static final String CLASSNAME = "InternalApiKeyFilter";
    private final Logger log = LoggerFactory.getLogger(InternalApiKeyFilter.class);
    private static final String API_KEY_HEADER = "Utm-Internal-Key";
    private static Boolean apiKeyHeaderInUse=false;

    private final InternalApiKeyProvider internalApiKeyProvider;

    public InternalApiKeyFilter(InternalApiKeyProvider internalApiKeyProvider) {
        this.internalApiKeyProvider = internalApiKeyProvider;
    }

    @Override
    protected void doFilterInternal(@NonNull HttpServletRequest request,
                                    @NonNull HttpServletResponse response,
                                    @NonNull FilterChain filterChain) throws ServletException, IOException {
        apiKeyHeaderInUse = false;
        final String ctx = CLASSNAME + ".doFilterInternal";
        String apiKeyHeader = request.getHeader(API_KEY_HEADER);
        String envApiKey = System.getenv("INTERNAL_KEY");

        if (!StringUtils.hasText(envApiKey)) {
            log.error(ctx + ": The environment variable that stores the internal communication key does not exist or has no value");
        } else if (StringUtils.hasText(apiKeyHeader) && apiKeyHeader.equals(envApiKey)) {
            UsernamePasswordAuthenticationToken authentication = internalApiKeyProvider.getAuthentication(apiKeyHeader);
            authentication.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));
            SecurityContextHolder.getContext().setAuthentication(authentication);
            apiKeyHeaderInUse = true;
        }
        filterChain.doFilter(request, response);
    }
    public static Boolean isApiKeyHeaderInUse(){
        return apiKeyHeaderInUse;
    }
}

