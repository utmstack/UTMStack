package com.park.utmstack.service.validators.allowlist;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class ThreatIntelAllowlistValidatorTest {

    @Test
    void acceptsValidIpsCidrsAndDomainsAndDeduplicates() {
        String raw = " 198.51.100.10, 203.0.113.0/24\n2001:db8::/32, Example.COM., *.Example.net, 198.51.100.10 ";
        String normalized = ThreatIntelAllowlistValidator.validateAndNormalize(raw);
        assertEquals("198.51.100.10,203.0.113.0/24,2001:db8:0:0:0:0:0:0/32,example.com,*.example.net", normalized);
    }

    @Test
    void acceptsEmptyOrBlank() {
        assertEquals("", ThreatIntelAllowlistValidator.validateAndNormalize(null));
        assertEquals("", ThreatIntelAllowlistValidator.validateAndNormalize("   , \n "));
    }

    @Test
    void rejectsMalformedEntries() {
        assertThrows(IllegalArgumentException.class, () -> ThreatIntelAllowlistValidator.validateAndNormalize("999.1.1.1"));
        assertThrows(IllegalArgumentException.class, () -> ThreatIntelAllowlistValidator.validateAndNormalize("198.51.100.10/33"));
        assertThrows(IllegalArgumentException.class, () -> ThreatIntelAllowlistValidator.validateAndNormalize("*.com"));
        assertThrows(IllegalArgumentException.class, () -> ThreatIntelAllowlistValidator.validateAndNormalize("bad host name"));
        assertThrows(IllegalArgumentException.class, () -> ThreatIntelAllowlistValidator.validateAndNormalize("https://example.com"));
    }
}
