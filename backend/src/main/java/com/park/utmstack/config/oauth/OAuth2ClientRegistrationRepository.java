package com.park.utmstack.config.oauth;

import com.park.utmstack.domain.idp_provider.enums.ClientAuthMethod;
import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import org.springframework.security.oauth2.client.registration.ClientRegistration;
import org.springframework.security.oauth2.client.registration.ClientRegistrationRepository;
import org.springframework.security.oauth2.core.AuthorizationGrantType;
import org.springframework.security.oauth2.core.ClientAuthenticationMethod;
import org.springframework.util.StringUtils;

import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

public class OAuth2ClientRegistrationRepository implements ClientRegistrationRepository {

    private final Map<String, ClientRegistration> registrations = new ConcurrentHashMap<>();

    public OAuth2ClientRegistrationRepository(IdentityProviderConfigRepository jpaClientRegistrationRepository) {

        jpaClientRegistrationRepository.findAll().forEach(entity -> {
            ClientAuthenticationMethod authMethod = Optional.ofNullable(entity.getClientAuthMethod())
                    .map(ClientAuthMethod::toSpringMethod)
                    .orElse(ClientAuthenticationMethod.CLIENT_SECRET_BASIC);

            ClientRegistration.Builder builder = ClientRegistration.withRegistrationId(entity.getProviderType().name().toLowerCase())
                    .clientId(entity.getClientId())
                    .clientSecret(entity.getClientSecret())
                    .clientAuthenticationMethod(authMethod)
                    .authorizationGrantType(AuthorizationGrantType.AUTHORIZATION_CODE)
                    .redirectUri(entity.getRedirectUri())
                    .scope(StringUtils.commaDelimitedListToStringArray(entity.getScopes()))
                    .authorizationUri(entity.getAuthUri())
                    .tokenUri(entity.getTokenUri())
                    .clientName(entity.getName());

            if (entity.getUserInfoUri() != null) {
                builder.userInfoUri(entity.getUserInfoUri());
            }

            if (entity.getJwksUri() != null) {
                builder.jwkSetUri(entity.getJwksUri());
            }

            registrations.put(entity.getProviderType().name().toLowerCase(), builder.build());
        });

    }

    @Override
    public ClientRegistration findByRegistrationId(String registrationId) {
        return registrations.get(registrationId);
    }
}
