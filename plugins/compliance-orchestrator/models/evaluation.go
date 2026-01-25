package models

type Evaluation struct {
	ReportID int           `json:"reportId"`
	Status   string        `json:"status"`
	Results  []QueryResult `json:"results"`
}
