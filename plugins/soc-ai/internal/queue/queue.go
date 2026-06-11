package queue

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/agent"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/alert"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

const maxAlertContentSize = 100000

type Item struct {
	Alert       *plugins.Alert
	AlertFields *schema.AlertFields // for manual submissions (already converted)
	Timestamp   time.Time
	IsManual    bool
}

type AlertQueue struct {
	queue   chan *Item
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	processedCount int64
	droppedCount   int64
	errorCount     int64
	queueSize      int64

	consecutiveDrops int64
	lastDropAlert    time.Time
}

const (
	DefaultQueueSize   = 1000
	DefaultWorkerCount = 5
	QueueFullTimeout   = 100 * time.Millisecond
)

var instance *AlertQueue

// Initialize creates and starts the alert processing queue.
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

// Enqueue adds an alert to the processing queue (from gRPC).
func Enqueue(pluginAlert *plugins.Alert) bool {
	if instance == nil {
		return false
	}
	return enqueueItem(&Item{Alert: pluginAlert, Timestamp: time.Now()}, pluginAlert.Id)
}

// EnqueueManual adds an alert to the processing queue (from the HTTP API).
func EnqueueManual(alertFields *schema.AlertFields) bool {
	if instance == nil {
		return false
	}
	return enqueueItem(&Item{AlertFields: alertFields, Timestamp: time.Now(), IsManual: true}, alertFields.Id)
}

func enqueueItem(item *Item, alertID string) bool {
	select {
	case instance.queue <- item:
		atomic.AddInt64(&instance.queueSize, 1)
		atomic.StoreInt64(&instance.consecutiveDrops, 0)
		return true
	case <-time.After(QueueFullTimeout):
		atomic.AddInt64(&instance.droppedCount, 1)
		atomic.AddInt64(&instance.consecutiveDrops, 1)
		_ = catcher.Error("Alert dropped due to full queue", nil, map[string]any{
			"process":           "plugin_com.utmstack.soc-ai",
			"id":                alertID,
			"total_dropped":     atomic.LoadInt64(&instance.droppedCount),
			"consecutive_drops": atomic.LoadInt64(&instance.consecutiveDrops),
		})
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
	switch {
	case item.IsManual && item.AlertFields != nil:
		alertFields = alert.Clean(*item.AlertFields)
	case item.Alert != nil:
		alertFields = alert.Clean(alert.ToAlertFields(item.Alert))
	default:
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
			})
		}
	}()

	cfg := config.GetConfig()
	if cfg == nil || !cfg.ModuleActive {
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	ag := agent.Current()
	if ag == nil {
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	alertJSON, err := json.Marshal(&alertFields)
	if err != nil {
		atomic.AddInt64(&aq.errorCount, 1)
		_ = catcher.Error("failed to marshal alert", err, map[string]any{"process": "plugin_com.utmstack.soc-ai", "id": alertFields.Id})
		return
	}
	content := string(alertJSON)
	if len(content) > maxAlertContentSize {
		content = content[:maxAlertContentSize] + "...[TRUNCATED]"
	}

	task := agent.RunTask{
		System:        agent.TriagePrompt(),
		Input:         "Triage this alert and record your assessment as a note.\n\nALERT:\n" + content,
		EnabledGroups: cfg.Capabilities,
		// The assessment note is the triage output channel — always permitted.
		AlwaysAllow: []string{"alerts.update_notes"},
		MaxIters:    cfg.MaxToolIterations,
	}

	if _, err := ag.Run(aq.ctx, task, nil); err != nil {
		atomic.AddInt64(&aq.errorCount, 1)
		_ = catcher.Error("agent triage failed", err, map[string]any{"process": "plugin_com.utmstack.soc-ai", "id": alertFields.Id})
		return
	}

	atomic.AddInt64(&aq.processedCount, 1)
}

func (aq *AlertQueue) metricsLogger() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-aq.ctx.Done():
			return
		case <-ticker.C:
			catcher.Info("SOC-AI queue metrics", map[string]any{
				"process":   "plugin_com.utmstack.soc-ai",
				"processed": atomic.LoadInt64(&aq.processedCount),
				"dropped":   atomic.LoadInt64(&aq.droppedCount),
				"errors":    atomic.LoadInt64(&aq.errorCount),
				"queueSize": atomic.LoadInt64(&aq.queueSize),
			})
		}
	}
}
