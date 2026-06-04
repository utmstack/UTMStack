package com.park.utmstack.config;

import com.github.benmanes.caffeine.cache.Cache;
import com.github.benmanes.caffeine.cache.Caffeine;
import com.park.utmstack.domain.tfa.TfaSetupState;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.time.Duration;

@Configuration
public class TfaCacheConfig {

    @Bean
    public Cache<String, TfaSetupState> tfaSetupCache() {
        return Caffeine.newBuilder()
                .expireAfterWrite(Duration.ofMinutes(10))
                .maximumSize(1000)
                .build();
    }
}
