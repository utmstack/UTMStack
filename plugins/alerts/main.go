package main

import (
	"context"
	"fmt"
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
	Timestamp         string           `json:"@timestamp"`
	ID                string           `json:"id"`
	ParentID          *string          `json:"parentId,omitempty"`
	Status            int              `json:"status"`
	StatusLabel       string           `json:"statusLabel"`
	StatusObservation string           `json:"statusObservation"`
	IsIncident        bool             `json:"isIncident"`
	IncidentDetail    IncidentDetail   `json:"incidentDetail"`
	Name              string           `json:"name"`
	Category          string           `json:"category"`
	Severity          int              `json:"severity"`
	SeverityLabel     string           `json:"severityLabel"`
	Description       string           `json:"description"`
	Solution          string           `json:"solution"`
	Technique         string           `json:"technique"`
	Reference         []string         `json:"reference"`
	DataType          string           `json:"dataType"`
	Impact            *plugins.Impact  `json:"impact"`
	ImpactScore       uint32           `json:"impactScore"`
	DataSource        string           `json:"dataSource"`
	Adversary         *plugins.Side    `json:"adversary"`
	Target            *plugins.Side    `json:"target"`
	Events            []*plugins.Event `json:"events"`
	LastEvent         *plugins.Event   `json:"lastEvent"`
	Tags              []string         `json:"tags"`
	Notes             string           `json:"notes"`
	TagRulesApplied   []int            `json:"tagRulesApplied"`
	DeduplicatedBy    []string         `json:"deduplicatedBy"`
	GroupedBy         []string         `json:"groupedBy"`
}

func main() {
	openSearchUrl := plugins.PluginCfg("org.opensearch", false).Get("opensearch").String()
	err := sdkos.Connect([]string{openSearchUrl}, "", "")
	if err != nil {
		_ = catcher.Error("cannot connect to OpenSearch", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
		os.Exit(1)
	}

	err = plugins.InitCorrelationPlugin("com.utmstack.alerts", correlate)
	if err != nil {
		_ = catcher.Error("com.utmstack.alerts", err, map[string]any{
			"process": "plugin_com.utmstack.alerts",
		})
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

	parentId := getPreviousAlertId(alert)

	if parentId != nil {
		if isDuplicate(alert) {
			return nil, nil
		}
		return nil, newAlert(alert, parentId)
	}

	if len(alert.DeduplicateBy) > 0 {
		if isDuplicate(alert) {
			return nil, nil
		}
	}

	return nil, newAlert(alert, nil)
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
	bb.FilterTerm("name.keyword", alert.Name)

	// Compile regex for array index stripping
	reArrayIndex := regexp.MustCompile(`\.[0-9]+(\.|$)`)

	for _, d := range alert.DeduplicateBy {
		d = strings.TrimSuffix(d, ".keyword")

		value := gjson.Get(*alertString, d)
		if value.Type == gjson.Null {
			continue
		}

		// Calculate OpenSearch field name by removing array indices
		searchField := reArrayIndex.ReplaceAllStringFunc(d, func(s string) string {
			if strings.HasSuffix(s, ".") {
				return "."
			}
			return ""
		})

		if value.Type == gjson.String {
			bb.FilterTerm(fmt.Sprintf("%s.keyword", searchField), value.String())
		} else if value.Type == gjson.Number {
			bb.FilterTerm(searchField, value.Float())
		} else if value.IsBool() {
			bb.FilterTerm(searchField, value.Bool())
		}
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

	searchFields := alert.GroupBy
	if len(searchFields) == 0 {
		searchFields = alert.DeduplicateBy
	}

	if len(searchFields) == 0 {
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
	bb.FilterTerm("name.keyword", alert.Name)

	// 2. Must NOT match existing ParentId (we want strictly the parent, or another orphan, not a child)
	// Original logic: MustNot exists field "parentId"
	bb.MustNotExists("parentId")

	// Compile regex for array index stripping
	reArrayIndex := regexp.MustCompile(`\.[0-9]+(\.|$)`)

	for _, d := range searchFields {
		d = strings.TrimSuffix(d, ".keyword")

		value := gjson.Get(*alertString, d)
		if value.Type == gjson.Null {
			continue
		}

		// Calculate OpenSearch field name by removing array indices
		searchField := reArrayIndex.ReplaceAllStringFunc(d, func(s string) string {
			if strings.HasSuffix(s, ".") {
				return "."
			}
			return ""
		})

		if value.Type == gjson.String {
			bb.FilterTerm(fmt.Sprintf("%s.keyword", searchField), value.String())
		} else if value.Type == gjson.Number {
			bb.FilterTerm(searchField, value.Float())
		} else if value.IsBool() {
			bb.FilterTerm(searchField, value.Bool())
		}
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
		ID:            alert.Id,
		ParentID:      parentId,
		Status:        1,
		StatusLabel:   "Automatic review",
		Name:          alert.Name,
		Category:      alert.Category,
		Severity:      severityN,
		SeverityLabel: severityLabel,
		Description:   alert.Description,
		Technique:     alert.Technique,
		Reference:     alert.References,
		DataType:      alert.DataType,
		DataSource:    alert.DataSource,
		Adversary:     alert.Adversary,
		Target:        alert.Target,
		LastEvent: func() *plugins.Event {
			l := len(alert.Events)
			if l == 0 {
				return nil
			}
			return alert.Events[l-1]
		}(),
		Events:         alert.Events,
		Impact:         alert.Impact,
		ImpactScore:    alert.ImpactScore,
		DeduplicatedBy: alert.DeduplicateBy,
		GroupedBy:      alert.GroupBy,
	}

	// Retry logic for indexing operation
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		cancelableContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		err := sdkos.IndexDoc(cancelableContext, a, sdkos.BuildCurrentIndex("v11", "alert"), alert.Id)
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
