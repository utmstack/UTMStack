package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/elastic"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/alert"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/llm"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/processor"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

// Item represents an item in the processing queue
type Item struct {
	Alert       *plugins.Alert
	AlertFields *schema.AlertFields // For manual submissions (already converted)
	Timestamp   time.Time
	IsManual    bool // True if submitted via HTTP API
}

// AlertQueue manages the alert processing queue with workers
type AlertQueue struct {
	queue   chan *Item
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	// Metrics
	processedCount int64
	droppedCount   int64
	errorCount     int64
	queueSize      int64

	// Track consecutive drops for critical alerts
	consecutiveDrops int64
	lastDropAlert    time.Time
}

const (
	DefaultQueueSize   = 1000
	DefaultWorkerCount = 5
	QueueFullTimeout   = 100 * time.Millisecond
)

var instance *AlertQueue

// Initialize creates and starts the alert processing queue
func Initialize() {
	ctx, cancel := context.WithCancel(context.Background())

	instance = &AlertQueue{
		queue:   make(chan *Item, DefaultQueueSize),
		workers: DefaultWorkerCount,
		ctx:     ctx,
		cancel:  cancel,
	}

	for i := range DefaultWorkerCount {
		instance.wg.Add(1)
		go instance.worker(i)
	}

	go instance.metricsLogger()
}

// Enqueue adds an alert to the processing queue (from gRPC)
func Enqueue(pluginAlert *plugins.Alert) bool {
	if instance == nil {
		return false
	}

	item := &Item{
		Alert:     pluginAlert,
		Timestamp: time.Now(),
		IsManual:  false,
	}

	return enqueueItem(item, pluginAlert.Id)
}

// EnqueueManual adds an alert to the processing queue (from HTTP API)
func EnqueueManual(alertFields *schema.AlertFields) bool {
	if instance == nil {
		return false
	}

	item := &Item{
		AlertFields: alertFields,
		Timestamp:   time.Now(),
		IsManual:    true,
	}

	return enqueueItem(item, alertFields.Id)
}

// enqueueItem is the internal function to add items to the queue
func enqueueItem(item *Item, alertID string) bool {
	select {
	case instance.queue <- item:
		atomic.AddInt64(&instance.queueSize, 1)
		atomic.StoreInt64(&instance.consecutiveDrops, 0)
		return true
	case <-time.After(QueueFullTimeout):
		atomic.AddInt64(&instance.droppedCount, 1)
		atomic.AddInt64(&instance.consecutiveDrops, 1)

		currentQueueSize := atomic.LoadInt64(&instance.queueSize)
		totalDropped := atomic.LoadInt64(&instance.droppedCount)
		consecutiveDrops := atomic.LoadInt64(&instance.consecutiveDrops)

		_ = catcher.Error("Alert Dropped due to queue full", nil, map[string]any{
			"process":           "plugin_com.utmstack.soc-ai",
			"id":                alertID,
			"total_dropped":     totalDropped,
			"consecutive_drops": consecutiveDrops,
		})

		elastic.RegisterError(fmt.Sprintf("Alert dropped - Queue FULL (%d/%d)", currentQueueSize, DefaultQueueSize), alertID)
		instance.lastDropAlert = time.Now()
		return false
	}
}

func (aq *AlertQueue) worker(workerID int) {
	defer aq.wg.Done()

	for {
		select {
		case <-aq.ctx.Done():
			return
		case item := <-aq.queue:
			if item == nil {
				continue
			}

			atomic.AddInt64(&aq.queueSize, -1)
			aq.processAlert(workerID, item)
		}
	}
}

func (aq *AlertQueue) processAlert(workerID int, item *Item) {
	var alertFields schema.AlertFields

	// Handle both gRPC and manual (HTTP) submissions
	if item.IsManual && item.AlertFields != nil {
		// Manual submission: already has AlertFields, just clean it
		alertFields = alert.Clean(*item.AlertFields)
	} else if item.Alert != nil {
		// gRPC submission: convert and clean
		alertFields = alert.Clean(alert.ToAlertFields(item.Alert))
	} else {
		// Invalid item
		atomic.AddInt64(&aq.errorCount, 1)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			atomic.AddInt64(&aq.errorCount, 1)
			_ = catcher.Error("recovered from panic in alert processing", nil, map[string]any{
				"process":  "plugin_com.utmstack.soc-ai",
				"panic":    r,
				"alert":    alertFields.Name,
				"workerID": workerID,
				"isManual": item.IsManual,
			})
			elastic.RegisterError(fmt.Sprintf("Panic in worker %d: %v", workerID, r), alertFields.Id)
		}
	}()

	if config.GetConfig() == nil || !config.GetConfig().ModuleActive {
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	// Skip alerts already tagged "False positive". Manual path carries Tags
	// in AlertFields; gRPC path has none (proto lacks the field) so look up
	// the alert doc from ES. Fail-open if the lookup fails.
	tags := alertFields.Tags
	if len(tags) == 0 && !item.IsManual {
		tags = fetchAlertTags(alertFields.Id)
	}
	if hasFalsePositiveTag(tags) {
		catcher.Info("Skipping LLM: alert tagged False positive", map[string]any{
			"process":  "plugin_com.utmstack.soc-ai",
			"alert_id": alertFields.Id,
		})
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	// Check connection to LLM endpoint
	if config.GetConfig().URL != "" {
		if err := utils.ConnectionChecker(config.GetConfig().URL); err != nil {
			atomic.AddInt64(&aq.errorCount, 1)
			_ = catcher.Error("Failed to establish connection to LLM", err, map[string]any{"process": "plugin_com.utmstack.soc-ai"})
			elastic.RegisterError("Failed to establish connection to LLM", alertFields.Id)
			return
		}
	}

	err := llm.SendRequest(&alertFields)
	if err != nil {
		atomic.AddInt64(&aq.errorCount, 1)
		elastic.RegisterError(err.Error(), alertFields.Id)
		return
	}

	err = processor.SaveToElastic(&alertFields)
	if err != nil {
		atomic.AddInt64(&aq.errorCount, 1)
		elastic.RegisterError(err.Error(), alertFields.Id)
		return
	}

	atomic.AddInt64(&aq.processedCount, 1)
}

// Matches backend Constants.FALSE_POSITIVE_TAG exactly.
const falsePositiveTag = "False positive"

func hasFalsePositiveTag(tags []string) bool {
	for _, t := range tags {
		if t == falsePositiveTag {
			return true
		}
	}
	return false
}

func fetchAlertTags(id string) []string {
	result, err := elastic.ElasticSearchWithLimit(config.ALERT_INDEX_PATTERN, "id", id, 1)
	if err != nil {
		return nil
	}
	var alerts []schema.AlertFields
	if err := json.Unmarshal(result, &alerts); err != nil || len(alerts) == 0 {
		return nil
	}
	return alerts[0].Tags
}

func (aq *AlertQueue) metricsLogger() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-aq.ctx.Done():
			return
		case <-ticker.C:
			processed := atomic.LoadInt64(&aq.processedCount)
			dropped := atomic.LoadInt64(&aq.droppedCount)
			errors := atomic.LoadInt64(&aq.errorCount)
			queueSize := atomic.LoadInt64(&aq.queueSize)

			catcher.Info("SOC-AI queue metrics", map[string]any{
				"process":   "plugin_com.utmstack.soc-ai",
				"processed": processed,
				"dropped":   dropped,
				"errors":    errors,
				"queueSize": queueSize,
			})
		}
	}
}
