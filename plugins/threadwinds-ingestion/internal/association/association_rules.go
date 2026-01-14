package association

type AssociationMode string

const (
	Association AssociationMode = "association"
	Aggregation AssociationMode = "aggregation"
)

type AssociationRule struct {
	Name       string
	SourceType string
	TargetType string
	Mode       AssociationMode
	Enabled    bool
}

var DefaultRules = []*AssociationRule{
	// ==================== Network Associations ====================
	// Bidirectional IP-Domain relationships
	{
		Name:       "ip-to-domain",
		SourceType: "ip",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "domain-to-ip",
		SourceType: "domain",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},

	// Bidirectional IP-Hostname relationships
	{
		Name:       "hostname-to-ip",
		SourceType: "hostname",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ip-to-hostname",
		SourceType: "ip",
		TargetType: "hostname",
		Mode:       Association,
		Enabled:    true,
	},

	// IP Network Infrastructure
	{
		Name:       "ip-to-port",
		SourceType: "ip",
		TargetType: "port",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "ip-to-mac",
		SourceType: "ip",
		TargetType: "mac-address",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "mac-to-ip",
		SourceType: "mac-address",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ip-to-asn",
		SourceType: "ip",
		TargetType: "asn",
		Mode:       Association,
		Enabled:    true,
	},

	// URL relationships
	{
		Name:       "url-to-domain",
		SourceType: "url",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "url-to-ip",
		SourceType: "url",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},

	// ==================== Email Associations ====================
	{
		Name:       "email-to-emailaddress",
		SourceType: "email",
		TargetType: "email-address",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "email-to-domain",
		SourceType: "email",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "email-to-ip",
		SourceType: "email",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "emailaddress-to-domain",
		SourceType: "email-address",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "domain-to-emailaddress",
		SourceType: "domain",
		TargetType: "email-address",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "dkim-to-domain",
		SourceType: "dkim",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "dkim-signature-to-email",
		SourceType: "dkim-signature",
		TargetType: "email",
		Mode:       Aggregation,
		Enabled:    true,
	},

	// ==================== File Associations ====================
	{
		Name:       "file-to-path",
		SourceType: "file",
		TargetType: "path",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "file-to-mimetype",
		SourceType: "file",
		TargetType: "mime-type",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "file-to-ip",
		SourceType: "file",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "file-to-url",
		SourceType: "file",
		TargetType: "url",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "file-to-domain",
		SourceType: "file",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},

	// File Hash Associations
	{
		Name:       "file-to-md5",
		SourceType: "file",
		TargetType: "md5",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "file-to-sha1",
		SourceType: "file",
		TargetType: "sha1",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "file-to-sha256",
		SourceType: "file",
		TargetType: "sha256",
		Mode:       Aggregation,
		Enabled:    true,
	},

	// ==================== Malware Associations ====================
	{
		Name:       "malware-to-domain",
		SourceType: "malware",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "malware-to-ip",
		SourceType: "malware",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "malware-to-file",
		SourceType: "malware",
		TargetType: "file",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "file-to-malware",
		SourceType: "file",
		TargetType: "malware",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "malware-to-url",
		SourceType: "malware",
		TargetType: "url",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "malware-to-process",
		SourceType: "malware",
		TargetType: "process",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "process-to-malware",
		SourceType: "process",
		TargetType: "malware",
		Mode:       Association,
		Enabled:    true,
	},

	// ==================== Certificate and Fingerprint Associations ====================
	{
		Name:       "certificate-to-domain",
		SourceType: "certificate-fingerprint",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "domain-to-certificate",
		SourceType: "domain",
		TargetType: "certificate-fingerprint",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "certificate-to-ip",
		SourceType: "certificate-fingerprint",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ja3-to-ip",
		SourceType: "ja3-fingerprint",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "jarm-to-ip",
		SourceType: "jarm-fingerprint",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "jarm-to-domain",
		SourceType: "jarm-fingerprint",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ssh-fingerprint-to-ip",
		SourceType: "ssh-fingerprint",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ssh-banner-to-ip",
		SourceType: "ssh-banner",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},

	// ==================== System and Process Associations ====================
	{
		Name:       "process-to-ip",
		SourceType: "process",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ip-to-process",
		SourceType: "ip",
		TargetType: "process",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "process-to-file",
		SourceType: "process",
		TargetType: "file",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "file-to-process",
		SourceType: "file",
		TargetType: "process",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "process-to-command",
		SourceType: "process",
		TargetType: "command",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "process-to-user",
		SourceType: "process",
		TargetType: "username",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "user-to-process",
		SourceType: "username",
		TargetType: "process",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "command-to-user",
		SourceType: "command",
		TargetType: "username",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "user-to-command",
		SourceType: "username",
		TargetType: "command",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "windows-task-to-user",
		SourceType: "windows-scheduled-task",
		TargetType: "username",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "windows-service-to-file",
		SourceType: "windows-service-name",
		TargetType: "file",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "hostname-to-user",
		SourceType: "hostname",
		TargetType: "username",
		Mode:       Association,
		Enabled:    true,
	},

	// ==================== Vulnerability Associations ====================
	{
		Name:       "cve-to-cpe",
		SourceType: "cve",
		TargetType: "cpe",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "cpe-to-cve",
		SourceType: "cpe",
		TargetType: "cve",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "cve-to-file",
		SourceType: "cve",
		TargetType: "file",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "file-to-cve",
		SourceType: "file",
		TargetType: "cve",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "cve-to-ip",
		SourceType: "cve",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ip-to-cve",
		SourceType: "ip",
		TargetType: "cve",
		Mode:       Association,
		Enabled:    true,
	},

	// ==================== Web Associations ====================
	{
		Name:       "cookie-to-domain",
		SourceType: "cookie",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "domain-to-cookie",
		SourceType: "domain",
		TargetType: "cookie",
		Mode:       Association,
		Enabled:    true,
	},

	// ==================== Identity Associations ====================
	{
		Name:       "user-to-ip",
		SourceType: "username",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ip-to-user",
		SourceType: "ip",
		TargetType: "username",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "user-to-group",
		SourceType: "username",
		TargetType: "group",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "group-to-user",
		SourceType: "group",
		TargetType: "username",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "user-to-hostname",
		SourceType: "username",
		TargetType: "hostname",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "hostname-to-group",
		SourceType: "hostname",
		TargetType: "group",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "jabber-to-ip",
		SourceType: "jabber-id",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
}

func GetEnabledRules() []*AssociationRule {
	rules := make([]*AssociationRule, 0, len(DefaultRules))
	for _, rule := range DefaultRules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	return rules
}
