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
import java.net.InetAddress;
import java.net.UnknownHostException;
import java.security.MessageDigest;
import java.util.Set;

/**
 * Grants full administrative authentication to machine-to-machine callers that
 * present the shared {@code Utm-Internal-Key} header.
 *
 * <p>Security constraints (CVE-2026-82042):
 * <ul>
 *   <li>The key is only honoured on a fixed allowlist of internal routes
 *       ({@link #INTERNAL_KEY_ALLOWED_PATHS}) — the exact set of endpoints the
 *       sibling containers (agent-manager, plugins, installer/updater) call.
 *       Any other path falls through to normal JWT authentication, so a stolen
 *       key can no longer be used to enumerate users, create admins, or read
 *       arbitrary data.</li>
 *   <li>Every accepted key-authenticated request is audit-logged with the
 *       caller IP, method, and path.</li>
 *   <li>The key comparison is constant-time to avoid a timing oracle.</li>
 *   <li>Optionally, the key is further restricted to configured source CIDRs
 *       when {@code INTERNAL_KEY_ALLOWED_CIDRS} is set (comma-separated). When
 *       unset, no IP restriction is applied, so deployments behind a proxy whose
 *       container networking the operator has not documented are not broken by
 *       default; the path allowlist above is the always-on control.</li>
 * </ul>
 */
public class InternalApiKeyFilter extends OncePerRequestFilter {
    private static final String CLASSNAME = "InternalApiKeyFilter";
    private final Logger log = LoggerFactory.getLogger(InternalApiKeyFilter.class);
    private static final String API_KEY_HEADER = "Utm-Internal-Key";
    private static final String ENV_ALLOWED_CIDRS = "INTERNAL_KEY_ALLOWED_CIDRS";
    private static Boolean apiKeyHeaderInUse = false;

    /**
     * The only endpoints the internal key may authenticate against. These map
     * 1:1 to the machine-to-machine calls made by agent-manager (via
     * config-client-go), the event-processor plugins (modules-config, feeds,
     * inputs, soc-ai) and the installer/updater. Anything not listed here is
     * rejected for key auth and must use a normal user JWT.
     */
    private static final Set<String> INTERNAL_KEY_ALLOWED_PATHS = Set.of(
            // agent-manager (config-client GetUTMConfig) + plugins (modules-config)
            "/api/utm-modules/module-details-decrypted",
            "/api/utm-modules/moduleDetails",
            // feeds + soc-ai: incident / alert reads and writes
            "/api/utm-incidents",
            "/api/utm-incidents/add-alerts",
            "/api/utm-incident-alerts",
            "/api/utm-configuration-parameters",
            "/api/utm-alerts/status",
            "/api/elasticsearch/search",
            // federation / panel connection keys
            "/api/federation-service/token",
            "/api/federation-service/generateApiToken"
    );

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
        } else if (StringUtils.hasText(apiKeyHeader) && MessageDigest.isEqual(
                apiKeyHeader.getBytes(java.nio.charset.StandardCharsets.UTF_8),
                envApiKey.getBytes(java.nio.charset.StandardCharsets.UTF_8))) {
            String path = requestPath(request);

            if (!INTERNAL_KEY_ALLOWED_PATHS.contains(path)) {
                log.warn(ctx + ": Internal key presented for a non-internal path and ignored. method={} path={} ip={}",
                        request.getMethod(), path, clientIp(request));
            } else if (!isSourceIpAllowed(request)) {
                log.warn(ctx + ": Internal key rejected, source IP not in " + ENV_ALLOWED_CIDRS + ". method={} path={} ip={}",
                        request.getMethod(), path, clientIp(request));
            } else {
                UsernamePasswordAuthenticationToken authentication = internalApiKeyProvider.getAuthentication(apiKeyHeader);
                authentication.setDetails(new WebAuthenticationDetailsSource().buildDetails(request));
                SecurityContextHolder.getContext().setAuthentication(authentication);
                apiKeyHeaderInUse = true;
                log.info(ctx + ": Request authenticated via internal key. method={} path={} ip={}",
                        request.getMethod(), path, clientIp(request));
            }
        }
        filterChain.doFilter(request, response);
    }

    public static Boolean isApiKeyHeaderInUse() {
        return apiKeyHeaderInUse;
    }

    /**
     * Returns the request path without query string or trailing slash, so
     * {@code /api/utm-modules/moduleDetails?nameShort=X} matches the allowlist
     * entry {@code /api/utm-modules/moduleDetails}.
     */
    private static String requestPath(HttpServletRequest request) {
        String uri = request.getRequestURI();
        if (uri == null) {
            return "";
        }
        // Strip context path if the app is deployed under one.
        String contextPath = request.getContextPath();
        if (contextPath != null && !contextPath.isEmpty() && uri.startsWith(contextPath)) {
            uri = uri.substring(contextPath.length());
        }
        if (uri.length() > 1 && uri.endsWith("/")) {
            uri = uri.substring(0, uri.length() - 1);
        }
        return uri;
    }

    /**
     * Enforces an optional source-IP allowlist. If {@code INTERNAL_KEY_ALLOWED_CIDRS}
     * is not set or is empty, every source IP is allowed (path allowlist still
     * applies). When set, the comma-separated list of CIDRs must contain the
     * caller IP.
     */
    private boolean isSourceIpAllowed(HttpServletRequest request) {
        String cidrs = System.getenv(ENV_ALLOWED_CIDRS);
        if (!StringUtils.hasText(cidrs)) {
            return true;
        }
        String ip = clientIp(request);
        if (ip == null) {
            return false;
        }
        for (String entry : cidrs.split(",")) {
            String c = entry.trim();
            if (c.isEmpty()) {
                continue;
            }
            if (ipInCidr(ip, c)) {
                return true;
            }
        }
        return false;
    }

    private static String clientIp(HttpServletRequest request) {
        String xf = request.getHeader("X-Forwarded-For");
        if (StringUtils.hasText(xf)) {
            // First (leftmost) is the originating client.
            return xf.split(",")[0].trim();
        }
        return request.getRemoteAddr();
    }

    private static boolean ipInCidr(String ip, String cidr) {
        try {
            InetAddress addr = InetAddress.getByName(ip);
            int slash = cidr.indexOf('/');
            if (slash < 0) {
                return cidr.equals(ip);
            }
            InetAddress net = InetAddress.getByName(cidr.substring(0, slash));
            int prefix = Integer.parseInt(cidr.substring(slash + 1));
            byte[] a = addr.getAddress();
            byte[] n = net.getAddress();
            if (a.length != n.length) {
                return false;
            }
            int fullBytes = prefix / 8;
            int remBits = prefix % 8;
            for (int i = 0; i < fullBytes; i++) {
                if (a[i] != n[i]) {
                    return false;
                }
            }
            if (remBits > 0) {
                int mask = (0xFF << (8 - remBits)) & 0xFF;
                if ((a[fullBytes] & mask) != (n[fullBytes] & mask)) {
                    return false;
                }
            }
            return true;
        } catch (UnknownHostException | NumberFormatException e) {
            return false;
        }
    }
}
