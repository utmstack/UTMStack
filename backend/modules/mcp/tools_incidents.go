package mcp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func registerIncidents(m *Module) {
	registerIncidentMain(m)
	registerIncidentAlerts(m)
	registerIncidentNotes(m)
	registerIncidentHistory(m)
}

// ---- incidents.* -----------------------------------------------------------

type incidentsCreateInput struct {
	Name        string              `json:"name" jsonschema:"Incident display name"`
	Description string              `json:"description,omitempty"`
	AssignedTo  string              `json:"assigned_to,omitempty"`
	Observation string              `json:"observation,omitempty"`
	AlertList   []dto.AlertLinkItem `json:"alert_list" jsonschema:"At least one alert link"`
}

type incidentsAddAlertsInput struct {
	IncidentID uuid.UUID           `json:"incident_id"`
	AlertList  []dto.AlertLinkItem `json:"alert_list"`
}

type incidentsChangeStatusInput struct {
	IncidentID  uuid.UUID `json:"incident_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status" jsonschema:"one of: Open, In review, Completed, Merged"`
	Solution    string    `json:"solution,omitempty"`
}

type incidentsListInput struct {
	Name       string `json:"name,omitempty"`
	Status     string `json:"status,omitempty"`
	AssignedTo string `json:"assigned_to,omitempty"`
	From       string `json:"from,omitempty" jsonschema:"RFC3339 lower bound on createdDate"`
	To         string `json:"to,omitempty"`
	Page       int    `json:"page,omitempty"`
	Size       int    `json:"size,omitempty"`
	Sort       string `json:"sort,omitempty"`
}

type incidentIDInput struct {
	ID uuid.UUID `json:"id"`
}

func registerIncidentMain(m *Module) {
	uc := m.deps.Incidents.GetIncidentUsecase()

	Add(m, &mcp.Tool{
		Name:        "incidents.create",
		Title:       "Create incident",
		Description: "Create a new incident with the given alert links.",
	}, Gate{Permission: "incidents.write"},
		func(ctx context.Context, actor *authz.Actor, in incidentsCreateInput) (any, error) {
			if in.Name == "" || len(in.AlertList) == 0 {
				return nil, fmt.Errorf("name and alert_list are required")
			}
			req := dto.CreateIncidentRequest{
				IncidentName: in.Name,
				AlertList:    in.AlertList,
			}
			if in.Description != "" {
				req.IncidentDescription = &in.Description
			}
			req.IncidentAssignedTo = in.AssignedTo
			if in.Observation != "" {
				req.IncidentObservation = &in.Observation
			}
			return uc.Create(ctx, actor.Email, req)
		})

	Add(m, &mcp.Tool{
		Name:        "incidents.add_alerts",
		Title:       "Attach alerts to an existing incident",
		Description: "Append additional alerts to an existing incident.",
	}, Gate{Permission: "incidents.write"},
		func(ctx context.Context, actor *authz.Actor, in incidentsAddAlertsInput) (any, error) {
			return uc.AddAlerts(ctx, actor.Email, dto.AddAlertsRequest{
				IncidentID: in.IncidentID,
				AlertList:  in.AlertList,
			})
		})

	Add(m, &mcp.Tool{
		Name:        "incidents.change_status",
		Title:       "Change incident status / metadata",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "incidents.write"},
		func(ctx context.Context, actor *authz.Actor, in incidentsChangeStatusInput) (any, error) {
			req := dto.ChangeStatusRequest{
				IncidentID:     in.IncidentID,
				IncidentName:   in.Name,
				IncidentStatus: in.Status,
			}
			if in.Description != "" {
				req.IncidentDescription = &in.Description
			}
			if in.Solution != "" {
				req.IncidentSolution = &in.Solution
			}
			return uc.ChangeStatus(ctx, actor.Email, req)
		})

	Add(m, &mcp.Tool{
		Name:        "incidents.list",
		Title:       "List incidents",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "incidents.read"},
		func(ctx context.Context, _ *authz.Actor, in incidentsListInput) (any, error) {
			q := dto.IncidentListQuery{
				Page: in.Page, Size: clampPageSize(in.Size), Sort: in.Sort,
			}
			if in.Name != "" {
				q.IncidentName = &in.Name
			}
			if in.Status != "" {
				q.IncidentStatus = &in.Status
			}
			if in.AssignedTo != "" {
				q.IncidentAssignedTo = &in.AssignedTo
			}
			items, total, err := uc.List(ctx, q)
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name:        "incidents.get",
		Title:       "Get incident",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "incidents.read"},
		func(ctx context.Context, _ *authz.Actor, in incidentIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name:        "incidents.assignees",
		Title:       "List the assignees incidents are currently handed to",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "incidents.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.GetAssignees(ctx)
		})
}

// ---- incident_alerts.* -----------------------------------------------------

type incidentAlertsCreateInput struct {
	IncidentID    uuid.UUID `json:"incident_id"`
	AlertID       string    `json:"alert_id"`
	AlertName     string    `json:"alert_name"`
	AlertSeverity string    `json:"alert_severity" jsonschema:"one of: low, medium, high"`
	AlertStatus   *string   `json:"alert_status,omitempty"`
}

type incidentAlertsUpdateStatusInput struct {
	IncidentID  uuid.UUID `json:"incident_id"`
	AlertIDs    []string  `json:"alert_ids"`
	AlertStatus string    `json:"status"`
}

type incidentAlertsUpdateInput struct {
	ID            uuid.UUID `json:"id"`
	IncidentID    uuid.UUID `json:"incident_id"`
	AlertID       string    `json:"alert_id"`
	AlertName     string    `json:"alert_name,omitempty"`
	AlertSeverity string    `json:"alert_severity,omitempty" jsonschema:"one of: low, medium, high"`
	AlertStatus   *string   `json:"alert_status,omitempty"`
}

