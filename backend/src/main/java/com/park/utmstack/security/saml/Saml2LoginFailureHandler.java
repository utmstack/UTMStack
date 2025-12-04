package com.park.utmstack.security.saml;

import org.springframework.security.core.AuthenticationException;
import org.springframework.security.web.authentication.AuthenticationFailureHandler;
import org.springframework.web.util.UriComponentsBuilder;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.net.URI;
import java.util.Objects;

/**
 * Failure handler for SAML2 login.
 * Redirects the user to the frontend with an error parameter.
 */
public class Saml2LoginFailureHandler implements AuthenticationFailureHandler {

    @Override
    public void onAuthenticationFailure(HttpServletRequest request,
                                        HttpServletResponse response,
                                        AuthenticationException exception) throws IOException {

        String scheme = Objects.requireNonNullElse(request.getHeader("X-Forwarded-Proto"), request.getScheme());
        String host = Objects.requireNonNullElse(request.getHeader("Host"), request.getServerName());

        String frontBaseUrl = scheme + "://" + host;

        URI redirectUri = UriComponentsBuilder.fromHttpUrl(frontBaseUrl)
                .queryParam("error", "saml2")
                .queryParam("message", exception.getMessage())
                .build().toUri();

        response.sendRedirect(redirectUri.toString());
    }
}
