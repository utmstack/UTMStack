package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/opensearch"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type correlationServer struct {
	plugins.UnimplementedCorrelationServer
}

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
}

func main() {
	// Recover from panics to ensure the main function doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in alerts main function", nil, map[string]any{
				"panic": r,
			})
			// Restart the main function after a brief delay
			time.Sleep(5 * time.Second)
			go main()
		}
	}()

	// Initialize with retry logic instead of exiting
	var socketsFolder utils.Folder
	var err error
	var socketFile string
	var unixAddress *net.UnixAddr
	var listener *net.UnixListener

	// Retry loop for initialization
	for {
		socketsFolder, err = utils.MkdirJoin(plugins.WorkDir, "sockets")
		if err != nil {
			_ = catcher.Error("cannot create socket directory", err, nil)
			time.Sleep(5 * time.Second)
			continue
		}

		socketFile = socketsFolder.FileJoin("com.utmstack.alerts_correlation.sock")
		_ = os.Remove(socketFile)

		unixAddress, err = net.ResolveUnixAddr("unix", socketFile)
		if err != nil {
			_ = catcher.Error("cannot resolve unix address", err, nil)
			time.Sleep(5 * time.Second)
			continue
		}

		listener, err = net.ListenUnix("unix", unixAddress)
		if err != nil {
			_ = catcher.Error("cannot listen to unix socket", err, nil)
			time.Sleep(5 * time.Second)
			continue
		}

		// If we got here, initialization was successful
		break
	}

	grpcServer := grpc.NewServer()
	plugins.RegisterCorrelationServer(grpcServer, &correlationServer{})

	// Connect to OpenSearch with retry logic
	for {
		osUrl := plugins.PluginCfg("com.utmstack", false).Get("opensearch").String()
		err = opensearch.Connect([]string{osUrl})
		if err != nil {
			_ = catcher.Error("cannot connect to OpenSearch", err, nil)
			time.Sleep(5 * time.Second)
			continue
		}
		// If we got here, connection was successful
		break
	}

	// Serve with error handling
	if err := grpcServer.Serve(listener); err != nil {
		_ = catcher.Error("cannot serve grpc", err, nil)
		// Instead of exiting, restart the main function
		time.Sleep(5 * time.Second)
		go main()
		return
	}
}

func (p *correlationServer) Correlate(_ context.Context,
	alert *plugins.Alert) (*emptypb.Empty, error) {
	// Recover from panics to ensure the method doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in Correlate method", nil, map[string]any{
				"panic": r,
				"alert": alert.Name,
			})
		}
	}()

	parentId := getPreviousAlertId(alert)

	return nil, newAlert(alert, parentId)
}

func getPreviousAlertId(alert *plugins.Alert) *string {
	// Recover from panics to ensure the function doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in getPreviousAlertId", nil, map[string]any{
				"panic": r,
				"alert": alert.Name,
			})
		}
	}()

	if len(alert.DeduplicateBy) == 0 {
		return nil
	}

	alertString, err := utils.ToString(alert)
	if err != nil {
		_ = catcher.Error("cannot convert alert to string", err, map[string]any{"alert": alert.Name})
		return nil
	}

	var filters []opensearch.Query
	var mustNot []opensearch.Query

	filters = append(filters, opensearch.Query{
		Term: map[string]map[string]interface{}{
			"name.keyword": {
				"value": alert.Name,
			},
		},
	})

	mustNot = append(mustNot, opensearch.Query{
		Exists: map[string]string{
			"field": "parentId",
		},
	})

	for _, d := range alert.DeduplicateBy {
		d = strings.TrimSuffix(d, ".keyword")

		value := gjson.Get(*alertString, d)
		if value.Type == gjson.Null {
			continue
		}

		if value.Type == gjson.String {
			filters = append(filters, opensearch.Query{
				Term: map[string]map[string]interface{}{
					fmt.Sprintf("%s.keyword", d): {
						"value": value.String(),
					},
				},
			})
		}

		if value.Type == gjson.Number {
			filters = append(filters, opensearch.Query{
				Term: map[string]map[string]interface{}{
					d: {
						"value": value.Float(),
					},
				},
			})
		}

		if value.IsBool() {
			filters = append(filters, opensearch.Query{
				Term: map[string]map[string]interface{}{
					d: {
						"value": value.Bool(),
					},
				},
			})
		}
	}

	searchQuery := opensearch.SearchRequest{
		Size:    1,
		From:    0,
		Version: true,
		Query: &opensearch.Query{
			Bool: &opensearch.Bool{
				Filter:  filters,
				MustNot: mustNot,
			},
		},
		StoredFields: []string{"*"},
		Source:       &opensearch.Source{Excludes: []string{}},
	}

	// Retry logic for search operation
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		hits, err := searchQuery.SearchIn(ctx, []string{opensearch.BuildIndexPattern("v11", "alert")})
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
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	// If we get here, all retries failed
	_ = catcher.Error("all retries failed when searching for previous alerts", nil, map[string]any{
		"alert": alert.Name,
	})
	return nil
}

func newAlert(alert *plugins.Alert, parentId *string) error {
	// Recover from panics to ensure the function doesn't terminate
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in newAlert", nil, map[string]any{
				"panic": r,
				"alert": alert.Name,
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
	}

	// Retry logic for indexing operation
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		cancelableContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		err := opensearch.IndexDoc(cancelableContext, a, opensearch.BuildCurrentIndex("v11", "alert"), alert.Id)
		if err == nil {
			cancel()
			return nil
		}
		cancel()

		_ = catcher.Error("cannot index document, retrying", err, map[string]any{
			"alert":      alert.Name,
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, return the error
			return catcher.Error("all retries failed when indexing document", err, map[string]any{
				"alert": alert.Name,
			})
		}
	}

	// This should never be reached, but just in case
	return nil
}

func updateParentAlertToOpen(parentHit opensearch.Hit) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in updateParentAlertToOpen", nil, map[string]any{
				"panic":    r,
				"parentId": parentHit.ID,
			})
		}
	}()

	var parentAlert AlertFields
	err := parentHit.Source.ParseSource(&parentAlert)
	if err != nil {
		_ = catcher.Error("cannot parse parent alert source", err, map[string]any{
			"parentId": parentHit.ID,
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
		})
	}
}
