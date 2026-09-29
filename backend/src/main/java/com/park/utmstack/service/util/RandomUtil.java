package com.park.utmstack.service.util;

import java.security.SecureRandom;

/**
 * Utility class for generating secure random strings.
 *
 * <p>All generators use a cryptographically secure source
 * ({@link SecureRandom}) — never {@code java.util.Random} / commons-lang3
 * {@code RandomStringUtils}, whose shared 48-bit LCG is predictable and would
 * undermine the reset/activation keys that protect unauthenticated endpoints.
 */
public final class RandomUtil {

    private static final int DEF_COUNT = 20;
    private static final SecureRandom SECURE_RANDOM = new SecureRandom();

    private RandomUtil() {
    }

    /**
     * Generate a password (printable graph characters).
     */
    public static String generatePassword() {
        return randomGraph(DEF_COUNT);
    }

    /**
     * Generate an activation key (numeric).
     */
    public static String generateActivationKey() {
        return randomNumeric(DEF_COUNT);
    }

    /**
     * Generate a password-reset key (numeric).
     */
    public static String generateResetKey() {
        return randomNumeric(DEF_COUNT);
    }

    private static String randomNumeric(int count) {
        StringBuilder sb = new StringBuilder(count);
        for (int i = 0; i < count; i++) {
            sb.append(SECURE_RANDOM.nextInt(10));
        }
        return sb.toString();
    }

    private static String randomGraph(int count) {
        // Same printable "graph" character set as RandomStringUtils.randomGraph:
        // ASCII 33-126.
        StringBuilder sb = new StringBuilder(count);
        for (int i = 0; i < count; i++) {
            int c = 33 + SECURE_RANDOM.nextInt(126 - 33 + 1);
            sb.append((char) c);
        }
        return sb.toString();
    }
}
