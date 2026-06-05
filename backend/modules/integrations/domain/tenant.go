package domain

// Tenant is one configured instance of an integration as stored in its YAML
// file. This is the file-persistence model only — the API uses dto.TenantRequest
// / dto.TenantResponse. Name is the identity; Config is the field→value bag
// (sensitive values encrypted on disk).
type Tenant struct {
	Name   string            `yaml:"name"`
	Config map[string]string `yaml:",inline"`
}
