package com.park.utmstack.config;

import com.park.utmstack.security.AuthoritiesConstants;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.support.ChannelInterceptor;
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
 *
 * <p>{@code @EnableWebSocketSecurity} also installs a default
 * {@code XorCsrfChannelInterceptor} that rejects a STOMP CONNECT when the
 * websocket session carries no CSRF token. This app authenticates the websocket
 * handshake with a JWT in the query string (see WebsocketConfiguration's
 * determineUser) and runs http.csrf().disable() globally (stateless, no
 * cookie/session), so a CSRF token is never present and the default would break
 * every real-time connection. We therefore register an explicit no-op
 * {@code csrfChannelInterceptor} (the documented override point) to restore the
 * pre-Boot-4 sameOriginDisabled()=true posture. Message authorization for the
 * /user/topic/** and /app/command/** destinations is still enforced by the
 * messageAuthorizationManager bean below.
 */
@Configuration
@EnableWebSocketSecurity
public class WebsocketSecurityConfiguration {

    /**
     * No-op websocket CSRF interceptor. The bean name "csrfChannelInterceptor" is the
     * exact override point WebSocketMessageBrokerSecurityConfiguration looks up; without
     * it the Security 7 default (XorCsrfChannelInterceptor) is used, which throws
     * MissingCsrfTokenException on CONNECT for this JWT-authenticated, csrf-disabled app.
     */
    @Bean(name = "csrfChannelInterceptor")
    public ChannelInterceptor csrfChannelInterceptor() {
        return new ChannelInterceptor() { };
    }

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
