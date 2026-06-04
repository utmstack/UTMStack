package com.park.utmstack.service.login_attempts;

import com.google.common.cache.CacheBuilder;
import com.google.common.cache.CacheLoader;
import com.google.common.cache.LoadingCache;
import org.jetbrains.annotations.NotNull;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import javax.servlet.http.HttpServletRequest;
import java.util.concurrent.TimeUnit;

@Service
public class LoginAttemptService {
    private static final String CLASSNAME = "LoginAttemptService";

    public static final int MAX_ATTEMPT = 10;

    private final LoadingCache<String, Integer> attemptsCache;
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

    public String getClientIP() {
        final String ctx = CLASSNAME + ".getClientIP";
        try {
            String xfHeader = request.getHeader("X-Forwarded-For");
            if (StringUtils.hasText(xfHeader))
                return xfHeader.split(",")[0];
            return request.getRemoteAddr();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}
