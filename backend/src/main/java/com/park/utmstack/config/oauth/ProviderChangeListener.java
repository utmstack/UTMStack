package com.park.utmstack.config.oauth;

import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import com.park.utmstack.util.events.ProviderChangedEvent;
import lombok.RequiredArgsConstructor;
import org.springframework.context.event.EventListener;
import org.springframework.security.oauth2.client.registration.ClientRegistrationRepository;
import org.springframework.stereotype.Component;

@Component
@RequiredArgsConstructor
public class ProviderChangeListener {

    private final ClientRegistrationRepository repository;
    private final IdentityProviderConfigRepository identityProviderConfigRepository;

    @EventListener
    public void handleProviderChanged(ProviderChangedEvent event) {
        if (repository instanceof OAuth2ClientRegistrationRepository customRepo) {
            customRepo.reloadProviders(identityProviderConfigRepository);
        }
    }
}
