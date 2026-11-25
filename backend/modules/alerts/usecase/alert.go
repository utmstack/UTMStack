package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type alertUsecase struct {
	repo     connectors.AlertRepository
	mailer   connectors.AlertMailer
	resolver connectors.CorrelationResolver // injected post-construction; may be nil
	hooks    []connectors.RaisedHook        // registered at start, before any alert is notified
}

func NewAlertUsecase(repo connectors.AlertRepository, mailer connectors.AlertMailer) connectors.AlertUsecase {
	return &alertUsecase{repo: repo, mailer: mailer}
}

// NotifyRaised tells what should know that an alert was raised. The event processor calls it once per new
// alert and never for the echo of a grouped one.
//
// Reading the alert is the costly part: the table is keyed by name and time, so finding one by id reads a
// large share of it. So nothing reads it that does not need it: the hooks are given the id and the
// tenant, which is all they start from, and the e-mail, which does need the alert, is only prepared when
// somebody is set to receive it.
func (u *alertUsecase) NotifyRaised(ctx context.Context, alertID string) error {
	if alertID == "" {
		return nil
	}

	tenant := authz.TenantIDFromContext(ctx)

	var errs []error

	for _, h := range u.hooks {
		if err := h(ctx, tenant, alertID); err != nil {
			errs = append(errs, err)
		}
	}

	if u.mailer != nil && u.mailer.Enabled(ctx) {
		alert, err := u.awaitAlert(ctx, alertID)

		switch {
		case err != nil:
			errs = append(errs, err)
		case alert.ParentID == "":
			if err := u.mailer.SendAlertRaised(ctx, alert); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// notifyWindow is how far back to look for an alert that was just raised.
const notifyWindow = 6 * time.Hour

// notifyRetryDelays is how long to wait between looks for an alert that was just written. The event
// processor tells the backend as soon as it has written the alert, and the store takes a moment to make
// a new row readable, so the first look often finds nothing. A variable so tests do not wait.
var notifyRetryDelays = []time.Duration{250 * time.Millisecond, 750 * time.Millisecond, 2 * time.Second, 4 * time.Second}

// awaitAlert reads the alert, trying again while it is not there yet.
func (u *alertUsecase) awaitAlert(ctx context.Context, alertID string) (domain.UtmAlert, error) {
	for attempt := 0; ; attempt++ {
		alerts, err := u.repo.SearchRecentByIDs(ctx, []string{alertID}, notifyWindow)
		if err != nil {
			return domain.UtmAlert{}, err
		}

		if len(alerts) > 0 {
			return alerts[0], nil
		}

		if attempt >= len(notifyRetryDelays) {
			return domain.UtmAlert{}, domain.ErrAlertNotFound
		}

		select {
		case <-ctx.Done():
			return domain.UtmAlert{}, ctx.Err()
		case <-time.After(notifyRetryDelays[attempt]):
		}
	}
}

func (u *alertUsecase) AddRaisedHook(h connectors.RaisedHook) {
	u.hooks = append(u.hooks, h)
}

func (u *alertUsecase) SetCorrelationResolver(r connectors.CorrelationResolver) {
	u.resolver = r
}

// resolveUser falls back to "system" when no authenticated user is present
// (e.g. internal callers). The email is supplied by the handler from the
// request context, matching the audit/identity-at-the-boundary convention.
func resolveUser(userEmail string) string {
	if userEmail == "" {
		return "system"
	}
	return userEmail
}

func (u *alertUsecase) UpdateStatus(ctx context.Context, userEmail string, req dto.UpdateAlertStatusRequest) error {
	if !domain.IsValid(req.Status) {
		return domain.ErrInvalidAlertStatus
	}

	user := resolveUser(userEmail)

	// Read first: the history says what the status changed from, and once the
	// change lands there is nothing left to read it off.
	oldAlerts, _ := u.repo.SearchByIDs(ctx, req.AlertIDs)

	tagFalsePositive := req.Status == domain.AlertStatusCompleted && req.AddFalsePositiveTag

	return u.repo.UpdateStatus(ctx, req.AlertIDs, req.Status, req.StatusObservation,
		tagFalsePositive, buildStatusEntries(oldAlerts, user, req))
}

func (u *alertUsecase) UpdateNotes(ctx context.Context, userEmail string, alertID string, notes string) error {
	if alertID == "" {
		return domain.ErrMissingAlertID
	}

	user := resolveUser(userEmail)

	return u.repo.UpdateNotes(ctx, alertID, notes,
		[]connectors.HistoryEntry{buildNotesEntry(alertID, user, notes)})
}

func (u *alertUsecase) RecordAssessment(ctx context.Context, userEmail string, alertID string, assessment string) error {
	if alertID == "" {
		return domain.ErrMissingAlertID
	}

	user := resolveUser(userEmail)

	return u.repo.MergeAssessment(ctx, alertID, assessment,
		[]connectors.HistoryEntry{buildNotesEntry(alertID, user, assessment)})
}

func (u *alertUsecase) UpdateAssignee(ctx context.Context, userEmail string, alertID string, assignee string) error {
	if alertID == "" {
		return domain.ErrMissingAlertID
	}

	user := resolveUser(userEmail)

	return u.repo.UpdateAssignee(ctx, alertID, assignee,
		[]connectors.HistoryEntry{buildAssigneeEntry(alertID, user, assignee)})
}

func (u *alertUsecase) UpdateTags(ctx context.Context, userEmail string, req dto.UpdateAlertTagsRequest) error {
	user := resolveUser(userEmail)

	return u.repo.UpdateTags(ctx, req.AlertIDs, req.Tags,
		buildTagEntries(user, req.AlertIDs, req.Tags))
}

func (u *alertUsecase) ConvertToIncident(ctx context.Context, userEmail string, req dto.ConvertToIncidentRequest) error {
	createdBy := resolveUser(userEmail)
	createdAt := time.Now().UTC()

	return u.repo.ConvertToIncident(ctx, req.AlertIDs, req.IncidentName, req.IncidentID,
		createdAt, createdBy, req.IncidentSource, buildIncidentEntries(createdBy, createdAt, req))
}

func (u *alertUsecase) CountOpenAlerts(ctx context.Context) (*dto.CountOpenAlertsResponse, error) {
	count, err := u.repo.CountOpenAlerts(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.CountOpenAlertsResponse{Count: count}, nil
}

func (u *alertUsecase) ListEchoes(ctx context.Context, parentID string, page, size int, sortBy, sortOrder string) ([]domain.UtmAlert, int64, error) {
	if parentID == "" {
		return nil, 0, domain.ErrMissingAlertID
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	if sortBy == "" {
		sortBy = "@timestamp"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	from := (page - 1) * size
	return u.repo.ListEchoes(ctx, parentID, from, size, sortBy, sortOrder)
}

// ---------------------------------------------------------------------------
// Internal helpers for HistoryEntry construction
// ---------------------------------------------------------------------------

// buildStatusEntries records the change, not a sentence describing it. The
// previous status is part of it because "from Open to Completed" is what a
// reader wants and it cannot be recovered afterwards.
func buildStatusEntries(oldAlerts []domain.UtmAlert, user string, req dto.UpdateAlertStatusRequest) []connectors.HistoryEntry {
	now := time.Now().UTC()

	oldByID := make(map[string]domain.UtmAlert, len(oldAlerts))
	for _, a := range oldAlerts {
		oldByID[a.ID] = a
	}

	entries := make([]connectors.HistoryEntry, 0, len(req.AlertIDs))
	for _, id := range req.AlertIDs {
		newVal := map[string]any{
			"status":            string(req.Status),
			"statusObservation": req.StatusObservation,
		}
		old, ok := oldByID[id]
		if ok {
			newVal["previousStatus"] = string(old.Status)
		} else {
			continue
		}

		newValJSON, _ := json.Marshal(newVal)

		entries = append(entries, connectors.HistoryEntry{
			AlertID:  id,
			User:     user,
			Action:   domain.ActionUpdateStatus,
			NewValue: string(newValJSON),
			At:       now,
		})
	}
	return entries
}

func buildAssigneeEntry(alertID, user, assignee string) connectors.HistoryEntry {
	newVal := map[string]any{"assignee": assignee}
	newValJSON, _ := json.Marshal(newVal)

	return connectors.HistoryEntry{
		AlertID:  alertID,
		User:     user,
		Action:   domain.ActionUpdateAssignee,
		NewValue: string(newValJSON),
		At:       time.Now().UTC(),
	}
}

func buildNotesEntry(alertID, user, notes string) connectors.HistoryEntry {
	newVal := map[string]any{"notes": notes}
	newValJSON, _ := json.Marshal(newVal)

	return connectors.HistoryEntry{
		AlertID:  alertID,
		User:     user,
		Action:   domain.ActionUpdateNotes,
		NewValue: string(newValJSON),
		At:       time.Now().UTC(),
	}
}

func buildTagEntries(user string, alertIDs, tags []string) []connectors.HistoryEntry {
	tagsCSV := strings.Join(tags, ",")

	var newVal map[string]any
	if len(tags) > 0 {
		newVal = map[string]any{"tags": tagsCSV}
	} else {
		newVal = map[string]any{"tags": ""}
	}
	newValJSON, _ := json.Marshal(newVal)

	now := time.Now().UTC()

	entries := make([]connectors.HistoryEntry, 0, len(alertIDs))
	for _, id := range alertIDs {
		entries = append(entries, connectors.HistoryEntry{
			AlertID:  id,
			User:     user,
			Action:   domain.ActionUpdateTags,
			NewValue: string(newValJSON),
			At:       now,
		})
	}
	return entries
}

func buildIncidentEntries(createdBy string, createdAt time.Time, req dto.ConvertToIncidentRequest) []connectors.HistoryEntry {
	newVal := map[string]any{
		"isIncident":                  true,
		"incidentDetail.incidentName": req.IncidentName,
		"incidentDetail.incidentId":   req.IncidentID,
		"incidentDetail.createdBy":    createdBy,
		"incidentDetail.creationDate": createdAt.UTC().Format(time.RFC3339),
		"incidentDetail.source":       req.IncidentSource,
	}
	newValJSON, _ := json.Marshal(newVal)

	now := time.Now().UTC()

	entries := make([]connectors.HistoryEntry, 0, len(req.AlertIDs))
	for _, id := range req.AlertIDs {
		entries = append(entries, connectors.HistoryEntry{
			AlertID:  id,
			User:     createdBy,
			Action:   domain.ActionMarkAsIncident,
			NewValue: string(newValJSON),
			At:       now,
		})
	}
	return entries
}
