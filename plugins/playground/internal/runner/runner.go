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

type RunState struct {
	Active bool
	UUID   string
}

type RunResult struct {
	UUID     string          `json:"uuid"`
	Event    json.RawMessage `json:"event,omitempty"`
	Alert    json.RawMessage `json:"alert,omitempty"`
	TimedOut bool            `json:"timedOut"`
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

	alert, err := r.awaitAlert(ctx, ws.resultingAlert, log.Id)
	if err != nil {
		return nil, err
	}

	return &RunResult{UUID: log.Id, Event: event, Alert: alert}, nil
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
	return r.pollUntil(ctx, r.eventTimeout, func() json.RawMessage {
		return findByID(path, id)
	})
}

func (r *Runner) awaitAlert(ctx context.Context, path, eventID string) (json.RawMessage, error) {
	result, _, err := r.pollUntil(ctx, r.alertGrace, func() json.RawMessage {
		return findAlertByEventID(path, eventID)
	})
	return result, err
}

func (r *Runner) pollUntil(ctx context.Context, deadline time.Duration, probe func() json.RawMessage) (json.RawMessage, bool, error) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	end := time.Now().Add(deadline)
	for {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		case <-ticker.C:
		}
		if result := probe(); result != nil {
			return result, false, nil
		}
		if !time.Now().Before(end) {
			return nil, true, nil
		}
	}
}
