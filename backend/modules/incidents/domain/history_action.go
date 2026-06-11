package domain

type HistoryAction struct {
	Type  string // action_type column  e.g. "INCIDENT_CREATED"
	Label string // action column       e.g. "Incident created"
}

var (
	ActionCreated            = HistoryAction{"INCIDENT_CREATED", "Incident created"}
	ActionAlertAdd           = HistoryAction{"INCIDENT_ALERT_ADD", "New alerts added to incident"}
	ActionAlertStatusChanged = HistoryAction{"INCIDENT_ALERT_STATUS_CHANGED", "Alert change status"}
	ActionStatusChange       = HistoryAction{"INCIDENT_STATUS_CHANGE", "Alert change status"}
	ActionNoteAdd            = HistoryAction{"INCIDENT_NOTE_ADD", "Note added"}
	ActionNoteChange         = HistoryAction{"INCIDENT_NOTE_CHANGE", "Note changed"}
	ActionAssigned           = HistoryAction{"INCIDENT_ASSIGNED", "Incident assigned"}
)
