package domain

type Pipeline struct {
	RelPath  string
	Content  []byte
	System   bool
	Active   bool
	TenantID string
}

type PipelineFile struct {
	Pipeline []map[string]any `yaml:"pipeline"`
}

type PipelineEntry struct {
	DataTypes []string         `yaml:"dataTypes"`
	Order     int32            `yaml:"order"`
	Steps     []map[string]any `yaml:"steps"`
	TenantID  string           `yaml:"tenantId,omitempty"`
}

type PipelineSpec struct {
	Pipeline []PipelineEntry `yaml:"pipeline"`
}
