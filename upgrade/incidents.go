package upgrade

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

var (
	incidentNamespace        = uuid.MustParse("ffa82061-9bd3-56c0-8f37-e34e33e2f9d0")
	incidentAlertNamespace   = uuid.MustParse("b685a259-ebac-589d-9425-cfc09cf2e9b9")
	incidentHistoryNamespace = uuid.MustParse("c7c9699c-a814-5e85-ba9d-ad11ddac670d")
	incidentNoteNamespace    = uuid.MustParse("351bf618-4905-5df6-bf66-2e7280a205d0")
)

type v11Incident struct {
	ID          int64
	Name        string    `gorm:"column:incident_name"`
	Description *string   `gorm:"column:incident_description"`
	Status      string    `gorm:"column:incident_status"`
	Severity    int64     `gorm:"column:incident_severity"`
	AssignedTo  *string   `gorm:"column:incident_assigned_to"`
	Solution    *string   `gorm:"column:incident_solution"`
	CreatedDate time.Time `gorm:"column:incident_created_date"`
}

type v11IncidentAlert struct {
	ID            int64
	IncidentID    int64
	AlertID       string
	AlertName     string
	AlertStatus   int64
	AlertSeverity int64
}

type v11IncidentHistory struct {
	ID         int64
	IncidentID int64
	ActionType string    `gorm:"column:action_type"`
	ActionDate time.Time `gorm:"column:action_date"`
	CreatedBy  string    `gorm:"column:action_created_by"`
}

type v11IncidentNote struct {
	ID           int64
	IncidentID   int64
	NoteText     string    `gorm:"column:note_text"`
	NoteSendDate time.Time `gorm:"column:note_send_date"`
	NoteSendBy   *string
}

func convertIncidents(db *gorm.DB) error {
	if !tableExists(db, "utm_incident") {
		logSkipTable("utm_incident")
		return nil
	}

	var src []v11Incident
	if err := db.Raw(`SELECT id, incident_name, incident_description, incident_status,
		incident_severity, incident_assigned_to, incident_solution, incident_created_date
		FROM utm_incident ORDER BY id`).Scan(&src).Error; err != nil {
		return err
	}

	count := 0
	for _, i := range src {
		desc := strFromPtr(i.Description)
		var descPtr *string
		if desc != "" {
			descPtr = &desc
		}
		var solPtr *string
		if s := strFromPtr(i.Solution); s != "" {
			solPtr = &s
		}
		assigned := strFrom(i.AssignedTo)

		if err := db.Exec(
			`INSERT INTO incident (id, tenant_id, incident_name, incident_description, incident_status, incident_severity, incident_assigned_to, incident_solution, incident_created_date)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO UPDATE SET
				tenant_id = EXCLUDED.tenant_id,
				incident_name = EXCLUDED.incident_name,
				incident_description = EXCLUDED.incident_description,
				incident_status = EXCLUDED.incident_status,
				incident_severity = EXCLUDED.incident_severity,
				incident_assigned_to = EXCLUDED.incident_assigned_to,
				incident_solution = EXCLUDED.incident_solution`,
			incidentUUID(i.ID), defaultTenantID, i.Name, descPtr,
			socStatus(i.Status), socSeverity(i.Severity), assigned, solPtr, i.CreatedDate,
		).Error; err != nil {
			return err
		}
		count++
	}

	alerts, err := convertIncidentAlerts(db)
	if err != nil {
		return err
	}
	history, err := convertIncidentHistory(db)
	if err != nil {
		return err
	}
	notes, err := convertIncidentNotes(db)
	if err != nil {
		return err
	}

	catcher.Info("v11 upgrade: incidents converted", map[string]any{
		"incidents": count, "alerts": alerts, "history": history, "notes": notes,
	})
	return nil
}

func incidentUUID(v11ID int64) string {
	return uuid.NewSHA1(incidentNamespace, []byte(strconv.FormatInt(v11ID, 10))).String()
}

