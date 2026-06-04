package com.park.utmstack.service.tfa;

import com.google.zxing.BarcodeFormat;
import com.google.zxing.MultiFormatWriter;
import com.google.zxing.client.j2se.MatrixToImageWriter;
import com.google.zxing.common.BitMatrix;
import com.park.utmstack.aop.logging.Loggable;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.tfa.TfaMethod;
import com.park.utmstack.domain.tfa.TfaSetupState;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.dto.tfa.init.Delivery;
import com.park.utmstack.service.dto.tfa.init.TfaInitResponse;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyResponse;
import com.park.utmstack.util.exceptions.TooManyRequestsException;
import com.warrenstrange.googleauth.GoogleAuthenticator;
import org.springframework.stereotype.Service;

import javax.imageio.ImageIO;
import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.util.Base64;
import java.util.concurrent.TimeUnit;

import static com.park.utmstack.config.Constants.TFA_ISSUER;

@Service
public class TotpTfaService implements TfaMethodService {

    private final GoogleAuthenticator authenticator;
    private final CacheService cache;
    private final UserService userService;

    TotpTfaService(CacheService cache, UserService userService) {
        this.userService = userService;
        this.authenticator = new GoogleAuthenticator();
        this.cache = cache;
    }

    @Override
    public TfaMethod getMethod() {
        return TfaMethod.TOTP;
    }

    @Override
    public TfaInitResponse initiateSetup(User user) {
        String secret = authenticator.createCredentials().getKey();
        long expiresAt = System.currentTimeMillis() + TimeUnit.SECONDS.toMillis(Constants.EXPIRES_IN_SECONDS_TOTP * 10);
        TfaSetupState state = new TfaSetupState(secret, expiresAt);
        cache.storeState(user.getLogin(), TfaMethod.TOTP, state);

        String uri = String.format("otpauth://totp/%s:%s?secret=%s&issuer=%s",
                TFA_ISSUER, user.getLogin(), secret, TFA_ISSUER);

        String qrBase64 = generateQrBase64(uri);
        Delivery delivery = new Delivery(TfaMethod.TOTP, qrBase64);

        return new TfaInitResponse("pending", delivery, Constants.EXPIRES_IN_SECONDS_TOTP * 10);
    }

    @Override
    public TfaVerifyResponse verifyCode(User user, String code) {
        TfaSetupState tfaSetupState = cache.getState(user.getLogin(), TfaMethod.TOTP)
                .orElseThrow(() -> new IllegalStateException("No TFA setup found for user: " + user.getLogin()));

        boolean expired = tfaSetupState.isExpired();
        boolean valid = !expired && authenticator.authorize(tfaSetupState.getSecret(), Integer.parseInt(code)) && !code.equals(tfaSetupState.getLastUsedCode());
        return new TfaVerifyResponse(
                valid,
                expired,
                tfaSetupState.getRemainingSeconds(),
                expired ? "Code expired" : "Code verification " + (valid ? "successful" : "failed")
        );
    }


    @Override
    public void persistConfiguration(User user) {
        String secret = cache.getState(user.getLogin(), TfaMethod.TOTP)
                .orElseThrow(() -> new IllegalStateException("No TFA setup found for user: " + user.getLogin()))
                .getSecret();
        userService.updateUserTfaSecret(user.getLogin(), secret, TfaMethod.TOTP.toString());
        cache.clear(user.getLogin(), TfaMethod.TOTP);
    }

    @Override
    public void generateChallenge(User user) {
        cache.clear(user.getLogin(), TfaMethod.TOTP);
        String secret = user.getTfaSecret();
        TfaSetupState state = new TfaSetupState(secret, System.currentTimeMillis() + TimeUnit.SECONDS.toMillis(Constants.EXPIRES_IN_SECONDS_TOTP));
        cache.storeState(user.getLogin(), TfaMethod.TOTP, state);
    }

    @Override
    public void regenerateChallenge(User user) {

        TfaSetupState state = cache.getState(user.getLogin(), TfaMethod.TOTP)
                .orElseThrow(() -> new IllegalStateException("No TFA setup found for user: " + user.getLogin()));

        if (!state.canRequestChallenge()){
            throw new TooManyRequestsException("Challenge request too soon. Please wait " + state.getCooldownRemainingSeconds() + " seconds.");
        }

        state.setExpiresAt(System.currentTimeMillis() + TimeUnit.SECONDS.toMillis(Constants.EXPIRES_IN_SECONDS_TOTP));
        state.markChallengeRequested();

        cache.storeState(user.getLogin(), TfaMethod.TOTP, state);
    }

    private String generateQrBase64(String uri) {
        try {
            BitMatrix matrix = new MultiFormatWriter().encode(uri, BarcodeFormat.QR_CODE, 200, 200);
            BufferedImage image = MatrixToImageWriter.toBufferedImage(matrix);

            ByteArrayOutputStream baos = new ByteArrayOutputStream();
            ImageIO.write(image, "png", baos);
            return Base64.getEncoder().encodeToString(baos.toByteArray());
        } catch (Exception e) {
            throw new RuntimeException("Error al generar QR", e);
        }
    }

    @Override
    public long expirationTimeSeconds() {
        return Constants.EXPIRES_IN_SECONDS_TOTP;
    }

}


