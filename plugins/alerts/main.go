package main

import (
	"context"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"

	"google.golang.org/protobuf/types/known/emptypb"
)

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	CreationDate string `json:"creationDate"`
	Source       string `json:"source"`
}

type AlertFields struct {
	Timestamp         string         `json:"@timestamp"`
	Status            int            `json:"status"`
	StatusLabel       string         `json:"statusLabel"`
	StatusObservation string         `json:"statusObservation"`
	IsIncident        bool           `json:"isIncident"`
	IncidentDetail    IncidentDetail `json:"incidentDetail"`
	Severity          int            `json:"severity"`
	SeverityLabel     string         `json:"severityLabel"`
	Solution          string         `json:"solution"`
	Reference         []string       `json:"reference"`
	LastEvent         *plugins.Event `json:"lastEvent"`
	Tags              []string       `json:"tags"`
	Notes             string         `json:"notes"`
	TagRulesApplied   []int          `json:"tagRulesApplied"`
	DeduplicatedBy    []string       `json:"deduplicatedBy"`
	GroupedBy         []string       `json:"groupedBy"`
	plugins.Alert
}

// rules holds the active tag-rule snapshot the plugin evaluates against each
// freshly indexed alert. It is refreshed in the background; a missing or empty
// snapshot just means no rules fire — release-to-Open still runs.
var rules *ruleCache

func main() {
	cfg := plugins.PluginCfg("org.opensearch").Get("opensearch")
	host := cfg.Get("host").String()
	port := cfg.Get("port").String()
	user := cfg.Get("user").String()
	password := cfg.Get("password").String()
	openSearchUrl := "https://" + host + ":" + port

	err := sdkos.Connect([]string{openSearchUrl}, user, password)
	if err != nil {
		_ = catcher.Error("cannot connect to OpenSearch", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

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

	if isDuplicate(alert) {
		return nil, nil
	}

	parentId := getPreviousAlertId(alert)

	if err := newAlert(alert, parentId); err != nil {
		return nil, err
	}

	// Tag-rule evaluation runs detached from the SDK ack so a slow OpenSearch
	// hop never stalls intake. Errors are logged inside applyTagRulesAndRelease.
	if rules != nil {
		go applyTagRulesAndRelease(context.Background(), alert.Id, rules.Snapshot())
	}

	return nil, nil
}

func isDuplicate(alert *plugins.Alert) bool {
	// Recover from panics to ensure the function doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in isDuplicate", nil, map[string]any{
				"panic":   r,
				"alert":   alert.Name,
				"process": "plugin_com.utmstack.alerts",
			})
		}
	}()

	if len(alert.DeduplicateBy) == 0 {
		return false
	}

	alertString, err := utils.ProtoMessageToString(alert)
	if err != nil {
		_ = catcher.Error("cannot convert alert to string", err, map[string]any{"alert": alert.Name, "process": "plugin_com.utmstack.alerts"})
		return false
	}

	ctx := context.Background()
	indices := []string{sdkos.BuildIndexPattern("v11", "alert")}

	// Create BoolBuilder
	bb := sdkos.NewBoolBuilder(ctx, indices, "plugin_com.utmstack.alerts")

	// 1. Filter by Name (always)
	bb.FilterTerm("name", alert.Name)

	bb.FilterRange("@timestamp", "gte", time.Now().UTC().Add(-24*7*time.Hour).Format(time.RFC3339Nano))
	bb.FilterRange("@timestamp", "lte", time.Now().UTC().Format(time.RFC3339Nano))

	// Compile regex for array index stripping
	reArrayIndex := regexp.MustCompile(`\.[0-9]+(\.|$)`)

	var execute bool = false

	for _, d := range alert.DeduplicateBy {
		d = strings.TrimSuffix(d, ".keyword")

		value := gjson.Get(*alertString, d)
		if value.Type == gjson.Null {
			continue
		}

		execute = true

		// Calculate OpenSearch field name by removing array indices
		searchField := reArrayIndex.ReplaceAllStringFunc(d, func(s string) string {
			if strings.HasSuffix(s, ".") {
				return "."
			}
			return ""
		})

		if value.Type == gjson.String {
			bb.FilterTerm(searchField, value.String())
		} else if value.Type == gjson.Number {
			bb.FilterTerm(searchField, value.Float())
		} else if value.IsBool() {
			bb.FilterTerm(searchField, value.Bool())
		}
	}

	if !execute {
		return false
	}

	// Create QueryBuilder and inject the Bool query
	qb := sdkos.NewQueryBuilder(ctx, indices, "plugin_com.utmstack.alerts")
	qb.Size(1)
	qb.From(0)
	qb.IncludeSource("id")

	qb.Filter(bb.Build())

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchRequest := qb.Build()

	// Ensure latest data is visible
	_ = sdkos.RefreshIndex(ctxTimeout, indices[0])

	hits, err := searchRequest.WideSearchIn(ctxTimeout, indices)

	if err == nil && hits.Hits.Total.Value != 0 {
		return true
	}

	return false
}

