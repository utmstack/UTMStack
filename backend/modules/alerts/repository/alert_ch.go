package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/threatwinds/go-sdk/store"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const falsePositiveTag = "False positive"

const (
	patchLimit = 10000
	patchChunk = 250
)

type chAlertRepo struct {
	store  *eventstore.Store
	locker *alertLocker
}

func NewCHAlertRepository(s *eventstore.Store, db *gorm.DB) connectors.AlertRepository {
	return &chAlertRepo{store: s, locker: newAlertLocker(db)}
}

func alertScope(ctx context.Context) (store.Scope, error) {
	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" {
		if tenancy.Enabled() {
			return store.Scope{}, ErrNoTenantScope
		}
		tenant = store.AllTenants
	}
	return store.Scope{Tenant: tenant, Dataset: eventstore.DatasetAlerts}, nil
}

var ErrNoTenantScope = errors.New("alerts: no tenant in scope")

func idIn(ids []string) store.Filter {
	return store.Filter{Field: "id", Op: store.OpIn, Value: ids}
}

func parentIn(ids []string) store.Filter {
	return store.Filter{Field: "parentId", Op: store.OpIn, Value: ids}
}

func (r *chAlertRepo) patch(ctx context.Context, filters []store.Filter, mutate func(doc map[string]any) bool) error {
	scope, err := alertScope(ctx)
	if err != nil {
		return err
	}

	buckets, err := r.store.TopValues(ctx, scope, "id", filters, patchLimit+1)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(buckets))
	for _, b := range buckets {
		if b.Key != "" {
			ids = append(ids, b.Key)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	if len(ids) > patchLimit {
		return domain.ErrTooManyAlerts
	}

	for start := 0; start < len(ids); start += patchChunk {
		end := min(start+patchChunk, len(ids))
		if err := r.rewrite(ctx, scope, ids[start:end], mutate); err != nil {
			return err
		}
	}
	return nil
}

func (r *chAlertRepo) rewrite(ctx context.Context, scope store.Scope, ids []string, mutate func(map[string]any) bool) error {
	release, err := r.locker.lock(ctx, ids)
	if err != nil {
		return err
	}
	defer release()

	rows, err := r.store.FetchN(ctx, scope, []store.Filter{idIn(ids)}, len(ids))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	w, err := r.store.BulkWriter(eventstore.DatasetAlerts)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	written := 0

	for _, raw := range rows {
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			continue
		}
		if !mutate(doc) {
			continue
		}

		tenant, _ := doc["tenantId"].(string)
		id, _ := doc["id"].(string)
		if tenant == "" || id == "" {
			continue
		}

		doc["lastUpdate"] = now

		out, err := json.Marshal(doc)
		if err != nil {
			continue
		}
		if err := w.Write(store.Scope{Tenant: tenant, Dataset: eventstore.DatasetAlerts}, out); err != nil {
			return err
		}
		written++
	}

	if written == 0 {
		return nil
	}
	return w.Close(ctx)
}

func (r *chAlertRepo) patchByIDOrParent(ctx context.Context, ids []string, mutate func(map[string]any) bool) error {
	if err := r.patch(ctx, []store.Filter{idIn(ids)}, mutate); err != nil {
		return err
	}
	return r.patch(ctx, []store.Filter{parentIn(ids)}, mutate)
}

func withHistory(entries []connectors.HistoryEntry, mutate func(map[string]any) bool) func(map[string]any) bool {
	if len(entries) == 0 {
		return mutate
	}

	byID := make(map[string][]map[string]any, len(entries))
	for _, e := range entries {
		at := e.At
		if at.IsZero() {
			at = time.Now().UTC()
		}
		byID[e.AlertID] = append(byID[e.AlertID], map[string]any{
			"user":      e.User,
			"action":    string(e.Action),
			"newValue":  e.NewValue,
			"timestamp": at.UTC().Format(time.RFC3339),
		})
	}

	return func(doc map[string]any) bool {
		changed := mutate(doc)

		id, _ := doc["id"].(string)
		pending := byID[id]
		if len(pending) == 0 {
			return changed
		}
		existing, _ := doc["history"].([]any)
		for _, entry := range pending {
			existing = append(existing, entry)
		}
		doc["history"] = existing
		return true
	}
}

func (r *chAlertRepo) UpdateStatus(ctx context.Context, alertIDs []string, status domain.AlertStatus, observation string, addFalsePositiveTag bool, history []connectors.HistoryEntry) error {
	return r.patchByIDOrParent(ctx, alertIDs, withHistory(history, func(doc map[string]any) bool {
		doc["status"] = string(status)
		doc["statusObservation"] = observation
		if addFalsePositiveTag {
			tags := stringSlice(doc["tags"])
			if !contains(tags, falsePositiveTag) {
				doc["tags"] = append(tags, falsePositiveTag)
			}
		}
		return true
	}))
}

