package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/protobuf/types/known/emptypb"
)

var statisticsQueue chan map[plugins.Topic]plugins.DataProcessingMessage
var statsMap map[plugins.Topic]map[string]map[string]int64
var statsLock sync.Mutex

func main() {
	statisticsQueue = make(chan map[plugins.Topic]plugins.DataProcessingMessage, runtime.NumCPU()*100)
	statsMap = make(map[plugins.Topic]map[string]map[string]int64)

	cfg := plugins.PluginCfg("org.opensearch").Get("opensearch")
	host := cfg.Get("host").String()
	port := cfg.Get("port").String()
	user := cfg.Get("user").String()
	password := cfg.Get("password").String()
	osUrl := "https://" + host + ":" + port

	err := sdkos.Connect([]string{osUrl}, user, password)
	if err != nil {
		_ = catcher.Error("failed when connecting to OpenSearch", err, map[string]any{"process": "plugin_com.utmstack.stats"})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

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

	err = plugins.InitNotificationPlugin("com.utmstack.stats", notify)
	if err != nil {
		_ = catcher.Error("failed to start notification plugin", err, map[string]any{
			"process": "plugin_com.utmstack.stats",
		})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	cancel()
	wg.Wait()
}

func notify(_ context.Context, msg *plugins.Message) (*emptypb.Empty, error) {
	switch plugins.Topic(msg.Topic) {
	case plugins.TopicEnqueueSuccess, plugins.TopicParsingDropped, plugins.TopicAnalysisDropped, plugins.TopicCorrelationDropped:
	default:
		return &emptypb.Empty{}, nil
	}

	messageBytes := []byte(msg.Message)

	var pMsg plugins.DataProcessingMessage

	err := json.Unmarshal(messageBytes, &pMsg)
	if err != nil {
		return &emptypb.Empty{}, catcher.Error("cannot unmarshal message", err, map[string]any{"process": "plugin_com.utmstack.stats"})
	}

	statisticsQueue <- map[plugins.Topic]plugins.DataProcessingMessage{plugins.Topic(msg.Topic): pMsg}

	return &emptypb.Empty{}, nil
}

func processStatistics(ctx context.Context) {
	for {
		select {
		case msg := <-statisticsQueue:
			for topic, v := range msg {
				statsLock.Lock()
				if _, ok := statsMap[topic]; !ok {
					statsMap[topic] = make(map[string]map[string]int64)
				}
				if _, ok := statsMap[topic][v.DataSource]; !ok {
					statsMap[topic][v.DataSource] = make(map[string]int64)
				}
				if _, ok := statsMap[topic][v.DataSource][v.DataType]; !ok {
					statsMap[topic][v.DataSource][v.DataType] = 0
				}
				statsMap[topic][v.DataSource][v.DataType]++
				statsLock.Unlock()
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

func extractStats() []Statistic {
	statsLock.Lock()
	defer statsLock.Unlock()

	var result []Statistic

	for topic, sourceMap := range statsMap {
		for dataSource, typeMap := range sourceMap {
			for dataType, count := range typeMap {
				result = append(result, Statistic{
					Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
					DataSource: dataSource,
					DataType:   dataType,
					Count:      count,
					Type:       string(topic),
				})
			}
		}
	}

	statsMap = make(map[plugins.Topic]map[string]map[string]int64)

	return result
}

func sendStatistic(t string) {
	stats := extractStats()
	for _, s := range stats {
		saveToOpenSearch(s)
	}
}

func saveToOpenSearch[Data any](data Data) {
	// Retry logic for indexing a document
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		oCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		err := sdkos.IndexDoc(oCtx, &data, fmt.Sprintf("v11-statistics-%s", time.Now().UTC().Format("2006.01")), uuid.NewString())
		cancel()

		if err == nil {
			// Successfully indexed document
			return
		}

		_ = catcher.Error("cannot index document, retrying", err, map[string]any{
			"process":    "plugin_com.utmstack.stats",
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
		"process":    "plugin_com.utmstack.stats",
		"maxRetries": maxRetries,
	})
}
