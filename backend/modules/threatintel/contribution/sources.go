// Package contribution sends the entities found in an instance's incidents to ThreatWinds, through the
// Customers Manager's proxy, so that what one installation learns helps the others. It runs inside
// the backend as a periodic task.
package contribution

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/utmstack/go-sdk/v2/store"

	incidents_connectors "github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	incidents_domain "github.com/utmstack/utmstack/backend/modules/incidents/domain"
	incidents_dto "github.com/utmstack/utmstack/backend/modules/incidents/dto"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/models"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

// pageSize is the most the incidents lists return in one page.
const pageSize = 200

// maxPages stops a list that never ends, which would otherwise keep a cycle going until its deadline.
const maxPages = 500

// IncidentSource reads incidents and their alerts through the incident use cases. It reads the
// platform tenant's, which is what an internal caller without a tenant has always seen.
type IncidentSource struct {
	incidents incidents_connectors.IncidentUsecase
	links     incidents_connectors.IncidentAlertUsecase
}

func NewIncidentSource(incidents incidents_connectors.IncidentUsecase, links incidents_connectors.IncidentAlertUsecase) *IncidentSource {
	return &IncidentSource{incidents: incidents, links: links}
}

func platform(ctx context.Context) context.Context {
	return authz.WithTenantID(ctx, authz.DefaultTenantID)
}

// RecentIncidents returns the incidents that are open or under review, newest first within each.
func (s *IncidentSource) RecentIncidents(ctx context.Context) ([]*models.Incident, error) {
	ctx = platform(ctx)

	var all []*models.Incident

	for _, status := range []string{string(incidents_domain.StatusOpen), string(incidents_domain.StatusInReview)} {
		status := status

		for page := 1; page <= maxPages; page++ {
			batch, _, err := s.incidents.List(ctx, incidents_dto.IncidentListQuery{
				IncidentStatus: &status,
				Sort:           "incidentCreatedDate,desc",
				Page:           page,
				Size:           pageSize,
			})
			if err != nil {
				return nil, fmt.Errorf("listing %s incidents: %w", status, err)
			}

			for _, i := range batch {
				all = append(all, &models.Incident{
					ID:          i.ID.String(),
					Name:        i.Name,
					Description: i.Description,
					Status:      string(i.Status),
					Severity:    string(i.Severity),
					CreatedDate: i.CreatedDate,
					AlertCount:  i.AlertCount,
				})
			}

			if len(batch) < pageSize {
				break
			}
		}
	}

	return all, nil
}

// IncidentAlerts returns the alerts linked to an incident.
func (s *IncidentSource) IncidentAlerts(ctx context.Context, incidentID string) ([]*models.IncidentAlert, error) {
	id, err := uuid.Parse(incidentID)
	if err != nil {
		return nil, fmt.Errorf("incident id %q: %w", incidentID, err)
	}

	ctx = platform(ctx)

	var all []*models.IncidentAlert

	for page := 1; page <= maxPages; page++ {
		batch, _, err := s.links.List(ctx, incidents_dto.IncidentAlertListQuery{IncidentID: &id, Page: page, Size: pageSize})
		if err != nil {
			return nil, fmt.Errorf("listing the alerts of incident %s: %w", incidentID, err)
		}

		for _, a := range batch {
			all = append(all, &models.IncidentAlert{
				ID:            a.ID.String(),
				IncidentID:    a.IncidentID.String(),
				AlertID:       a.AlertID,
				AlertName:     a.AlertName,
				AlertSeverity: string(a.AlertSeverity),
				AlertStatus:   a.AlertStatus,
			})
		}

		if len(batch) < pageSize {
			break
		}
	}

	return all, nil
}

// AlertFetcher is the part of the event store that reads one document.
type AlertFetcher interface {
	FetchByID(ctx context.Context, scope store.Scope, id string) (json.RawMessage, error)
}

// AlertSource reads alerts from the event store. An incident carries no tenant of its own, so the
// lookup is across tenants; the identifier is a UUID, so it stays unambiguous.
type AlertSource struct {
	store AlertFetcher
}

func NewAlertSource(s AlertFetcher) *AlertSource { return &AlertSource{store: s} }

func (s *AlertSource) AlertByID(ctx context.Context, alertID string) (*models.Alert, error) {
	raw, err := s.store.FetchByID(ctx, store.Scope{Tenant: store.AllTenants, Dataset: eventstore.DatasetAlerts}, alertID)
	if err != nil {
		return nil, fmt.Errorf("alert %s: %w", alertID, err)
	}

	var alert models.Alert
	if err := json.Unmarshal(raw, &alert); err != nil {
		return nil, fmt.Errorf("decoding alert %s: %w", alertID, err)
	}

	if alert.ID == "" {
		return nil, fmt.Errorf("alert %s came back without an id", alertID)
	}

	return &alert, nil
}