func convertIncidentAlerts(db *gorm.DB) (int, error) {
	if !tableExists(db, "utm_incident_alert") {
		logSkipTable("utm_incident_alert")
		return 0, nil
	}

	var src []v11IncidentAlert
	if err := db.Raw(`SELECT id, incident_id, alert_id, alert_name, alert_status, alert_severity
		FROM utm_incident_alert ORDER BY id`).Scan(&src).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, a := range src {
		if err := db.Exec(
			`INSERT INTO incident_alert (id, tenant_id, incident_id, alert_id, alert_name, alert_severity, alert_status)
			 VALUES (?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO UPDATE SET
				tenant_id = EXCLUDED.tenant_id,
				incident_id = EXCLUDED.incident_id,
				alert_name = EXCLUDED.alert_name,
				alert_severity = EXCLUDED.alert_severity,
				alert_status = EXCLUDED.alert_status`,
			uuid.NewSHA1(incidentAlertNamespace, []byte(strconv.FormatInt(a.ID, 10))).String(), defaultTenantID,
			incidentUUID(a.IncidentID), a.AlertID, a.AlertName,
			socSeverity(a.AlertSeverity), socAlertStatus(a.AlertStatus),
		).Error; err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

func convertIncidentHistory(db *gorm.DB) (int, error) {
	if !tableExists(db, "utm_incident_history") {
		logSkipTable("utm_incident_history")
		return 0, nil
	}

	var src []v11IncidentHistory
	if err := db.Raw(`SELECT id, incident_id, action_type, action_date, action_created_by
		FROM utm_incident_history ORDER BY id`).Scan(&src).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, h := range src {
		action := socHistoryAction(h.ActionType)
		if action == "" {
			continue
		}
		if err := db.Exec(
			`INSERT INTO incident_history (id, tenant_id, incident_id, action, action_created_date, action_created_by)
			 VALUES (?, ?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO UPDATE SET
				tenant_id = EXCLUDED.tenant_id,
				incident_id = EXCLUDED.incident_id,
				action = EXCLUDED.action,
				action_created_date = EXCLUDED.action_created_date,
				action_created_by = EXCLUDED.action_created_by`,
			uuid.NewSHA1(incidentHistoryNamespace, []byte(strconv.FormatInt(h.ID, 10))).String(), defaultTenantID,
			incidentUUID(h.IncidentID), action, h.ActionDate, h.CreatedBy,
		).Error; err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

func convertIncidentNotes(db *gorm.DB) (int, error) {
	if !tableExists(db, "utm_incident_note") {
		logSkipTable("utm_incident_note")
		return 0, nil
	}

	var src []v11IncidentNote
	if err := db.Raw(`SELECT id, incident_id, note_text, note_send_date, note_send_by
		FROM utm_incident_note ORDER BY id`).Scan(&src).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, n := range src {
		if err := db.Exec(
			`INSERT INTO incident_note (id, tenant_id, incident_id, note_text, note_send_date, note_send_by)
			 VALUES (?, ?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO UPDATE SET
				tenant_id = EXCLUDED.tenant_id,
				incident_id = EXCLUDED.incident_id,
				note_text = EXCLUDED.note_text,
				note_send_date = EXCLUDED.note_send_date,
				note_send_by = EXCLUDED.note_send_by`,
			uuid.NewSHA1(incidentNoteNamespace, []byte(strconv.FormatInt(n.ID, 10))).String(), defaultTenantID,
			incidentUUID(n.IncidentID), n.NoteText, n.NoteSendDate, strFrom(n.NoteSendBy),
		).Error; err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

func socStatus(v11 string) string {
	switch strings.ToUpper(strings.TrimSpace(v11)) {
	case "OPEN":
		return "Open"
	case "IN_REVIEW":
		return "In review"
	case "COMPLETED":
		return "Completed"
	case "MERGED":
		return "Merged"
	default:
		return "Open"
	}
}

func socSeverity(n int64) string {
	switch n {
	case 1:
		return "low"
	case 3:
		return "high"
	default:
		return "medium"
	}
}

func socAlertStatus(n int64) string {
	switch n {
	case 1:
		return "Automatic review"
	case 2:
		return "Open"
	case 3:
		return "In review"
	case 5:
		return "Completed"
	default:
		return ""
	}
}

func socHistoryAction(v11 string) string {
	switch strings.TrimSpace(v11) {
	case "INCIDENT_MERGED", "INCIDENT_COMPLETED":
		return "INCIDENT_STATUS_CHANGE"
	case "INCIDENT_ASSIGNED_TO", "INCIDENT_ASSIGNED_CHANGE":
		return "INCIDENT_ASSIGNED"
	case "INCIDENT_ALERT_DELETED", "INCIDENT_ALERT_MOVED", "INCIDENT_COMMAND_EXECUTED":
		return ""
	default:
		return strings.TrimSpace(v11)
	}
}

func strFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
