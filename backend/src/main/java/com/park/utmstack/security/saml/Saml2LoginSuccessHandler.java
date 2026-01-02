package com.park.utmstack.security.saml;

import com.park.utmstack.domain.User;
import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.security.jwt.TokenProvider;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.saml2.provider.service.authentication.Saml2AuthenticatedPrincipal;
import org.springframework.security.saml2.provider.service.authentication.Saml2Authentication;
import org.springframework.security.web.authentication.AuthenticationSuccessHandler;
import org.springframework.web.util.UriComponentsBuilder;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.net.URI;
import java.util.Collection;
import java.util.Objects;
import java.util.Optional;

import static com.park.utmstack.config.Constants.FRONT_BASE_URL;

/**
 * Success handler for SAML2 login.
 * Extracts NameID and attributes from the SAML assertion,
 * generates a JWT, and redirects to the frontend with the token.
 */

@RequiredArgsConstructor
@Slf4j
public class Saml2LoginSuccessHandler implements AuthenticationSuccessHandler {

    private final TokenProvider tokenProvider;
    private final UserRepository userRepository;
    private final Saml2LoginFailureHandler failureHandler;


    @Override
    public void onAuthenticationSuccess(HttpServletRequest request,
                                        HttpServletResponse response,
                                        Authentication authentication) throws IOException {

        String scheme = Objects.requireNonNullElse(request.getHeader("X-Forwarded-Proto"), request.getScheme());
        String host = Objects.requireNonNullElse(request.getHeader("Host"), request.getServerName());

        String frontBaseUrl = scheme + "://" + host;

        Saml2AuthenticatedPrincipal samlUser = (Saml2AuthenticatedPrincipal) authentication.getPrincipal();
        String username = samlUser.getName();

        User user = userRepository.findOneByLogin(username)
                .orElseThrow(() -> new BadCredentialsException("The provided credentials do not match any active user account."));

        Collection<? extends GrantedAuthority> authorities = Objects.requireNonNull(user.getAuthorities())
                .stream()
                .map(a -> new SimpleGrantedAuthority(a.getName()))
                .toList();

        UsernamePasswordAuthenticationToken auth =
                new UsernamePasswordAuthenticationToken(username, null, authorities);

        SecurityContextHolder.getContext().setAuthentication(auth);

        String token = tokenProvider.createToken(auth, false, true);

        URI redirectUri = UriComponentsBuilder.fromUriString(frontBaseUrl)
                .path("/")
                .queryParam("token", token)
                .build()
                .toUri();

        response.sendRedirect(redirectUri.toString());
    }
}
