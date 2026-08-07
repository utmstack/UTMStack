package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/threatwinds/go-sdk/catcher"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type incidentUsecase struct {
	incidentRepo  connectors.IncidentRepository
	alertRepo     connectors.IncidentAlertRepository
	historyRepo   connectors.IncidentHistoryRepository
	mailer        connectors.IncidentMailer
	alertsGateway connectors.AlertsGateway
	audit         audit_connectors.Logger
}

func NewIncidentUsecase(
	incidentRepo connectors.IncidentRepository,
	alertRepo connectors.IncidentAlertRepository,
	historyRepo connectors.IncidentHistoryRepository,
	mailer connectors.IncidentMailer,
	alertsGateway connectors.AlertsGateway,
	audit audit_connectors.Logger,
) connectors.IncidentUsecase {
	return &incidentUsecase{
		incidentRepo:  incidentRepo,
		alertRepo:     alertRepo,
		historyRepo:   historyRepo,
		mailer:        mailer,
		alertsGateway: alertsGateway,
		audit:         audit,
	}
}

func (u *incidentUsecase) Create(ctx context.Context, userEmail string, req dto.CreateIncidentRequest) (*domain.Incident, error) {
	u.audit.Log(ctx, audit_connectors.Event{
		Action:    "incident.create.attempt",
		EventType: audit_domain.INCIDENT_CREATION_ATTEMPT,
		Status:    audit_domain.StatusSuccess,
	})

	if len(req.AlertList) == 0 {
		return nil, fmt.Errorf("at least one alert is required")
	}
	if err := validateAlertList(req.AlertList); err != nil {
		return nil, err
	}

	alertIDs := make([]string, len(req.AlertList))
	for i, item := range req.AlertList {
		alertIDs[i] = item.AlertID
	}
	existing, err := u.alertRepo.FindByAlertIDs(ctx, alertIDs)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return nil, domain.ErrAlertAlreadyLinked
	}

	incident := &domain.Incident{
		Name:        req.IncidentName,
		Description: req.IncidentDescription,
		Status:      domain.StatusOpen,
		Severity:    maxRequestedSeverity(req.AlertList),
		AssignedTo:  strings.TrimSpace(req.IncidentAssignedTo),
		CreatedDate: time.Now().UTC(),
	}
	if err := u.incidentRepo.Create(ctx, incident, alertRows(incident, req.AlertList)); err != nil {
		u.audit.Log(ctx, audit_connectors.Event{
			Action:       "incident.create.fail",
			EventType:    audit_domain.INCIDENT_CREATION_ATTEMPT,
			Status:       audit_domain.StatusFailure,
			ErrorMessage: err.Error(),
		})
		return nil, err
	}

	currentUser := resolveUser(userEmail)
	if err := u.saveHistory(ctx, incident.ID, domain.ActionCreated, currentUser); err != nil {
		catcher.Warn("incidents: failed to write history", map[string]any{"error": err.Error()})
	}

	u.audit.Log(ctx, audit_connectors.Event{
		Action:     "incident.create.success",
		EventType:  audit_domain.INCIDENT_CREATION_SUCCESS,
		Status:     audit_domain.StatusSuccess,
		ResourceID: incident.ID.String(),
		Metadata: map[string]any{
			"alertCount": len(req.AlertList),
			"severity":   incident.Severity,
		},
	})

	if err := u.mailer.SendIncidentCreated(ctx, *incident); err != nil {
		catcher.Warn("incidents: mail notification failed", map[string]any{"error": err.Error()})
	}

	return incident, nil
}

