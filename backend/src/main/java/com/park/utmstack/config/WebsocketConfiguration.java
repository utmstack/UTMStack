package com.park.utmstack.config;


import com.park.utmstack.security.jwt.TokenProvider;
import org.jetbrains.annotations.NotNull;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.server.ServerHttpRequest;
import org.springframework.http.server.ServerHttpResponse;
import org.springframework.http.server.ServletServerHttpRequest;
import org.springframework.messaging.simp.config.MessageBrokerRegistry;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.util.StringUtils;
import org.springframework.web.socket.WebSocketHandler;
import org.springframework.web.socket.config.annotation.EnableWebSocketMessageBroker;
import org.springframework.web.socket.config.annotation.StompEndpointRegistry;
import org.springframework.web.socket.config.annotation.WebSocketMessageBrokerConfigurer;
import org.springframework.web.socket.server.HandshakeInterceptor;
import org.springframework.web.socket.server.support.DefaultHandshakeHandler;

import java.util.Map;

@Configuration
@EnableWebSocketMessageBroker
public class WebsocketConfiguration implements WebSocketMessageBrokerConfigurer {
    private static final String CLASSNAME = "WebsocketConfiguration";
    private static final String IP_ADDRESS = "IP_ADDRESS";
    private final TokenProvider tokenProvider;

    public WebsocketConfiguration(TokenProvider tokenProvider) {
        this.tokenProvider = tokenProvider;
    }

    @Override
    public void configureMessageBroker(MessageBrokerRegistry config) {
        config.enableSimpleBroker("/topic");
        config.setApplicationDestinationPrefixes("/app");
    }

    @Override
    public void registerStompEndpoints(StompEndpointRegistry registry) {
        registry.addEndpoint("/ws")
            .addInterceptors(httpSessionHandshakeInterceptor())
            .setHandshakeHandler(defaultHandshakeHandler())
            .setAllowedOriginPatterns("*")
            .withSockJS();
    }

    private DefaultHandshakeHandler defaultHandshakeHandler() {
        return new DefaultHandshakeHandler() {
            @Override
            protected UsernamePasswordAuthenticationToken determineUser(@NotNull ServerHttpRequest request,
                                                                        @NotNull WebSocketHandler wsHandler,
                                                                        @NotNull Map<String, Object> attributes) {
                final String ctx = CLASSNAME + ".determineUser";
                final String ACCESS_TOKEN = "access_token";

                String query = request.getURI().getQuery();
                if (!StringUtils.hasText(query) || !query.contains(ACCESS_TOKEN))
                    throw new BadCredentialsException(ctx + ": Access token not found");

                String[] accessToken = query.split("=");
                if (accessToken.length != 2 || !accessToken[0].equals(ACCESS_TOKEN) || !StringUtils.hasText(accessToken[1]))
                    throw new BadCredentialsException(ctx + ": Access token not found");

                if (tokenProvider.validateToken(accessToken[1]))
                    return tokenProvider.getAuthentication(accessToken[1]);

                throw new BadCredentialsException("Invalid token");
            }
        };
    }

    public HandshakeInterceptor httpSessionHandshakeInterceptor() {
        return new HandshakeInterceptor() {
            @Override
            public boolean beforeHandshake(@NotNull ServerHttpRequest request, @NotNull ServerHttpResponse response,
                                           @NotNull WebSocketHandler wsHandler, @NotNull Map<String, Object> attributes) {
                if (request instanceof ServletServerHttpRequest)
                    attributes.put(IP_ADDRESS, request.getRemoteAddress());
                return true;
            }

            @Override
            public void afterHandshake(@NotNull ServerHttpRequest request, @NotNull ServerHttpResponse response,
                                       @NotNull WebSocketHandler wsHandler, Exception exception) {
            }
        };
    }

}
