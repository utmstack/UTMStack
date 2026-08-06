package domain

// Tenant is a configured connector instance for a module (e.g. one GCP
// project, one Azure subscription) — "tenant" here is a legacy name for a
// connector instance label, not a UTMStack platform tenant. TenantId is
// UTMStack's own platform tenant this connector instance belongs to — empty
// for every on-prem/single-tenant install, only ever set for a SaaS tenant's
// own connector config. Name+TenantId together are the real identity: two
// different platform tenants may legitimately pick the same Name.
type Tenant struct {
	Name     string            `yaml:"name"`
	TenantId string            `yaml:"tenantId,omitempty"`
	Config   map[string]string `yaml:",inline"`
}
