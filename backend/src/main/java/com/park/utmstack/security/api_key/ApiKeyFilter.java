package com.park.utmstack.security.api_key;


import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.api_keys.ApiKey;
import com.park.utmstack.loggin.api_key.ApiKeyUsageLoggingService;
import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.service.api_key.ApiKeyService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.exceptions.ApiKeyInvalidAccessException;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.net.util.SubnetUtils;
import org.springframework.lang.NonNull;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.User;
import org.springframework.security.web.authentication.WebAuthenticationDetailsSource;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.springframework.web.filter.OncePerRequestFilter;
import org.springframework.web.util.ContentCachingRequestWrapper;
import org.springframework.web.util.ContentCachingResponseWrapper;

import javax.servlet.FilterChain;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.time.Instant;
import java.util.List;
import java.util.Objects;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.regex.Pattern;

import static com.park.utmstack.config.Constants.API_ENDPOINT_IGNORE;

@Slf4j
@Component
@AllArgsConstructor
public class ApiKeyFilter extends OncePerRequestFilter {

    private static final String LOG_USAGE_FLAG = "LOG_USAGE_DONE";
    private static final Pattern CIDR_PATTERN = Pattern.compile(
            "^((25[0-5]|2[0-4]\\d|1\\d{2}|[1-9]?\\d)\\.){3}(25[0-5]|2[0-4]\\d|1\\d{2}|[1-9]?\\d)/(\\d|[1-2]\\d|3[0-2])$"
    );


    private final UserRepository userRepository;
    private final ApiKeyService apiKeyService;
    private final ConcurrentMap<String, Boolean> invalidApiKeyBlackList = new ConcurrentHashMap<>();
    private final ConcurrentMap<String, SubnetUtils> cidrCache = new ConcurrentHashMap<>();
    private final ApiKeyUsageLoggingService apiKeyUsageLoggingService;


    @Override
    protected void doFilterInternal(@NonNull HttpServletRequest request,
                                    @NonNull HttpServletResponse response,
                                    @NonNull FilterChain filterChain) throws ServletException, IOException {

        if (API_ENDPOINT_IGNORE.contains(request.getRequestURI())) {
            filterChain.doFilter(request, response);
            return;
        }

        if (request.getAttribute(LOG_USAGE_FLAG) != null) {
            filterChain.doFilter(request, response);
            return;
        }

        String apiKey = request.getHeader(Constants.API_KEY_HEADER);

        if (!StringUtils.hasText(apiKey)) {
            filterChain.doFilter(request, response);
            return;
        }

        String ipAddress = request.getRemoteAddr();
        var key = getApiKey(apiKey);

        var wrappedRequest = new ContentCachingRequestWrapper(request);
        var wrappedResponse = new ContentCachingResponseWrapper(response);

        UsernamePasswordAuthenticationToken authentication;

        try {
            authentication = getAuthentication(key, ipAddress);
            authentication.setDetails(new WebAuthenticationDetailsSource().buildDetails(wrappedRequest));
            SecurityContextHolder.getContext().setAuthentication(authentication);

        } catch (ApiKeyInvalidAccessException e) {
            apiKeyUsageLoggingService.logUsage(wrappedRequest, wrappedResponse, key, ipAddress, e.getMessage());
            throw e;
        }

        filterChain.doFilter(wrappedRequest, wrappedResponse);
        wrappedResponse.copyBodyToResponse();

        apiKeyUsageLoggingService.logUsage(wrappedRequest, wrappedResponse, key, ipAddress, null);
    }

    private ApiKey getApiKey(String apiKey) {
        if (invalidApiKeyBlackList.containsKey(apiKey)) {
            log.warn("Access attempt with invalid API key (cached)");
            throw new ApiKeyInvalidAccessException("Invalid API key");
        }

        return apiKeyService.findOneByApiKey(apiKey)
                .orElseGet(() -> {
                    invalidApiKeyBlackList.put(apiKey, Boolean.TRUE);
                    log.warn("Access attempt with invalid API key (not found in DB)");
                    throw new ApiKeyInvalidAccessException("Invalid API key");
                });
    }

    public UsernamePasswordAuthenticationToken getAuthentication(ApiKey apiKey, String remoteIpAddress) {
        Objects.requireNonNull(apiKey, "API key must not be null");
        Objects.requireNonNull(remoteIpAddress, "Remote IP address must not be null");

        if (!allowAccessToRemoteIp(apiKey.getAllowedIp(), remoteIpAddress)) {
            log.warn("Access denied: IP [{}] not allowed for API key [{}]", remoteIpAddress, apiKey.getApiKey());
            throw new ApiKeyInvalidAccessException(
                    "Invalid IP address: " + remoteIpAddress + ". If you recognize this IP, add it to allowed IP list."
            );
        }

        if (apiKey.getExpiresAt() != null && !apiKey.getExpiresAt().isAfter(Instant.now())) {
            log.warn("Access denied: API key [{}] expired at {}", apiKey.getApiKey(), apiKey.getExpiresAt());
            throw new ApiKeyInvalidAccessException("API key expired at " + apiKey.getExpiresAt());
        }

        var userEntityOpt = userRepository.findById(apiKey.getUserId());
        if (userEntityOpt.isEmpty()) {
            log.warn("Access denied: User [{}] not found for API key [{}]", apiKey.getUserId(), apiKey.getApiKey());
            throw new ApiKeyInvalidAccessException("User not found for API key");
        }

        var userEntity = userEntityOpt.get();

        if (!userEntity.getActivated()) {
            log.warn("Access denied: User [{}] not activated", userEntity.getLogin());
            throw new ApiKeyInvalidAccessException("User not activated");
        }

        List<SimpleGrantedAuthority> authorities = userEntity.getAuthorities().stream()
                .map(auth -> new SimpleGrantedAuthority(auth.getName()))
                .toList();

        User principal = new User(userEntity.getLogin(), "", authorities);

        return new UsernamePasswordAuthenticationToken(principal, apiKey.getApiKey(), authorities);
    }

    public boolean allowAccessToRemoteIp(String allowedIpList, String remoteIp) {
        if (allowedIpList == null || allowedIpList.trim().isEmpty()) {
            return true;
        }
        String[] whitelistIps = allowedIpList.split(",");
        for (String ip : whitelistIps) {
            String allowed = ip.trim();
            if (allowed.isEmpty()) {
                continue;
            }
            if (CIDR_PATTERN.matcher(allowed).matches()) {
                try {
                    SubnetUtils subnetUtils = cidrCache.computeIfAbsent(allowed, key -> {
                        SubnetUtils su = new SubnetUtils(key);
                        su.setInclusiveHostCount(true);
                        return su;
                    });
                    if (subnetUtils.getInfo().isInRange(remoteIp)) {
                        return true;
                    }
                } catch (IllegalArgumentException e) {
                    log.error("Invalid CIDR notation: {}", allowed);
                }
            } else if (allowed.equals(remoteIp)) {
                return true;
            }
        }
        return false;
    }
}
