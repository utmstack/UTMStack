package dto

type SqlSearchDto struct {
	Query     string `json:"query"`
	FetchSize *int   `json:"fetchSize,omitempty"`
}
