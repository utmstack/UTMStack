package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store"
	ch "github.com/threatwinds/go-sdk/store/clickhouse"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	processName = "plugin_com.utmstack.stats"

	datasetStats      = store.Dataset("statistics")
	defaultStatsTable = "statistics"

	cycleInterval = 10 * time.Minute

	// A cycle's rows go in one batch, so a retry re-sends the whole batch.
	maxRetries   = 3
	retryDelay   = 2 * time.Second
	writeTimeout = 60 * time.Second
)

var errQueueFull = errors.New("queue is full")

type statsKey struct {
	Topic      plugins.Topic
	TenantID   string
	DataSource string
	DataType   string
}

type statsValue struct {
	Count int64
	Bytes int64
}

type statsEntry struct {
	Topic   plugins.Topic
	Message plugins.DataProcessingMessage
}

var statisticsQueue chan statsEntry
var statsMap map[statsKey]*statsValue
var statsLock sync.Mutex
var statsStore *ch.Driver

func main() {
	statisticsQueue = make(chan statsEntry, runtime.NumCPU()*100)
	statsMap = make(map[statsKey]*statsValue)

	cfg := plugins.PluginCfg("clickhouse")
	table := cfg.Get("statisticsTable").String()
	if table == "" {
		table = defaultStatsTable
	}

	var err error
	statsStore, err = ch.New(ch.Config{
		Addr:     []string{cfg.Get("host").String() + ":" + cfg.Get("port").String()},
		Database: cfg.Get("database").String(),
		Username: cfg.Get("user").String(),
		Password: cfg.Get("password").String(),
		Tables: map[store.Dataset]string{
			datasetStats: table,
		},
		TenantColumn:   "tenantId",
		TimeColumn:     "@timestamp",
		DataTypeColumn: "dataType",
	})
	if err != nil {
		_ = catcher.Error("failed to build the statistics store", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	defer statsStore.Close()

	ctx, cancel := context.WithCancel(context.Background())

	var accumulators, writer sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		accumulators.Add(1)
		go func() {
			defer accumulators.Done()
			processStatistics(ctx)
		}()
	}

	writer.Add(1)
	go func() {
		defer writer.Done()
		saveToDB(ctx)
	}()

	err = plugins.InitNotificationPlugin("com.utmstack.stats", notify)
	if err != nil {
		_ = catcher.Error("failed to start notification plugin", err, map[string]any{
			"process": processName,
		})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	cancel()
	accumulators.Wait()
	writer.Wait()
	sendStatistic(context.WithoutCancel(ctx))
}

func notify(_ context.Context, msg *plugins.Message) (*emptypb.Empty, error) {
	switch plugins.Topic(msg.Topic) {
	case plugins.TopicEnqueueSuccess, plugins.TopicParsingDropped, plugins.TopicAnalysisDropped, plugins.TopicCorrelationDropped:
	default:
		return &emptypb.Empty{}, nil
	}

	var pMsg plugins.DataProcessingMessage

	if err := json.Unmarshal([]byte(msg.Message), &pMsg); err != nil {
		return &emptypb.Empty{}, catcher.Error("cannot unmarshal message", err, map[string]any{"process": "plugin_com.utmstack.stats"})
	}

	select {
	case statisticsQueue <- statsEntry{Topic: plugins.Topic(msg.Topic), Message: pMsg}:
	default:
		_ = catcher.Error("cannot enqueue statistic", errQueueFull, map[string]any{
			"process": processName,
			"queue":   "statistics",
		})
	}

	return &emptypb.Empty{}, nil
}

func accumulate(e statsEntry) {
	k := statsKey{
		Topic:      e.Topic,
		TenantID:   e.Message.TenantID,
		DataSource: e.Message.DataSource,
		DataType:   e.Message.DataType,
	}

	statsLock.Lock()
	defer statsLock.Unlock()

	v, ok := statsMap[k]
	if !ok {
		v = &statsValue{}
		statsMap[k] = v
	}
	v.Count++
	v.Bytes += e.Message.Bytes
}

func processStatistics(ctx context.Context) {
	for {
		select {
		case e := <-statisticsQueue:
			accumulate(e)
		case <-ctx.Done():
			for {
				select {
				case e := <-statisticsQueue:
					accumulate(e)
				default:
					return
				}
			}
		}
	}
}

type Statistic struct {
	Timestamp  string `json:"@timestamp"`
	TenantID   string `json:"tenantId,omitempty"`
	DataSource string `json:"dataSource"`
	DataType   string `json:"dataType"`
	Count      int64  `json:"count"`
	Bytes      int64  `json:"bytes"`
	Type       string `json:"type"`
	BatchID    string `json:"batchId"`
}

func saveToDB(ctx context.Context) {
	for {
		select {
		case <-time.After(cycleInterval):
			sendStatistic(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func extractStats() []Statistic {
	statsLock.Lock()
	defer statsLock.Unlock()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	result := make([]Statistic, 0, len(statsMap))

	for k, v := range statsMap {
		result = append(result, Statistic{
			Timestamp:  now,
			TenantID:   k.TenantID,
			DataSource: k.DataSource,
			DataType:   k.DataType,
			Count:      v.Count,
			Bytes:      v.Bytes,
			Type:       string(k.Topic),
		})
	}

	statsMap = make(map[statsKey]*statsValue)

	return result
}

func sendStatistic(ctx context.Context) {
	stats := extractStats()
	if len(stats) == 0 {
		return
	}

	rows := make([]Statistic, 0, len(stats))
	var orphans int64
	var orphanBytes int64
	for _, s := range stats {
		if s.TenantID == "" {
			orphans++
			orphanBytes += s.Bytes
			continue
		}
		rows = append(rows, s)
	}
	if orphans > 0 {
		_ = catcher.Error("dropping statistics with no tenant", nil, map[string]any{
			"process": processName,
			"rows":    orphans,
			"bytes":   orphanBytes,
		})
	}
	if len(rows) == 0 {
		return
	}

	batchID := uuid.NewString()
	for i := range rows {
		rows[i].BatchID = batchID
	}

	delay := retryDelay
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 && batchLanded(ctx, rows[0].TenantID, batchID) {
			return
		}

		err := writeBatch(ctx, rows)
		if err == nil {
			return
		}

		_ = catcher.Error("cannot write statistics, retrying", err, map[string]any{
			"process":    processName,
			"retry":      attempt,
			"maxRetries": maxRetries,
			"batchId":    batchID,
		})

		if attempt < maxRetries {
			time.Sleep(delay)
			delay *= 2
		}
	}

	var lostCount, lostBytes int64
	for _, s := range rows {
		lostCount += s.Count
		lostBytes += s.Bytes
	}
	_ = catcher.Error("all retries failed, statistics lost", nil, map[string]any{
		"process":    processName,
		"rows":       len(rows),
		"lostEvents": lostCount,
		"lostBytes":  lostBytes,
		"maxRetries": maxRetries,
	})
}

func batchLanded(ctx context.Context, tenantID, batchID string) bool {
	cCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()

	n, err := statsStore.Count(cCtx,
		store.Scope{Tenant: tenantID, Dataset: datasetStats},
		[]store.Filter{{Field: "batchId", Op: store.OpEq, Value: batchID}},
	)
	if err != nil {
		_ = catcher.Error("cannot tell whether the batch landed", err, map[string]any{
			"process": processName,
			"batchId": batchID,
		})
		return false
	}

	if n > 0 {
		catcher.Info("the batch had already landed, not re-sending", map[string]any{
			"process": processName,
			"batchId": batchID,
		})
	}
	return n > 0
}

func writeBatch(ctx context.Context, rows []Statistic) error {
	w, err := statsStore.BulkWriter(datasetStats)
	if err != nil {
		return err
	}

	for _, s := range rows {
		doc, err := json.Marshal(s)
		if err != nil {
			return err
		}
		if err := w.Write(store.Scope{Tenant: s.TenantID, Dataset: datasetStats}, doc); err != nil {
			return err
		}
	}

	wCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return w.Close(wCtx)
}
