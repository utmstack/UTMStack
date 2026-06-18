package domain

type SystemIndexPattern int

const (
	SysPatternLogs        SystemIndexPattern = 0
	SysPatternAlerts      SystemIndexPattern = 1
	SysPatternLogsWindows SystemIndexPattern = 2
)

type IndexPatternRegistry struct {
	Logs        string // db id=1
	Alerts      string // db id=2
	LogsWindows string // db id=8
}

func (r *IndexPatternRegistry) Get(p SystemIndexPattern) string {
	switch p {
	case SysPatternLogs:
		return r.Logs
	case SysPatternAlerts:
		return r.Alerts
	case SysPatternLogsWindows:
		return r.LogsWindows
	default:
		return ""
	}
}
