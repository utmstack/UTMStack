package com.park.utmstack.security.internalApiKey;

import com.park.utmstack.security.jwt.JWTFilter;
import org.springframework.security.config.annotation.SecurityConfigurerAdapter;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.web.DefaultSecurityFilterChain;

public class InternalApiKeyConfigurer extends SecurityConfigurerAdapter<DefaultSecurityFilterChain, HttpSecurity> {

    private final InternalApiKeyProvider internalApiKeyProvider;

    public InternalApiKeyConfigurer(InternalApiKeyProvider apiKeyProvider) {
        this.internalApiKeyProvider = apiKeyProvider;
    }

    @Override
    public void configure(HttpSecurity builder) {
        InternalApiKeyFilter internalApiKeyFilter = new InternalApiKeyFilter(internalApiKeyProvider);
        builder.addFilterBefore(internalApiKeyFilter, JWTFilter.class);
    }
}
