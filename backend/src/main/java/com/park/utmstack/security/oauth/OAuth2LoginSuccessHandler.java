package com.park.utmstack.security.oauth;

import com.park.utmstack.domain.idp_provider.CustomOidcUser;
import com.park.utmstack.security.jwt.TokenProvider;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.web.authentication.AuthenticationSuccessHandler;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;


public class OAuth2LoginSuccessHandler implements AuthenticationSuccessHandler {

    private final TokenProvider tokenProvider;

    public OAuth2LoginSuccessHandler(TokenProvider tokenProvider) {
        this.tokenProvider = tokenProvider;
    }

    @Override
    public void onAuthenticationSuccess(HttpServletRequest request, HttpServletResponse response, Authentication authentication) throws IOException {
        CustomOidcUser oAuth2User = (CustomOidcUser) authentication.getPrincipal();

        UsernamePasswordAuthenticationToken auth =
                new UsernamePasswordAuthenticationToken(oAuth2User.getName(), null, oAuth2User.getAuthorities());

        SecurityContextHolder.getContext().setAuthentication(auth);

        String token = tokenProvider.createToken(auth, false, true);

        response.sendRedirect("http://localhost:4200?token=" + token);
    }
}

