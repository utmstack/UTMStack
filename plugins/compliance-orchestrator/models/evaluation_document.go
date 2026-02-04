package models

import "time"

type EvaluationDocument struct {
	ControlID        int64                   `json:"control_id"`
	ControlName      string                  `json:"control_name"`
	Status           ControlEvaluationStatus `json:"status"`
	Timestamp        time.Time               `json:"timestamp"`
	QueryEvaluations []QueryEvaluation       `json:"query_evaluations"`
}