type incidentAlertsListInput struct {
	IncidentID  *uuid.UUID `json:"incident_id,omitempty"`
	AlertID     *string    `json:"alert_id,omitempty"`
	AlertStatus *string    `json:"alert_status,omitempty"`
	Page        int        `json:"page,omitempty"`
	Size        int        `json:"size,omitempty"`
	Sort        string     `json:"sort,omitempty"`
}

func registerIncidentAlerts(m *Module) {
	uc := m.deps.Incidents.GetIncidentAlertUsecase()

	Add(m, &mcp.Tool{
		Name: "incident_alerts.create", Title: "Add alert row to incident",
	}, Gate{Permission: "incidents.write"},
		func(ctx context.Context, _ *authz.Actor, in incidentAlertsCreateInput) (any, error) {
			return uc.Create(ctx, dto.IncidentAlertRequest{
				IncidentID: in.IncidentID, AlertID: in.AlertID,
				AlertName: in.AlertName, AlertSeverity: in.AlertSeverity, AlertStatus: in.AlertStatus,
			})
		})

	Add(m, &mcp.Tool{
		Name: "incident_alerts.update_status", Title: "Update incident-alert status",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "incidents.write"},
		func(ctx context.Context, actor *authz.Actor, in incidentAlertsUpdateStatusInput) (any, error) {
			err := uc.UpdateStatus(ctx, actor.Email, dto.UpdateAlertStatusRequest{
				IncidentID: in.IncidentID, AlertIds: in.AlertIDs, AlertStatus: in.AlertStatus,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"updated": len(in.AlertIDs)}, nil
		})

	Add(m, &mcp.Tool{
		Name: "incident_alerts.update", Title: "Update incident-alert row",
	}, Gate{Permission: "incidents.write"},
		func(ctx context.Context, _ *authz.Actor, in incidentAlertsUpdateInput) (any, error) {
			return uc.Update(ctx, dto.UpdateIncidentAlertRequest{
				ID: in.ID, IncidentID: in.IncidentID, AlertID: in.AlertID,
				AlertName: in.AlertName, AlertSeverity: in.AlertSeverity, AlertStatus: in.AlertStatus,
			})
		})

	Add(m, &mcp.Tool{
		Name: "incident_alerts.list", Title: "List incident-alert links",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "incidents.read"},
		func(ctx context.Context, _ *authz.Actor, in incidentAlertsListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.IncidentAlertListQuery{
				IncidentID: in.IncidentID, AlertID: in.AlertID, AlertStatus: in.AlertStatus,
				Page: in.Page, Size: clampPageSize(in.Size), Sort: in.Sort,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "incident_alerts.delete", Title: "Remove alert from incident",
		Annotations: &mcp.ToolAnnotations{},
	}, Gate{Permission: "incidents.write"},
		func(ctx context.Context, _ *authz.Actor, in incidentIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

// ---- incident_notes.* ------------------------------------------------------

type incidentNotesCreateInput struct {
	IncidentID uuid.UUID `json:"incident_id"`
	NoteText   string    `json:"note_text"`
}

type incidentNotesListInput struct {
	IncidentID *uuid.UUID `json:"incident_id,omitempty"`
	Page       int        `json:"page,omitempty"`
	Size       int        `json:"size,omitempty"`
	Sort       string     `json:"sort,omitempty"`
}

func registerIncidentNotes(m *Module) {
	uc := m.deps.Incidents.GetIncidentNoteUsecase()

	Add(m, &mcp.Tool{
		Name: "incident_notes.create", Title: "Add note to incident",
	}, Gate{Permission: "incidents.write"},
		func(ctx context.Context, actor *authz.Actor, in incidentNotesCreateInput) (any, error) {
			return uc.Create(ctx, actor.Email, dto.CreateNoteRequest{
				IncidentID: in.IncidentID, NoteText: in.NoteText,
			})
		})

	Add(m, &mcp.Tool{
		Name: "incident_notes.list", Title: "List incident notes",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "incidents.read"},
		func(ctx context.Context, _ *authz.Actor, in incidentNotesListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.IncidentNoteListQuery{
				IncidentID: in.IncidentID, Page: in.Page, Size: clampPageSize(in.Size), Sort: in.Sort,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})
}

// ---- incident_history.* ----------------------------------------------------

type incidentHistoryListInput struct {
	IncidentID *uuid.UUID `json:"incident_id,omitempty"`
	Action     *string    `json:"action,omitempty"`
	Page       int        `json:"page,omitempty"`
	Size       int        `json:"size,omitempty"`
	Sort       string     `json:"sort,omitempty"`
}

func registerIncidentHistory(m *Module) {
	uc := m.deps.Incidents.GetIncidentHistoryUsecase()

	Add(m, &mcp.Tool{
		Name: "incident_history.list", Title: "List incident history",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "incidents.read"},
		func(ctx context.Context, _ *authz.Actor, in incidentHistoryListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.HistoryListQuery{
				IncidentID: in.IncidentID, Action: in.Action,
				Page: in.Page, Size: clampPageSize(in.Size), Sort: in.Sort,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "incident_history.count", Title: "Count history rows",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "incidents.read"},
		func(ctx context.Context, _ *authz.Actor, in incidentHistoryListInput) (any, error) {
			c, err := uc.Count(ctx, dto.HistoryListQuery{
				IncidentID: in.IncidentID, Action: in.Action,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"count": c}, nil
		})

	Add(m, &mcp.Tool{
		Name: "incident_history.get", Title: "Get incident history row",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "incidents.read"},
		func(ctx context.Context, _ *authz.Actor, in incidentIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})
}
