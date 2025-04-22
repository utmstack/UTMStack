package main

import (
	"context"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/opensearch"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"net"
	"os"
	"time"

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
	ParentID          string           `json:"parentId"`
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
	ImpactScore       int32            `json:"impactScore"`
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
	socketsFolder, err := utils.MkdirJoin(plugins.WorkDir, "sockets")
	if err != nil {
		_ = catcher.Error("cannot create socket directory", err, nil)
		os.Exit(1)
	}

	socketFile := socketsFolder.FileJoin("com.utmstack.alerts_correlation.sock")
	_ = os.Remove(socketFile)

	unixAddress, err := net.ResolveUnixAddr("unix", socketFile)
	if err != nil {
		_ = catcher.Error("cannot resolve unix address", err, nil)
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", unixAddress)
	if err != nil {
		_ = catcher.Error("cannot listen to unix socket", err, nil)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	plugins.RegisterCorrelationServer(grpcServer, &correlationServer{})

	osUrl := plugins.PluginCfg("com.utmstack", false).Get("opensearch").String()
	err = opensearch.Connect([]string{osUrl})
	if err != nil {
		_ = catcher.Error("cannot connect to OpenSearch", err, nil)
		os.Exit(1)
	}

	if err := grpcServer.Serve(listener); err != nil {
		_ = catcher.Error("cannot serve grpc", err, nil)
		os.Exit(1)
	}
}

func (p *correlationServer) Correlate(_ context.Context,
	alert *plugins.Alert) (*emptypb.Empty, error) {

	parentId := getPreviousAlertId(alert)

	return nil, newAlert(alert, parentId)
}

func getPreviousAlertId(alert *plugins.Alert) string {
	alertString, err := utils.ToString(alert)
	if err != nil {
		_ = catcher.Error("cannot convert alert to string", err, map[string]any{"alert": alert.Name})
		return ""
	}

	var filters []opensearch.Query

	filters = append(filters, opensearch.Query{
		Term: map[string]map[string]interface{}{
			"name.keyword": {
				"value": alert.Name,
			},
		},
	})

	for _, d := range alert.DeduplicateBy {
		value := gjson.Get(*alertString, d)
		if value.Type == gjson.Null {
			continue
		}

		if value.Type == gjson.String {
			filters = append(filters, opensearch.Query{
				Term: map[string]map[string]interface{}{
					d: {
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
				Filter: filters,
			},
		},
		StoredFields: []string{"*"},
		Source:       &opensearch.Source{Excludes: []string{}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	hits, err := searchQuery.SearchIn(ctx, []string{opensearch.BuildIndexPattern("v11", "alert")})
	if err != nil {
		_ = catcher.Error("cannot search for previous alerts", err, map[string]any{"alert": alert.Name})
		return ""
	}

	if hits.Hits.Total.Value != 0 {
		return hits.Hits.Hits[0].ID
	}

	return ""
}

func newAlert(alert *plugins.Alert, parentId string) error {
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

	cancelableContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := opensearch.IndexDoc(cancelableContext, a, opensearch.BuildCurrentIndex("v11", "alert"), alert.Id)
	if err != nil {
		return catcher.Error("cannot index document", err, map[string]any{
			"alert": alert.Name,
		})
	}

	return nil
}
