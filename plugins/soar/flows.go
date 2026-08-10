package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

const (
	flowExt         = ".yaml"
	enabledFileName = "enabled.yaml"
	systemSubdir    = "system"
	userSubdir      = "user"
	reloadInterval  = 60 * time.Second
)

func flowDirs() (systemDir, userRoot string) {
	root := envOr("SOAR_FLOWS_DIR", filepath.Join(plugins.WorkDir, "soar"))
	return filepath.Join(root, systemSubdir), filepath.Join(root, userSubdir)
}

type condition struct {
	Operator string `yaml:"operator"`
	Field    string `yaml:"field"`
	Value    any    `yaml:"value"`
}

type flow struct {
	RelPath    string
	Conditions []condition
	tenant     string
}

var (
	flowsMu   sync.RWMutex
	flows     []flow
	enabledBy map[string]map[string]bool
	ownedBy   map[string]map[string]bool
)

func activeFlowsFor(tenant string) []flow {
	flowsMu.RLock()
	defer flowsMu.RUnlock()

	on, owned := enabledBy[tenant], ownedBy[tenant]

	out := make([]flow, 0, len(flows))
	for _, f := range flows {
		if !on[f.RelPath] {
			continue
		}
		switch {
		case f.tenant == tenant:
			out = append(out, f)
		case f.tenant == "" && !owned[f.RelPath]:
			out = append(out, f)
		}
	}
	return out
}

func loadFlows(systemDir, userRoot string) {
	out := make([]flow, 0, 128)
	on := map[string]map[string]bool{}
	owned := map[string]map[string]bool{}

	system := readFlowDir(systemDir, systemDir, "")
	if len(system) == 0 {
		_ = catcher.Error("soar: no shipped flows found; everything a tenant enables from the catalog is inert", nil,
			map[string]any{"dir": systemDir})
	}
	out = append(out, system...)

	entries, _ := os.ReadDir(userRoot)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		tenant := e.Name()
		dir := filepath.Join(userRoot, tenant)

		on[tenant] = readEnabled(dir)
		own := map[string]bool{}
		for _, f := range readFlowDir(dir, dir, tenant) {
			own[f.RelPath] = true
			out = append(out, f)
		}
		owned[tenant] = own
	}

	flowsMu.Lock()
	flows, enabledBy, ownedBy = out, on, owned
	flowsMu.Unlock()
}

func readFlowDir(dir, base, tenant string) []flow {
	var out []flow
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, flowExt) {
			return nil
		}
		if filepath.Base(path) == enabledFileName {
			return nil
		}
		rel, rerr := filepath.Rel(base, path)
		if rerr != nil {
			return nil
		}
		conds, cerr := parseConditions(path)
		if cerr != nil {
			return nil
		}
		if len(conds) == 0 {
			_ = catcher.Error("soar: ignoring a flow with no conditions — it would match every alert", nil,
				map[string]any{"path": path})
			return nil
		}
		out = append(out, flow{RelPath: filepath.ToSlash(rel), Conditions: conds, tenant: tenant})
		return nil
	})
	return out
}

func readEnabled(dir string) map[string]bool {
	data, err := os.ReadFile(filepath.Join(dir, enabledFileName))
	if err != nil {
		return map[string]bool{}
	}
	var content struct {
		Enabled []string `yaml:"enabled"`
	}
	if yaml.Unmarshal(data, &content) != nil {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(content.Enabled))
	for _, rel := range content.Enabled {
		out[rel] = true
	}
	return out
}

func watchFlows(ctx context.Context, systemDir, userRoot string) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		_ = catcher.Error("soar: fsnotify unavailable; periodic reload only", err, nil)
		w = nil
	} else {
		defer func() { _ = w.Close() }()
		addWatches(w, systemDir)
		addWatches(w, userRoot)
	}

	ticker := time.NewTicker(reloadInterval)
	defer ticker.Stop()
	var events chan fsnotify.Event
	if w != nil {
		events = w.Events
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-events:
			loadFlows(systemDir, userRoot)
		case <-ticker.C:
			if w != nil {
				addWatches(w, systemDir)
				addWatches(w, userRoot)
			}
			loadFlows(systemDir, userRoot)
		}
	}
}

func addWatches(w *fsnotify.Watcher, root string) {
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			_ = w.Add(path)
		}
		return nil
	})
}

func matches(alertJSON string, conds []condition) bool {
	for _, c := range conds {
		actual := gjson.Get(alertJSON, strings.ReplaceAll(c.Field, ".keyword", "")).String()
		switch c.Operator {
		case "IS":
			if actual != toString(c.Value) {
				return false
			}
		case "IS_ONE_OF":
			list := toStringList(c.Value)
			if len(list) == 0 || !slices.Contains(list, actual) {
				return false
			}
		case "IS_NOT_ONE_OF":
			if slices.Contains(toStringList(c.Value), actual) {
				return false
			}
		case "IS_NOT":
			if actual == toString(c.Value) {
				return false
			}
		case "CONTAINS":
			if !strings.Contains(actual, toString(c.Value)) {
				return false
			}
		case "NOT_CONTAINS":
			if strings.Contains(actual, toString(c.Value)) {
				return false
			}
		case "START_WITH":
			if !strings.HasPrefix(actual, toString(c.Value)) {
				return false
			}
		case "NOT_START_WITH":
			if strings.HasPrefix(actual, toString(c.Value)) {
				return false
			}
		case "ENDS_WITH":
			if !strings.HasSuffix(actual, toString(c.Value)) {
				return false
			}
		case "NOT_ENDS_WITH":
			if strings.HasSuffix(actual, toString(c.Value)) {
				return false
			}
		case "EXISTS":
			if !gjson.Get(alertJSON, strings.ReplaceAll(c.Field, ".keyword", "")).Exists() {
				return false
			}
		case "NOT_EXISTS":
			if gjson.Get(alertJSON, strings.ReplaceAll(c.Field, ".keyword", "")).Exists() {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func toStringList(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, toString(it))
	}
	return out
}

func toString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), `"`)
	}
}

func parseConditions(path string) ([]condition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var list []struct {
		Conditions []condition `yaml:"conditions"`
	}
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fs.ErrNotExist
	}
	return list[0].Conditions, nil
}
