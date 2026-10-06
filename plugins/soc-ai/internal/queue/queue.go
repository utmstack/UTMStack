package queue

import (
	"context"
	"encoding/json"
	"strings"
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

// maxNoteLength bounds anything the model writes before it's either stored
// as an alert note or carried into another prompt. The required note format
// is one line with six short " | "-separated sections — generous slack over
// that, but nowhere near unbounded. Without this, a model that ignores the
// one-line instruction would write an arbitrarily long note, and that note
// is read back as "prior verdicts" context by future triage runs on the
// same alert name — one oversized reply would otherwise keep inflating
// every later prompt it's included in, not just this one.
const maxNoteLength = 2000

// truncateForReuse caps text the model produced before it's written
// anywhere or fed into another prompt, so a model ignoring the length
// instruction can't compound across writes or future prompts.
func truncateForReuse(s string) string {
	if len(s) <= maxNoteLength {
		return s
	}
	return s[:maxNoteLength] + "...[TRUNCATED]"
}

type Item struct {
	Alert       *plugins.Alert
	AlertFields *schema.AlertFields // for manual submissions (already converted)
	Timestamp   time.Time
	IsManual    bool
	TenantID    string
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
	return enqueueItem(&Item{
		Alert:     pluginAlert,
		Timestamp: time.Now(),
		TenantID:  pluginAlert.TenantId,
	}, pluginAlert.Id)
}

// EnqueueManual adds an alert to the processing queue (from the HTTP API).
func EnqueueManual(alertFields *schema.AlertFields, tenantID string) bool {
	if instance == nil {
		return false
	}
	return enqueueItem(&Item{
		AlertFields: alertFields,
		Timestamp:   time.Now(),
		IsManual:    true,
		TenantID:    tenantID,
	}, alertFields.Id)
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

	cfg := config.GetConfig(item.TenantID)
	if cfg == nil || !cfg.ModuleActive {
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	ag := agent.For(item.TenantID)
	if ag == nil {
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	if !item.IsManual && !consumeQuota(aq.ctx, cfg.Backend, cfg.InternalKey, item.TenantID) {
		atomic.AddInt64(&aq.processedCount, 1)
		catcher.Info("skipping automatic analysis: the tenant has used its AI allowance for today", map[string]any{
			"process": "plugin_com.utmstack.soc-ai",
			"tenant":  item.TenantID,
			"alert":   alertFields.Id,
		})
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

	// The deterministic score is never a choice the model makes — it's always
	// wanted, so fetching it never needs to cost an LLM round trip. Go calls
	// the same scoring tool directly here, same as the model would have.
	scoreArgs, _ := json.Marshal(map[string]string{"alert_id": alertFields.Id})
	scoreText, scoreIsErr, scoreErr := ag.Broker().Call(aq.ctx, "alerts.score", scoreArgs)
	if scoreErr != nil || scoreIsErr {
		_ = catcher.Error("alert scoring failed, escalating straight to full triage", scoreErr, map[string]any{
			"process": "plugin_com.utmstack.soc-ai", "id": alertFields.Id,
		})
		aq.runEscalatedTriage(ag, cfg, alertFields.Id, content, "the deterministic score was unavailable: "+truncateForReuse(scoreText))
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	fastInput := "ALERT:\n" + content + "\n\nDETERMINISTIC SCORE:\n" + scoreText

	reply, err := ag.QuickComplete(aq.ctx, agent.FastTriagePrompt(), fastInput)
	if err != nil {
		atomic.AddInt64(&aq.errorCount, 1)
		_ = catcher.Error("fast triage completion failed", err, map[string]any{"process": "plugin_com.utmstack.soc-ai", "id": alertFields.Id})
		return
	}
	reply = strings.TrimSpace(reply)

	if reason, escalate := strings.CutPrefix(reply, agent.EscalateSentinel); escalate {
		reason = truncateForReuse(strings.TrimSpace(reason))
		aq.runEscalatedTriage(ag, cfg, alertFields.Id, content, "DETERMINISTIC SCORE:\n"+scoreText+"\n\nA first pass found this unclear: "+reason)
		atomic.AddInt64(&aq.processedCount, 1)
		return
	}

	// reply IS the note line — write it directly. No LLM call needed to
	// confirm a write that already succeeded or failed; we just log it.
	// Capped in case the model ignored the one-line instruction: this note
	// is read back as "prior verdicts" context by future triage runs on the
	// same alert name, so an oversized one would keep inflating every later
	// prompt it's included in, not just this one.
	reply = truncateForReuse(reply)
	notesArgs, _ := json.Marshal(map[string]string{"alert_id": alertFields.Id, "notes": reply})
	if _, isErr, err := ag.Broker().Call(aq.ctx, "alerts.update_notes", notesArgs); err != nil || isErr {
		atomic.AddInt64(&aq.errorCount, 1)
		_ = catcher.Error("failed to write fast-triage note", err, map[string]any{"process": "plugin_com.utmstack.soc-ai", "id": alertFields.Id})
		return
	}

	atomic.AddInt64(&aq.processedCount, 1)
}

// runEscalatedTriage is the fallback path for an alert the fast pass
// couldn't confidently classify: a full agentic run with investigation
// tools, anchored on the same score (or the reason it's missing) so the
// escalated agent never needs to re-fetch it as a tool call either.
func (aq *AlertQueue) runEscalatedTriage(ag *agent.Agent, cfg *config.Config, alertID, alertContent, scoreContext string) {
	task := agent.RunTask{
		System:        agent.TriagePrompt(),
		Input:         "Triage this alert and record your assessment as a note.\n\nALERT:\n" + alertContent + "\n\n" + scoreContext,
		EnabledGroups: cfg.TriageCapabilities(),
		// The assessment note is the triage output channel — always permitted.
		AlwaysAllow: []string{"alerts.update_notes"},
		MaxIters:    cfg.MaxToolIterations,
	}

	if _, err := ag.Run(aq.ctx, task, nil); err != nil {
		atomic.AddInt64(&aq.errorCount, 1)
		_ = catcher.Error("escalated agent triage failed", err, map[string]any{"process": "plugin_com.utmstack.soc-ai", "id": alertID})
	}
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
