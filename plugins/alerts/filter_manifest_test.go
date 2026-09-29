package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type filterContract struct {
	Technology string    `json:"technology"`
	Filters    []string  `json:"filters"`
	Rules      []string  `json:"rules"`
	Fixtures   []Fixture `json:"fixtures"`
}

func loadFilterContracts(t *testing.T) []filterContract {
	t.Helper()
	paths, err := filepath.Glob("testdata/filter-contracts/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no filter contract manifests")
	}
	var out []filterContract
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var manifest filterContract
		if err := json.Unmarshal(data, &manifest); err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		if manifest.Technology == "" || len(manifest.Filters) == 0 || len(manifest.Fixtures) == 0 {
			t.Fatalf("%s: technology, filters and fixtures are required", p)
		}
		out = append(out, manifest)
	}
	return out
}
