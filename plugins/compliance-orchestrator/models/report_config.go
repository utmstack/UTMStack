package models

type ControlConfig struct {
	ID                 int64              `json:"id"`
	StandardSectionID  int64              `json:"standardSectionId"`
	Section            Section            `json:"section"`
	ControlName        string             `json:"controlName"`
	ControlSolution    string             `json:"controlSolution"`
	ControlRemediation string             `json:"controlRemediation"`
	ControlStrategy    ComplianceStrategy `json:"controlStrategy"`
	QueriesConfigs     []QueryConfig      `json:"queriesConfigs"`
}

type Section struct {
	ID                         int64  `json:"id"`
	StandardSectionName        string `json:"standardSectionName"`
	StandardSectionDescription string `json:"standardSectionDescription"`
	StandardID                 int64  `json:"standardId"`
}

type QueryConfig struct {
	ID               int64          `json:"id"`
	QueryName        string         `json:"queryName"`
	QueryDescription string         `json:"queryDescription"`
	SQLQuery         string         `json:"sqlQuery"`
	EvaluationRule   EvaluationRule `json:"evaluationRule"`
	RuleValue        *int           `json:"ruleValue"`
	IndexPatternID   int64          `json:"indexPatternId"`
	ControlConfigID  int64          `json:"controlConfigId"`
}
