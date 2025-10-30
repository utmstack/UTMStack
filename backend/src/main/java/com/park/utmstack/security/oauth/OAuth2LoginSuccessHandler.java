package com.park.utmstack.security.oauth;

import com.park.utmstack.domain.idp_provider.CustomOAuth2User;
import com.park.utmstack.security.jwt.TokenProvider;
import org.springframework.security.core.Authentication;
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
        CustomOAuth2User oAuth2User = (CustomOAuth2User) authentication.getPrincipal();

        String jwt = tokenProvider.createToken(authentication, false, true);

        response.setContentType("application/json");
        response.getWriter().write("{\"id_token\": \"" + jwt + "\"}");
    }
}

