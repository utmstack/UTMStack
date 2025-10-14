package com.park.utmstack.web.rest;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.Authority;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.federation_service.UtmFederationServiceClient;
import com.park.utmstack.loggin.LogContextBuilder;
import com.park.utmstack.repository.federation_service.UtmFederationServiceClientRepository;
import com.park.utmstack.security.TooMuchLoginAttemptsException;
import com.park.utmstack.security.jwt.JWTFilter;
import com.park.utmstack.security.jwt.TokenProvider;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.auth.JWTToken;
import com.park.utmstack.service.login_attempts.LoginAttemptService;
import com.park.utmstack.service.tfa.TfaService;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.exceptions.InvalidConnectionKeyException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.vm.LoginVM;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.util.Base64Utils;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import javax.validation.Valid;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * Controller to authenticate users.
 */
@RequiredArgsConstructor
@Slf4j
@RestController
@RequestMapping("/api")
public class UserJWTController {

    private final TokenProvider tokenProvider;
    private final AuthenticationManager authenticationManager;
    private final ApplicationEventService applicationEventService;
    private final UserService userService;
    private final TfaService tfaService;
    private final MailService mailService;
    private final LoginAttemptService loginAttemptService;
    private final UtmFederationServiceClientRepository fsClientRepository;
    private final LogContextBuilder logContextBuilder;


    @AuditEvent(
            attemptType = ApplicationEventType.AUTH_ATTEMPT,
            attemptMessage = "Authentication attempt registered",
            successType = ApplicationEventType.UNDEFINED,
            successMessage = ""
    )
    @PostMapping("/authenticate")
    public ResponseEntity<JWTToken> authorize(@Valid @RequestBody LoginVM loginVM, HttpServletRequest request) {

            if (loginAttemptService.isBlocked()) {
                String ip = loginAttemptService.getClientIP();
                throw new TooMuchLoginAttemptsException(String.format("Authentication blocked: IP %s exceeded login attempt threshold", ip));
            }

            boolean authenticated = !Boolean.parseBoolean(Constants.CFG.get(Constants.PROP_TFA_ENABLE));

            UsernamePasswordAuthenticationToken authenticationToken =
                    new UsernamePasswordAuthenticationToken(loginVM.getUsername(), loginVM.getPassword());
            boolean rememberMe = loginVM.isRememberMe() != null && loginVM.isRememberMe();

            Authentication authentication = this.authenticationManager.authenticate(authenticationToken);
            SecurityContextHolder.getContext().setAuthentication(authentication);

            String jwt = tokenProvider.createToken(authentication, rememberMe, authenticated);
            Map<String, Object> args = logContextBuilder.buildArgs(request);

            if (!authenticated) {
                String secret = tfaService.generateSecret();
                String code = tfaService.generateCode(secret);
                User user = userService.updateUserTfaSecret(loginVM.getUsername(), secret);

                applicationEventService.createEvent(
                        "TFA challenge issued for user '" + user.getLogin(),
                        ApplicationEventType.TFA_CODE_SENT,
                        args
                );
                mailService.sendTfaVerificationCode(user, code);
            } else {
                applicationEventService.createEvent(
                        "Login successfully completed for user '" + loginVM.getUsername() + "'",
                        ApplicationEventType.AUTH_SUCCESS,
                        args);
            }

            HttpHeaders httpHeaders = new HttpHeaders();
            httpHeaders.add(JWTFilter.AUTHORIZATION_HEADER, "Bearer " + jwt);
            return new ResponseEntity<>(new JWTToken(jwt, authenticated), httpHeaders, HttpStatus.OK);

    }

    @GetMapping("/check-credentials")
    public ResponseEntity<String> checkPassword(@Valid @RequestParam String password, @RequestParam String checkUUID) {

            User user = userService.getCurrentUserLogin();
            UsernamePasswordAuthenticationToken authenticationToken =
                    new UsernamePasswordAuthenticationToken(user.getLogin(), password);
            Authentication authentication = this.authenticationManager.authenticate(authenticationToken);
            if (authentication.isAuthenticated()) {
                return new ResponseEntity<>(checkUUID, HttpStatus.OK);
            } else {
                return new ResponseEntity<>(checkUUID, HttpStatus.BAD_REQUEST);
            }

    }

    @PostMapping("/authenticateFederationServiceManager")
    public ResponseEntity<JWTToken> authorizeFederationServiceManager(@Valid @RequestBody String token) {

            if (!StringUtils.hasText(token))
                throw new InvalidConnectionKeyException("It's needed to provide a connection key");

            UtmFederationServiceClient fsToken = fsClientRepository.findByFsClientToken(token)
                    .orElseThrow(() -> new InvalidConnectionKeyException("Unrecognized connection key"));

            String[] tokenInfo = new String(Base64Utils.decodeFromUrlSafeString(fsToken.getFsClientToken())).split("\\|");

            if (tokenInfo.length != 2)
                throw new InvalidConnectionKeyException("Connection key is corrupt, length is invalid");

            /*if (!tokenInfo[0].equals(System.getenv(Constants.ENV_SERVER_NAME)))
                throw new InvalidConnectionKeyException("Connection key is corrupt, unrecognized instance");*/

            UsernamePasswordAuthenticationToken authenticationToken =
                    new UsernamePasswordAuthenticationToken(Constants.FS_USER, CipherUtil.decrypt(tokenInfo[1], System.getenv(Constants.ENV_ENCRYPTION_KEY)));

            Authentication authentication = this.authenticationManager.authenticate(authenticationToken);
            SecurityContextHolder.getContext().setAuthentication(authentication);

            String jwt = tokenProvider.createToken(authentication, true, true);

            HttpHeaders httpHeaders = new HttpHeaders();
            httpHeaders.add(JWTFilter.AUTHORIZATION_HEADER, "Bearer " + jwt);

            return new ResponseEntity<>(new JWTToken(jwt, true), httpHeaders, HttpStatus.OK);

    }

    @AuditEvent(
            attemptType = ApplicationEventType.TFA_CODE_VERIFY_ATTEMPT,
            attemptMessage = "Verification attempt for second-factor authentication",
            successType = ApplicationEventType.AUTH_SUCCESS,
            successMessage = "Login successfully completed"
    )
    @GetMapping("/tfa/verifyCode")
    public ResponseEntity<JWTToken> verifyCode(@RequestParam String code) {

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
    }

}
