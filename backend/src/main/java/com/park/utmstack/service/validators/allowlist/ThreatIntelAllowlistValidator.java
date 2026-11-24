package com.park.utmstack.service.validators.allowlist;

import java.net.Inet6Address;
import java.net.InetAddress;
import java.net.UnknownHostException;
import java.util.ArrayList;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Locale;
import java.util.Set;
import java.util.regex.Pattern;

/**
 * Validates and normalizes Threat Intelligence allowlist entries:
 * IPv4/IPv6 addresses, IPv4/IPv6 CIDR blocks, and exact or wildcard domains/hostnames.
 */
public final class ThreatIntelAllowlistValidator {

    private static final Pattern IPV4_PATTERN = Pattern.compile(
        "^(25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)(\\.(25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)){3}$"
    );
    private static final Pattern IPV6_SYNTAX = Pattern.compile("^[0-9a-fA-F:.]+$");
    private static final Pattern DOMAIN_LABEL = Pattern.compile("^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$");

    private ThreatIntelAllowlistValidator() {}

    public static String validateAndNormalize(String rawValue) {
        if (rawValue == null || rawValue.isBlank()) {
            return "";
        }

        String[] tokens = rawValue.split("[,\\r\\n]+");
        Set<String> seen = new LinkedHashSet<>();
        List<String> normalized = new ArrayList<>();

        for (String token : tokens) {
            String trimmed = token.trim();
            if (trimmed.isEmpty()) {
                continue;
            }
            String canonical = validateEntry(trimmed);
            if (seen.add(canonical.toLowerCase(Locale.ROOT))) {
                normalized.add(canonical);
            }
        }
        return String.join(",", normalized);
    }

    public static String validateEntry(String entry) {
        if (entry == null || entry.isBlank()) {
            throw new IllegalArgumentException("Allowlist entry must not be blank");
        }
        String trimmed = entry.trim();

        if (trimmed.contains("/")) {
            return validateCidr(trimmed);
        }
        if (IPV4_PATTERN.matcher(trimmed).matches()) {
            return trimmed;
        }
        if (trimmed.contains(":")) {
            return validateIpv6(trimmed);
        }
        return validateDomain(trimmed);
    }

    private static String validateCidr(String cidr) {
        int slash = cidr.indexOf('/');
        if (slash <= 0 || slash != cidr.lastIndexOf('/') || slash == cidr.length() - 1) {
            throw new IllegalArgumentException("Invalid CIDR entry: '" + cidr + "'");
        }
        String ipPart = cidr.substring(0, slash);
        String prefixPart = cidr.substring(slash + 1);
        int prefix;
        try {
            prefix = Integer.parseInt(prefixPart);
        } catch (NumberFormatException e) {
            throw new IllegalArgumentException("Invalid CIDR prefix: '" + cidr + "'");
        }

        if (IPV4_PATTERN.matcher(ipPart).matches()) {
            if (prefix < 0 || prefix > 32) {
                throw new IllegalArgumentException("IPv4 CIDR prefix must be between 0 and 32: '" + cidr + "'");
            }
            return ipPart + "/" + prefix;
        }
        if (ipPart.contains(":")) {
            String normalizedIp = validateIpv6(ipPart);
            if (prefix < 0 || prefix > 128) {
                throw new IllegalArgumentException("IPv6 CIDR prefix must be between 0 and 128: '" + cidr + "'");
            }
            return normalizedIp + "/" + prefix;
        }
        throw new IllegalArgumentException("Invalid CIDR IP address: '" + cidr + "'");
    }

    private static String validateIpv6(String value) {
        if (!IPV6_SYNTAX.matcher(value).matches()) {
            throw new IllegalArgumentException("Invalid IPv6 address: '" + value + "'");
        }
        try {
            InetAddress addr = InetAddress.getByName(value);
            if (!(addr instanceof Inet6Address)) {
                throw new IllegalArgumentException("Invalid IPv6 address: '" + value + "'");
            }
            return addr.getHostAddress();
        } catch (UnknownHostException e) {
            throw new IllegalArgumentException("Invalid IPv6 address: '" + value + "'");
        }
    }

    private static String validateDomain(String raw) {
        String lower = raw.toLowerCase(Locale.ROOT);
        if (lower.endsWith(".")) {
            lower = lower.substring(0, lower.length() - 1);
        }
        boolean wildcard = lower.startsWith("*.");
        String host = wildcard ? lower.substring(2) : lower;

        if (host.isEmpty() || host.length() > 253) {
            throw new IllegalArgumentException("Invalid domain/hostname entry: '" + raw + "'");
        }
        if (wildcard && !host.contains(".")) {
            throw new IllegalArgumentException("Wildcard domain must contain a dot (e.g. *.example.com): '" + raw + "'");
        }

        String[] labels = host.split("\\.", -1);
        boolean allNumeric = true;
        for (String label : labels) {
            if (!DOMAIN_LABEL.matcher(label).matches()) {
                throw new IllegalArgumentException("Invalid domain/hostname entry: '" + raw + "'");
            }
            if (!label.chars().allMatch(Character::isDigit)) {
                allNumeric = false;
            }
        }
        if (allNumeric) {
            throw new IllegalArgumentException("Invalid IP or domain entry: '" + raw + "'");
        }
        return wildcard ? "*." + host : host;
    }
}