func (r *chAlertRepo) UpdateNotes(ctx context.Context, alertID, notes string, history []connectors.HistoryEntry) error {
	return r.patch(ctx, []store.Filter{idIn([]string{alertID})}, withHistory(history, func(doc map[string]any) bool {
		doc["notes"] = notes
		return true
	}))
}

func (r *chAlertRepo) UpdateAssignee(ctx context.Context, alertID, assignee string, history []connectors.HistoryEntry) error {
	return r.patch(ctx, []store.Filter{idIn([]string{alertID})}, withHistory(history, func(doc map[string]any) bool {
		doc["assignee"] = assignee
		return true
	}))
}

func (r *chAlertRepo) UpdateTags(ctx context.Context, alertIDs []string, tags []string, history []connectors.HistoryEntry) error {
	return r.patch(ctx, []store.Filter{idIn(alertIDs)}, withHistory(history, func(doc map[string]any) bool {
		if len(tags) == 0 {
			doc["tags"] = []string{}
		} else {
			doc["tags"] = tags
		}
		return true
	}))
}

func contains(vals []string, want string) bool {
	for _, v := range vals {
		if v == want {
			return true
		}
	}
	return false
}

func (r *chAlertRepo) ConvertToIncident(ctx context.Context, alertIDs []string, name, id string, createdAt time.Time, createdBy, source string, history []connectors.HistoryEntry) error {
	filters := []store.Filter{
		idIn(alertIDs),
		{Field: "isIncident", Op: store.OpEq, Value: false},
	}
	return r.patch(ctx, filters, withHistory(history, func(doc map[string]any) bool {
		doc["isIncident"] = true
		doc["incidentDetail"] = map[string]any{
			"incidentName": name,
			"incidentId":   id,
			"creationDate": createdAt.UTC().Format(time.RFC3339),
			"createdBy":    createdBy,
			"source":       source,
		}
		return true
	}))
}

func (r *chAlertRepo) CountOpenAlerts(ctx context.Context) (int64, error) {
	filters := []store.Filter{
		{Field: "status", Op: store.OpEq, Value: string(domain.AlertStatusOpen)},
		{Field: "parentId", Op: store.OpEq, Value: ""},
		{Field: "tags", Op: store.OpNotContains, Value: falsePositiveTag},
	}
	scope, err := alertScope(ctx)
	if err != nil {
		return 0, err
	}
	return r.store.Count(ctx, scope, filters)
}

func (r *chAlertRepo) CountByStatus(ctx context.Context, status domain.AlertStatus) (int64, error) {
	scope, err := alertScope(ctx)
	if err != nil {
		return 0, err
	}
	filters := []store.Filter{{Field: "status", Op: store.OpEq, Value: string(status)}}
	return r.store.Count(ctx, scope, filters)
}

func (r *chAlertRepo) SearchByIDs(ctx context.Context, alertIDs []string) ([]domain.UtmAlert, error) {
	scope, err := alertScope(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.store.FetchN(ctx, scope, []store.Filter{idIn(alertIDs)}, patchLimit)
	if err != nil {
		return nil, err
	}
	return decodeAlerts(rows), nil
}

func (r *chAlertRepo) ListEchoes(ctx context.Context, parentID string, from, size int, sortBy, sortOrder string) ([]domain.UtmAlert, int64, error) {
	page := store.Page{
		Offset: from,
		Limit:  size,
		SortBy: sortBy,
		Order:  store.Desc,
	}
	if sortOrder == "asc" {
		page.Order = store.Asc
	}

	scope, err := alertScope(ctx)
	if err != nil {
		return nil, 0, err
	}
	filters := []store.Filter{{Field: "parentId", Op: store.OpEq, Value: parentID}}
	rows, total, err := r.store.FetchPage(ctx, scope, filters, page)
	if err != nil {
		return nil, 0, err
	}
	return decodeAlerts(rows), total, nil
}

func (r *chAlertRepo) GetRawByID(ctx context.Context, alertID string) (json.RawMessage, error) {
	scope, err := alertScope(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.store.FetchN(ctx, scope, []store.Filter{idIn([]string{alertID})}, 1)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

func decodeAlerts(rows []json.RawMessage) []domain.UtmAlert {
	out := make([]domain.UtmAlert, 0, len(rows))
	for _, raw := range rows {
		var a domain.UtmAlert
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		out = append(out, a)
	}
	return out
}

func stringSlice(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
