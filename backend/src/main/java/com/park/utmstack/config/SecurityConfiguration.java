package com.park.utmstack.config;

import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.security.AuthoritiesConstants;
import com.park.utmstack.security.api_key.ApiKeyConfigurer;
import com.park.utmstack.security.api_key.ApiKeyFilter;
import com.park.utmstack.security.internalApiKey.InternalApiKeyConfigurer;
import com.park.utmstack.security.internalApiKey.InternalApiKeyProvider;
import com.park.utmstack.security.jwt.JWTConfigurer;
import com.park.utmstack.security.jwt.TokenProvider;
import com.park.utmstack.security.saml.Saml2LoginFailureHandler;
import com.park.utmstack.security.saml.Saml2LoginSuccessHandler;
import com.park.utmstack.service.tfa.TfaService;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Lazy;
import org.springframework.context.annotation.Import;
import org.springframework.http.HttpMethod;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.ProviderManager;
import org.springframework.security.authentication.dao.DaoAuthenticationProvider;
import org.springframework.security.config.annotation.method.configuration.EnableMethodSecurity;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configuration.WebSecurityCustomizer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.filter.CorsFilter;
import org.zalando.problem.spring.web.advice.security.SecurityProblemSupport;

/**
 * Spring Security 6 (Spring Boot 3) configuration. Migrated from the JHipster 7
 * {@code WebSecurityConfigurerAdapter} style: the http rules now live in a
 * {@link SecurityFilterChain} bean using the lambda DSL, method security uses
 * {@link EnableMethodSecurity}, and the {@link AuthenticationManager} is built
 * explicitly with a {@link DaoAuthenticationProvider} (Security 6 removed
 * {@code AuthenticationManagerBuilder}).
 */
@Configuration
@EnableWebSecurity
@EnableMethodSecurity(prePostEnabled = true, securedEnabled = true)
@Import(SecurityProblemSupport.class)
public class SecurityConfiguration {

    private final UserDetailsService userDetailsService;
    private final TokenProvider tokenProvider;
    private final CorsFilter corsFilter;
    private final InternalApiKeyProvider internalApiKeyProvider;
    private final ApiKeyFilter apiKeyFilter;
    private final UserRepository userRepository;
    private final TfaService tfaService;

