package com.park.utmstack.service.idp_provider;

import com.park.utmstack.domain.idp_provider.CustomOAuth2User;
import com.park.utmstack.repository.UserRepository;
import org.springframework.security.oauth2.client.userinfo.DefaultOAuth2UserService;
import org.springframework.security.oauth2.client.userinfo.OAuth2UserRequest;
import org.springframework.security.oauth2.core.OAuth2AuthenticationException;
import org.springframework.security.oauth2.core.user.OAuth2User;
import org.springframework.stereotype.Service;

@Service
public class CustomOAuth2UserService extends DefaultOAuth2UserService {

    private final UserRepository userRepository;

    public CustomOAuth2UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    @Override
    public OAuth2User loadUser(OAuth2UserRequest userRequest) throws OAuth2AuthenticationException {
        OAuth2User oAuth2User = super.loadUser(userRequest);
        String email = oAuth2User.getAttribute("email");

        return userRepository.findOneByEmailIgnoreCase(email)
                .map(user -> new CustomOAuth2User(oAuth2User, user))
                .orElseThrow(() -> new OAuth2AuthenticationException("User with email " + email + " not found"));
    }
}
