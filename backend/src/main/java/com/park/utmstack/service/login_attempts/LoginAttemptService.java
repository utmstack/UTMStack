package com.park.utmstack.service.login_attempts;

import com.google.common.cache.CacheBuilder;
import com.google.common.cache.CacheLoader;
import com.google.common.cache.LoadingCache;
import org.jetbrains.annotations.NotNull;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import javax.servlet.http.HttpServletRequest;
import java.net.InetAddress;
import java.util.concurrent.TimeUnit;

@Service
public class LoginAttemptService {
    private static final String CLASSNAME = "LoginAttemptService";

    public static final int MAX_ATTEMPT = 10;
    public static final int MAX_TFA_ATTEMPT = 5;

    private final LoadingCache<String, Integer> attemptsCache;
    private final LoadingCache<String, Integer> tfaAttemptsCache;
    private final HttpServletRequest request;

    public LoginAttemptService(HttpServletRequest request) {
        this.request = request;
        attemptsCache = CacheBuilder.newBuilder().expireAfterWrite(10, TimeUnit.MINUTES).build(new CacheLoader<>() {
            @NotNull
            @Override
            public Integer load(@NotNull final String key) {
                return 0;
            }
        });

        tfaAttemptsCache = CacheBuilder.newBuilder().expireAfterWrite(10, TimeUnit.MINUTES).build(new CacheLoader<>() {
            @NotNull
            @Override
            public Integer load(@NotNull final String key) {
                return 0;
            }
        });
    }

    public void registerFailedLogin(String clientIp) {
        final String ctx = CLASSNAME + ".registerFailedLogin";
        try {
            int attempts;
            try {
                attempts = attemptsCache.get(clientIp);
            } catch (Exception e) {
                attempts = 0;
            }
            attempts++;
            attemptsCache.put(clientIp, attempts);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    public void registerSuccessfulLogin(String clientIp) {
        final String ctx = CLASSNAME + ".registerSuccessfulLogin";
        try {
            attemptsCache.put(clientIp, 0);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    public boolean isBlocked() {
        final String ctx = CLASSNAME + ".isBlocked";
        try {
            return attemptsCache.get(getClientIP()) >= MAX_ATTEMPT;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Resolves the caller address used to key the lockout.
     *
     * <p>X-Forwarded-For is client-supplied, so taking its first entry let an
     * attacker move to a fresh bucket on every request (and pin the lockout on
     * someone else). The panel runs behind its own nginx, which appends the peer
     * address, so we walk the chain from the right and return the first address
     * that is not one of our own proxies. If the direct peer is not a trusted
     * proxy the header is ignored outright.
     */
    public String getClientIP() {
        final String ctx = CLASSNAME + ".getClientIP";
        try {
            String remoteAddr = request.getRemoteAddr();
            String xfHeader = request.getHeader("X-Forwarded-For");

            if (!StringUtils.hasText(xfHeader) || !isTrustedProxy(remoteAddr))
                return remoteAddr;

            String[] forwarded = xfHeader.split(",");
            for (int i = forwarded.length - 1; i >= 0; i--) {
                String candidate = forwarded[i].trim();
                if (StringUtils.hasText(candidate) && !isTrustedProxy(candidate))
                    return candidate;
            }
            return remoteAddr;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Proxies whose X-Forwarded-For entries are ours, not a caller's: loopback
     * and the private ranges the compose network lives on, overridable with
     * LOGIN_TRUSTED_PROXIES (comma-separated CIDRs).
     */
    private static boolean isTrustedProxy(String ip) {
        if (!StringUtils.hasText(ip))
            return false;

        String configured = System.getenv("LOGIN_TRUSTED_PROXIES");
        String[] cidrs = StringUtils.hasText(configured)
                ? configured.split(",")
                : new String[]{"127.0.0.0/8", "::1/128", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"};

        for (String cidr : cidrs) {
            if (ipInCidr(ip, cidr.trim()))
                return true;
        }
        return false;
    }

    private static boolean ipInCidr(String ip, String cidr) {
        try {
            if (!StringUtils.hasText(cidr))
                return false;

            int slash = cidr.indexOf('/');
            if (slash < 0)
                return cidr.equals(ip);

            byte[] address = InetAddress.getByName(ip).getAddress();
            byte[] network = InetAddress.getByName(cidr.substring(0, slash)).getAddress();
            int prefix = Integer.parseInt(cidr.substring(slash + 1));
            if (address.length != network.length)
                return false;

            int fullBytes = prefix / 8;
            int remainingBits = prefix % 8;
            for (int i = 0; i < fullBytes; i++) {
                if (address[i] != network[i])
                    return false;
            }
            if (remainingBits > 0) {
                int mask = (0xFF << (8 - remainingBits)) & 0xFF;
                return (address[fullBytes] & mask) == (network[fullBytes] & mask);
            }
            return true;
        } catch (Exception e) {
            return false;
        }
    }

    /**
     * Second-factor attempts are counted per user, not per IP: the code is six
     * digits, so an attacker holding the password could otherwise walk the whole
     * space by rotating source addresses.
     */
    public void registerFailedTfa(String login) {
        if (!StringUtils.hasText(login))
            return;
        int attempts;
        try {
            attempts = tfaAttemptsCache.get(login);
        } catch (Exception e) {
            attempts = 0;
        }
        tfaAttemptsCache.put(login, attempts + 1);
    }

    public void registerSuccessfulTfa(String login) {
        if (StringUtils.hasText(login))
            tfaAttemptsCache.put(login, 0);
    }

    public boolean isTfaBlocked(String login) {
        if (!StringUtils.hasText(login))
            return false;
        try {
            return tfaAttemptsCache.get(login) >= MAX_TFA_ATTEMPT;
        } catch (Exception e) {
            return false;
        }
    }
}
