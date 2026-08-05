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
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

const (
	flowExt        = ".yaml"
	disabledSuffix = ".disabled"
	reloadInterval = 60 * time.Second
)

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
	flowsMu sync.RWMutex
	flows   []flow
)

func activeFlowsFor(tenant string) []flow {
	flowsMu.RLock()
	defer flowsMu.RUnlock()

	out := make([]flow, 0, len(flows))
	for _, f := range flows {
		if f.tenant == "" || f.tenant == tenant {
			out = append(out, f)
		}
	}
	return out
}

func loadFlows(root string) {
	enabled := map[string]flow{}
	disabled := map[string]bool{}
	{
		dir := filepath.Join(root, "user")
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			logical, off := path, false
			if strings.HasSuffix(path, disabledSuffix) {
				logical, off = strings.TrimSuffix(path, disabledSuffix), true
			}
			if !strings.HasSuffix(logical, flowExt) {
				return nil
			}
			rel, err := filepath.Rel(dir, logical)
			if err != nil {
				return nil
			}
			relPath := filepath.ToSlash(rel)
			i := strings.Index(relPath, "/")
			if i <= 0 {
				return nil // a file naming no tenant belongs to nobody
			}
			tenant := relPath[:i]
			relPath = strings.TrimPrefix(relPath[i+1:], "system/")

			key := tenant + "\x00" + relPath
			if off {
				disabled[key] = true
				delete(enabled, key)
				return nil
			}
			conds, err := parseConditions(path)
			if err != nil {
				return nil
			}
			delete(disabled, key)
			enabled[key] = flow{RelPath: relPath, Conditions: conds, tenant: tenant}
			return nil
		})
	}

	out := make([]flow, 0, len(enabled))
	for key, f := range enabled {
		if !disabled[key] {
			out = append(out, f)
		}
	}
	flowsMu.Lock()
	flows = out
	flowsMu.Unlock()
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

func watchFlows(ctx context.Context, root string) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		_ = catcher.Error("soar: fsnotify unavailable; periodic reload only", err, nil)
		w = nil
	} else {
		defer func() { _ = w.Close() }()
		addWatches(w, root)
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
			loadFlows(root)
		case <-ticker.C:
			if w != nil {
				addWatches(w, root)
			}
			loadFlows(root)
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
