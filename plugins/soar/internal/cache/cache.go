package cache

import (
	"context"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/UTMStack/plugins/soar/internal/client"
)

const refreshInterval = 30 * time.Second

type CachedRule struct {
	client.Rule
	AgentNames []string
}

var (
	mu      sync.RWMutex
	rules   []CachedRule
	byPath  = map[string]CachedRule{}
	backend *client.BackendClient
)

func Init(c *client.BackendClient) { backend = c }

func ActiveRules() []CachedRule {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]CachedRule, len(rules))
	copy(out, rules)
	return out
}

// GetRule returns the cached active rule for relPath. Missing means the flow was
// disabled or deleted between create-execution and dispatch — caller should skip.
func GetRule(relPath string) (CachedRule, bool) {
	mu.RLock()
	defer mu.RUnlock()
	r, ok := byPath[relPath]
	return r, ok
}

func refresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fetched, err := backend.ListActiveRules(ctx)
	if err != nil {
		_ = catcher.Error("failed to refresh rule cache", err, nil)
		return
	}

	out := make([]CachedRule, 0, len(fetched))
	for _, r := range fetched {
		agents, err := backend.ListAgentsByPlatform(ctx, r.AgentPlatform)
		if err != nil {
			_ = catcher.Error("failed to load agents for rule", err, map[string]any{"rule": r.RelPath})
			continue
		}
		out = append(out, CachedRule{Rule: r, AgentNames: agents})
	}

	idx := make(map[string]CachedRule, len(out))
	for _, r := range out {
		idx[r.RelPath] = r
	}

	mu.Lock()
	rules = out
	byPath = idx
	mu.Unlock()
}

func RefreshLoop() {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in RefreshLoop", nil, map[string]any{"panic": r})
		}
	}()
	refresh()
	t := time.NewTicker(refreshInterval)
	defer t.Stop()
	for range t.C {
		refresh()
	}
}
