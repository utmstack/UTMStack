package dto

type IndexListResponse struct {
	Indices []IndexInfo `json:"indices"`
	Total   int         `json:"total"`
}

type IndexPropertyResponse struct {
	Properties []IndexPropertyType `json:"properties"`
}
