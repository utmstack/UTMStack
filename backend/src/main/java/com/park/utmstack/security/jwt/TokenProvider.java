package com.park.utmstack.security.jwt;


import com.park.utmstack.security.AuthoritiesConstants;
import com.park.utmstack.util.CipherUtil;
import io.jsonwebtoken.*;
import io.jsonwebtoken.io.Decoders;
import io.jsonwebtoken.security.Keys;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.userdetails.User;
import org.springframework.stereotype.Component;
import tech.jhipster.config.JHipsterProperties;

import java.security.Key;
import java.util.Arrays;
import java.util.Collection;
import java.util.Date;
import java.util.stream.Collectors;

@Component
public class TokenProvider {

    private static final String CLASSNAME = "TokenProvider";
    private final Logger log = LoggerFactory.getLogger(TokenProvider.class);

    private static final String AUTHORITIES_KEY = "auth";
    private static final String AUTHENTICATED = "authenticated";
    private static final String SECRET = CipherUtil.generateSafeToken();
    public static final long TEMP_TOKEN_VALIDITY_IN_MILLIS = 300000;

    private final Key key;
    private final JwtParser jwtParser;
    private final long tokenValidityInMilliseconds;
    private final long tokenValidityInMillisecondsForRememberMe;

    public TokenProvider(JHipsterProperties jHipsterProperties) {
        this.key = Keys.hmacShaKeyFor(Decoders.BASE64.decode(SECRET));
        jwtParser = Jwts.parserBuilder().setSigningKey(key).build();
        this.tokenValidityInMilliseconds =
            1000 * jHipsterProperties.getSecurity().getAuthentication().getJwt().getTokenValidityInSeconds();
        this.tokenValidityInMillisecondsForRememberMe =
            1000 * jHipsterProperties.getSecurity().getAuthentication().getJwt()
                .getTokenValidityInSecondsForRememberMe();
    }

    /**
     * @param authentication
     * @param rememberMe
     * @param authenticated
     * @return
     */
    public String createToken(Authentication authentication, boolean rememberMe, boolean authenticated) {
        final String ctx = CLASSNAME + ".createToken";

        try {
            String authorities = !authenticated ? AuthoritiesConstants.PRE_VERIFICATION_USER : authentication.getAuthorities().stream()
                .map(GrantedAuthority::getAuthority).collect(Collectors.joining(","));

            long now = (new Date()).getTime();
            Date validity;

            if (!authenticated) {
                validity = new Date(now + TEMP_TOKEN_VALIDITY_IN_MILLIS);
            } else {
                if (rememberMe) {
                    validity = new Date(now + this.tokenValidityInMillisecondsForRememberMe);
                } else {
                    validity = new Date(now + this.tokenValidityInMilliseconds);
                }
            }
            return Jwts.builder().setSubject(authentication.getName()).claim(AUTHORITIES_KEY, authorities)
                .claim(AUTHENTICATED, authenticated).signWith(key, SignatureAlgorithm.HS512).setExpiration(validity)
                .compact();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * @param token
     * @return
     */
    public UsernamePasswordAuthenticationToken getAuthentication(String token) {
        Claims claims = jwtParser.parseClaimsJws(token).getBody();
        Collection<? extends GrantedAuthority> authorities = Arrays
            .stream(claims.get(AUTHORITIES_KEY).toString().split(","))
            .filter(auth -> !auth.trim().isEmpty())
            .map(SimpleGrantedAuthority::new)
            .collect(Collectors.toList());
        User principal = new User(claims.getSubject(), "", authorities);
        return new UsernamePasswordAuthenticationToken(principal, token, authorities);
    }

    public Boolean isAuthenticated(String token) {
        Claims claims = jwtParser.parseClaimsJws(token).getBody();
        return claims.get(AUTHENTICATED, Boolean.class);
    }

    public String getUserLoginFromToken(String token) {
        Claims claims = jwtParser.parseClaimsJws(token).getBody();
        return claims.getSubject();
    }

    public boolean validateToken(String authToken) {
        try {
            jwtParser.parseClaimsJws(authToken);
            return true;
        } catch (JwtException | IllegalArgumentException e) {
            log.info("Invalid JWT token.");
            log.trace("Invalid JWT token trace.", e);
        }
        return false;
    }
}
