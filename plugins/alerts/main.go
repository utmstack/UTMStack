package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store"
	ch "github.com/threatwinds/go-sdk/store/clickhouse"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"

	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	processName = "plugin_com.utmstack.alerts"

	datasetAlerts      = store.Dataset("alerts")
	defaultAlertsTable = "alerts"

	dedupWindow = 7 * 24 * time.Hour

	searchTimeout = 10 * time.Second
	writeTimeout  = 10 * time.Second

	maxRetries        = 3
	initialRetryDelay = 2 * time.Second
)

var alertStore *ch.Driver

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	CreationDate string `json:"creationDate"`
	Source       string `json:"source"`
}

type AlertFields struct {
	Timestamp         string         `json:"@timestamp"`
	Status            string         `json:"status"`
	StatusObservation string         `json:"statusObservation"`
	IsIncident        bool           `json:"isIncident"`
	IncidentDetail    IncidentDetail `json:"incidentDetail"`
	Solution          string         `json:"solution"`
	Tags              []string       `json:"tags,omitempty"`
	Notes             string         `json:"notes"`
	TagRulesApplied   []string       `json:"tagRulesApplied,omitempty"`
	plugins.Alert
}

var rules *ruleCache

func main() {
	cfg := plugins.PluginCfg("clickhouse")
	table := cfg.Get("alertsTable").String()
	if table == "" {
		table = defaultAlertsTable
	}

	var err error
	alertStore, err = ch.New(ch.Config{
		Addr:     []string{cfg.Get("host").String() + ":" + cfg.Get("port").String()},
		Database: cfg.Get("database").String(),
		Username: cfg.Get("user").String(),
		Password: cfg.Get("password").String(),
		Tables: map[store.Dataset]string{
			datasetAlerts: table,
		},
		TenantColumn:   "tenantId",
		TimeColumn:     "@timestamp",
		DataTypeColumn: "dataType",
	})
	if err != nil {
		_ = catcher.Error("cannot build the alert store", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	defer alertStore.Close()

	rules = newRuleCache()
	initialCtx, initialCancel := context.WithTimeout(context.Background(), rulesRequestTimeout)
	if err := rules.Refresh(initialCtx); err != nil {
		// First-pass failure is non-fatal — the background loop keeps trying.
		_ = catcher.Error("initial tag-rule cache refresh failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}
	initialCancel()
	go rules.Run(context.Background())

	err = plugins.InitCorrelationPlugin("com.utmstack.alerts", correlate)
	if err != nil {
		_ = catcher.Error("com.utmstack.alerts", err, map[string]any{
			"process": "plugin_com.utmstack.alerts",
		})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func correlate(ctx context.Context,
	alert *plugins.Alert) (*emptypb.Empty, error) {
	// Recover from panics to ensure the method doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in Correlate method", nil, map[string]any{
				"panic":   r,
				"alert":   alert.Name,
				"process": "plugin_com.utmstack.alerts",
			})
		}
	}()

	if alert.TenantId == "" {
		_ = catcher.Error("dropping an alert with no tenant", nil, map[string]any{
			"alert":   alert.Name,
			"process": processName,
		})
		return nil, nil
	}

	if alert.LastEvent == nil && len(alert.Events) > 0 {
		alert.LastEvent = alert.Events[len(alert.Events)-1]
	}

	raw, err := utils.ProtoMessageToString(alert)
	if err != nil {
		_ = catcher.Error("cannot convert alert to string", err, map[string]any{"alert": alert.Name, "process": processName})
		return nil, nil
	}
	alertJSON := *raw

	if isDuplicate(alert, alertJSON) {
		return nil, nil
	}

	parentId := getPreviousAlertId(alert, alertJSON)

	var snapshot []RuleSnapshot
	if rules != nil {
		snapshot = rules.Snapshot(alert.TenantId)
	}

	if err := newAlert(alert, alertJSON, parentId, snapshot); err != nil {
		return nil, err
	}

	return nil, nil
}

var reArrayIndex = regexp.MustCompile(`\.[0-9]+(\.|$)`)

func normalizeSeverity(s string) string {
	switch s {
	case "low", "medium", "high":
		return s
	default:
		return "low"
	}
}

func matchFilters(fields []string, alertString *string) []store.Filter {
	out := make([]store.Filter, 0, len(fields))

	for _, d := range fields {
		value := gjson.Get(*alertString, d)
		if value.Type == gjson.Null {
			continue
		}

		field := reArrayIndex.ReplaceAllStringFunc(d, func(s string) string {
			if strings.HasSuffix(s, ".") {
				return "."
			}
			return ""
		})

		switch {
		case value.Type == gjson.String:
			out = append(out, store.Filter{Field: field, Op: store.OpEq, Value: value.String()})
		case value.Type == gjson.Number:
			out = append(out, store.Filter{Field: field, Op: store.OpEq, Value: value.Float()})
		case value.IsBool():
			out = append(out, store.Filter{Field: field, Op: store.OpEq, Value: value.Bool()})
		}
	}

	return out
}

func deduplicationToken(alert *plugins.Alert, alertJSON string) string {
	if len(alert.DeduplicateBy) == 0 {
		return ""
	}

	filters := matchFilters(alert.DeduplicateBy, &alertJSON)
	if len(filters) == 0 {
		return ""
	}

	parts := make([]string, 0, len(filters))
	for _, f := range filters {
		parts = append(parts, fmt.Sprintf("%s=%v", f.Field, f.Value))
	}
	sort.Strings(parts)

	sum := sha256.Sum256([]byte(alert.TenantId + "\x00" + alert.Name + "\x00" + strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}

func isDuplicate(alert *plugins.Alert, alertJSON string) bool {
	// Recover from panics to ensure the function doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in isDuplicate", nil, map[string]any{
				"panic":   r,
				"alert":   alert.Name,
				"process": processName,
			})
		}
	}()

	if len(alert.DeduplicateBy) == 0 {
		return false
	}

	filters := matchFilters(alert.DeduplicateBy, &alertJSON)
	if len(filters) == 0 {
		return false
	}
	filters = append(filters, store.Filter{Field: "name", Op: store.OpEq, Value: alert.Name})

	now := time.Now().UTC()
	scope := store.Scope{
		Tenant:  alert.TenantId,
		Dataset: datasetAlerts,
		From:    now.Add(-dedupWindow),
		To:      now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	n, err := alertStore.Count(ctx, scope, filters)
	if err != nil {
		// Treating a failed lookup as "not a duplicate" keeps a store problem
		// from swallowing alerts; the cost is a possible duplicate.
		_ = catcher.Error("cannot check for duplicates", err, map[string]any{
			"alert":   alert.Name,
			"process": processName,
		})
		return false
	}

	return n > 0
}

func getPreviousAlertId(alert *plugins.Alert, alertJSON string) *string {
	// Recover from panics to ensure the function doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in getPreviousAlertId", nil, map[string]any{
				"panic":   r,
				"alert":   alert.Name,
				"process": "plugin_com.utmstack.alerts",
			})
		}
	}()

	if len(alert.GroupBy) == 0 {
		return nil
	}

	filters := matchFilters(alert.GroupBy, &alertJSON)
	if len(filters) == 0 {
		return nil
	}
	filters = append(filters,
		store.Filter{Field: "name", Op: store.OpEq, Value: alert.Name},
		store.Filter{Field: "parentId", Op: store.OpEq, Value: ""},
	)

	scope := store.Scope{Tenant: alert.TenantId, Dataset: datasetAlerts}

	retryDelay := initialRetryDelay

	for retry := 0; retry < maxRetries; retry++ {
		ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
		docs, err := alertStore.FetchN(ctx, scope, filters, 1)
		cancel()

		if err == nil {
			if len(docs) == 0 {
				return nil
			}

			id := gjson.GetBytes(docs[0], "id").String()
			if id == "" {
				return nil
			}

			if gjson.GetBytes(docs[0], "status").String() == statusCompleted {
				go updateParentAlertToOpen(alert.TenantId, id)
			}

			return utils.PointerOf(id)
		}

		_ = catcher.Error("cannot search for previous alerts, retrying", err, map[string]any{
			"alert":      alert.Name,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"process":    processName,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	// If we get here, all retries failed
	_ = catcher.Error("all retries failed when searching for previous alerts", nil, map[string]any{
		"alert":   alert.Name,
		"process": processName,
	})
	return nil
}

func applyTagRules(a *AlertFields, rules []RuleSnapshot) {
	a.StatusObservation = tagRuleOpenObservation
	if len(rules) == 0 {
		return
	}

	raw, err := json.Marshal(a)
	if err != nil {
		_ = catcher.Error("tag rules: cannot marshal alert", err, map[string]any{"alert": a.Name, "process": "plugin_com.utmstack.alerts"})
		return
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		_ = catcher.Error("tag rules: cannot decode alert", err, map[string]any{"alert": a.Name, "process": "plugin_com.utmstack.alerts"})
		return
	}

	d := evaluateTagRules(doc, rules)
	a.Tags = d.Tags
	a.TagRulesApplied = d.RuleIDs
	if d.FalsePositive {
		a.Status = statusCompleted
		a.StatusObservation = tagRuleCompletedObservation
	}
}

func buildAlert(alert *plugins.Alert, parentId *string) *AlertFields {
	a := &AlertFields{
		Status: statusOpen,
		Timestamp: func() string {
			if alert.Timestamp != "" {
				return alert.Timestamp
			}
			return time.Now().UTC().Format(time.RFC3339Nano)
		}(),
	}

	a.Id = alert.Id
	a.TenantId = alert.TenantId
	a.TenantName = alert.TenantName
	// The table replaces by this: a row written without it collapses to 1970 and
	// loses to every later version of the same alert.
	a.LastUpdate = a.Timestamp
	a.ParentId = func() string {
		if parentId != nil {
			return *parentId
		}
		return ""
	}()
	a.Name = alert.Name
	a.Category = alert.Category
	a.Description = alert.Description
	a.Technique = alert.Technique
	a.DataSource = alert.DataSource
	a.DataType = alert.DataType
	a.Severity = normalizeSeverity(alert.Severity)
	a.Adversary = alert.Adversary
	a.Target = alert.Target
	a.Events = alert.Events
	a.LastEvent = alert.LastEvent
	a.Impact = alert.Impact
	a.ImpactScore = alert.ImpactScore
	a.References = alert.References
	a.DeduplicateBy = alert.DeduplicateBy
	a.GroupBy = alert.GroupBy
	a.Errors = alert.Errors

	return a
}

func newAlert(alert *plugins.Alert, alertJSON string, parentId *string, ruleSnapshot []RuleSnapshot) error {
	// Recover from panics to ensure the function doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in newAlert", nil, map[string]any{
				"panic":   r,
				"alert":   alert.Name,
				"process": "plugin_com.utmstack.alerts",
			})
		}
	}()

	a := buildAlert(alert, parentId)

	applyTagRules(a, ruleSnapshot)

	scope := store.Scope{Tenant: alert.TenantId, Dataset: datasetAlerts}
	retryDelay := initialRetryDelay
	token := deduplicationToken(alert, alertJSON)

	for retry := 0; retry < maxRetries; retry++ {
		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		if token != "" {
			ctx = ch.WithDeduplicationToken(ctx, token)
		}

		err := alertStore.Insert(ctx, scope, alert.Id, a)
		cancel()

		if err == nil {
			return nil
		}

		_ = catcher.Error("cannot write alert, retrying", err, map[string]any{
			"alert":      alert.Name,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"process":    processName,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, return the error
			return catcher.Error("all retries failed when writing alert", err, map[string]any{
				"alert":   alert.Name,
				"process": processName,
			})
		}
	}

	// This should never be reached, but just in case
	return nil
}

func updateParentAlertToOpen(tenantID, parentID string) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in updateParentAlertToOpen", nil, map[string]any{
				"panic":    r,
				"parentId": parentID,
				"process":  processName,
			})
		}
	}()

	scope := store.Scope{Tenant: tenantID, Dataset: datasetAlerts}
	filters := []store.Filter{{Field: "id", Op: store.OpEq, Value: parentID}}
	patch := map[string]any{"status": statusOpen}

	retryDelay := initialRetryDelay

	for retry := 0; retry < maxRetries; retry++ {
		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		_, err := alertStore.UpdateWhere(ctx, scope, filters, patch)
		cancel()

		if err == nil {
			return
		}

		_ = catcher.Error("failed to update parent alert to Open, retrying", err, map[string]any{
			"parentId":   parentID,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"process":    processName,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	_ = catcher.Error("all retries failed when updating parent alert to Open", nil, map[string]any{
		"parentId": parentID,
		"process":  processName,
	})
}
