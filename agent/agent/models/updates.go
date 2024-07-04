package models

type Version struct {
	ServiceVersion      string `json:"service_version"`
	DependenciesVersion string `json:"dependencies_version"`
}

type DependencyUpdateResponse struct {
	Message     string `json:"message"`
	Version     string `json:"version"`
	FileContent []byte `json:"fileContent,omitempty"`
}
