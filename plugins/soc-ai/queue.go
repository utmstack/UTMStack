package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/elastic"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

// AlertQueueItem represents an item in the processing queue
type AlertQueueItem struct {
	Alert     *plugins.Alert
	Timestamp time.Time
}

// AlertQueue manages the alert processing queue with workers
type AlertQueue struct {
	queue   chan *AlertQueueItem
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

// Global queue instance
var alertQueue *AlertQueue

const (
	DefaultQueueSize   = 1000
	DefaultWorkerCount = 5
	QueueFullTimeout   = 100 * time.Millisecond
)

func InitializeQueue() {
	ctx, cancel := context.WithCancel(context.Background())

	alertQueue = &AlertQueue{
		queue:   make(chan *AlertQueueItem, DefaultQueueSize),
		workers: DefaultWorkerCount,
		ctx:     ctx,
		cancel:  cancel,
	}

	for i := range DefaultWorkerCount {
		alertQueue.wg.Add(1)
		go alertQueue.worker(i)
	}

	go alertQueue.metricsLogger()

	utils.Logger.LogF(100, "Alert queue initialized with %d workers and queue size %d", DefaultWorkerCount, DefaultQueueSize)
}

func EnqueueAlert(alert *plugins.Alert) bool {
	if alertQueue == nil {
		utils.Logger.LogF(500, "Alert queue not initialized")
		return false
	}

	item := &AlertQueueItem{
		Alert:     alert,
		Timestamp: time.Now(),
	}

	select {
	case alertQueue.queue <- item:
		atomic.AddInt64(&alertQueue.queueSize, 1)
		// Reset consecutive drops counter on successful enqueue
		atomic.StoreInt64(&alertQueue.consecutiveDrops, 0)
		utils.Logger.LogF(100, "Alert %s enqueued for processing", alert.Id)
		return true
	case <-time.After(QueueFullTimeout):
		atomic.AddInt64(&alertQueue.droppedCount, 1)
		atomic.AddInt64(&alertQueue.consecutiveDrops, 1)

		currentQueueSize := atomic.LoadInt64(&alertQueue.queueSize)
		totalDropped := atomic.LoadInt64(&alertQueue.droppedCount)
		consecutiveDrops := atomic.LoadInt64(&alertQueue.consecutiveDrops)

		_ = plugins.EnqueueNotification(plugins.TopicIntegrationFailure, plugins.Message{
			Id: uuid.NewString(),
			Message: catcher.Error("Alert Dropped", nil, map[string]any{
				"id":                alert.Id,
				"total_dropped":     totalDropped,
				"consecutive_drops": consecutiveDrops,
			}).Error(),
		})
		utils.Logger.ErrorF("QUEUE FULL - Alert %s DROPPED! Queue size: %d/%d, Total dropped: %d, Consecutive: %d.",
			alert.Id, currentQueueSize, DefaultQueueSize, totalDropped, consecutiveDrops)

		elastic.RegisterError(fmt.Sprintf("Alert dropped - Queue FULL (%d/%d)", currentQueueSize, DefaultQueueSize), alert.Id)
		alertQueue.lastDropAlert = time.Now()
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

func (aq *AlertQueue) processAlert(workerID int, item *AlertQueueItem) {
	startTime := time.Now()
	alert := cleanAlerts(alertToAlertFields(item.Alert))

	utils.Logger.LogF(100, "Worker %d processing alert: %s", workerID, alert.ID)

	defer func() {
		if r := recover(); r != nil {
			atomic.AddInt64(&aq.errorCount, 1)
			_ = catcher.Error("recovered from panic in alert processing", nil, map[string]any{
				"panic":    r,
				"alert":    alert.Name,
				"workerID": workerID,
			})
			elastic.RegisterError(fmt.Sprintf("Panic in worker %d: %v", workerID, r), alert.ID)
		}
	}()

	if config.GetConfig() == nil || !config.GetConfig().ModuleActive {
		utils.Logger.LogF(100, "SOC-AI module is disabled, skipping alert: %s", alert.ID)
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	if config.GetConfig().Provider == "openai" {
		if err := utils.ConnectionChecker(config.GPT_API_ENDPOINT); err != nil {
			atomic.AddInt64(&aq.errorCount, 1)
			_ = catcher.Error("Failed to establish internet connection", err, nil)
			elastic.RegisterError("Failed to establish internet connection", alert.ID)
			return
		}
	}

	err := sendRequestToLLM(&alert)
	if err != nil {
		atomic.AddInt64(&aq.errorCount, 1)
		elastic.RegisterError(err.Error(), alert.ID)
		return
	}

	err = processAlertToElastic(&alert)
	if err != nil {
		atomic.AddInt64(&aq.errorCount, 1)
		elastic.RegisterError(err.Error(), alert.ID)
		return
	}

	atomic.AddInt64(&aq.processedCount, 1)
	duration := time.Since(startTime)
	queueTime := startTime.Sub(item.Timestamp)

	utils.Logger.LogF(100, "Worker %d completed alert %s in %v (queue time: %v)",
		workerID, alert.ID, duration, queueTime)
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

			utils.Logger.LogF(200, "Queue metrics - Processed: %d, Dropped: %d, Errors: %d, Current queue size: %d",
				processed, dropped, errors, queueSize)
		}
	}
}
