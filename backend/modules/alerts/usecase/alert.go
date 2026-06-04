package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
)

type alertUsecase struct {
	repo    connectors.AlertRepository
	history connectors.HistoryRecorder
}

func NewAlertUsecase(
	repo connectors.AlertRepository,
	history connectors.HistoryRecorder,
) connectors.AlertUsecase {
	return &alertUsecase{repo: repo, history: history}
}

// resolveUser falls back to "system" when no authenticated user is present
// (e.g. internal callers). The login is supplied by the handler from the
// request context, matching the audit/identity-at-the-boundary convention.
func resolveUser(userLogin string) string {
	if userLogin == "" {
		return "system"
	}
	return userLogin
}

func (u *alertUsecase) UpdateStatus(ctx context.Context, userLogin string, req dto.UpdateAlertStatusRequest) error {
	if !domain.IsValid(domain.AlertStatus(req.Status)) {
		return domain.ErrInvalidAlertStatus
	}

	user := resolveUser(userLogin)

	label := domain.StatusName(domain.AlertStatus(req.Status))

	var oldAlerts []domain.UtmAlert
	if u.history != nil {
		oldAlerts, _ = u.repo.SearchByIDs(ctx, req.AlertIDs)
	}

	if req.Status == int(domain.AlertStatusCompleted) && req.AddFalsePositiveTag {
		if err := u.repo.UpdateStatus(ctx, req.AlertIDs, req.Status, label, req.StatusObservation); err != nil {
			return err
		}
		if err := u.repo.UpdateStatusAndTag(ctx, req.AlertIDs); err != nil {
			return err
		}
		if err := u.repo.UpdateStatus(ctx, req.AlertIDs, req.Status, label, req.StatusObservation); err != nil {
			return err
		}
	} else {
		if err := u.repo.UpdateStatus(ctx, req.AlertIDs, req.Status, label, req.StatusObservation); err != nil {
			return err
		}
	}

	if u.history != nil {
		entries := buildStatusEntries(oldAlerts, user, label, req)
		if err := u.history.Record(ctx, entries); err != nil {
			return err
		}
	}

	return nil
}

func (u *alertUsecase) UpdateNotes(ctx context.Context, userLogin string, alertID string, notes string) error {
	if alertID == "" {
		return domain.ErrMissingAlertID
	}

	user := resolveUser(userLogin)

	if err := u.repo.UpdateNotes(ctx, alertID, notes); err != nil {
		return err
	}

	if u.history != nil {
		entry := buildNotesEntry(alertID, user, notes)
		if err := u.history.Record(ctx, []connectors.HistoryEntry{entry}); err != nil {
			return err
		}
	}
	return nil
}

func (u *alertUsecase) UpdateTags(ctx context.Context, userLogin string, req dto.UpdateAlertTagsRequest) error {
	user := resolveUser(userLogin)

	if err := u.repo.UpdateTags(ctx, req.AlertIDs, req.Tags); err != nil {
		return err
	}

	if u.history != nil {
		entries := buildTagEntries(user, req.AlertIDs, req.Tags)
		if err := u.history.Record(ctx, entries); err != nil {
			return err
		}
	}
	return nil
}

func (u *alertUsecase) ConvertToIncident(ctx context.Context, userLogin string, req dto.ConvertToIncidentRequest) error {
	createdBy := resolveUser(userLogin)
	createdAt := time.Now().UTC()

	if err := u.repo.ConvertToIncident(ctx, req.AlertIDs, req.IncidentName, req.IncidentID, createdAt, createdBy, req.IncidentSource); err != nil {
		return err
	}

	if u.history != nil {
		entries := buildIncidentEntries(createdBy, createdAt, req)
		if err := u.history.Record(ctx, entries); err != nil {
			return err
		}
	}
	return nil
}

