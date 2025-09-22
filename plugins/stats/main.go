package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/opensearch"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type notificationServer struct {
	plugins.UnimplementedNotificationServer
}

var statisticsQueue chan map[string]plugins.DataProcessingMessage
var success map[string]map[string]int64
var successLock sync.Mutex

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Retry logic for initialization
	var filePath utils.Folder
	var err error
	var socketPath string
	var unixAddress *net.UnixAddr
	var listener *net.UnixListener

	// Retry logic for creating socket directory
	maxRetries := 10
	retryDelay := 5 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		filePath, err = utils.MkdirJoin(plugins.WorkDir, "sockets")
		if err != nil {
			_ = catcher.Error("cannot create directory, retrying", err, map[string]any{
				"retry":      retry + 1,
				"maxRetries": maxRetries,
			})
			time.Sleep(retryDelay)
			continue
		}

		socketPath = filePath.FileJoin("com.utmstack.stats_notification.sock")
		_ = os.Remove(socketPath)

		unixAddress, err = net.ResolveUnixAddr("unix", socketPath)
		if err != nil {
			_ = catcher.Error("cannot resolve unix address, retrying", err, map[string]any{
				"retry":      retry + 1,
				"maxRetries": maxRetries,
			})
			time.Sleep(retryDelay)
			continue
		}

		listener, err = net.ListenUnix("unix", unixAddress)
		if err != nil {
			_ = catcher.Error("cannot listen to unix socket, retrying", err, map[string]any{
				"retry":      retry + 1,
				"maxRetries": maxRetries,
			})
			time.Sleep(retryDelay)
			continue
		}

		// If we got here, initialization was successful
		break
	}

	// If all retries failed, log a final error and exit
	if listener == nil {
		_ = catcher.Error("all retries failed when initializing socket", nil, map[string]any{
			"maxRetries": maxRetries,
		})
		os.Exit(1)
	}

	statisticsQueue = make(chan map[string]plugins.DataProcessingMessage, runtime.NumCPU()*100)
	success = make(map[string]map[string]int64)

	grpcServer := grpc.NewServer()
	plugins.RegisterNotificationServer(grpcServer, &notificationServer{})

	pCfg := plugins.PluginCfg("com.utmstack", false)
	osUrl := pCfg.Get("opensearch").String()

	// Retry logic for connecting to OpenSearch
	maxOSRetries := 10
	osRetryDelay := 5 * time.Second
	var osConnected bool

	for retry := 0; retry < maxOSRetries; retry++ {
		err := opensearch.Connect([]string{osUrl})
		if err == nil {
			osConnected = true
			break
		}
		_ = catcher.Error("cannot connect to ElasticSearch/OpenSearch, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxOSRetries,
		})
		time.Sleep(osRetryDelay)
	}

	// If all retries failed, log a final error and exit
	if !osConnected {
		_ = catcher.Error("all retries failed when connecting to OpenSearch", nil, map[string]any{
			"maxRetries": maxOSRetries,
		})
		os.Exit(1)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(listener); err != nil {
			_ = catcher.Error("cannot serve grpc", err, nil)
			// Instead of exiting, just log the error and let the main function handle it
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

	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	<-signs

	grpcServer.GracefulStop()
	cancel()

	wg.Wait()
}

func (p *notificationServer) Notify(_ context.Context, msg *plugins.Message) (*emptypb.Empty, error) {
	switch plugins.Topic(msg.Topic) {
	case plugins.TopicEnqueueSuccess:
	default:
		return &emptypb.Empty{}, nil
	}

	messageBytes := []byte(msg.Message)

	var pMsg plugins.DataProcessingMessage

	err := json.Unmarshal(messageBytes, &pMsg)
	if err != nil {
		return &emptypb.Empty{}, catcher.Error("cannot unmarshal message", err, nil)
	}

	statisticsQueue <- map[string]plugins.DataProcessingMessage{msg.Topic: pMsg}

	return &emptypb.Empty{}, nil
}

func processStatistics(ctx context.Context) {
	for {
		select {
		case msg := <-statisticsQueue:
			for _, v := range msg {
				successLock.Lock()
				if _, ok := success[v.DataSource]; !ok {
					success[v.DataSource] = make(map[string]int64)
				}
				if _, ok := success[v.DataSource][v.DataType]; !ok {
					success[v.DataSource][v.DataType] = 0
				}
				success[v.DataSource][v.DataType]++
				successLock.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

type Statistic struct {
	Timestamp  string `json:"@timestamp"`
	DataSource string `json:"dataSource"`
	DataType   string `json:"dataType"`
	Count      int64  `json:"count"`
	Type       string `json:"type"`
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

func sendStatistic(t string) {
	success := extractSuccess()
	for _, s := range success {
		saveToOpenSearch(s)
	}
}

func saveToOpenSearch[Data any](data Data) {
	// Retry logic for indexing a document
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		oCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		err := opensearch.IndexDoc(oCtx, &data, fmt.Sprintf("v11-statistics-%s", time.Now().UTC().Format("2006.01")), uuid.NewString())
		cancel()

		if err == nil {
			// Successfully indexed document
			return
		}

		_ = catcher.Error("cannot index document, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry (exponential backoff)
			retryDelay *= 2
		}
	}

	// After all retries, log a final error
	_ = catcher.Error("all retries failed when indexing document", nil, map[string]any{
		"maxRetries": maxRetries,
	})
}
