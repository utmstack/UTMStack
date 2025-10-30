package com.park.utmstack.config.oauth;

import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.oauth2.client.registration.ClientRegistrationRepository;

@Configuration
public class OAuth2ClientConfig {

    @Bean
    public ClientRegistrationRepository clientRegistrationRepository(IdentityProviderConfigRepository repo) {
        return new OAuth2ClientRegistrationRepository(repo);
    }
}
