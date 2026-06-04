package repository

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
)

// appendHistoryScript appends each alert's pending change-history entries to
// its embedded `history` array. Entries are keyed by alert id in params.byId so
// a single update_by_query can fan out to many alerts at once: each matched
// document looks itself up by `id` and appends only the entries meant for it.
const appendHistoryScript = `
String aid = ctx._source.id;
if (aid != null && params.byId.containsKey(aid)) {
  if (ctx._source.history == null) { ctx._source.history = []; }
  ctx._source.history.addAll(params.byId.get(aid));
}
`

// osHistoryRecorder appends change-history entries to the alert documents in
// OpenSearch, replacing the former Postgres utm_alert_log table.
type osHistoryRecorder struct {
	client *ospkg.Client
}

func NewHistoryRecorder(client *ospkg.Client) connectors.HistoryRecorder {
	return &osHistoryRecorder{client: client}
}

func (r *osHistoryRecorder) Record(ctx context.Context, entries []connectors.HistoryEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Group entries by alert id; each becomes one element of the doc's log array.
	byID := make(map[string][]map[string]any, len(entries))
	for _, e := range entries {
		at := e.At
		if at.IsZero() {
			at = time.Now().UTC()
		}
		entry := map[string]any{
			"user":      e.User,
			"action":    string(e.Action),
			"message":   e.Message,
			"newValue":  e.NewValue,
			"timestamp": at.UTC().Format(time.RFC3339),
		}
		byID[e.AlertID] = append(byID[e.AlertID], entry)
	}

	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}

	script := ospkg.Script{
		Source: appendHistoryScript,
		Params: map[string]any{"byId": byID},
	}
	return r.client.UpdateByQuery(ctx, alertIndex, idOnlyFilter(ids), script)
}