func (u *alertUsecase) CountOpenAlerts(ctx context.Context) (*dto.CountOpenAlertsResponse, error) {
	count, err := u.repo.CountOpenAlerts(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.CountOpenAlertsResponse{Count: count}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers for HistoryEntry construction
// ---------------------------------------------------------------------------

func buildStatusEntries(oldAlerts []domain.UtmAlert, user, newLabel string, req dto.UpdateAlertStatusRequest) []connectors.HistoryEntry {
	// Build new-value map that matches Java's logManualAlertStatusChange.
	newVal := map[string]any{
		"status":            req.Status,
		"statusLabel":       newLabel,
		"statusObservation": req.StatusObservation,
	}
	newValJSON, _ := json.Marshal(newVal)

	now := time.Now().UTC()

	// Index old alerts by ID for fast lookup.
	oldByID := make(map[string]domain.UtmAlert, len(oldAlerts))
	for _, a := range oldAlerts {
		oldByID[a.ID] = a
	}

	entries := make([]connectors.HistoryEntry, 0, len(req.AlertIDs))
	for _, id := range req.AlertIDs {
		e := connectors.HistoryEntry{
			AlertID:  id,
			User:     user,
			Action:   domain.ActionUpdateStatus,
			NewValue: string(newValJSON),
			At:       now,
		}

		if old, ok := oldByID[id]; ok {
			oldLabel := old.StatusLabel
			if oldLabel == "" {
				oldLabel = domain.StatusName(domain.AlertStatus(old.Status))
			}
			if strings.TrimSpace(req.StatusObservation) == "" {
				e.Message = fmt.Sprintf(domain.MsgStatusChangedNoObs, user, oldLabel, newLabel)
			} else {
				e.Message = fmt.Sprintf(domain.MsgStatusChangedWithObs, user, oldLabel, newLabel, req.StatusObservation)
			}
		}

		entries = append(entries, e)
	}
	return entries
}

func buildNotesEntry(alertID, user, notes string) connectors.HistoryEntry {
	newVal := map[string]any{"notes": notes}
	newValJSON, _ := json.Marshal(newVal)

	var msg string
	if strings.TrimSpace(notes) != "" {
		msg = fmt.Sprintf(domain.MsgNotesUpdated, user, notes)
	} else {
		msg = fmt.Sprintf(domain.MsgNotesCleared, user)
	}

	return connectors.HistoryEntry{
		AlertID:  alertID,
		User:     user,
		Action:   domain.ActionUpdateNotes,
		NewValue: string(newValJSON),
		Message:  msg,
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
		var msg string
		if len(tags) > 0 {
			msg = fmt.Sprintf(domain.MsgTagsManual, user, tagsCSV)
		} else {
			msg = fmt.Sprintf(domain.MsgTagsManualCleared, user)
		}

		entries = append(entries, connectors.HistoryEntry{
			AlertID:  id,
			User:     user,
			Action:   domain.ActionUpdateTags,
			NewValue: string(newValJSON),
			Message:  msg,
			At:       now,
		})
	}
	return entries
}

func buildIncidentEntries(createdBy string, createdAt time.Time, req dto.ConvertToIncidentRequest) []connectors.HistoryEntry {
	incidentRef := fmt.Sprintf("%s(%d)", req.IncidentName, req.IncidentID)

	newVal := map[string]any{
		"isIncident":                  true,
		"incidentDetail.incidentName": req.IncidentName,
		"incidentDetail.incidentId":   req.IncidentID,
		"incidentDetail.createdBy":    createdBy,
		"incidentDetail.creationDate": createdAt.UTC().Format(time.RFC3339),
		"incidentDetail.source":       req.IncidentSource,
	}
	newValJSON, _ := json.Marshal(newVal)

	msg := fmt.Sprintf(domain.MsgConvertToIncident, incidentRef, createdBy)

	now := time.Now().UTC()

	entries := make([]connectors.HistoryEntry, 0, len(req.AlertIDs))
	for _, id := range req.AlertIDs {
		entries = append(entries, connectors.HistoryEntry{
			AlertID:  id,
			User:     createdBy,
			Action:   domain.ActionMarkAsIncident,
			NewValue: string(newValJSON),
			Message:  msg,
			At:       now,
		})
	}
	return entries
}
