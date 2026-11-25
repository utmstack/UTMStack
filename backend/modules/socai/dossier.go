package socai

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/utmstack/go-sdk/v2/store"

	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

// What the dossier carries of an alert. The SOC AI service reads these paths and no others, so the
// message stays small however large the alert is: its events and its history never leave the database.
// The service builds the same dossier from the same fields when it has to ask for it, so keep the two
// lists the same (soc-ai/src/socai/tools.py: ALERT_FIELDS and RELATED_FIELDS).
var (
	alertFields = []string{
		"id", "name", "severity", "status", "category", "technique", "@timestamp", "dataSource", "dataType",
		"description", "adversary.ip", "adversary.host", "adversary.user", "adversary.port", "target.ip",
		"target.host", "target.user", "target.port", "tags", "notes", "impactScore", "echoes", "tenantId", "parentId",
	}
	relatedFields = []string{"id", "name", "severity", "status", "@timestamp", "adversary.ip", "target.ip", "target.host", "tags"}
)

const relatedLimit = 8

// recentAlert is how far back to look for the alert a dossier is about.
const recentAlert = 6 * time.Hour

// dossier is everything needed to judge one alert, as the service expects it: the alert, its score, the
// alerts from the same adversary and the same target in the last 7 days, and how many alerts with its
// name there were in 30 days. A part that cannot be had is left out; the alert itself is required.
func (t *Triage) dossier(ctx context.Context, tenantID, alertID string) (map[string]any, error) {
	now := time.Now()

	// The alert was raised a moment ago. Saying so lets the store skip the old months: it is keyed by name
	// and time, so an id alone makes it read everything the tenant has.
	recent := store.Scope{Tenant: tenantID, Dataset: eventstore.DatasetAlerts, From: now.Add(-recentAlert), To: now.Add(time.Minute)}

	raw, err := t.docs.FetchByID(ctx, recent, alertID)
	if err != nil {
		return nil, err
	}

	var alert map[string]any
	if err := json.Unmarshal(raw, &alert); err != nil {
		return nil, err
	}

	if alert["id"] == nil {
		return nil, errors.New("the alert came back without an id")
	}

	out := map[string]any{"alert": project(alert, alertFields)}

	if t.score != nil {
		if score, err := t.score(ctx, alertID); err == nil && score != nil {
			out["score"] = score
		}
	}

	week := store.Scope{Tenant: tenantID, Dataset: eventstore.DatasetAlerts, From: now.Add(-7 * 24 * time.Hour), To: now}

	for label, field := range map[string]string{"same_adversary": "adversary.ip", "same_target": "target.ip"} {
		ip, _ := path(alert, field).(string)
		if ip == "" {
			continue
		}

		docs, err := t.docs.FetchN(ctx, week, []store.Filter{{Field: field, Op: store.OpEq, Value: ip}}, relatedLimit+1)
		if err != nil {
			continue
		}

		related := make([]map[string]any, 0, relatedLimit)

		for _, d := range docs {
			var m map[string]any
			if json.Unmarshal(d, &m) != nil || m["id"] == alertID {
				continue
			}

			if len(related) < relatedLimit {
				related = append(related, project(m, relatedFields))
			}
		}

		out[label] = related
	}

	if name, _ := alert["name"].(string); name != "" {
		month := store.Scope{Tenant: tenantID, Dataset: eventstore.DatasetAlerts, From: now.Add(-30 * 24 * time.Hour), To: now}

		if n, err := t.docs.Count(ctx, month, []store.Filter{{Field: "name", Op: store.OpEq, Value: name}}); err == nil {
			out["same_name_30d"] = n
		}
	}

	return out, nil
}

func path(doc map[string]any, dotted string) any {
	var cur any = doc

	for _, part := range strings.Split(dotted, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}

		if cur = m[part]; cur == nil {
			return nil
		}
	}

	return cur
}

func empty(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	}

	return false
}

// project keeps only the dotted paths, in the nesting the document has.
func project(doc map[string]any, paths []string) map[string]any {
	out := map[string]any{}

	for _, p := range paths {
		v := path(doc, p)
		if empty(v) {
			continue
		}

		parts := strings.Split(p, ".")
		node := out

		for _, part := range parts[:len(parts)-1] {
			next, ok := node[part].(map[string]any)
			if !ok {
				next = map[string]any{}
				node[part] = next
			}

			node = next
		}

		node[parts[len(parts)-1]] = v
	}

	return out
}
