package extractor

import (
	"fmt"

	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/models"
)

type FieldExtractor struct{}

func NewFieldExtractor() *FieldExtractor {
	return &FieldExtractor{}
}

func (e *FieldExtractor) ExtractFromEvent(event *models.Event) []*models.FlattenedField {
	fields := make([]*models.FlattenedField, 0, 50)

	if event.Origin != nil {
		fields = append(fields, e.extractFromSide(event.Origin, fmt.Sprintf("event.%s.origin", event.ID))...)
	}

	if event.Target != nil {
		fields = append(fields, e.extractFromSide(event.Target, fmt.Sprintf("event.%s.target", event.ID))...)
	}

	return fields
}

func (e *FieldExtractor) extractFromSide(side *models.Side, prefix string) []*models.FlattenedField {
	fields := make([]*models.FlattenedField, 0, 60)

	// ==================== Network Identification ====================
	if side.IP != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".ip",
			Key:   "ip",
			Value: side.IP,
		})
	}

	if side.Port != 0 {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".port",
			Key:   "port",
			Value: side.Port,
		})
	}

	if side.Host != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".host",
			Key:   "hostname",
			Value: side.Host,
		})
	}

	if side.Domain != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".domain",
			Key:   "domain",
			Value: side.Domain,
		})
	}

	if side.Mac != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".mac",
			Key:   "mac-address",
			Value: side.Mac,
		})
	}

	if side.URL != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".url",
			Key:   "url",
			Value: side.URL,
		})
	}

	if side.Cidr != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".cidr",
			Key:   "cidr",
			Value: side.Cidr,
		})
	}

	// ==================== Geolocation ====================
	if side.Geolocation != nil && side.Geolocation.ASN != 0 {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".geolocation.asn",
			Key:   "asn",
			Value: side.Geolocation.ASN,
		})
	}

	// ==================== Certificates & Fingerprints ====================
	if side.CertificateFingerprint != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".certificateFingerprint",
			Key:   "certificate-fingerprint",
			Value: side.CertificateFingerprint,
		})
	}

	if side.Ja3Fingerprint != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".ja3Fingerprint",
			Key:   "ja3-fingerprint",
			Value: side.Ja3Fingerprint,
		})
	}

	if side.JarmFingerprint != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".jarmFingerprint",
			Key:   "jarm-fingerprint",
			Value: side.JarmFingerprint,
		})
	}

	if side.SshBanner != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sshBanner",
			Key:   "ssh-banner",
			Value: side.SshBanner,
		})
	}

	if side.SshFingerprint != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sshFingerprint",
			Key:   "ssh-fingerprint",
			Value: side.SshFingerprint,
		})
	}

	// ==================== Web Attributes ====================
	if side.Cookie != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".cookie",
			Key:   "cookie",
			Value: side.Cookie,
		})
	}

	if side.JabberId != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".jabberId",
			Key:   "jabber-id",
			Value: side.JabberId,
		})
	}

	// ==================== Email Attributes ====================
	if side.EmailAddress != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".emailAddress",
			Key:   "email-address",
			Value: side.EmailAddress,
		})
	}

	if side.EmailBody != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".emailBody",
			Key:   "email-body",
			Value: side.EmailBody,
		})
	}

	if side.EmailDisplayName != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".emailDisplayName",
			Key:   "email-display-name",
			Value: side.EmailDisplayName,
		})
	}

	if side.EmailSubject != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".emailSubject",
			Key:   "email-subject",
			Value: side.EmailSubject,
		})
	}

	if side.EmailThreadIndex != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".emailThreadIndex",
			Key:   "email-thread-index",
			Value: side.EmailThreadIndex,
		})
	}

	if side.EmailXMailer != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".emailXMailer",
			Key:   "email-x-mailer",
			Value: side.EmailXMailer,
		})
	}

	if side.Dkim != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".dkim",
			Key:   "dkim",
			Value: side.Dkim,
		})
	}

	if side.DkimSignature != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".dkimSignature",
			Key:   "dkim-signature",
			Value: side.DkimSignature,
		})
	}

	// ==================== WHOIS Attributes ====================
	if side.WhoisRegistrant != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".whoisRegistrant",
			Key:   "whois-registrant",
			Value: side.WhoisRegistrant,
		})
	}

	if side.WhoisRegistrar != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".whoisRegistrar",
			Key:   "whois-registrar",
			Value: side.WhoisRegistrar,
		})
	}

	// ==================== Identity Attributes ====================
	if side.User != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".user",
			Key:   "username",
			Value: side.User,
		})
	}

	// ==================== Process Attributes ====================
	if side.Process != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".process",
			Key:   "process",
			Value: side.Process,
		})
	}

	if side.ProcessState != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".processState",
			Key:   "process-state",
			Value: side.ProcessState,
		})
	}

	if side.WindowsScheduledTask != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".windowsScheduledTask",
			Key:   "windows-scheduled-task",
			Value: side.WindowsScheduledTask,
		})
	}

	if side.WindowsServiceDisplayName != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".windowsServiceDisplayName",
			Key:   "windows-service-display-name",
			Value: side.WindowsServiceDisplayName,
		})
	}

	if side.WindowsServiceName != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".windowsServiceName",
			Key:   "windows-service-name",
			Value: side.WindowsServiceName,
		})
	}

	// ==================== File Attributes ====================
	if side.File != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".file",
			Key:   "file",
			Value: side.File,
		})
	}

	if side.Path != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".path",
			Key:   "path",
			Value: side.Path,
		})
	}

	if side.Filename != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".filename",
			Key:   "filename",
			Value: side.Filename,
		})
	}

	if side.SizeInBytes != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sizeInBytes",
			Key:   "size-in-bytes",
			Value: side.SizeInBytes,
		})
	}

	if side.MimeType != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".mimeType",
			Key:   "mime-type",
			Value: side.MimeType,
		})
	}

	// ==================== Hash Attributes ====================
	if side.Authentihash != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".authentihash",
			Key:   "authentihash",
			Value: side.Authentihash,
		})
	}

	if side.Cdhash != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".cdhash",
			Key:   "cdhash",
			Value: side.Cdhash,
		})
	}

	if side.MD5 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".md5",
			Key:   "md5",
			Value: side.MD5,
		})
	}

	if side.SHA1 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha1",
			Key:   "sha1",
			Value: side.SHA1,
		})
	}

	if side.SHA224 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha224",
			Key:   "sha224",
			Value: side.SHA224,
		})
	}

	if side.SHA256 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha256",
			Key:   "sha256",
			Value: side.SHA256,
		})
	}

	if side.SHA384 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha384",
			Key:   "sha384",
			Value: side.SHA384,
		})
	}

	if side.SHA3224 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha3224",
			Key:   "sha3-224",
			Value: side.SHA3224,
		})
	}

	if side.SHA3256 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha3256",
			Key:   "sha3-256",
			Value: side.SHA3256,
		})
	}

	if side.SHA3384 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha3384",
			Key:   "sha3-384",
			Value: side.SHA3384,
		})
	}

	if side.SHA3512 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha3512",
			Key:   "sha3-512",
			Value: side.SHA3512,
		})
	}

	if side.SHA512 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha512",
			Key:   "sha512",
			Value: side.SHA512,
		})
	}

	if side.SHA512224 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha512224",
			Key:   "sha512-224",
			Value: side.SHA512224,
		})
	}

	if side.SHA512256 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".sha512256",
			Key:   "sha512-256",
			Value: side.SHA512256,
		})
	}

	if side.Hex != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".hex",
			Key:   "hex",
			Value: side.Hex,
		})
	}

	if side.Base64 != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".base64",
			Key:   "base64",
			Value: side.Base64,
		})
	}

	// ==================== System Attributes ====================
	if side.ChromeExtension != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".chromeExtension",
			Key:   "chrome-extension-id",
			Value: side.ChromeExtension,
		})
	}

	if side.MobileAppId != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".mobileAppId",
			Key:   "mobile-app-id",
			Value: side.MobileAppId,
		})
	}

	// ==================== Vulnerability Attributes ====================
	if side.Cpe != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".cpe",
			Key:   "cpe",
			Value: side.Cpe,
		})
	}

	if side.Cve != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".cve",
			Key:   "cve",
			Value: side.Cve,
		})
	}

	// ==================== Malware Attributes ====================
	if side.Malware != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".malware",
			Key:   "malware",
			Value: side.Malware,
		})
	}

	if side.MalwareFamily != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".malwareFamily",
			Key:   "malware-family",
			Value: side.MalwareFamily,
		})
	}

	if side.MalwareType != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".malwareType",
			Key:   "malware-type",
			Value: side.MalwareType,
		})
	}

	// ==================== Key Attributes ====================
	if side.PgpPrivateKey != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".pgpPrivateKey",
			Key:   "pgp-private-key",
			Value: side.PgpPrivateKey,
		})
	}

	if side.PgpPublicKey != "" {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".pgpPublicKey",
			Key:   "pgp-public-key",
			Value: side.PgpPublicKey,
		})
	}

	return fields
}
