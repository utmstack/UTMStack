package com.park.utmstack.security.saml;

import org.springframework.security.core.AuthenticationException;
import org.springframework.security.web.authentication.AuthenticationFailureHandler;
import org.springframework.web.util.UriComponentsBuilder;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.net.URI;

import static com.park.utmstack.config.Constants.FRONT_BASE_URL;

/**
 * Failure handler for SAML2 login.
 * Redirects the user to the frontend with an error parameter.
 */
public class Saml2LoginFailureHandler implements AuthenticationFailureHandler {

    @Override
    public void onAuthenticationFailure(HttpServletRequest request,
                                        HttpServletResponse response,
                                        AuthenticationException exception) throws IOException {
        URI redirectUri = UriComponentsBuilder.fromHttpUrl(FRONT_BASE_URL)
                .queryParam("error", "saml2")
                .queryParam("message", exception.getMessage()) // optional: include error details
                .build().toUri();

        response.sendRedirect(redirectUri.toString());
    }
}
