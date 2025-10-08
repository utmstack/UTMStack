package com.park.utmstack.service.tfa;

import com.park.utmstack.aop.logging.Loggable;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.tfa.TfaMethod;
import com.park.utmstack.domain.tfa.TfaSetupState;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.dto.tfa.init.Delivery;
import com.park.utmstack.service.dto.tfa.init.TfaInitResponse;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyResponse;
import com.park.utmstack.util.exceptions.TooManyRequestsException;
import com.park.utmstack.util.exceptions.UtmMailException;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.concurrent.TimeUnit;

@Service
@RequiredArgsConstructor
public class EmailTfaService implements TfaMethodService {

    private static final String CLASSNAME = "EmailTfaService";
    private final CacheService cache;
    private final UserService userService;
    private final EmailTotpService tfaService;
    private final MailService mailService;

    @Override
    public TfaMethod getMethod() {
        return TfaMethod.EMAIL;
    }

    @Override
    @Loggable
    public TfaInitResponse initiateSetup(User user) {
        final String ctx = CLASSNAME + ".initiateSetup";
        try {
            mailService.sendCheckEmail(List.of(user.getEmail()));

            String secret = tfaService.generateSecret();
            String code = tfaService.generateCode(secret);

            long expiresAt = System.currentTimeMillis() + TimeUnit.SECONDS.toMillis(30);
            TfaSetupState state = new TfaSetupState(secret, expiresAt);
            cache.storeState(user.getLogin(), TfaMethod.EMAIL, state);

            mailService.sendTfaVerificationCode(user, code);

            Delivery delivery = new Delivery(TfaMethod.EMAIL, "Code sent to email " + user.getEmail());
            return new TfaInitResponse("pending", delivery, 300);
        }
        catch (Exception e) {
            throw new UtmMailException(ctx + ": " + e.getLocalizedMessage());
        }
    }


    @Override
    @Loggable
    public TfaVerifyResponse verifyCode(User user, String code) {

        TfaSetupState tfaSetupState = cache.getState(user.getLogin(), TfaMethod.EMAIL)
                .orElseThrow(() -> new IllegalStateException("No TFA setup found for user: " + user.getLogin()));

        boolean expired = tfaSetupState.isExpired();
        boolean valid = !expired && tfaService.validateCode(tfaSetupState.getSecret(), code);

        return new TfaVerifyResponse(
                valid,
                expired,
                tfaSetupState.getRemainingSeconds(),
                expired ? "Code expired" : "Code verification " + (valid ? "successful" : "failed")
        );
    }

    @Override
    public void persistConfiguration(User user) {
        String secret = cache.getState(user.getLogin(), TfaMethod.EMAIL)
                .orElseThrow(() -> new IllegalStateException("No TFA setup found for user: " + user.getLogin()))
                .getSecret();
        userService.updateUserTfaSecret(user.getLogin(), secret, TfaMethod.EMAIL.toString());
        cache.clear(user.getLogin(), TfaMethod.EMAIL);
    }

    @Override
    public void generateChallenge(User user) {
        String secret = user.getTfaSecret();
        String code = tfaService.generateCode(secret);

        TfaSetupState state = new TfaSetupState(secret, System.currentTimeMillis() + (Constants.EXPIRES_IN_SECONDS_EMAIL * 4) * 1000 * 10);
        cache.storeState(user.getLogin(), TfaMethod.EMAIL, state);

        mailService.sendTfaVerificationCode(user, code);
    }

    @Override
    public void regenerateChallenge(User user) {

        TfaSetupState state = cache.getState(user.getLogin(), TfaMethod.EMAIL)
                .orElseThrow(() -> new IllegalStateException("No TFA setup found for user: " + user.getLogin()));

        if (!state.canRequestChallenge()){
            throw new TooManyRequestsException("Challenge request too soon. Please wait " + state.getCooldownRemainingSeconds() + " seconds.");
        }

        state.setExpiresAt(System.currentTimeMillis() + TimeUnit.SECONDS.toMillis(Constants.EXPIRES_IN_SECONDS_EMAIL));
        state.markChallengeRequested();

        mailService.sendTfaVerificationCode(user, tfaService.generateCode(state.getSecret()));
        cache.storeState(user.getLogin(), TfaMethod.EMAIL, state);

    }

    @Override
    public long expirationTimeSeconds() {
        return Constants.EXPIRES_IN_SECONDS_EMAIL;
    }
}

