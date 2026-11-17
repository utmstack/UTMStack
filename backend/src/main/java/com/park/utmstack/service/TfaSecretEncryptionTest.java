package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Security regression guard (2026-11-17): TFA secrets (TOTP/email seeds) must be
 * encrypted at rest via the house CipherUtil, never stored as the raw seed, and
 * legacy plaintext rows must keep working until the one-time migration runs.
 *
 * <p>The cipher helpers are static and key-parameterized, so these tests run
 * hermetically — no Spring context, no environment variable, no DB.
 */
class TfaSecretEncryptionTest {

    private static final String KEY = "tfa-unit-test-key-not-a-real-secret";
    private static final String OTHER_KEY = "a-different-key-entirely";

    @Test
    void encryptedPredicateRecognizesPrefixOnly() {
        assertTrue(UserService.isEncryptedTfaSecret(Constants.TFA_SECRET_ENCRYPTED_PREFIX + "abc123"));
        assertFalse(UserService.isEncryptedTfaSecret("JBSWY3DPEHPK3PXP"));
        assertFalse(UserService.isEncryptedTfaSecret(null));
        assertFalse(UserService.isEncryptedTfaSecret(""));
    }

    @Test
    void decryptPassesThroughLegacyPlaintextAndEmpty() {
        assertEquals("JBSWY3DPEHPK3PXP", UserService.decryptTfaSecret("JBSWY3DPEHPK3PXP", KEY));
        assertEquals(null, UserService.decryptTfaSecret(null, KEY));
        assertEquals("", UserService.decryptTfaSecret("", KEY));
    }

    @Test
    void encryptThenDecryptRoundTrips() {
        String seed = "JBSWY3DPEHPK3PXP";
        String stored = UserService.encryptTfaSecret(seed, KEY);
        assertTrue(UserService.isEncryptedTfaSecret(stored), "stored value must carry the enc:: prefix");
        assertNotEquals(seed, stored, "stored value must not be the raw seed");
        assertEquals(seed, UserService.decryptTfaSecret(stored, KEY));
    }

    @Test
    void encryptIsIdempotentOnEncryptedValues() {
        String once = UserService.encryptTfaSecret("KRSXG5C4ZVZL6Y4Q", KEY);
        assertEquals(once, UserService.encryptTfaSecret(once, KEY),
                "re-encrypting an encrypted value must be a no-op (migration idempotency)");
    }

    @Test
    void encryptLeavesNullAndBlankUntouched() {
        assertEquals(null, UserService.encryptTfaSecret(null, KEY));
        assertEquals("   ", UserService.encryptTfaSecret("   ", KEY));
    }

    @Test
    void differentKeysProduceDifferentCiphertexts() {
        String fromKeyA = UserService.encryptTfaSecret("JBSWY3DPEHPK3PXP", KEY);
        String fromKeyB = UserService.encryptTfaSecret("JBSWY3DPEHPK3PXP", OTHER_KEY);
        assertNotEquals(fromKeyA, fromKeyB, "same seed under different keys must not collide");
    }

    @Test
    void decryptWithWrongKeyNeverReturnsTheSeed() {
        String seed = "JBSWY3DPEHPK3PXP";
        String stored = UserService.encryptTfaSecret(seed, KEY);
        try {
            String out = UserService.decryptTfaSecret(stored, OTHER_KEY);
            assertNotEquals(seed, out, "decryption under a different key must never reproduce the seed");
        } catch (RuntimeException e) {
            // AES/CBC PKCS5 padding check fails on a wrong key — the guarantee holds.
        }
    }
}
