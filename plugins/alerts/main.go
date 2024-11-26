package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	go_sdk_os "github.com/threatwinds/go-sdk/opensearch"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type correlationServer struct {
	go_sdk.UnimplementedCorrelationServer
}

type Config struct {
	Elasticsearch string `yaml:"elasticsearch"`
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
	Impact            *go_sdk.Impact `json:"impact"`
	ImpactScore       int32          `json:"impactScore"`
	DataSource        string         `json:"dataSource"`
	Adversary         *go_sdk.Side   `json:"adversary"`
	Target            *go_sdk.Side   `json:"target"`
	Logs              []string       `json:"logs"`
	Tags              []string       `json:"tags"`
	Notes             string         `json:"notes"`
	TagRulesApplied   []int          `json:"tagRulesApplied"`
}

func init() {
	pCfg, e := go_sdk.PluginCfg[Config]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	go_sdk_os.Connect([]string{pCfg.Elasticsearch})
}

func main() {
	os.Remove(path.Join(go_sdk.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.alerts_correlation.sock"))

	laddr, err := net.ResolveUnixAddr(
		"unix", path.Join(go_sdk.GetCfg().Env.Workdir,
			"sockets", "com.utmstack.alerts_correlation.sock"))

	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	go_sdk.RegisterCorrelationServer(grpcServer, &correlationServer{})

	if err := grpcServer.Serve(listener); err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}
}

func (p *correlationServer) Correlate(ctx context.Context,
	alert *go_sdk.Alert) (*emptypb.Empty, error) {

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
		Logs:          alert.Events,
		Impact:        alert.Impact,
		ImpactScore:   alert.ImpactScore,
	}

	cancelableContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := go_sdk_os.IndexDoc(cancelableContext, a, indexBuilder("v11-alert", time.Now().UTC()), uuid.NewString())

	return nil, err
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
