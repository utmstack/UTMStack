package com.park.utmstack.repository.idp_provider;

import org.springframework.context.annotation.Configuration;
import org.springframework.security.oauth2.client.registration.ClientRegistration;
import org.springframework.security.oauth2.client.registration.ClientRegistrationRepository;
import org.springframework.security.oauth2.core.AuthorizationGrantType;
import org.springframework.security.oauth2.core.ClientAuthenticationMethod;
import org.springframework.util.StringUtils;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Configuration
public class OAuth2ClientRegistrationRepository implements ClientRegistrationRepository {

    private final Map<String, ClientRegistration> registrations = new ConcurrentHashMap<>();
    private final IdentityProviderConfigRepository jpaClientRegistrationRepository;

    public OAuth2ClientRegistrationRepository(IdentityProviderConfigRepository jpaClientRegistrationRepository) {

        this.jpaClientRegistrationRepository = jpaClientRegistrationRepository;

        this.jpaClientRegistrationRepository.findAll().forEach(entity -> {
            ClientRegistration registration = ClientRegistration.withRegistrationId(entity.getProviderType().name().toLowerCase())
                    .clientId(entity.getClientId())
                    .clientSecret(entity.getClientSecret())
                    .clientAuthenticationMethod(ClientAuthenticationMethod.CLIENT_SECRET_BASIC)
                    .authorizationGrantType(AuthorizationGrantType.AUTHORIZATION_CODE)
                    .redirectUri(entity.getRedirectUri())
                    .scope(StringUtils.commaDelimitedListToStringArray(entity.getScopes()))
                    .authorizationUri(entity.getAuthUri())
                    .tokenUri(entity.getTokenUri())
                    .clientName(entity.getProviderType().name().toLowerCase())
                    .build();
            registrations.put(entity.getProviderType().name().toLowerCase(), registration);
        });

    }

    @Override
    public ClientRegistration findByRegistrationId(String registrationId) {
        return registrations.get(registrationId);
    }
}
