package models

type ReportConfig struct {
	ID                  int64      `json:"id"`
	ConfigSolution      string     `json:"configSolution"`
	ConfigRemediation   *string    `json:"configRemediation"`
	StandardSectionID   int64      `json:"standardSectionId"`
	DashboardID         int64      `json:"dashboardId"`
	ConfigType          string     `json:"configType"`
	ConfigReportName    string     `json:"configReportName"`
	Section             *Section   `json:"section"`
	AssociatedDashboard *Dashboard `json:"associatedDashboard"`
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

type Dashboard struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	RefreshTime   *string `json:"refreshTime"`
	CreatedDate   string  `json:"createdDate"`
	ModifiedDate  string  `json:"modifiedDate"`
	UserCreated   string  `json:"userCreated"`
	UserModified  string  `json:"userModified"`
	Filters       string  `json:"filters"`
	DashboardType *string `json:"dashboardType"`
	SystemOwner   bool    `json:"systemOwner"`
}
