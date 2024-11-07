package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/threatwinds/go-sdk/opensearch"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type notificationServer struct {
	go_sdk.UnimplementedNotificationServer
}

type Config struct {
	RulesFolder   string `yaml:"rules_folder"`
	GeoIPFolder   string `yaml:"geoip_folder"`
	Elasticsearch string `yaml:"elasticsearch"`
	PostgreSQL    struct {
		Server   string `yaml:"server"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"postgresql"`
}

var statisticsQueue chan map[string]go_sdk.DataProcessingMessage
var fails map[string]map[string]map[string]map[string]int64
var failsLock sync.Mutex
var success map[string]map[string]int64
var successLock sync.Mutex

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	os.Remove(path.Join(go_sdk.GetCfg().Env.Workdir,
		"sockets", "com.utmstack.stats_notification.sock"))

	laddr, err := net.ResolveUnixAddr(
		"unix", path.Join(go_sdk.GetCfg().Env.Workdir,
			"sockets", "com.utmstack.stats_notification.sock"))

	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	statisticsQueue = make(chan map[string]go_sdk.DataProcessingMessage, runtime.NumCPU()*100)
	fails = make(map[string]map[string]map[string]map[string]int64)
	success = make(map[string]map[string]int64)

	grpcServer := grpc.NewServer()
	go_sdk.RegisterNotificationServer(grpcServer, &notificationServer{})

	pCfg, e := go_sdk.PluginCfg[Config]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	if err := opensearch.Connect([]string{pCfg.Elasticsearch}); err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(listener); err != nil {
			go_sdk.Logger().ErrorF(err.Error())
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
	go_sdk.Logger().Info("shutting down...")
}

func (p *notificationServer) Notify(ctx context.Context, msg *go_sdk.Message) (*emptypb.Empty, error) {
	go_sdk.Logger().LogF(100, "%s: %s", msg.Topic, msg.Message)

	switch msg.Topic {
	case go_sdk.TOPIC_ENQUEUE_SUCCESS:
	case go_sdk.TOPIC_ENQUEUE_FAILURE:
	case go_sdk.TOPIC_PARSING_FAILURE:
	case go_sdk.TOPIC_ANALYSIS_FAILURE:
	case go_sdk.TOPIC_CORRELATION_FAILURE:
	default:
		return &emptypb.Empty{}, nil
	}
	
	mbytes := []byte(msg.Message)

	var pMsg go_sdk.DataProcessingMessage

	err := json.Unmarshal(mbytes, &pMsg)
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
		return &emptypb.Empty{}, err
	}

	statisticsQueue <- map[string]go_sdk.DataProcessingMessage{msg.Topic: pMsg}

	return &emptypb.Empty{}, nil
}

func processStatistics(ctx context.Context) {
	for {
		select {
		case msg := <-statisticsQueue:
			for k, v := range msg {
				switch k {
				case go_sdk.TOPIC_ENQUEUE_SUCCESS:
					successLock.Lock()
					if _, ok := success[v.DataSource]; !ok {
						success[v.DataSource] = make(map[string]int64)
					}

					if _, ok := success[v.DataSource][v.DataType]; !ok {
						success[v.DataSource][v.DataType] = 0
					}

					go_sdk.Logger().LogF(100, "success: %s", v.DataType)

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

					if _, ok := fails[k][v.DataSource][v.DataType][*v.Cause]; !ok {
						fails[k][v.DataSource][v.DataType][*v.Cause] = 0
					}

					go_sdk.Logger().LogF(100, "failure: %s", v.DataType)

					fails[k][v.DataSource][v.DataType][*v.Cause]++
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
			go_sdk.Logger().Info("sending %s statistics", t)
			sendStatistic(t)
		case <-ctx.Done():
			go_sdk.Logger().Info("shutting down %s statistics", t)
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
				Type:       go_sdk.TOPIC_ENQUEUE_SUCCESS,
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
						Cause:      go_sdk.PointerOf(cause),
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
		go_sdk.Logger().Info("sending %d success statistics", len(success))
		for _, s := range success {
			saveToOpensearch(s)
		}

	case "failure":
		fails := extractFails()
		go_sdk.Logger().Info("sending %d failure statistics", len(fails))
		for _, f := range fails {
			saveToOpensearch(f)
		}
	}
}

func saveToOpensearch[Data any](data Data) {
	oCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := opensearch.IndexDoc(oCtx, &data, fmt.Sprintf("statistics-%s", time.Now().UTC().Format("2006.01")), uuid.NewString())
	if err != nil {
		go_sdk.Logger().ErrorF(err.Error())
	}
}
