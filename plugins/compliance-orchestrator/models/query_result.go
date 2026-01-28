package models

type QueryResult struct {
	QueryID int    `json:"queryId"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}
