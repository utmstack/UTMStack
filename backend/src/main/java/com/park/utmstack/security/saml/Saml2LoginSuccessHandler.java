package com.park.utmstack.security.saml;

import com.park.utmstack.security.jwt.TokenProvider;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.saml2.provider.service.authentication.Saml2AuthenticatedPrincipal;
import org.springframework.security.web.authentication.AuthenticationSuccessHandler;
import org.springframework.web.util.UriComponentsBuilder;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.net.URI;

import static com.park.utmstack.config.Constants.FRONT_BASE_URL;

/**
 * Success handler for SAML2 login.
 * Extracts NameID and attributes from the SAML assertion,
 * generates a JWT, and redirects to the frontend with the token.
 */
public class Saml2LoginSuccessHandler implements AuthenticationSuccessHandler {

    private final TokenProvider tokenProvider;

    public Saml2LoginSuccessHandler(TokenProvider tokenProvider) {
        this.tokenProvider = tokenProvider;
    }

    @Override
    public void onAuthenticationSuccess(HttpServletRequest request,
                                        HttpServletResponse response,
                                        Authentication authentication) throws IOException {
        Saml2AuthenticatedPrincipal samlUser = (Saml2AuthenticatedPrincipal) authentication.getPrincipal();

        // Extract NameID (default identifier)
        String username = samlUser.getName();

        // Example: extract email attribute if provided by IdP
        String email = samlUser.getFirstAttribute("email");

        UsernamePasswordAuthenticationToken auth =
                new UsernamePasswordAuthenticationToken(username, null, authentication.getAuthorities());

        SecurityContextHolder.getContext().setAuthentication(auth);

        // Generate JWT
        String token = tokenProvider.createToken(auth, false, true);

        // Redirect to frontend with token
        URI redirectUri = UriComponentsBuilder.fromHttpUrl(FRONT_BASE_URL)
                .queryParam("token", token)
                .build().toUri();

        response.sendRedirect(redirectUri.toString());
    }
}