func getPreviousAlertId(alert *plugins.Alert) *string {
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

	alertString, err := utils.ProtoMessageToString(alert)
	if err != nil {
		_ = catcher.Error("cannot convert alert to string", err, map[string]any{"alert": alert.Name, "process": "plugin_com.utmstack.alerts"})
		return nil
	}

	ctx := context.Background()
	indices := []string{sdkos.BuildIndexPattern("v11", "alert")}

	// Create BoolBuilder
	bb := sdkos.NewBoolBuilder(ctx, indices, "plugin_com.utmstack.alerts")

	// 1. Filter by Name (always)
	bb.FilterTerm("name", alert.Name)

	// 2. Must NOT match existing ParentId (we want strictly the parent, or another orphan, not a child)
	// Original logic: MustNot exists field "parentId"
	bb.MustNotExists("parentId")

	// Compile regex for array index stripping
	reArrayIndex := regexp.MustCompile(`\.[0-9]+(\.|$)`)

	var execute bool = false

	for _, d := range alert.GroupBy {
		d = strings.TrimSuffix(d, ".keyword")

		value := gjson.Get(*alertString, d)
		if value.Type == gjson.Null {
			continue
		}

		execute = true

		// Calculate OpenSearch field name by removing array indices
		searchField := reArrayIndex.ReplaceAllStringFunc(d, func(s string) string {
			if strings.HasSuffix(s, ".") {
				return "."
			}
			return ""
		})

		if value.Type == gjson.String {
			bb.FilterTerm(searchField, value.String())
		} else if value.Type == gjson.Number {
			bb.FilterTerm(searchField, value.Float())
		} else if value.IsBool() {
			bb.FilterTerm(searchField, value.Bool())
		}
	}

	if !execute {
		return nil
	}

	// Create QueryBuilder and inject the Bool query
	qb := sdkos.NewQueryBuilder(ctx, indices, "plugin_com.utmstack.alerts")
	qb.Size(1)
	qb.From(0)
	qb.Version(true)
	qb.IncludeSource("*") // Previously StoredFields("*")

	// We use Filter(...) method of QueryBuilder which takes varargs of Query.
	// bb.Build() returns a Query struct that wraps the Bool query.
	// Since we built a full Bool query with Filter/MustNot clauses inside bb,
	// we just need to add this whole Bool query to the QueryBuilder.
	// qb wraps everything in its own top-level Bool query.
	// So we can add our 'bb' as a Must or Filter clause of the top-level query.
	// Since 'bb' contains the logic "Match THIS AND THAT AND NOT THIS", it should be a Must/Filter clause.
	qb.Filter(bb.Build())

	// Retry logic for search operation
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		searchRequest := qb.Build()
		hits, err := searchRequest.WideSearchIn(ctxTimeout, indices)
		cancel()

		if err == nil {
			if hits.Hits.Total.Value != 0 {
				go updateParentAlertToOpen(hits.Hits.Hits[0])
				return utils.PointerOf(hits.Hits.Hits[0].ID)
			}
			return nil
		}

		_ = catcher.Error("cannot search for previous alerts, retrying", err, map[string]any{
			"alert":      alert.Name,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"process":    "plugin_com.utmstack.alerts",
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	// If we get here, all retries failed
	_ = catcher.Error("all retries failed when searching for previous alerts", nil, map[string]any{
		"alert":   alert.Name,
		"process": "plugin_com.utmstack.alerts",
	})
	return nil
}

func newAlert(alert *plugins.Alert, parentId *string) error {
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

	var severityN int
	var severityLabel string
	switch alert.Severity {
	case "low":
		severityN = 1
		severityLabel = "Low"
	case "medium":
		severityN = 2
		severityLabel = "Medium"
	case "high":
		severityN = 3
		severityLabel = "High"
	default:
		severityN = 1
		severityLabel = "Low"
	}

	a := AlertFields{
		Timestamp:     alert.Timestamp,
		Status:        1,
		StatusLabel:   "Automatic review",
		Severity:      severityN,
		SeverityLabel: severityLabel,
		Reference:     alert.References,
		LastEvent: func() *plugins.Event {
			l := len(alert.Events)
			if l == 0 {
				return nil
			}
			return alert.Events[l-1]
		}(),
		DeduplicatedBy: alert.DeduplicateBy,
		GroupedBy:      alert.GroupBy,
	}

	a.Id = alert.Id
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
	a.Adversary = alert.Adversary
	a.Target = alert.Target
	a.Events = alert.Events
	a.Impact = alert.Impact
	a.ImpactScore = alert.ImpactScore
	a.Errors = alert.Errors

	// Retry logic for indexing operation
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		cancelableContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		err := sdkos.IndexDoc(cancelableContext, a, sdkos.BuildCurrentDayIndex("v11", "alert"), alert.Id)
		if err == nil {
			cancel()
			return nil
		}
		cancel()

		_ = catcher.Error("cannot index document, retrying", err, map[string]any{
			"alert":      alert.Name,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"process":    "plugin_com.utmstack.alerts",
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, return the error
			return catcher.Error("all retries failed when indexing document", err, map[string]any{
				"alert":   alert.Name,
				"process": "plugin_com.utmstack.alerts",
			})
		}
	}

	// This should never be reached, but just in case
	return nil
}

func updateParentAlertToOpen(parentHit sdkos.Hit) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in updateParentAlertToOpen", nil, map[string]any{
				"panic":    r,
				"parentId": parentHit.ID,
				"process":  "plugin_com.utmstack.alerts",
			})
		}
	}()

	var parentAlert AlertFields
	err := parentHit.Source.ParseSource(&parentAlert)
	if err != nil {
		_ = catcher.Error("cannot parse parent alert source", err, map[string]any{
			"parentId": parentHit.ID,
			"process":  "plugin_com.utmstack.alerts",
		})
		return
	}

	// Only update if it is Completed status
	if parentAlert.Status == 5 {
		parentAlert.Status = 2
		parentAlert.StatusLabel = "Open"

		err := parentHit.Source.SetSource(parentAlert)
		if err != nil {
			_ = catcher.Error("cannot set updated parent alert source", err, map[string]any{
				"parentId": parentHit.ID,
				"process":  "plugin_com.utmstack.alerts",
			})
			return
		}

		maxRetries := 3
		retryDelay := 2 * time.Second

		for retry := 0; retry < maxRetries; retry++ {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			err := parentHit.Save(ctx)
			cancel()

			if err != nil {
				_ = catcher.Error("failed to update parent alert to Open, retrying", err, map[string]any{
					"parentId":   parentHit.ID,
					"retry":      retry + 1,
					"maxRetries": maxRetries,
					"process":    "alerts-plugin",
				})

				if retry < maxRetries-1 {
					time.Sleep(retryDelay)
					retryDelay *= 2
				}
				continue
			}

			return
		}

		_ = catcher.Error("all retries failed when updating parent alert to Open", nil, map[string]any{
			"parentId": parentHit.ID,
			"process":  "alerts-plugin",
		})
	}
}
