package com.park.utmstack.security.saml;

import com.park.utmstack.config.Constants;
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
import org.springframework.security.web.authentication.AuthenticationSuccessHandler;
import org.springframework.util.StringUtils;
import org.springframework.web.util.UriComponentsBuilder;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.net.URI;
import java.util.Collection;
import java.util.Objects;

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


    @Override
    public void onAuthenticationSuccess(HttpServletRequest request,
                                        HttpServletResponse response,
                                        Authentication authentication) throws IOException {

        String frontBaseUrl = resolveFrontBaseUrl(request);

        Saml2AuthenticatedPrincipal samlUser = (Saml2AuthenticatedPrincipal) authentication.getPrincipal();
        String username = samlUser.getName();

        User user = userRepository.findOneByLogin(username)
                .orElseThrow(() -> {
                    log.warn("SAML2 authentication successful for '{}' but user not found in local database", username);
                    return new BadCredentialsException("User not provisioned in local system");
                });

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

        log.info("SAML2 login successful for user: {}", username);
        response.sendRedirect(redirectUri.toString());
    }

    /**
     * Builds the base URL the browser is sent back to after SAML login.
     *
     * <p>The configured panel URL wins. X-Forwarded-Host / X-Forwarded-Proto are
     * client-supplied: trusting them let anyone who could reach the backend with
     * their own headers steer the post-login redirect — which carries a token —
     * at a host of their choosing. Without configuration we fall back to the
     * request's own scheme and server name.
     */
    private String resolveFrontBaseUrl(HttpServletRequest request) {
        String configured = Constants.CFG.get(Constants.PROP_MAIL_BASE_URL);
        if (StringUtils.hasText(configured)) {
            return StringUtils.trimTrailingCharacter(configured.trim(), '/');
        }

        String scheme = request.isSecure() ? "https" : request.getScheme();
        return scheme + "://" + request.getServerName();
    }
}
