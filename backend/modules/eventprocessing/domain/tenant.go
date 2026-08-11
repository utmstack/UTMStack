package domain

type TenantAssets struct {
	ID     string
	Name   string
	Assets []Asset
}

type Asset struct {
	Name            string   `yaml:"name"`
	Hostnames       []string `yaml:"hostnames,omitempty"`
	Ips             []string `yaml:"ips,omitempty"`
	Confidentiality uint32   `yaml:"confidentiality"`
	Availability    uint32   `yaml:"availability"`
	Integrity       uint32   `yaml:"integrity"`
}

type TenantsFile struct {
	Tenants []TenantEntry `yaml:"tenants"`
}

type TenantEntry struct {
	ID                string   `yaml:"id"`
	Name              string   `yaml:"name"`
	Assets            []Asset  `yaml:"assets,omitempty"`
	DisabledRules     []string `yaml:"disabledRules,omitempty"`
	DisabledPipelines []string `yaml:"disabledPipelines,omitempty"`

	// PipelineOrder names the pipelines in the order this tenant wants them
	// applied. The key must stay exactly this: the engine parses the file with
	// DiscardUnknown, so a misspelling is ignored in silence and the tenant
	// quietly falls back to the shipped order.
	PipelineOrder []string `yaml:"pipelineOrder,omitempty"`
}
