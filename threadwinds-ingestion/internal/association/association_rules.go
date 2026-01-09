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
	// Network Infrastructure
	{
		Name:       "ip-to-port",
		SourceType: "ip",
		TargetType: "port",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "port-to-service",
		SourceType: "port",
		TargetType: "service",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "domain-to-ip",
		SourceType: "domain",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "ip-to-domain",
		SourceType: "ip",
		TargetType: "domain",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "subdomain-to-domain",
		SourceType: "domain",
		TargetType: "domain",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "url-to-domain",
		SourceType: "url",
		TargetType: "domain",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "url-to-ip",
		SourceType: "url",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},

	// Geographic and ASN
	{
		Name:       "ip-to-asn",
		SourceType: "ip",
		TargetType: "asn",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "asn-to-organization",
		SourceType: "asn",
		TargetType: "organization",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "domain-to-asn",
		SourceType: "domain",
		TargetType: "asn",
		Mode:       Association,
		Enabled:    true,
	},

	// Identity and Access
	{
		Name:       "user-to-ip",
		SourceType: "user",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "user-to-hostname",
		SourceType: "user",
		TargetType: "hostname",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "user-to-account",
		SourceType: "user",
		TargetType: "account",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "email-to-user",
		SourceType: "email",
		TargetType: "user",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "email-to-domain",
		SourceType: "email",
		TargetType: "domain",
		Mode:       Aggregation,
		Enabled:    true,
	},

	// Threat Intelligence
	{
		Name:       "malware-to-ip",
		SourceType: "malware",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "malware-to-domain",
		SourceType: "malware",
		TargetType: "domain",
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
		Name:       "hash-to-malware",
		SourceType: "hash",
		TargetType: "malware",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "ip-to-threat-actor",
		SourceType: "ip",
		TargetType: "threat-actor",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "domain-to-threat-actor",
		SourceType: "domain",
		TargetType: "threat-actor",
		Mode:       Association,
		Enabled:    true,
	},
	{
		Name:       "cve-to-exploit",
		SourceType: "cve",
		TargetType: "exploit",
		Mode:       Aggregation,
		Enabled:    true,
	},
	{
		Name:       "vulnerability-to-ip",
		SourceType: "vulnerability",
		TargetType: "ip",
		Mode:       Association,
		Enabled:    true,
	},

	// Legacy (backward compatibility)
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
	{
		Name:       "hostname-to-user",
		SourceType: "hostname",
		TargetType: "user",
		Mode:       Association,
		Enabled:    true,
	},
}

func GetEnabledRules() []*AssociationRule {
	rules := make([]*AssociationRule, 0)
	for _, rule := range DefaultRules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	return rules
}
