package com.park.utmstack.security.api_key;

import com.park.utmstack.security.jwt.JWTFilter;
import org.springframework.security.config.annotation.SecurityConfigurerAdapter;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.web.DefaultSecurityFilterChain;

public class ApiKeyConfigurer extends SecurityConfigurerAdapter<DefaultSecurityFilterChain, HttpSecurity> {

    private final ApiKeyFilter apiKeyFilter;

    public ApiKeyConfigurer(ApiKeyFilter apiKeyFilter) {
        this.apiKeyFilter = apiKeyFilter;
    }

    @Override
    public void configure(HttpSecurity http) throws Exception {
        http.addFilterAfter(apiKeyFilter, JWTFilter.class);
    }
}
