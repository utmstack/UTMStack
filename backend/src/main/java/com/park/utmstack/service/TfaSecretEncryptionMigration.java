package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

/**
 * Boots the one-time TFA-secret-at-rest migration: legacy plaintext seeds in
 * jhi_user.tfa_secret are re-encrypted so a DB dump no longer contains usable
 * 2FA material. Idempotent (already-encrypted rows are skipped) and safe to run
 * on every boot. If ENCRYPTION_KEY is not configured the migration is skipped
 * with a warning — TFA continues to work on the plaintext fallback so this can
 * never brick logins.
 */
@Component
public class TfaSecretEncryptionMigration implements CommandLineRunner {

    private static final String CLASSNAME = "TfaSecretEncryptionMigration";
    private final Logger log = LoggerFactory.getLogger(TfaSecretEncryptionMigration.class);

    private final UserService userService;

    public TfaSecretEncryptionMigration(UserService userService) {
        this.userService = userService;
    }

    @Override
    public void run(String... args) {
        log.info(CLASSNAME + ": starting TFA-secret at-rest migration check");
        if (!org.springframework.util.StringUtils.hasText(System.getenv(Constants.ENV_ENCRYPTION_KEY))) {
            log.warn(CLASSNAME + ": " + Constants.ENV_ENCRYPTION_KEY + " not configured; TFA secrets remain plaintext. Skipping migration.");
            return;
        }
        int migrated = userService.migrateLegacyTfaSecrets();
        if (migrated > 0) {
            log.info(CLASSNAME + ": encrypted {} legacy TFA secret(s) at rest", migrated);
        } else {
            log.info(CLASSNAME + ": all TFA secrets already encrypted at rest");
        }
    }
}
