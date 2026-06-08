package domain

type ComplianceType = string
type ComplianceStatus = string
type ComplianceStrategy = string
type EvaluationRule = string

const (
	ComplianceTypeCustom   ComplianceType = "CUSTOM"
	ComplianceTypeTemplate ComplianceType = "TEMPLATE"

	ComplianceStatusCompliant    ComplianceStatus = "COMPLIANT"
	ComplianceStatusNonCompliant ComplianceStatus = "NON_COMPLIANT"

	ComplianceStrategyAll ComplianceStrategy = "ALL"
	ComplianceStrategyAny ComplianceStrategy = "ANY"

	EvaluationRuleNoHitsAllowed   EvaluationRule = "NO_HITS_ALLOWED"
	EvaluationRuleMinHitsRequired EvaluationRule = "MIN_HITS_REQUIRED"
	EvaluationRuleThresholdMax    EvaluationRule = "THRESHOLD_MAX"
	EvaluationRuleMatchFieldValue EvaluationRule = "MATCH_FIELD_VALUE"
)
