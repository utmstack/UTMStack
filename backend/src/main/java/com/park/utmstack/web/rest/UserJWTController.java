package com.park.utmstack.web.rest;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.Authority;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.federation_service.UtmFederationServiceClient;
import com.park.utmstack.repository.federation_service.UtmFederationServiceClientRepository;
import com.park.utmstack.security.TooMuchLoginAttemptsException;
import com.park.utmstack.security.jwt.JWTFilter;
import com.park.utmstack.security.jwt.TokenProvider;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.login_attempts.LoginAttemptService;
import com.park.utmstack.service.tfa.TfaService;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.util.exceptions.InvalidConnectionKeyException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.vm.LoginVM;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.util.Base64Utils;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;
import java.util.stream.Collectors;

/**
 * Controller to authenticate users.
 */
@RestController
@RequestMapping("/api")
public class UserJWTController {

    private static final String CLASSNAME = "UserJWTController";
    private final Logger log = LoggerFactory.getLogger(UserJWTController.class);

    private final TokenProvider tokenProvider;
    private final AuthenticationManager authenticationManager;
    private final ApplicationEventService applicationEventService;
    private final UserService userService;
    private final TfaService tfaService;
    private final MailService mailService;
    private final LoginAttemptService loginAttemptService;
    private final UtmFederationServiceClientRepository fsClientRepository;
    private final PasswordEncoder passwordEncoder;

    public UserJWTController(TokenProvider tokenProvider,
                             AuthenticationManager authenticationManager,
                             ApplicationEventService applicationEventService,
                             UserService userService,
                             TfaService tfaService,
                             MailService mailService,
                             LoginAttemptService loginAttemptService,
                             UtmFederationServiceClientRepository fsClientRepository,
                             PasswordEncoder passwordEncoder) {
        this.tokenProvider = tokenProvider;
        this.authenticationManager = authenticationManager;
        this.applicationEventService = applicationEventService;
        this.userService = userService;
        this.tfaService = tfaService;
        this.mailService = mailService;
        this.loginAttemptService = loginAttemptService;
        this.fsClientRepository = fsClientRepository;
        this.passwordEncoder = passwordEncoder;
    }