func (u *incidentUsecase) AddAlerts(ctx context.Context, userEmail string, req dto.AddAlertsRequest) (*domain.Incident, error) {
	u.audit.Log(ctx, audit_connectors.Event{
		Action:    "incident.alert.add.attempt",
		EventType: audit_domain.INCIDENT_ALERT_ADD_ATTEMPT,
		Status:    audit_domain.StatusSuccess,
	})

	if err := validateAlertList(req.AlertList); err != nil {
		return nil, err
	}

	incident, err := u.incidentRepo.FindByID(ctx, req.IncidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrNotFound
	}

	addAlertIDs := make([]string, len(req.AlertList))
	for i, item := range req.AlertList {
		addAlertIDs[i] = item.AlertID
	}
	existingAlerts, err := u.alertRepo.FindByAlertIDs(ctx, addAlertIDs)
	if err != nil {
		return nil, err
	}
	if len(existingAlerts) > 0 {
		return nil, domain.ErrAlertAlreadyLinked
	}

	// An incident is as bad as its worst alert, so adding one can raise the
	// severity but never lower it. The new rows and the severity they imply are
	// stored together — a half-applied add leaves the incident understating what
	// it holds.
	if newMax := maxRequestedSeverity(req.AlertList); newMax.Rank() > incident.Severity.Rank() {
		incident.Severity = newMax
	}
	if err := u.incidentRepo.LinkAlerts(ctx, incident, alertRows(incident, req.AlertList)); err != nil {
		return nil, err
	}

	currentUser := resolveUser(userEmail)
	if err := u.saveHistory(ctx, incident.ID, domain.ActionAlertAdd, currentUser); err != nil {
		catcher.Warn("incidents: failed to write history", map[string]any{"error": err.Error()})
	}

	u.audit.Log(ctx, audit_connectors.Event{
		Action:     "incident.alert.add.success",
		EventType:  audit_domain.INCIDENT_ALERT_ADD_SUCCESS,
		Status:     audit_domain.StatusSuccess,
		ResourceID: incident.ID.String(),
		Metadata:   map[string]any{"alertCount": len(req.AlertList)},
	})

	return incident, nil
}

func (u *incidentUsecase) ChangeStatus(ctx context.Context, userEmail string, req dto.ChangeStatusRequest) (*domain.Incident, error) {
	u.audit.Log(ctx, audit_connectors.Event{
		Action:    "incident.status.change.attempt",
		EventType: audit_domain.INCIDENT_UPDATE_ATTEMPT,
		Status:    audit_domain.StatusSuccess,
	})

	newStatus := domain.IncidentStatus(req.IncidentStatus)
	if !newStatus.Valid() {
		return nil, fmt.Errorf("%w: %q", domain.ErrInvalidStatus, req.IncidentStatus)
	}

	incident, err := u.incidentRepo.FindByID(ctx, req.IncidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrNotFound
	}

	oldStatus := incident.Status
	incident.Status = newStatus
	if req.IncidentSolution != nil {
		incident.Solution = req.IncidentSolution
	}

	if err := u.incidentRepo.Update(ctx, incident); err != nil {
		return nil, err
	}

	linkedAlerts, err := u.alertRepo.FindByIncidentID(ctx, incident.ID)
	if err != nil {
		return nil, err
	}

	if len(linkedAlerts) > 0 {
		alertIDs := make([]string, len(linkedAlerts))
		for i, a := range linkedAlerts {
			alertIDs[i] = a.AlertID
		}
		if err := u.alertRepo.BulkUpdateStatus(ctx, alertIDs, string(newStatus)); err != nil {
			return nil, err
		}

		observation := ""
		if req.IncidentSolution != nil {
			observation = *req.IncidentSolution
		}
		if err := u.alertsGateway.UpdateAlertStatus(ctx, alertIDs, newStatus, observation); err != nil {
			catcher.Warn("incidents: failed to sync alert status", map[string]any{"error": err.Error()})
		}
	}

	currentUser := resolveUser(userEmail)
	if err := u.saveHistory(ctx, incident.ID, domain.ActionStatusChange, currentUser); err != nil {
		catcher.Warn("incidents: failed to write history", map[string]any{"error": err.Error()})
	}

	u.audit.Log(ctx, audit_connectors.Event{
		Action:     "incident.status.change.success",
		EventType:  audit_domain.INCIDENT_UPDATE_SUCCESS,
		Status:     audit_domain.StatusSuccess,
		ResourceID: incident.ID.String(),
		Metadata:   map[string]any{"from": oldStatus, "to": newStatus},
	})

	return incident, nil
}

