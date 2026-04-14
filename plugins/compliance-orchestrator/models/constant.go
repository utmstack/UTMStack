package models

type QueryEvaluationStatus string

const (
	QueryStatusCompliant     QueryEvaluationStatus = "COMPLIANT"
	QueryStatusNonCompliant  QueryEvaluationStatus = "NON_COMPLIANT"
	QueryStatusNotApplicable QueryEvaluationStatus = "NOT_APPLICABLE"
	QueryStatusError         QueryEvaluationStatus = "ERROR"
)

type ControlEvaluationStatus string

const (
	ControlStatusCompliant     ControlEvaluationStatus = "COMPLIANT"
	ControlStatusNonCompliant  ControlEvaluationStatus = "NON_COMPLIANT"
	ControlStatusNotApplicable ControlEvaluationStatus = "NOT_APPLICABLE"
	ControlStatusNotEvaluated  ControlEvaluationStatus = "NOT_EVALUATED"
)

type EvaluationRule string

const (
	NoHitsAllowed   EvaluationRule = "NO_HITS_ALLOWED"
	MinHitsRequired EvaluationRule = "MIN_HITS_REQUIRED"
	ThresholdMax    EvaluationRule = "THRESHOLD_MAX"
)

type ComplianceStrategy string

const (
	StrategyAll ComplianceStrategy = "ALL"
	StrategyAny ComplianceStrategy = "ANY"
)
