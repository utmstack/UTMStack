package models

import "time"

type ControlEvaluation struct {
	ControlConfigID  int64                   `json:"controlConfigId"`
	ControlName      string                  `json:"controlName"`
	Status           ControlEvaluationStatus `json:"status"`
	QueryEvaluations []QueryEvaluation       `json:"queryevaluations"`
	EvaluatedAt      time.Time               `json:"evaluatedAt"`
}
