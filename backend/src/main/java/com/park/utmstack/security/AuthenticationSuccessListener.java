package com.park.utmstack.security;

import com.park.utmstack.service.login_attempts.LoginAttemptService;
import org.jetbrains.annotations.NotNull;
import org.springframework.context.ApplicationListener;
import org.springframework.security.authentication.event.AuthenticationSuccessEvent;
import org.springframework.stereotype.Component;

@Component
public class AuthenticationSuccessListener implements ApplicationListener<AuthenticationSuccessEvent> {

    private static final String CLASSNAME = "AuthenticationSuccessListener";

    private final LoginAttemptService loginAttemptService;

    public AuthenticationSuccessListener(LoginAttemptService loginAttemptService) {
        this.loginAttemptService = loginAttemptService;
    }

    @Override
    public void onApplicationEvent(@NotNull AuthenticationSuccessEvent evt) {
        final String ctx = CLASSNAME + ".onApplicationEvent";
        try {
            loginAttemptService.registerSuccessfulLogin(loginAttemptService.getClientIP());
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}
