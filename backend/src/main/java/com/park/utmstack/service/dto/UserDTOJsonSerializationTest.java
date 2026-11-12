package com.park.utmstack.service.dto;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Security regression guard (2026-11-12): {@code GET /api/account} and
 * {@code GET /api/users[/...]} serialize {@link UserDTO} and must NOT expose the
 * user's second-factor seed ({@code tfaSecret}). Anyone holding a valid session
 * or API key would otherwise be able to read the TOTP seed and forge valid TOTP
 * codes, defeating 2FA. The seed is delivered to the user only via the TFA
 * enrollment QR code, so it must never round-trip through a JSON response.
 */
class UserDTOJsonSerializationTest {

    private final ObjectMapper mapper = new ObjectMapper();

    private UserDTO dto() {
        UserDTO dto = new UserDTO();
        dto.setLogin("admin");
        dto.setEmail("admin@example.com");
        dto.setFirstName("Admin");
        dto.setLastName("User");
        dto.setActivated(true);
        dto.setLangKey("en");
        dto.setAuthorities(Set.of("ROLE_ADMIN", "ROLE_USER"));
        dto.setTfaSecret("JBSWY3DPEHPK3PXP"); // base32 TOTP seed
        return dto;
    }

    @Test
    void tfaSecretMustNotBeSerialized() throws Exception {
        String json = mapper.writeValueAsString(dto());
        // The field must be entirely absent (not merely empty).
        assertFalse(json.contains("tfaSecret"),
            "UserDTO JSON must not contain the tfaSecret field, got: " + json);
        // The seed value itself must not leak either.
        assertFalse(json.contains("JBSWY3DPEHPK3PXP"),
            "UserDTO JSON must not contain the TOTP seed value, got: " + json);
    }

    @Test
    void otherAccountFieldsStillSerialize() throws Exception {
        String json = mapper.writeValueAsString(dto());
        assertNotNull(json);
        assertTrue(json.contains("\"login\":\"admin\""), "login should still serialize: " + json);
        assertTrue(json.contains("ROLE_ADMIN"), "authorities should still serialize: " + json);
    }
}
