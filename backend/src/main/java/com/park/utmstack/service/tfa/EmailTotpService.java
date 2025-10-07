package com.park.utmstack.service.tfa;

import dev.samstevens.totp.code.CodeGenerator;
import dev.samstevens.totp.code.DefaultCodeGenerator;
import dev.samstevens.totp.code.DefaultCodeVerifier;
import dev.samstevens.totp.secret.DefaultSecretGenerator;
import dev.samstevens.totp.secret.SecretGenerator;
import dev.samstevens.totp.time.SystemTimeProvider;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;

import static com.park.utmstack.config.Constants.EXPIRES_IN_SECONDS_EMAIL;

@Service
public class EmailTotpService {
    private static final String CLASSNAME = "TfaService";

    private final SystemTimeProvider timeProvider = new SystemTimeProvider();

    /**
     * @return
     */
    public String generateSecret() {
        final String ctx = CLASSNAME + ".generateSecret";
        try {
            SecretGenerator secretGenerator = new DefaultSecretGenerator();
            return secretGenerator.generate();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     *
     */
    public String generateCode(String secret) {
        final String ctx = CLASSNAME + ".generateCode";
        try {
            Assert.hasText(secret, "Secret value is missing");
            CodeGenerator codeGenerator = new DefaultCodeGenerator();
            return codeGenerator.generate(secret, Math.floorDiv(timeProvider.getTime(), EXPIRES_IN_SECONDS_EMAIL));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     *
     * @param secret
     * @param code
     * @return
     */
    public boolean validateCode(String secret, String code) {
        final String ctx = CLASSNAME + ".validateCode";
        try {
            Assert.hasText(secret, "Secret value is missing");
            Assert.hasText(code, "Code value is missing");

            DefaultCodeVerifier verifier = new DefaultCodeVerifier(new DefaultCodeGenerator(), timeProvider);
            verifier.setTimePeriod(EXPIRES_IN_SECONDS_EMAIL);
            verifier.setAllowedTimePeriodDiscrepancy(1);

            return verifier.isValidCode(secret, code);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
