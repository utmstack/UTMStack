package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	gosdk "github.com/threatwinds/go-sdk"
	sdkos "github.com/threatwinds/go-sdk/opensearch"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type correlationServer struct {
	gosdk.UnimplementedCorrelationServer
}

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	CreationDate string `json:"creationDate"`
	Source       string `json:"source"`
}

type AlertFields struct {
	Timestamp         string         `json:"@timestamp"`
	ID                string         `json:"id"`
	Status            int            `json:"status"`
	StatusLabel       string         `json:"statusLabel"`
	StatusObservation string         `json:"statusObservation"`
	IsIncident        bool           `json:"isIncident"`
	IncidentDetail    IncidentDetail `json:"incidentDetail"`
	Name              string         `json:"name"`
	Category          string         `json:"category"`
	Severity          int            `json:"severity"`
	SeverityLabel     string         `json:"severityLabel"`
	Description       string         `json:"description"`
	Solution          string         `json:"solution"`
	Technique         string         `json:"technique"`
	Reference         []string       `json:"reference"`
	DataType          string         `json:"dataType"`
	Impact            *gosdk.Impact  `json:"impact"`
	ImpactScore       int32          `json:"impactScore"`
	DataSource        string         `json:"dataSource"`
	Adversary         *gosdk.Side    `json:"adversary"`
	Target            *gosdk.Side    `json:"target"`
	Events            []*gosdk.Event `json:"events"`
	Tags              []string       `json:"tags"`
	Notes             string         `json:"notes"`
	TagRulesApplied   []int          `json:"tagRulesApplied"`
}

func main() {
	_ = os.Remove(path.Join(gosdk.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.alerts_correlation.sock"))

	unixAddress, err := net.ResolveUnixAddr(
		"unix", path.Join(gosdk.GetCfg().Env.Workdir, "sockets", "com.utmstack.alerts_correlation.sock"))
	if err != nil {
		_ = gosdk.Error("cannot resolve unix address", err, nil)
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", unixAddress)
	if err != nil {
		_ = gosdk.Error("cannot listen to unix socket", err, nil)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	gosdk.RegisterCorrelationServer(grpcServer, &correlationServer{})

	elasticUrl := gosdk.PluginCfg("com.utmstack", false).Get("elasticsearch").String()
	err = sdkos.Connect([]string{elasticUrl})
	if err != nil {
		_ = gosdk.Error("cannot connect to ElasticSearch/OpenSearch", err, nil)
		os.Exit(1)
	}

	if err := grpcServer.Serve(listener); err != nil {
		_ = gosdk.Error("cannot serve grpc", err, nil)
		os.Exit(1)
	}
}

func (p *correlationServer) Correlate(_ context.Context,
	alert *gosdk.Alert) (*emptypb.Empty, error) {

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

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
		Timestamp:     timestamp,
		ID:            alert.Id,
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
		Events:        alert.Events,
		Impact:        alert.Impact,
		ImpactScore:   alert.ImpactScore,
	}

	cancelableContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := sdkos.IndexDoc(cancelableContext, a, indexBuilder("v11-alert", time.Now().UTC()), uuid.NewString())
	if err != nil {
		return nil, gosdk.Error("cannot index document", err, map[string]any{
			"alert": alert.Name,
		})
	}

	return nil, nil
}

func indexBuilder(name string, timestamp time.Time) string {
	fst := timestamp.Format("2006.01.02")

	index := fmt.Sprintf(
		"%s-%s",
		name,
		fst,
	)

	return index
}
