package com.park.utmstack.security;

import com.park.utmstack.service.login_attempts.LoginAttemptService;
import org.jetbrains.annotations.NotNull;
import org.springframework.context.ApplicationListener;
import org.springframework.security.authentication.event.AuthenticationFailureBadCredentialsEvent;
import org.springframework.stereotype.Component;

@Component
public class AuthenticationFailureListener implements ApplicationListener<AuthenticationFailureBadCredentialsEvent> {
    private static final String CLASSNAME = "AuthenticationFailureListener";
    private final LoginAttemptService loginAttemptService;

    public AuthenticationFailureListener(LoginAttemptService loginAttemptService) {
        this.loginAttemptService = loginAttemptService;
    }

    @Override
    public void onApplicationEvent(@NotNull AuthenticationFailureBadCredentialsEvent evt) {
        final String ctx = CLASSNAME + ".onApplicationEvent";
        try {
            loginAttemptService.registerFailedLogin(loginAttemptService.getClientIP());
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}
