package client

type SQLResult struct {
	Rows  [][]any
	Count int
}

type SQLResponse struct {
	Schema   []any   `json:"schema"`
	DataRows [][]any `json:"datarows"`
	Total    int     `json:"total"`
}
