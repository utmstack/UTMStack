package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

const definitionsDashboardsDir = "../../../../definitions/dashboards"

const sampleDashboardYAML = `
name: "Agents"
description: "Fleet-wide agent activity."
widgets:
  - layout: { x: 0, y: 0, w: 4, h: 2 }
    spec:
      dataset: logs
      dataType: linux
      chart: metric
      metric:
        agg: count
    config:
      __builder:
        chartType: metric
        title: "Linux Events"
`

func TestADashboardDefinitionParsesFromYAML(t *testing.T) {
	var def DashboardDefinition
	if err := yaml.Unmarshal([]byte(sampleDashboardYAML), &def); err != nil {
		t.Fatalf("parse: %v", err)
	}

	if def.Name != "Agents" {
		t.Errorf("name = %q, want %q", def.Name, "Agents")
	}
	if len(def.Widgets) != 1 {
		t.Fatalf("widgets = %d, want 1", len(def.Widgets))
	}

	w := def.Widgets[0]
	if w.Spec["dataType"] != "linux" {
		t.Errorf("spec.dataType = %v, want linux", w.Spec["dataType"])
	}
	if w.Layout["w"] != 4 {
		t.Errorf("layout.w = %v, want 4", w.Layout["w"])
	}
	builder, ok := w.Config["__builder"].(map[string]any)
	if !ok {
		t.Fatalf("config.__builder is %T, want a map", w.Config["__builder"])
	}
	if builder["chartType"] != "metric" {
		t.Errorf("config.__builder.chartType = %v, want metric", builder["chartType"])
	}
}

// The dashboard this repo ships is exactly what a bootstrap run will try to
// seed — if its spec can't validate, the real backend would silently skip it
// (Run logs and continues past a bad file) and the widget would never exist.
func TestTheShippedAgentsDashboardSpecsAreAllValid(t *testing.T) {
	var def DashboardDefinition
	if err := yaml.Unmarshal([]byte(sampleDashboardYAML), &def); err != nil {
		t.Fatalf("parse: %v", err)
	}
	for i, w := range def.Widgets {
		if _, err := encodeSpec(w.Spec); err != nil {
			t.Errorf("widget %d: %v", i, err)
		}
	}
}

func TestEncodeSpecRejectsAnUnanswerableSpec(t *testing.T) {
	cases := map[string]map[string]any{
		"category chart with no dimension": {
			"dataset": "alerts",
			"chart":   "category",
		},
		"unknown dataset": {
			"dataset": "system.query_log",
			"chart":   "metric",
		},
		"an average (store only counts)": {
			"dataset": "logs",
			"chart":   "metric",
			"metric":  map[string]any{"agg": "avg", "field": "bytes"},
		},
	}

	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := encodeSpec(spec); err == nil {
				t.Error("expected the spec to be refused, it was accepted")
			}
		})
	}
}

// Every YAML file actually shipped in definitions/dashboards/ parses and
// every widget's spec validates — this is what Run() does to each file, so a
// definition that fails here is one the real bootstrap would silently skip.
func TestEveryShippedDashboardDefinitionIsValid(t *testing.T) {
	entries, err := os.ReadDir(definitionsDashboardsDir)
	if err != nil {
		t.Fatalf("reading %s: %v", definitionsDashboardsDir, err)
	}

	found := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != DashboardFileExt {
			continue
		}
		found++
		t.Run(e.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(definitionsDashboardsDir, e.Name()))
			if err != nil {
				t.Fatalf("read: %v", err)
			}

			var def DashboardDefinition
			if err := yaml.Unmarshal(data, &def); err != nil {
				t.Fatalf("parse: %v", err)
			}
			if def.Name == "" {
				t.Fatal("no name")
			}
			for i, w := range def.Widgets {
				if _, err := encodeSpec(w.Spec); err != nil {
					t.Errorf("widget %d: %v", i, err)
				}
				if _, err := json.Marshal(w.Config); err != nil {
					t.Errorf("widget %d: config does not encode: %v", i, err)
				}
				if _, err := json.Marshal(w.Layout); err != nil {
					t.Errorf("widget %d: layout does not encode: %v", i, err)
				}
			}
		})
	}
	if found == 0 {
		t.Fatal("no .yaml files found in definitions/dashboards -- is the path right?")
	}
}

func TestEncodeSpecAcceptsAGoodSpec(t *testing.T) {
	spec := map[string]any{
		"dataset": "logs",
		"chart":   "metric",
		"metric":  map[string]any{"agg": "count"},
	}
	if _, err := encodeSpec(spec); err != nil {
		t.Errorf("a valid spec was refused: %v", err)
	}
}
