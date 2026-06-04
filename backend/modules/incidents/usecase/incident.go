package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	iamGateway    connectors.IAMGateway
	audit         audit_connectors.Logger
}

func NewIncidentUsecase(
	incidentRepo connectors.IncidentRepository,
	alertRepo connectors.IncidentAlertRepository,
	historyRepo connectors.IncidentHistoryRepository,
	mailer connectors.IncidentMailer,
	alertsGateway connectors.AlertsGateway,
	iamGateway connectors.IAMGateway,
	audit audit_connectors.Logger,
) connectors.IncidentUsecase {
	return &incidentUsecase{
		incidentRepo:  incidentRepo,
		alertRepo:     alertRepo,
		historyRepo:   historyRepo,
		mailer:        mailer,
		alertsGateway: alertsGateway,
		iamGateway:    iamGateway,
		audit:         audit,
	}
}

func (u *incidentUsecase) Create(ctx context.Context, userLogin string, req dto.CreateIncidentRequest) (*domain.UtmIncident, error) {
	u.audit.Log(ctx, audit_connectors.Event{
		Action:    "incident.create.attempt",
		EventType: audit_domain.INCIDENT_CREATION_ATTEMPT,
		Status:    audit_domain.StatusSuccess,
	})

	if len(req.AlertList) == 0 {
		return nil, fmt.Errorf("at least one alert is required")
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

	severity := maxSeverity(req.AlertList)

	now := time.Now().UTC()
	incident := &domain.UtmIncident{
		IncidentName:        req.IncidentName,
		IncidentDescription: req.IncidentDescription,
		IncidentStatus:      string(domain.StatusOpen),
		IncidentSeverity:    &severity,
		IncidentAssignedTo:  req.IncidentAssignedTo,
		IncidentCreatedDate: now,
	}
	if err := u.incidentRepo.Save(ctx, incident); err != nil {
		u.audit.Log(ctx, audit_connectors.Event{
			Action:       "incident.create.fail",
			EventType:    audit_domain.INCIDENT_CREATION_ATTEMPT,
			Status:       audit_domain.StatusFailure,
			ErrorMessage: err.Error(),
		})
		return nil, err
	}

	alertStatus := domain.StatusOpen.ToAlertStatus()
	for _, item := range req.AlertList {
		s := alertStatus
		row := &domain.UtmIncidentAlert{
			IncidentID:    incident.ID,
			AlertID:       item.AlertID,
			AlertName:     item.AlertName,
			AlertSeverity: item.AlertSeverity,
			AlertStatus:   &s,
		}
		if item.AlertStatus != nil {
			row.AlertStatus = item.AlertStatus
		}
		if err := u.alertRepo.Save(ctx, row); err != nil {
			return nil, err
		}
	}

	currentUser := resolveUser(userLogin)
	detail := fmt.Sprintf("Incident created with %d alerts", len(req.AlertList))
	if err := u.saveHistory(ctx, incident.ID, domain.ActionCreated, detail, currentUser); err != nil {
		catcher.Warn("incidents: failed to write history", map[string]any{"error": err.Error()})
	}

	u.audit.Log(ctx, audit_connectors.Event{
		Action:     "incident.create.success",
		EventType:  audit_domain.INCIDENT_CREATION_SUCCESS,
		Status:     audit_domain.StatusSuccess,
		ResourceID: strconv.FormatInt(incident.ID, 10),
	})

	if err := u.mailer.SendIncidentCreated(ctx, *incident); err != nil {
		catcher.Warn("incidents: mail notification failed", map[string]any{"error": err.Error()})
	}

	return incident, nil
}

func (u *incidentUsecase) AddAlerts(ctx context.Context, userLogin string, req dto.AddAlertsRequest) (*domain.UtmIncident, error) {
	u.audit.Log(ctx, audit_connectors.Event{
		Action:    "incident.alert.add.attempt",
		EventType: audit_domain.INCIDENT_ALERT_ADD_ATTEMPT,
		Status:    audit_domain.StatusSuccess,
	})

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

	alertStatus := domain.IncidentStatus(incident.IncidentStatus).ToAlertStatus()
	for _, item := range req.AlertList {
		s := alertStatus
		row := &domain.UtmIncidentAlert{
			IncidentID:    incident.ID,
			AlertID:       item.AlertID,
			AlertName:     item.AlertName,
			AlertSeverity: item.AlertSeverity,
			AlertStatus:   &s,
		}
		if item.AlertStatus != nil {
			row.AlertStatus = item.AlertStatus
		}
		if err := u.alertRepo.Save(ctx, row); err != nil {
			return nil, err
		}
	}

	currentUser := resolveUser(userLogin)
	detail := fmt.Sprintf("New %d alerts added to incident", len(req.AlertList))
	if err := u.saveHistory(ctx, incident.ID, domain.ActionAlertAdd, detail, currentUser); err != nil {
		catcher.Warn("incidents: failed to write history", map[string]any{"error": err.Error()})
	}

	u.audit.Log(ctx, audit_connectors.Event{
		Action:     "incident.alert.add.success",
		EventType:  audit_domain.INCIDENT_ALERT_ADD_SUCCESS,
		Status:     audit_domain.StatusSuccess,
		ResourceID: strconv.FormatInt(incident.ID, 10),
	})

	return incident, nil
}

func (u *incidentUsecase) ChangeStatus(ctx context.Context, userLogin string, req dto.ChangeStatusRequest) (*domain.UtmIncident, error) {
	u.audit.Log(ctx, audit_connectors.Event{
		Action:    "incident.status.change.attempt",
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

	oldStatus := domain.IncidentStatus(incident.IncidentStatus)
	newStatus := domain.IncidentStatus(req.IncidentStatus)

	incident.IncidentStatus = string(newStatus)
	if req.IncidentSolution != nil {
		incident.IncidentSolution = req.IncidentSolution
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
		if err := u.alertRepo.BulkUpdateStatus(ctx, alertIDs, newStatus.ToAlertStatus()); err != nil {
			return nil, err
		}

		observation := ""
		if req.IncidentSolution != nil {
			observation = *req.IncidentSolution
		}
		if err := u.alertsGateway.UpdateAlertStatus(ctx, alertIDs, newStatus.ToAlertStatus(), observation); err != nil {
			catcher.Warn("incidents: failed to sync OpenSearch alert status", map[string]any{"error": err.Error()})
		}
	}

	currentUser := resolveUser(userLogin)
	detail := fmt.Sprintf("Incident status changed from %s to %s", oldStatus.Label(), newStatus.Label())
	if newStatus == domain.StatusCompleted && req.IncidentSolution != nil {
		detail += fmt.Sprintf(" with solution: %s", *req.IncidentSolution)
	}
	if err := u.saveHistory(ctx, incident.ID, domain.ActionStatusChange, detail, currentUser); err != nil {
		catcher.Warn("incidents: failed to write history", map[string]any{"error": err.Error()})
	}

	u.audit.Log(ctx, audit_connectors.Event{
		Action:     "incident.status.change.success",
		EventType:  audit_domain.INCIDENT_UPDATE_SUCCESS,
		Status:     audit_domain.StatusSuccess,
		ResourceID: strconv.FormatInt(incident.ID, 10),
	})

	return incident, nil
}

func (u *incidentUsecase) List(ctx context.Context, query dto.IncidentListQuery) ([]domain.UtmIncident, int64, error) {
	return u.incidentRepo.FindAll(ctx, query)
}

func (u *incidentUsecase) GetByID(ctx context.Context, id int64) (*domain.UtmIncident, error) {
	incident, err := u.incidentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if incident == nil {
		return nil, domain.ErrNotFound
	}
	return incident, nil
}

func (u *incidentUsecase) GetUsersAssigned(ctx context.Context) ([]dto.UserAssignedDTO, error) {
	all, _, err := u.incidentRepo.FindAll(ctx, dto.IncidentListQuery{Page: 1, Size: 1000})
	if err != nil {
		return nil, err
	}

	seen := make(map[int64]struct{})
	var ids []int64

	for _, inc := range all {
		if inc.IncidentAssignedTo == nil || *inc.IncidentAssignedTo == "" {
			continue
		}
		parts := strings.Split(*inc.IncidentAssignedTo, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				continue
			}
			if _, dup := seen[id]; !dup {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}

	if len(ids) == 0 {
		return []dto.UserAssignedDTO{}, nil
	}

	return u.iamGateway.FindUsersByIDs(ctx, ids)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// resolveUser falls back to "system" when no authenticated user is present.
// The login is supplied by the handler from the request context, matching the
// identity-at-the-boundary convention used across the backend.
func resolveUser(userLogin string) string {
	if userLogin == "" {
		return "system"
	}
	return userLogin
}

func (u *incidentUsecase) saveHistory(
	ctx context.Context,
	incidentID int64,
	action domain.HistoryAction,
	detail string,
	by string,
) error {
	now := time.Now().UTC()
	h := &domain.UtmIncidentHistory{
		IncidentID:        incidentID,
		Action:            action.Label,
		ActionType:        action.Type,
		ActionDetail:      &detail,
		ActionCreatedDate: now,
		ActionCreatedBy:   &by,
	}
	return u.historyRepo.Save(ctx, h)
}

func maxSeverity(list []dto.AlertLinkItem) int {
	max := 0
	for _, item := range list {
		if item.AlertSeverity > max {
			max = item.AlertSeverity
		}
	}
	return max
}
