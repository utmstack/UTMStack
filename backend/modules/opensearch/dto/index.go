package dto

import ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"

type IndexListResponse struct {
	Indices []ospkg.IndexInfo `json:"indices"`
	Total   int               `json:"total"`
}

type IndexPropertyResponse struct {
	Properties []ospkg.IndexPropertyType `json:"properties"`
}
