package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

type SourceStats struct {
	Received      uint64    `json:"received"`
	Persisted     uint64    `json:"persisted"`
	PersistFailed uint64    `json:"persist_failed"`
	LastEventAt   time.Time `json:"last_event_at"`
}

type statsRegistry struct {
	mu      sync.Mutex
	sources map[string]*SourceStats
}

var registry = &statsRegistry{sources: make(map[string]*SourceStats)}

func RecordSourceEvent(source string, err error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	s, ok := registry.sources[source]
	if !ok {
		s = &SourceStats{}
		registry.sources[source] = s
	}
	s.Received++
	if err != nil {
		s.PersistFailed++
		return
	}
	s.Persisted++
	s.LastEventAt = time.Now()
}

func snapshotStats() map[string]SourceStats {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	out := make(map[string]SourceStats, len(registry.sources))
	for name, s := range registry.sources {
		out[name] = *s
	}
	return out
}

var statusFile = filepath.Join(fs.GetExecutablePath(), "agent_status.json")

const statusFlushInterval = 15 * time.Second

type StatusSnapshot struct {
	GeneratedAt time.Time              `json:"generated_at"`
	Sources     map[string]SourceStats `json:"sources"`
}

func flushStatus() {
	snapshot := StatusSnapshot{GeneratedAt: time.Now(), Sources: snapshotStats()}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		utils.Logger.ErrorF("status: error encoding snapshot: %v", err)
		return
	}
	if err := os.WriteFile(statusFile, data, 0o600); err != nil {
		utils.Logger.ErrorF("status: error writing %s: %v", statusFile, err)
	}
}

func RunStatusReporter(ctx context.Context) {
	flushStatus() // an initial snapshot immediately, not after a full interval
	ticker := time.NewTicker(statusFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			flushStatus()
			return
		case <-ticker.C:
			flushStatus()
		}
	}
}

func ReadStatus() (*StatusSnapshot, error) {
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return nil, err
	}
	var snapshot StatusSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}
