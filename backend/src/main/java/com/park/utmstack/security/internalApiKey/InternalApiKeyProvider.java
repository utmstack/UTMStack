package com.park.utmstack.security.internalApiKey;

import com.park.utmstack.config.Constants;
import com.park.utmstack.repository.UserRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.userdetails.User;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.stream.Collectors;

@Component
public class InternalApiKeyProvider {
    private static final String CLASSNAME = "InternalApiKeyProvider";
    private final Logger log = LoggerFactory.getLogger(InternalApiKeyProvider.class);

    private final UserRepository userRepository;

    public InternalApiKeyProvider(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    public UsernamePasswordAuthenticationToken getAuthentication(String apiKey) {
        final String ctx = CLASSNAME + ".getAuthentication";
        try {
            com.park.utmstack.domain.User user = userRepository.findOneWithAuthoritiesByLogin(Constants.ADMIN_USER)
                .orElseThrow(() -> new RuntimeException("User not found"));
            List<SimpleGrantedAuthority> authorities = user.getAuthorities().stream().map(d -> new SimpleGrantedAuthority(d.getName()))
                .collect(Collectors.toList());
            User principal = new User(user.getLogin(), "", authorities);
            return new UsernamePasswordAuthenticationToken(principal, apiKey, authorities);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }
}

