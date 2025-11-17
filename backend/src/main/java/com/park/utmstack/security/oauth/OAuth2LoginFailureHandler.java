package com.park.utmstack.security.oauth;

import org.springframework.security.core.AuthenticationException;
import org.springframework.security.web.authentication.AuthenticationFailureHandler;
import org.springframework.web.util.UriComponentsBuilder;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.net.URI;

import static com.park.utmstack.config.Constants.FRONT_BASE_URL;

public class OAuth2LoginFailureHandler implements AuthenticationFailureHandler {
    @Override
    public void onAuthenticationFailure(HttpServletRequest request, HttpServletResponse response, AuthenticationException exception) throws IOException {
        URI redirectUri = UriComponentsBuilder.fromHttpUrl(FRONT_BASE_URL)
                .queryParam("error", "oauth2")
                .build().toUri();
        response.sendRedirect(redirectUri.toString());
    }
}
