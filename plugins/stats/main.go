package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/opensearch"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type notificationServer struct {
	plugins.UnimplementedNotificationServer
}

var statisticsQueue chan map[string]plugins.DataProcessingMessage
var fails map[string]map[string]map[string]map[string]int64
var failsLock sync.Mutex
var success map[string]map[string]int64
var successLock sync.Mutex

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	_ = os.Remove(path.Join(plugins.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.stats_notification.sock"))

	unixAddress, err := net.ResolveUnixAddr(
		"unix", path.Join(plugins.GetCfg().Env.Workdir,
			"sockets", "com.utmstack.stats_notification.sock"))

	if err != nil {
		_ = catcher.Error("cannot resolve unix address", err, nil)
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", unixAddress)
	if err != nil {
		_ = catcher.Error("cannot listen to unix socket", err, nil)
		os.Exit(1)
	}

	statisticsQueue = make(chan map[string]plugins.DataProcessingMessage, runtime.NumCPU()*100)
	fails = make(map[string]map[string]map[string]map[string]int64)
	success = make(map[string]map[string]int64)

	grpcServer := grpc.NewServer()
	plugins.RegisterNotificationServer(grpcServer, &notificationServer{})

	pCfg := plugins.PluginCfg("com.utmstack", false)
	elasticUrl := pCfg.Get("elasticsearch").String()

	if err := opensearch.Connect([]string{elasticUrl}); err != nil {
		_ = catcher.Error("cannot connect to ElasticSearch/OpenSearch", err, nil)
		os.Exit(1)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(listener); err != nil {
			_ = catcher.Error("cannot serve grpc", err, nil)
			os.Exit(1)
		}
	}()

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processStatistics(ctx)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		saveToDB(ctx, "success")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		saveToDB(ctx, "failure")
	}()

	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	<-signs

	grpcServer.GracefulStop()
	cancel()
}

func (p *notificationServer) Notify(_ context.Context, msg *plugins.Message) (*emptypb.Empty, error) {
	switch plugins.Topic(msg.Topic) {
	case plugins.TopicEnqueueSuccess:
	case plugins.TopicEnqueueFailure:
	case plugins.TopicParsingFailure:
	case plugins.TopicAnalysisFailure:
	case plugins.TopicCorrelationFailure:
	default:
		return &emptypb.Empty{}, nil
	}

	messageBytes := []byte(msg.Message)

	var pMsg plugins.DataProcessingMessage

	err := json.Unmarshal(messageBytes, &pMsg)
	if err != nil {
		return &emptypb.Empty{}, catcher.Error("cannot unmarshal message", err, map[string]any{})
	}

	statisticsQueue <- map[string]plugins.DataProcessingMessage{msg.Topic: pMsg}

	return &emptypb.Empty{}, nil
}

func processStatistics(ctx context.Context) {
	for {
		select {
		case msg := <-statisticsQueue:
			for k, v := range msg {
				switch plugins.Topic(k) {
				case plugins.TopicEnqueueSuccess:
					successLock.Lock()
					if _, ok := success[v.DataSource]; !ok {
						success[v.DataSource] = make(map[string]int64)
					}

					if _, ok := success[v.DataSource][v.DataType]; !ok {
						success[v.DataSource][v.DataType] = 0
					}

					success[v.DataSource][v.DataType]++
					successLock.Unlock()
				default:
					failsLock.Lock()
					if _, ok := fails[k]; !ok {
						fails[k] = make(map[string]map[string]map[string]int64)
					}

					if _, ok := fails[k][v.DataSource]; !ok {
						fails[k][v.DataSource] = make(map[string]map[string]int64)
					}

					if _, ok := fails[k][v.DataSource][v.DataType]; !ok {
						fails[k][v.DataSource][v.DataType] = make(map[string]int64)
					}

					if _, ok := fails[k][v.DataSource][v.DataType][v.Error.Code]; !ok {
						fails[k][v.DataSource][v.DataType][v.Error.Code] = 0
					}

					fails[k][v.DataSource][v.DataType][v.Error.Code]++
					failsLock.Unlock()
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

type Statistic struct {
	Timestamp  string  `json:"@timestamp"`
	DataSource string  `json:"dataSource"`
	DataType   string  `json:"dataType"`
	Cause      *string `json:"cause,omitempty"`
	Count      int64   `json:"count"`
	Type       string  `json:"type"`
}

func saveToDB(ctx context.Context, t string) {
	for {
		select {
		case <-time.After(10 * time.Minute):
			sendStatistic(t)
		case <-ctx.Done():
			return
		}
	}
}

func extractSuccess() []Statistic {
	successLock.Lock()
	defer successLock.Unlock()

	var result []Statistic

	for dataSource, dataTypes := range success {
		for dataType, count := range dataTypes {
			result = append(result, Statistic{
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				DataSource: dataSource,
				DataType:   dataType,
				Count:      count,
				Type:       string(plugins.TopicEnqueueSuccess),
			})
		}
	}

	success = make(map[string]map[string]int64)

	return result
}

func extractFails() []Statistic {
	failsLock.Lock()
	defer failsLock.Unlock()

	var result []Statistic

	for topic, dataSources := range fails {
		for dataSource, dataTypes := range dataSources {
			for dataType, causes := range dataTypes {
				for cause, count := range causes {
					result = append(result, Statistic{
						Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
						DataSource: dataSource,
						DataType:   dataType,
						Cause:      utils.PointerOf(cause),
						Count:      count,
						Type:       topic,
					})
				}
			}
		}
	}

	fails = make(map[string]map[string]map[string]map[string]int64)

	return result
}

func sendStatistic(t string) {
	switch t {
	case "success":
		success := extractSuccess()
		for _, s := range success {
			saveToOpenSearch(s)
		}

	case "failure":
		fails := extractFails()
		for _, f := range fails {
			saveToOpenSearch(f)
		}
	}
}

func saveToOpenSearch[Data any](data Data) {
	oCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := opensearch.IndexDoc(oCtx, &data, fmt.Sprintf("v11-statistics-%s", time.Now().UTC().Format("2006.01")), uuid.NewString())
	if err != nil {
		_ = catcher.Error("cannot index document", err, map[string]any{})
	}
}
