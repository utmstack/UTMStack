package com.park.utmstack.config;

import com.park.utmstack.security.AuthoritiesConstants;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.security.authorization.AuthorizationManager;
import org.springframework.security.config.annotation.web.socket.EnableWebSocketSecurity;
import org.springframework.security.messaging.access.intercept.MessageMatcherDelegatingAuthorizationManager;
import org.springframework.security.messaging.access.intercept.MessageMatcherDelegatingAuthorizationManager.Builder;

/**
 * WebSocket message security (Spring Security 7 / Boot 4).
 *
 * <p>The pre-7 API (AbstractSecurityWebSocketMessageBrokerConfigurer +
 * MessageSecurityMetadataSourceRegistry) was removed. Security 7 exposes a
 * MessageMatcherDelegatingAuthorizationManager bean, auto-wired when
 * {@code @EnableWebSocketSecurity} is present, and the same matcher rules are
 * expressed with its builder DSL.
 */
@Configuration
@EnableWebSocketSecurity
public class WebsocketSecurityConfiguration {

    @Bean
    public AuthorizationManager<org.springframework.messaging.Message<?>> messageAuthorizationManager() {
        return new MessageMatcherDelegatingAuthorizationManager.Builder()
            .nullDestMatcher().authenticated()
            .simpDestMatchers("/user/topic/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
            .simpDestMatchers("/user/topic/**").authenticated()
            .simpDestMatchers("/app/command/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
            .simpDestMatchers("/app/command/**").authenticated()
            .simpTypeMatchers(SimpMessageType.MESSAGE, SimpMessageType.SUBSCRIBE).denyAll()
            .anyMessage().denyAll()
            .build();
    }
}
