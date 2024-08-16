package models

type Version struct {
	ServiceVersion      string `json:"service_version"`
	DependenciesVersion string `json:"dependencies_version"`
}

type DependencyUpdateResponse struct {
	Message     string `json:"message"`
	Version     string `json:"version"`
	TotalSize   int64  `json:"totalSize"`
	PartIndex   string `json:"partIndex,omitempty"`
	NParts      int    `json:"nParts,omitempty"`
	FileContent []byte `json:"fileContent,omitempty"`
}