func (u *incidentUsecase) Assign(ctx context.Context, userEmail string, req dto.AssignRequest) (*domain.Incident, error) {
	u.audit.Log(ctx, audit_connectors.Event{
		Action:    "incident.assign.attempt",
		EventType: audit_domain.INCIDENT_UPDATE_ATTEMPT,
		Status:    audit_domain.StatusSuccess,
	})

	incident, err := u.incidentRepo.FindByID(ctx, req.IncidentID)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrNotFound
	}

	// A blank AssignedTo unassigns the incident.
	old := incident.AssignedTo
	incident.AssignedTo = strings.TrimSpace(req.AssignedTo)

	if err := u.incidentRepo.Update(ctx, incident); err != nil {
		return nil, err
	}

	currentUser := resolveUser(userEmail)
	if err := u.saveHistory(ctx, incident.ID, domain.ActionAssigned, currentUser); err != nil {
		catcher.Warn("incidents: failed to write history", map[string]any{"error": err.Error()})
	}

	u.audit.Log(ctx, audit_connectors.Event{
		Action:     "incident.assign.success",
		EventType:  audit_domain.INCIDENT_UPDATE_SUCCESS,
		Status:     audit_domain.StatusSuccess,
		ResourceID: incident.ID.String(),
		Metadata:   map[string]any{"from": old, "to": incident.AssignedTo},
	})

	return incident, nil
}

func (u *incidentUsecase) List(ctx context.Context, query dto.IncidentListQuery) ([]domain.Incident, int64, error) {
	return u.incidentRepo.FindAll(ctx, query)
}

func (u *incidentUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	incident, err := u.incidentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrNotFound
	}
	return incident, nil
}

func (u *incidentUsecase) GetAssignees(ctx context.Context) ([]string, error) {
	names, err := u.incidentRepo.DistinctAssignees(ctx)
	if err != nil {
		return nil, err
	}
	if names == nil {
		return []string{}, nil
	}
	return names, nil
}

func resolveUser(userEmail string) string {
	if userEmail == "" {
		return "system"
	}
	return userEmail
}

// alertRows turns the requested links into rows. Their status mirrors the
// incident's: linking an alert into an open incident is what takes it off the
// queue.
func alertRows(incident *domain.Incident, list []dto.AlertLinkItem) []domain.IncidentAlert {
	rows := make([]domain.IncidentAlert, 0, len(list))
	for _, item := range list {
		row := domain.IncidentAlert{
			IncidentID:    incident.ID,
			AlertID:       item.AlertID,
			AlertName:     item.AlertName,
			AlertSeverity: domain.IncidentSeverity(item.AlertSeverity),
			AlertStatus:   string(incident.Status),
		}
		if item.AlertStatus != nil && *item.AlertStatus != "" {
			row.AlertStatus = *item.AlertStatus
		}
		rows = append(rows, row)
	}
	return rows
}

func (u *incidentUsecase) saveHistory(
	ctx context.Context,
	incidentID uuid.UUID,
	action domain.Action,
	by string,
) error {
	return u.historyRepo.Save(ctx, &domain.IncidentHistory{
		IncidentID:        incidentID,
		Action:            action,
		ActionCreatedDate: time.Now().UTC(),
		ActionCreatedBy:   &by,
	})
}

func validateAlertList(list []dto.AlertLinkItem) error {
	for _, item := range list {
		if !domain.IncidentSeverity(item.AlertSeverity).Valid() {
			return fmt.Errorf("alert %s: unknown severity %q", item.AlertID, item.AlertSeverity)
		}
	}
	return nil
}

func maxRequestedSeverity(list []dto.AlertLinkItem) domain.IncidentSeverity {
	worst := domain.IncidentSeverity("")
	for _, item := range list {
		if s := domain.IncidentSeverity(item.AlertSeverity); s.Rank() > worst.Rank() {
			worst = s
		}
	}
	return worst
}
