package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
)

// TODO(module-15): wire IndexPatternRegistry.Alerts here when all consumers are ported.
// alertIndex is the OpenSearch index pattern for alert documents.
// After Java Liquibase changeset 20241227001, 'alert-*' was prefixed to 'v11-alert-*'.
// TODO(module-15): wire IndexPatternRegistry.Alerts here when all consumers are ported.
const alertIndex = "v11-alert-*"

// falsePositiveTag is the canonical false-positive tag string.
const falsePositiveTag = "False positive"

// ---------------------------------------------------------------------------
// Painless script constants — all are static string literals; NO string concat.
// ---------------------------------------------------------------------------

const updateStatusScript = `
ctx._source.status = params.status;
ctx._source.statusLabel = params.statusLabel;
ctx._source.statusObservation = params.statusObservation;
`

const updateTagsScript = `ctx._source.tags = params.tags;`

const updateNotesScript = `ctx._source.notes = params.notes;`

const updateAssigneeScript = `ctx._source.assignee = params.assignee;`

// addFalsePositiveScript adds the "False positive" tag without disturbing others.
const addFalsePositiveScript = `
if (ctx._source.tags == null) { ctx._source.tags = []; }
if (!ctx._source.tags.contains(params.tag)) { ctx._source.tags.add(params.tag); }
`

const convertToIncidentScript = `
ctx._source.isIncident = true;
if (ctx._source.incidentDetail == null) { ctx._source.incidentDetail = [:]; }
ctx._source.incidentDetail.incidentName = params.incidentName;
ctx._source.incidentDetail.incidentId   = params.incidentId;
ctx._source.incidentDetail.creationDate = params.creationDate;
ctx._source.incidentDetail.createdBy    = params.createdBy;
ctx._source.incidentDetail.source       = params.source;
`

// ---------------------------------------------------------------------------
// osAlertRepo implements connectors.AlertRepository against the go-sdk `os` client.
// ---------------------------------------------------------------------------

type osAlertRepo struct{}

func NewOSAlertRepository() connectors.AlertRepository {
	return &osAlertRepo{}
}

func idOrParentFilter(alertIDs []string) map[string]any {
	return map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				termsQuery("id.keyword", alertIDs),
				termsQuery("parentId.keyword", alertIDs),
			},
			"minimum_should_match": 1,
		},
	}
}

func idOnlyFilter(alertIDs []string) map[string]any {
	return termsQuery("id.keyword", alertIDs)
}

func (r *osAlertRepo) UpdateStatus(ctx context.Context, alertIDs []string, status int, statusLabel, observation string) error {
	script := Script{
		Source: updateStatusScript,
		Params: map[string]any{
			"status":            status,
			"statusLabel":       statusLabel,
			"statusObservation": observation,
		},
	}
	return osUpdateByQuery(ctx, alertIndex, idOrParentFilter(alertIDs), script)
}

func (r *osAlertRepo) UpdateStatusAndTag(ctx context.Context, alertIDs []string) error {
	script := Script{
		Source: addFalsePositiveScript,
		Params: map[string]any{
			"tag": falsePositiveTag,
		},
	}
	return osUpdateByQuery(ctx, alertIndex, idOrParentFilter(alertIDs), script)
}

func (r *osAlertRepo) UpdateNotes(ctx context.Context, alertID, notes string) error {
	script := Script{
		Source: updateNotesScript,
		Params: map[string]any{
			"notes": notes,
		},
	}
	filter := termQuery("id.keyword", alertID)
	return osUpdateByQuery(ctx, alertIndex, filter, script)
}

func (r *osAlertRepo) UpdateAssignee(ctx context.Context, alertID, assignee string) error {
	script := Script{
		Source: updateAssigneeScript,
		Params: map[string]any{
			"assignee": assignee,
		},
	}
	filter := termQuery("id.keyword", alertID)
	return osUpdateByQuery(ctx, alertIndex, filter, script)
}

func (r *osAlertRepo) UpdateTags(ctx context.Context, alertIDs []string, tags []string) error {
	var tagsParam any
	if len(tags) > 0 {
		tagsParam = tags
	}
	script := Script{
		Source: updateTagsScript,
		Params: map[string]any{
			"tags": tagsParam,
		},
	}
	return osUpdateByQuery(ctx, alertIndex, idOnlyFilter(alertIDs), script)
}

func (r *osAlertRepo) ConvertToIncident(ctx context.Context, alertIDs []string, name string, id int, createdAt time.Time, createdBy, source string) error {
	filter := map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				idOnlyFilter(alertIDs),
				termQuery("isIncident", false),
			},
		},
	}
	script := Script{
		Source: convertToIncidentScript,
		Params: map[string]any{
			"incidentName": name,
			"incidentId":   id,
			"creationDate": createdAt.UTC().Format(time.RFC3339),
			"createdBy":    createdBy,
			"source":       source,
		},
	}
	return osUpdateByQuery(ctx, alertIndex, filter, script)
}

func (r *osAlertRepo) CountOpenAlerts(ctx context.Context) (int64, error) {
	query := map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				termQuery("status", int(domain.AlertStatusOpen)),
			},
			"must_not": []map[string]any{
				{
					"term": map[string]any{
						"tags": falsePositiveTag,
					},
				},
				{
					"exists": map[string]any{
						"field": "parentId.keyword",
					},
				},
			},
		},
	}
	return osCount(ctx, alertIndex, query)
}

func (r *osAlertRepo) CountByStatus(ctx context.Context, status int) (int64, error) {
	query := termQuery("status", status)
	return osCount(ctx, alertIndex, query)
}

func (r *osAlertRepo) SearchByIDs(ctx context.Context, alertIDs []string) ([]domain.UtmAlert, error) {
	query := idOnlyFilter(alertIDs)
	raws, err := osSearchSources(ctx, alertIndex, query, 10000)
	if err != nil {
		return nil, err
	}
	alerts := make([]domain.UtmAlert, 0, len(raws))
	for _, raw := range raws {
		var a domain.UtmAlert
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