    @PostMapping("/authenticate")
    public ResponseEntity<JWTToken> authorize(@Valid @RequestBody LoginVM loginVM) {
        final String ctx = CLASSNAME + ".authorize";
        try {
            if (loginAttemptService.isBlocked())
                throw new TooMuchLoginAttemptsException(String.format("Client IP %1$s blocked due to too many failed login attempts", loginAttemptService.getClientIP()));

            boolean authenticated = !Boolean.parseBoolean(Constants.CFG.get(Constants.PROP_TFA_ENABLE));

            UsernamePasswordAuthenticationToken authenticationToken =
                new UsernamePasswordAuthenticationToken(loginVM.getUsername(), loginVM.getPassword());
            boolean rememberMe = loginVM.isRememberMe() != null && loginVM.isRememberMe();

            Authentication authentication = this.authenticationManager.authenticate(authenticationToken);
            SecurityContextHolder.getContext().setAuthentication(authentication);

            String jwt = tokenProvider.createToken(authentication, rememberMe, authenticated);

            if (!authenticated) {
                String secret = tfaService.generateSecret();
                String code = tfaService.generateCode(secret);
                User user = userService.updateUserTfaSecret(loginVM.getUsername(), secret);
                mailService.sendTfaVerificationCode(user, code);
            }

            HttpHeaders httpHeaders = new HttpHeaders();
            httpHeaders.add(JWTFilter.AUTHORIZATION_HEADER, "Bearer " + jwt);
            return new ResponseEntity<>(new JWTToken(jwt, authenticated), httpHeaders, HttpStatus.OK);
        } catch (BadCredentialsException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildUnauthorizedResponse(msg);
        } catch (TooMuchLoginAttemptsException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildLockedResponse(msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildInternalServerErrorResponse(msg);
        }
    }

    @GetMapping("/check-credentials")
    public ResponseEntity<String> checkPassword(@Valid @RequestParam String password, @RequestParam String checkUUID) {
        final String ctx = CLASSNAME + ".checkPassword";
        try {
            User user = userService.getCurrentUserLogin();
            UsernamePasswordAuthenticationToken authenticationToken =
                new UsernamePasswordAuthenticationToken(user.getLogin(), password);
            Authentication authentication = this.authenticationManager.authenticate(authenticationToken);
            if (authentication.isAuthenticated()) {
                return new ResponseEntity<>(checkUUID, HttpStatus.OK);
            } else {
                return new ResponseEntity<>(checkUUID, HttpStatus.BAD_REQUEST);
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PostMapping("/authenticateFederationServiceManager")
    public ResponseEntity<JWTToken> authorizeFederationServiceManager(@Valid @RequestBody String token) {
        final String ctx = CLASSNAME + ".authorizeFederationServiceManager";
        try {
            if (!StringUtils.hasText(token))
                throw new InvalidConnectionKeyException("It's needed to provide a connection key");

            UtmFederationServiceClient fsToken = fsClientRepository.findByFsClientToken(token)
                .orElseThrow(() -> new InvalidConnectionKeyException("Unrecognized connection key"));

            String[] tokenInfo = new String(Base64Utils.decodeFromUrlSafeString(fsToken.getFsClientToken())).split("\\|");

            if (tokenInfo.length != 2)
                throw new InvalidConnectionKeyException("Connection key is corrupt, length is invalid");

            if (!tokenInfo[0].equals(System.getenv(Constants.ENV_SERVER_NAME)))
                throw new InvalidConnectionKeyException("Connection key is corrupt, unrecognized instance");

            UsernamePasswordAuthenticationToken authenticationToken =
                new UsernamePasswordAuthenticationToken(Constants.FS_USER, CipherUtil.decrypt(tokenInfo[1], System.getenv(Constants.ENV_ENCRYPTION_KEY)));

            Authentication authentication = this.authenticationManager.authenticate(authenticationToken);
            SecurityContextHolder.getContext().setAuthentication(authentication);

            String jwt = tokenProvider.createToken(authentication, true, true);

            HttpHeaders httpHeaders = new HttpHeaders();
            httpHeaders.add(JWTFilter.AUTHORIZATION_HEADER, "Bearer " + jwt);

            return new ResponseEntity<>(new JWTToken(jwt, true), httpHeaders, HttpStatus.OK);
        } catch (InvalidConnectionKeyException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildBadRequestResponse(msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildInternalServerErrorResponse(msg);
        }
    }

    @GetMapping("/tfa/verifyCode")
    public ResponseEntity<JWTToken> verifyCode(@RequestParam String code) {
        final String ctx = CLASSNAME + ".verifyCode";
        try {
            User user = userService.getCurrentUserLogin();
            if (!tfaService.validateCode(user.getTfaSecret(), code))
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED).headers(
                    HeaderUtil.createFailureAlert("", "", "Your secret code is invalid")).body(null);

            List<SimpleGrantedAuthority> authorities = user.getAuthorities().stream().map(Authority::getName)
                .map(SimpleGrantedAuthority::new).collect(Collectors.toList());

            org.springframework.security.core.userdetails.User principal = new org.springframework.security.core.userdetails.User(user.getLogin(), "", authorities);

            UsernamePasswordAuthenticationToken authentication = new UsernamePasswordAuthenticationToken(principal, "", authorities);

            String jwt = tokenProvider.createToken(authentication, true, true);

            HttpHeaders httpHeaders = new HttpHeaders();
            httpHeaders.add(JWTFilter.AUTHORIZATION_HEADER, "Bearer " + jwt);
            return new ResponseEntity<>(new JWTToken(jwt, true), httpHeaders, HttpStatus.OK);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }


    /**
     * Object to return as body in JWT Authentication.
     */
    public static class JWTToken {

        private String idToken;
        private boolean authenticated;

        JWTToken(String idToken, boolean authenticated) {
            this.idToken = idToken;
            this.authenticated = authenticated;
        }

        @JsonProperty("id_token")
        String getIdToken() {
            return idToken;
        }

        void setIdToken(String idToken) {
            this.idToken = idToken;
        }

        public boolean isAuthenticated() {
            return authenticated;
        }

        public void setAuthenticated(boolean authenticated) {
            this.authenticated = authenticated;
        }
    }
}