    /**
     * #3042: the SAML success handler needs TfaService, but TfaService
     * transitively depends on UserService -> PasswordEncoder, and PasswordEncoder
     * is a @Bean in THIS class — a circular reference that Spring Boot 4 rejects.
     * The TfaService dependency is only used at SAML login time, so it is injected
     * lazily to break the cycle at construction.
     */
    public SecurityConfiguration(UserDetailsService userDetailsService,
                                 TokenProvider tokenProvider,
                                 CorsFilter corsFilter,
                                 InternalApiKeyProvider internalApiKeyProvider,
                                 ApiKeyFilter apiKeyFilter,
                                 UserRepository userRepository,
                                 @Lazy TfaService tfaService) {
        this.userDetailsService = userDetailsService;
        this.tokenProvider = tokenProvider;
        this.corsFilter = corsFilter;
        this.internalApiKeyProvider = internalApiKeyProvider;
        this.apiKeyFilter = apiKeyFilter;
        this.userRepository = userRepository;
        this.tfaService = tfaService;
    }

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }

    @Bean
    public AuthenticationManager authenticationManager() {
        // Spring Security 7: DaoAuthenticationProvider has no no-arg constructor; it is built
        // with the UserDetailsService. It is an AuthenticationProvider, not an
        // AuthenticationManager, so wrap it in a ProviderManager.
        DaoAuthenticationProvider provider = new DaoAuthenticationProvider(userDetailsService);
        provider.setPasswordEncoder(passwordEncoder());
        return new ProviderManager(provider);
    }

    @Bean
    public WebSecurityCustomizer webSecurityCustomizer() {
        return web -> web.ignoring()
                .requestMatchers(HttpMethod.OPTIONS, "/**")
                .requestMatchers("/swagger-ui/**")
                .requestMatchers("/i18n/**");
    }

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        http
                .csrf(csrf -> csrf.disable())
                .addFilterBefore(corsFilter, UsernamePasswordAuthenticationFilter.class)
                .exceptionHandling(ex -> ex
                        .authenticationEntryPoint((request, response, authException) -> response.sendError(HttpServletResponse.SC_UNAUTHORIZED))
                        .accessDeniedHandler((request, response, accessDeniedException) -> response.sendError(HttpServletResponse.SC_FORBIDDEN)))
                .headers(headers -> headers
                        // Same-origin only: the panel frames its own PDF/report views,
                        // nothing else needs to frame backend responses.
                        .frameOptions(frame -> frame.sameOrigin()))
                .sessionManagement(sm -> sm.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                .authorizeHttpRequests(auth -> auth
                        .requestMatchers("/api/authenticate").permitAll()
                        .requestMatchers("/api/authenticateFederationServiceManager").permitAll()
                        .requestMatchers("/api/ping").permitAll()
                        .requestMatchers("/api/date-format").permitAll()
                        .requestMatchers("/api/v1/mode").permitAll()
                        .requestMatchers("/api/healthcheck").permitAll()
                        .requestMatchers("/api/releaseInfo").permitAll()
                        .requestMatchers("/api/account/reset-password/init").permitAll()
                        .requestMatchers("/api/account/reset-password/finish").permitAll()
                        .requestMatchers("/api/images/all").permitAll()
                        .requestMatchers("/api/info/version").permitAll()
                        .requestMatchers("/api/enrollment/**").hasAnyAuthority(AuthoritiesConstants.PRE_VERIFICATION_USER)
                        .requestMatchers("/api/tfa/verify-code").hasAnyAuthority(AuthoritiesConstants.PRE_VERIFICATION_USER, AuthoritiesConstants.USER, AuthoritiesConstants.ADMIN)
                        .requestMatchers("/api/tfa/refresh").hasAnyAuthority(AuthoritiesConstants.PRE_VERIFICATION_USER, AuthoritiesConstants.USER, AuthoritiesConstants.ADMIN)
                        .requestMatchers("/api/tfa/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
                        // Incident-response dispatch: security analysts (ROLE_USER) execute
                        // IR actions, not just platform admins.
                        .requestMatchers("/api/utm-incident-jobs").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
                        .requestMatchers("/api/utm-incident-jobs/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
                        .requestMatchers("/api/utm-incident-variables/**").hasAuthority(AuthoritiesConstants.ADMIN)
                        // Original (Security 5) used zero-arg .hasAnyAuthority() here = "any authenticated user"
                        // (subpaths above are admin-only). Security 6's zero-arg form throws "authorities
                        // cannot be empty" at bean-construction, so the DSL equivalent is .authenticated().
                        .requestMatchers(HttpMethod.GET, "/api/utm-incident-variables").authenticated()
                        .requestMatchers("/api/custom-reports/**").denyAll()
                        // ---- admin-only surfaces (audit: default was allow for ROLE_USER) ----
                        // Agent/collector enrollment master credential + federation key issuance.
                        .requestMatchers("/api/federation-service/**").hasAuthority(AuthoritiesConstants.ADMIN)
                        // IdP / SAML configuration (create, update, delete, list).
                        .requestMatchers("/api/identity-providers/**").hasAuthority(AuthoritiesConstants.ADMIN)
                        // Login page lists SSO providers before authentication; the payload is
                        // trimmed to IdentityProviderPublicDto. Everything else is admin-only.
                        .requestMatchers(HttpMethod.GET, "/api/utm-providers").permitAll()
                        .requestMatchers("/api/utm-providers/**").hasAuthority(AuthoritiesConstants.ADMIN)
                        // Alert-response automation rules can execute OS commands on agents.
                        .requestMatchers("/api/utm-alert-response-rules").hasAuthority(AuthoritiesConstants.ADMIN)
                        .requestMatchers("/api/utm-alert-response-rules/**").hasAuthority(AuthoritiesConstants.ADMIN)
                        .requestMatchers("/api/utm-alert-response-action-templates").hasAuthority(AuthoritiesConstants.ADMIN)
                        // Instance-wide configuration parameters (includes TFA + mail settings).
                        .requestMatchers(HttpMethod.PUT, "/api/utm-configuration-parameters").hasAuthority(AuthoritiesConstants.ADMIN)
                        // TFA enrollment completion (writes TFA enable/method, returns full token).
                        .requestMatchers("/api/tfa/complete").hasAuthority(AuthoritiesConstants.ADMIN)
                        // OpenSearch index management (destructive).
                        .requestMatchers("/api/elasticsearch/index/delete-index").hasAuthority(AuthoritiesConstants.ADMIN)
                        // Encryption endpoint (encryption oracle).
                        .requestMatchers("/api/encrypt").hasAuthority(AuthoritiesConstants.ADMIN)
                        .requestMatchers("/api/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER)
                        .requestMatchers("/ws/topic").hasAuthority(AuthoritiesConstants.ADMIN)
                        .requestMatchers("/ws/**").permitAll()
                        .requestMatchers("/management/info").permitAll()
                        // Runtime log level: TRACE on the wrong logger dumps request bodies
                        // and tokens into the logs.
                        .requestMatchers("/management/logs").hasAuthority(AuthoritiesConstants.ADMIN)
                        .requestMatchers("/management/**").hasAnyAuthority(AuthoritiesConstants.ADMIN, AuthoritiesConstants.USER));

        // In Spring Security 6, saml2Login() and apply(configurer) return the configurer, not
        // HttpSecurity, so they cannot be chained after authorizeHttpRequests(). Call them as
        // separate statements; each still registers on the same HttpSecurity builder.
        http.saml2Login(saml -> saml
                .successHandler(new Saml2LoginSuccessHandler(tokenProvider, userRepository, tfaService))
                .failureHandler(new Saml2LoginFailureHandler()));
        http.apply(securityConfigurerAdapterForJwt());
        http.apply(securityConfigurerAdapterForInternalApiKey());
        http.apply(securityConfigurerAdapterForApiKey());

        return http.build();
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

    @Bean
    public Saml2LoginFailureHandler saml2LoginFailureHandler() {
        return new Saml2LoginFailureHandler();
    }
}
