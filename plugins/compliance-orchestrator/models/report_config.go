package models

type ReportConfig struct {
	ID                int64       `json:"id"`
	ConfigSolution    string      `json:"configSolution"`
	ConfigRemediation *string     `json:"configRemediation"`
	StandardSectionID int64       `json:"standardSectionId"`
	DashboardID       int64       `json:"dashboardId"`
	ConfigType        string      `json:"configType"`
	ConfigReportName  string      `json:"configReportName"`
	Section           *Section    `json:"section"`
	Queries           []QuerySpec `json:"queries"`
}

type Section struct {
	ID                         int64     `json:"id"`
	StandardID                 int64     `json:"standardId"`
	StandardSectionName        string    `json:"standardSectionName"`
	StandardSectionDescription string    `json:"standardSectionDescription"`
	Standard                   *Standard `json:"standard"`
}

type Standard struct {
	ID                  int64  `json:"id"`
	StandardName        string `json:"standardName"`
	StandardDescription string `json:"standardDescription"`
	SystemOwner         bool   `json:"systemOwner"`
}

type QuerySpec struct {
	ID              int64  `json:"id"`
	Description     string `json:"queryDescription"`
	SQLQuery        string `json:"sqlQuery"`
	EvaluationRule  string `json:"evaluationRule"`
	IndexPatternID  int64  `json:"indexPatternId"`
	ControlConfigID int64  `json:"controlConfigId"`
}
