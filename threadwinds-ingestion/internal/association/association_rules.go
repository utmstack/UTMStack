package association

type AssociationMode string

const (
	Association AssociationMode = "association"
	Aggregation AssociationMode = "aggregation"
)

type RuleCategory string

const (
	CategoryNetwork  RuleCategory = "network"
	CategoryIdentity RuleCategory = "identity"
)

type AssociationRule struct {
	Name        string
	SourceType  string
	TargetType  string
	Mode        AssociationMode
	Category    RuleCategory
	Description string
	Enabled     bool
	Priority    int
}

var DefaultRules = []*AssociationRule{
	// Network Associations
	{
		Name:        "ip-to-port",
		SourceType:  "ip",
		TargetType:  "port",
		Mode:        Association,
		Category:    CategoryNetwork,
		Description: "IP exposes port",
		Enabled:     true,
		Priority:    10,
	},
	{
		Name:        "port-to-ip",
		SourceType:  "port",
		TargetType:  "ip",
		Mode:        Association,
		Category:    CategoryNetwork,
		Description: "Port exposed on IP",
		Enabled:     true,
		Priority:    10,
	},
	{
		Name:        "hostname-to-ip",
		SourceType:  "hostname",
		TargetType:  "ip",
		Mode:        Association,
		Category:    CategoryNetwork,
		Description: "Hostname resolves to IP",
		Enabled:     true,
		Priority:    10,
	},
	{
		Name:        "ip-to-hostname",
		SourceType:  "ip",
		TargetType:  "hostname",
		Mode:        Association,
		Category:    CategoryNetwork,
		Description: "IP resolves to hostname",
		Enabled:     true,
		Priority:    10,
	},

	// Identity Associations
	{
		Name:        "username-to-ip",
		SourceType:  "username",
		TargetType:  "ip",
		Mode:        Association,
		Category:    CategoryIdentity,
		Description: "User accessed from IP",
		Enabled:     true,
		Priority:    10,
	},
	{
		Name:        "ip-to-username",
		SourceType:  "ip",
		TargetType:  "username",
		Mode:        Association,
		Category:    CategoryIdentity,
		Description: "IP accessed by user",
		Enabled:     true,
		Priority:    10,
	},
	{
		Name:        "username-to-hostname",
		SourceType:  "username",
		TargetType:  "hostname",
		Mode:        Association,
		Category:    CategoryIdentity,
		Description: "User accessed from hostname",
		Enabled:     true,
		Priority:    9,
	},
	{
		Name:        "hostname-to-username",
		SourceType:  "hostname",
		TargetType:  "username",
		Mode:        Association,
		Category:    CategoryIdentity,
		Description: "Hostname accessed by user",
		Enabled:     true,
		Priority:    9,
	},

	// ASN Associations
	{
		Name:        "ip-to-asn",
		SourceType:  "ip",
		TargetType:  "asn",
		Mode:        Association,
		Category:    CategoryNetwork,
		Description: "IP belongs to ASN",
		Enabled:     true,
		Priority:    10,
	},
	{
		Name:        "asn-to-ip",
		SourceType:  "asn",
		TargetType:  "ip",
		Mode:        Association,
		Category:    CategoryNetwork,
		Description: "ASN contains IP",
		Enabled:     true,
		Priority:    10,
	},
	{
		Name:        "hostname-to-asn",
		SourceType:  "hostname",
		TargetType:  "asn",
		Mode:        Association,
		Category:    CategoryNetwork,
		Description: "Hostname resolves to IP in ASN",
		Enabled:     true,
		Priority:    9,
	},
	{
		Name:        "asn-to-hostname",
		SourceType:  "asn",
		TargetType:  "hostname",
		Mode:        Association,
		Category:    CategoryNetwork,
		Description: "ASN contains hostname",
		Enabled:     true,
		Priority:    9,
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
