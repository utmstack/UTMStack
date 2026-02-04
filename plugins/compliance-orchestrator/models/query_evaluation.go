package models

type QueryEvaluation struct {
	QueryConfigID int64                 `json:"queryConfigId"`
	QueryName     string                `json:"queryName"`
	Hits          int64                 `json:"hits"`
	Status        QueryEvaluationStatus `json:"status"`
	ErrorMessage  *string               `json:"errorMessage,omitempty"`
	Evidence      [][]any               `json:"evidence,omitempty"`
}
