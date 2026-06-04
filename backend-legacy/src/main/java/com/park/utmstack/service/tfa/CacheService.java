package com.park.utmstack.service.tfa;

import com.github.benmanes.caffeine.cache.Cache;
import com.park.utmstack.domain.tfa.TfaMethod;
import com.park.utmstack.domain.tfa.TfaSetupState;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
@RequiredArgsConstructor
public class CacheService {
    private final Cache<String, TfaSetupState> cache;

    public void storeState(String username, TfaMethod method, TfaSetupState state) {
        cache.put(key(username, method), state);
    }

    public Optional<TfaSetupState> getState(String username, TfaMethod method) {
        return Optional.ofNullable(cache.getIfPresent(key(username, method)));
    }

    public void clear(String username, TfaMethod method) {
        cache.invalidate(key(username, method));
    }

    private String key(String username, TfaMethod method) {
        return username + ":" + method.name();
    }
}

