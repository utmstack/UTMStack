package com.park.utmstack.web.rest.tfa;

import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.Authority;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.UtmConfigurationParameter;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.tfa.TfaMethod;
import com.park.utmstack.loggin.LogContextBuilder;
import com.park.utmstack.security.jwt.JWTFilter;
import com.park.utmstack.security.jwt.TokenProvider;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.UtmConfigurationParameterService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.jwt.JWTToken;
import com.park.utmstack.service.dto.tfa.init.TfaInitRequest;
import com.park.utmstack.service.dto.tfa.init.TfaInitResponse;
import com.park.utmstack.service.dto.tfa.save.TfaSaveRequest;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyRequest;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyResponse;
import com.park.utmstack.service.tfa.TfaService;
import com.park.utmstack.util.ResponseUtil;
import com.park.utmstack.util.exceptions.TfaVerificationException;
import com.park.utmstack.util.exceptions.UtmMailException;
import io.swagger.v3.oas.annotations.Hidden;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.stream.Collectors;

import static com.park.utmstack.config.Constants.PROP_TFA_METHOD;

@RestController
@RequiredArgsConstructor
@Slf4j
@Hidden
@RequestMapping("api/tfa")
public class TfaResource {

    private static final String CLASSNAME = "TfaController";

    private final TfaService tfaService;
    private final UserService userService;
    private final ApplicationEventService applicationEventService;
    private final UtmConfigurationParameterService utmConfigurationParameterService;
    private final TokenProvider tokenProvider;
    private final LogContextBuilder logContextBuilder;

    @PostMapping("/init")
    public ResponseEntity<TfaInitResponse> initTfa(@RequestBody TfaInitRequest request) {
        final String ctx = CLASSNAME + ".initTfa";
        try {
            User user = userService.getCurrentUserLogin();
            TfaInitResponse response = tfaService.initiateSetup(user, request.getMethod());
            return ResponseEntity.ok(response);
        } catch (Exception e){
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }

    }

    @PostMapping("/verify")
    public ResponseEntity<TfaVerifyResponse> verifyTfa(@RequestBody TfaVerifyRequest request) {
        final String ctx = CLASSNAME + ".verifyTfa";
        try {
            User user = userService.getCurrentUserLogin();
            TfaVerifyResponse response = tfaService.verifyCode(user, request);
            return ResponseEntity.ok(response);

        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @GetMapping("/refresh")
    public ResponseEntity<Void> generateChallenge() {
        final String ctx = CLASSNAME + ".generateChallenge";
        try {
            User user = userService.getCurrentUserLogin();
            tfaService.regenerateChallenge(user);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @PostMapping("/complete")
    public ResponseEntity<Void> completeTfa(@RequestBody TfaSaveRequest request) {
        final String ctx = CLASSNAME + ".completeTfa";
        try {

            List<UtmConfigurationParameter> tfaParams = utmConfigurationParameterService.getConfigParameterBySectionId(Constants.TFA_SETTING_ID);

            for (UtmConfigurationParameter param : tfaParams) {
                switch (param.getConfParamShort()) {
                    case PROP_TFA_METHOD:
                        param.setConfParamValue(String.valueOf(request.getMethod()));
                        break;
                    case Constants.PROP_TFA_ENABLE:
                        param.setConfParamValue(String.valueOf(request.isEnable()));
                        break;
                }
            }

            tfaService.persistConfiguration(request.getMethod());
            User user = userService.getCurrentUserLogin();
            utmConfigurationParameterService.saveAllConfigParams(tfaParams);
            tfaService.generateChallenge(user);
            return ResponseEntity.ok().build();
        } catch (UtmMailException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildPreconditionFailedResponse(msg);
        } catch (IllegalArgumentException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildBadRequestResponse(msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @AuditEvent(
            attemptType = ApplicationEventType.TFA_CODE_VERIFY_ATTEMPT,
            attemptMessage = "Verification attempt for second-factor authentication",
            successType = ApplicationEventType.AUTH_SUCCESS,
            successMessage = "Login successfully completed"
    )
    @PostMapping("/verify-code")
    public ResponseEntity<JWTToken> verifyCode(@RequestBody String code, HttpServletRequest request) {
        final String ctx = CLASSNAME + ".verifyCode";
        User user = userService.getCurrentUserLogin();
        TfaMethod method = TfaMethod.valueOf(user.getTfaMethod());
        TfaVerifyRequest tfaVerifyRequest = new TfaVerifyRequest(method, code);
        TfaVerifyResponse response = tfaService.verifyCode(user, tfaVerifyRequest);

        if (!response.isValid()) {
            throw new TfaVerificationException("TFA invalid for user '" + user.getLogin() + "': " + response.getMessage());
        }

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
