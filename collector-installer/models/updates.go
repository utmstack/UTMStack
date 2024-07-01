package models

type Version struct {
	ServiceVersion      string `json:"service_version"`
	DependenciesVersion string `json:"dependencies_version"`
}
