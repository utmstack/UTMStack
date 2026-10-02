package usecase

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
)

// idpNameRe keeps names URL-safe: they sit in /sso/<name>/login.
var idpNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// idpHostRe matches a bare hostname or IP: no scheme, no path, no spaces.
var idpHostRe = regexp.MustCompile(`^[A-Za-z0-9.-]+$`)

// Presence + secrets are decided by prepareSettings; this only rejects malformed values.
func validateIDPSettingsFormat(kind domain.ProviderType, raw json.RawMessage) error {
	switch kind {
	case domain.ProviderSAML:
		var s domain.SAMLSettings
		if err := json.Unmarshal(raw, &s); err != nil {
			return domain.ErrIDPSettingsInvalid
		}
		return validateSAMLFormat(s)
	case domain.ProviderOIDC:
		var s domain.OIDCSettings
		if err := json.Unmarshal(raw, &s); err != nil {
			return domain.ErrIDPSettingsInvalid
		}
		return validateOIDCFormat(s)
	case domain.ProviderLDAP:
		var s domain.LDAPSettings
		if err := json.Unmarshal(raw, &s); err != nil {
			return domain.ErrIDPSettingsInvalid
		}
		return validateLDAPFormat(s)
	default:
		return domain.ErrIDPTypeUnsupported
	}
}

func validateSAMLFormat(s domain.SAMLSettings) error {
	 if !idpIsHTTPURL(s.MetadataURL) || !idpIsHTTPURL(s.SpACSURL) ||
		 !idpIsURI(s.SpEntityID) ||
		!strings.Contains(s.SpCertificatePem, "-----BEGIN CERTIFICATE-----") ||
		!strings.Contains(s.SpCertificatePem, "-----END CERTIFICATE-----") {
		return domain.ErrIDPSettingsInvalid
	}
	return nil
}

func validateOIDCFormat(s domain.OIDCSettings) error {
	if !idpIsHTTPSURL(s.Issuer) || !idpIsRedirectURL(s.RedirectURL) {
		return domain.ErrIDPSettingsInvalid
	}
	return nil
}

func validateLDAPFormat(s domain.LDAPSettings) error {
	host := strings.TrimSpace(s.Host)
	filter := strings.TrimSpace(s.UserFilter)
	if host == "" || s.Port < 0 || s.Port > 65535 ||
		!idpHostRe.MatchString(host) ||
		strings.Count(filter, "(") != strings.Count(filter, ")") {
		return domain.ErrIDPSettingsInvalid
	}
	return nil
}

func idpIsHTTPURL(v string) bool {
	u, err := url.Parse(strings.TrimSpace(v))
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func idpIsHTTPSURL(v string) bool {
	u, err := url.Parse(strings.TrimSpace(v))
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "https"
}

// idpIsURI accepts http(s) URLs or urn: identifiers (SAML entity IDs).
func idpIsURI(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	if strings.HasPrefix(strings.ToLower(v), "urn:") && len(v) > 4 {
		return true
	}
	return idpIsHTTPURL(v)
}

func idpIsRedirectURL(v string) bool {
	u, err := url.Parse(strings.TrimSpace(v))
	if err != nil || u.Host == "" {
		return false
	}
	if u.Scheme == "https" {
		return true
	}
	return u.Scheme == "http" && isLoopbackHost(u.Hostname())
}

func isLoopbackHost(h string) bool {
	switch h {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return false
}
