package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/protobuf/encoding/protojson"
)

type StopReason string

const (
	StopReasonQuiet    StopReason = "quiet"
	StopReasonHardCap  StopReason = "hard_cap"
	StopReasonNoAlerts StopReason = "no_alerts"
)

type RunState struct {
	Active bool
	UUID   string
}

type RunResult struct {
	UUID       string            `json:"uuid"`
	Event      json.RawMessage   `json:"event,omitempty"`
	Alerts     []json.RawMessage `json:"alerts,omitempty"`
	TimedOut   bool              `json:"timedOut"`
	StopReason StopReason        `json:"stopReason,omitempty"`
}

type Runner struct {
	workDir      string
	eventTimeout time.Duration
	alertGrace   time.Duration
	pollInterval time.Duration

	mu     sync.Mutex
	active bool
	uuid   string
}

func New(workDir string, eventTimeout, alertGrace, pollInterval time.Duration) *Runner {
	return &Runner{
		workDir:      workDir,
		eventTimeout: eventTimeout,
		alertGrace:   alertGrace,
		pollInterval: pollInterval,
	}
}

func (r *Runner) WorkDir() string { return r.workDir }

func (r *Runner) State() RunState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return RunState{Active: r.active, UUID: r.uuid}
}

func (r *Runner) Lock() { r.mu.Lock() }

func (r *Runner) Unlock() { r.mu.Unlock() }

func (r *Runner) Run(ctx context.Context, log *plugins.Log) (*RunResult, error) {
	release := r.acquireRun(log)
	defer release()

	ws := r.workspace()
	if err := ws.reset(); err != nil {
		return nil, err
	}
	if err := ws.submitLog(log); err != nil {
		return nil, err
	}

	event, timedOut, err := r.awaitEvent(ctx, ws.resultingLog, log.Id)
	if err != nil {
		return nil, err
	}
	if timedOut {
		return &RunResult{UUID: log.Id, TimedOut: true}, nil
	}

	alerts, stop, err := r.awaitAlerts(ctx, ws.resultingAlert, log.Id)
	if err != nil {
		return nil, err
	}

	return &RunResult{
		UUID:       log.Id,
		Event:      event,
		Alerts:     alerts,
		StopReason: stop,
	}, nil
}

func (r *Runner) acquireRun(log *plugins.Log) func() {
	r.mu.Lock()
	if log.Id == "" {
		log.Id = uuid.NewString()
	}
	r.active = true
	r.uuid = log.Id
	return func() {
		r.active = false
		r.uuid = ""
		r.mu.Unlock()
	}
}

type workspace struct {
	inputDir       string
	resultingLog   string
	resultingAlert string
}

func (r *Runner) workspace() workspace {
	return workspace{
		inputDir:       filepath.Join(r.workDir, "input"),
		resultingLog:   filepath.Join(r.workDir, "output", "resulting_log.json"),
		resultingAlert: filepath.Join(r.workDir, "output", "resulting_alert.json"),
	}
}

func (ws workspace) reset() error {
	if err := DeleteFilesInDir(ws.inputDir); err != nil {
		return fmt.Errorf("clear input dir: %w", err)
	}
	if err := TruncateFile(ws.resultingLog); err != nil {
		return fmt.Errorf("truncate resulting_log: %w", err)
	}
	if err := TruncateFile(ws.resultingAlert); err != nil {
		return fmt.Errorf("truncate resulting_alert: %w", err)
	}
	return nil
}

func (ws workspace) submitLog(log *plugins.Log) error {
	bytes, err := protojson.Marshal(log)
	if err != nil {
		return fmt.Errorf("marshal log: %w", err)
	}
	if err := os.MkdirAll(ws.inputDir, 0755); err != nil {
		return fmt.Errorf("ensure input dir: %w", err)
	}
	dst := filepath.Join(ws.inputDir, log.Id+".json")
	if err := os.WriteFile(dst, bytes, 0644); err != nil {
		return fmt.Errorf("write input file: %w", err)
	}
	return nil
}

func (r *Runner) awaitEvent(ctx context.Context, path, id string) (json.RawMessage, bool, error) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	end := time.Now().Add(r.eventTimeout)
	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		case <-ticker.C:
		}
		if result := findByID(path, id); result != nil {
			return result, false, nil
		}
		if !time.Now().Before(end) {
			return nil, true, nil
		}
	}
}

func (r *Runner) awaitAlerts(ctx context.Context, path, eventID string) ([]json.RawMessage, StopReason, error) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	var alerts []json.RawMessage
	seenIDs := make(map[string]struct{})

	now := time.Now()
	hardDeadline := now.Add(r.eventTimeout)
	quietDeadline := now.Add(r.alertGrace)

	for {
		select {
		case <-ctx.Done():
			return alerts, "", ctx.Err()
		case <-ticker.C:
		}

		newAlerts := findAlertsByEventID(path, eventID, seenIDs)
		if len(newAlerts) > 0 {
			alerts = append(alerts, newAlerts...)
			quietDeadline = time.Now().Add(r.alertGrace)
		}

		now = time.Now()

		if !now.Before(hardDeadline) {
			if len(alerts) == 0 {
				return nil, StopReasonNoAlerts, nil
			}
			return alerts, StopReasonHardCap, nil
		}

		if !now.Before(quietDeadline) {
			if len(alerts) == 0 {
				return nil, StopReasonNoAlerts, nil
			}
			return alerts, StopReasonQuiet, nil
		}
	}
}
