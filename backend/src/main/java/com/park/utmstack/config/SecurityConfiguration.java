package com.park.utmstack.config;

import com.park.utmstack.loggin.api_key.ApiKeyUsageLoggingService;
import com.park.utmstack.loggin.filter.MdcCleanupFilter;
import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.security.AuthoritiesConstants;
import com.park.utmstack.security.api_key.ApiKeyConfigurer;
import com.park.utmstack.security.api_key.ApiKeyFilter;
import com.park.utmstack.security.internalApiKey.InternalApiKeyConfigurer;
import com.park.utmstack.security.internalApiKey.InternalApiKeyProvider;
import com.park.utmstack.security.jwt.JWTConfigurer;
import com.park.utmstack.security.jwt.JWTFilter;
import com.park.utmstack.security.jwt.TokenProvider;
import com.park.utmstack.service.api_key.ApiKeyService;
import lombok.RequiredArgsConstructor;
import org.apache.commons.net.util.SubnetUtils;
import org.springframework.beans.factory.BeanInitializationException;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Import;
import org.springframework.http.HttpMethod;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.config.annotation.authentication.builders.AuthenticationManagerBuilder;
import org.springframework.security.config.annotation.method.configuration.EnableGlobalMethodSecurity;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter;
import org.springframework.security.config.annotation.web.configuration.WebSecurityCustomizer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.filter.CorsFilter;
import org.zalando.problem.spring.web.advice.security.SecurityProblemSupport;

import javax.annotation.PostConstruct;
import javax.servlet.http.HttpServletResponse;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

@Configuration
@RequiredArgsConstructor
@EnableWebSecurity
@EnableGlobalMethodSecurity(prePostEnabled = true, securedEnabled = true)
@Import(SecurityProblemSupport.class)
public class SecurityConfiguration extends WebSecurityConfigurerAdapter {

    private final AuthenticationManagerBuilder authenticationManagerBuilder;
    private final UserDetailsService userDetailsService;
    private final TokenProvider tokenProvider;
    private final CorsFilter corsFilter;
    private final InternalApiKeyProvider internalApiKeyProvider;
    private final ApiKeyFilter apiKeyFilter;

    @PostConstruct
    public void init() {
        try {
            authenticationManagerBuilder
                    .userDetailsService(userDetailsService)
                    .passwordEncoder(passwordEncoder());
        } catch (Exception e) {
            throw new BeanInitializationException("Security configuration failed", e);
        }
    }

    @Override
    @Bean
    public AuthenticationManager authenticationManagerBean() throws Exception {
        return super.authenticationManagerBean();
    }

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }

    @Bean
    public WebSecurityCustomizer webSecurityCustomizer() {
        return web -> web.ignoring()
                .antMatchers(HttpMethod.OPTIONS, "/**")
                .antMatchers("/swagger-ui/**")
                .antMatchers("/i18n/**");
    }

    @Override
    public void configure(HttpSecurity http) throws Exception {
        http
                .csrf()
                .disable()
                .addFilterBefore(corsFilter, UsernamePasswordAuthenticationFilter.class)
                .exceptionHandling()
                .authenticationEntryPoint((request, response, authException) -> response.sendError(HttpServletResponse.SC_UNAUTHORIZED))
                .accessDeniedHandler((request, response, accessDeniedException) -> response.sendError(HttpServletResponse.SC_FORBIDDEN))
                .and()
                .headers()
                .frameOptions()
                .disable()
                .and()
                .sessionManagement()
                .sessionCreationPolicy(SessionCreationPolicy.STATELESS)
                .and()
                .authorizeRequests()
                .antMatchers("/api/authenticate").permitAll()
                .antMatchers("/api/authenticateFederationServiceManager").permitAll()
                .antMatchers("/api/ping").permitAll()
                .antMatchers("/api/date-format").permitAll()
                .antMatchers("/api/healthcheck").permitAll()
                .antMatchers("/api/releaseInfo").permitAll()
                .antMatchers("/api/account/reset-password/init").permitAll()
                .antMatchers("/api/account/reset-password/finish").permitAll()
                .antMatchers("/api/images/all").permitAll()
                .antMatchers("/api/enrollment/**").hasAnyAuthority(AuthoritiesConstants.PRE_VERIFICATION_USER)
                .antMatchers("/api/tfa/verify-code").hasAnyAuthority(AuthoritiesConstants.PRE_VERIFICATION_USER, AuthoritiesConstants.USER, AuthoritiesConstants.ADMIN)
                .antMatchers("/api/tfa/refresh").hasAnyAuthority(AuthoritiesConstants.PRE_VERIFICATION_USER, AuthoritiesConstants.USER, AuthoritiesConstants.ADMIN)
                .antMatchers("/api/tfa/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
                .antMatchers("/api/utm-incident-jobs").hasAuthority(AuthoritiesConstants.ADMIN)
                .antMatchers("/api/utm-incident-jobs/**").hasAuthority(AuthoritiesConstants.ADMIN)
                .antMatchers("/api/utm-incident-variables/**").hasAuthority(AuthoritiesConstants.ADMIN)
                .antMatchers(HttpMethod.GET, "/api/utm-incident-variables").hasAnyAuthority()
                .antMatchers("/api/custom-reports/**").denyAll()
                .antMatchers("/api/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
                .antMatchers("/ws/topic").hasAuthority(AuthoritiesConstants.ADMIN)
                .antMatchers("/ws/**").permitAll()
                .antMatchers("/management/info").permitAll()
                .antMatchers("/management/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
                .and()
                .apply(securityConfigurerAdapterForJwt())
                .and()
                .apply(securityConfigurerAdapterForInternalApiKey())
                .and()
                .apply(securityConfigurerAdapterForApiKey())    ;


    }

    private JWTConfigurer securityConfigurerAdapterForJwt() {
        return new JWTConfigurer(tokenProvider);
    }

    private InternalApiKeyConfigurer securityConfigurerAdapterForInternalApiKey() {
        return new InternalApiKeyConfigurer(internalApiKeyProvider);
    }

    private ApiKeyConfigurer securityConfigurerAdapterForApiKey() {
        return new ApiKeyConfigurer(apiKeyFilter);
    }

}
