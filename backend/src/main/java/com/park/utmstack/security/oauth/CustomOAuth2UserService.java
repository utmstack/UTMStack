package com.park.utmstack.security.oauth;

import com.park.utmstack.domain.idp_provider.CustomOidcUser;
import com.park.utmstack.repository.UserRepository;
import org.springframework.security.oauth2.client.oidc.userinfo.OidcUserRequest;
import org.springframework.security.oauth2.client.oidc.userinfo.OidcUserService;
import org.springframework.security.oauth2.client.userinfo.OAuth2UserService;
import org.springframework.security.oauth2.core.OAuth2AuthenticationException;
import org.springframework.security.oauth2.core.oidc.user.OidcUser;

public class CustomOAuth2UserService implements OAuth2UserService<OidcUserRequest, OidcUser> {

    private final UserRepository userRepository;

    public CustomOAuth2UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    @Override
    public OidcUser loadUser(OidcUserRequest userRequest) throws OAuth2AuthenticationException {
        OidcUserService delegate = new OidcUserService();
        OidcUser oidcUser = delegate.loadUser(userRequest);

        String email = oidcUser.getAttribute("email");
        if (email == null) {
            throw new OAuth2AuthenticationException("Email not found in OIDC token");
        }

        return userRepository.findOneWithAuthoritiesByEmail(email)
                .map(user -> new CustomOidcUser(oidcUser, user))
                .orElseThrow(() -> new OAuth2AuthenticationException("User with email " + email + " is not registered"));
    }

}

