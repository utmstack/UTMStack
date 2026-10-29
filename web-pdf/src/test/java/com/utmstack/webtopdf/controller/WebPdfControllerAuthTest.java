package com.utmstack.webtopdf.controller;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Unit tests for the web-pdf internal-key gate (C3 hardening).
 * Plain JUnit — the auth check is extracted as a static so no Spring
 * context or WebDriver is needed.
 */
class WebPdfControllerAuthTest {

    private static final String KEY = "s3cr3t-internal-key";

    @Test
    void matchingKeysPass() {
        assertTrue(WebPdfController.internalKeyMatches(KEY, KEY));
    }

    @Test
    void mismatchedKeysFail() {
        assertFalse(WebPdfController.internalKeyMatches("wrong-key", KEY));
    }

    @Test
    void differentLengthKeysFail() {
        assertFalse(WebPdfController.internalKeyMatches(KEY.substring(1), KEY));
    }

    @Test
    void prefixKeyFails() {
        assertFalse(WebPdfController.internalKeyMatches(KEY + "x", KEY));
    }

    @Test
    void nullHeaderFails() {
        assertFalse(WebPdfController.internalKeyMatches(null, KEY));
    }

    @Test
    void emptyHeaderFails() {
        assertFalse(WebPdfController.internalKeyMatches("", KEY));
    }

    @Test
    void nullEnvFailsClosed() {
        assertFalse(WebPdfController.internalKeyMatches(KEY, null));
    }

    @Test
    void emptyEnvFailsClosed() {
        assertFalse(WebPdfController.internalKeyMatches(KEY, ""));
    }

    @Test
    void bothBlankFail() {
        assertFalse(WebPdfController.internalKeyMatches("", ""));
    }

    @Test
    void whitespaceIsNotTreatedAsBlank() {
        // A header of spaces does not equal the key; only null/"" are treated as missing.
        assertFalse(WebPdfController.internalKeyMatches("   ", KEY));
    }
}
