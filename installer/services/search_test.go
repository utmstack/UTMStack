package services

import (
	"encoding/json"
	"testing"
)

func TestLogIndexMappingsValidJSON(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(logIndexMappings), &m); err != nil {
		t.Fatalf("logIndexMappings is not valid JSON: %v", err)
	}
	ev, _ := m["event"].(map[string]any)
	if ev["type"] != "flat_object" {
		t.Fatalf("event should be flat_object (OpenSearch): %v", m["event"])
	}
	ct, _ := m["controls"].(map[string]any)
	if ct["type"] != "keyword" {
		t.Fatalf("controls should be keyword: %v", m["controls"])
	}
}

func TestNewBagMappingsValidJSON(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(newBagMappings), &m); err != nil {
		t.Fatalf("newBagMappings is not valid JSON: %v", err)
	}
	if ev, _ := m["event"].(map[string]any); ev["type"] != "flat_object" {
		t.Fatalf("event should be flat_object (OpenSearch): %v", m["event"])
	}
	if ct, _ := m["controls"].(map[string]any); ct["type"] != "keyword" {
		t.Fatalf("controls should be keyword: %v", m["controls"])
	}
}

func TestAlertIndexMappingsValidJSON(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(alertIndexMappings), &m); err != nil {
		t.Fatalf("alertIndexMappings is not valid JSON: %v", err)
	}
	// lastEvent.event must be flat_object
	le, ok := m["lastEvent"].(map[string]any)
	if !ok {
		t.Fatalf("lastEvent should be an object: %v", m["lastEvent"])
	}
	ev, _ := le["properties"].(map[string]any)
	if ev == nil {
		t.Fatalf("lastEvent.properties missing")
	}
	if evType, _ := ev["event"].(map[string]any); evType["type"] != "flat_object" {
		t.Fatalf("lastEvent.event should be flat_object (OpenSearch): %v", ev["event"])
	}
	if ct, _ := ev["controls"].(map[string]any); ct["type"] != "keyword" {
		t.Fatalf("lastEvent.controls should be keyword: %v", ev["controls"])
	}
	// events[*].event must be flat_object
	ea, ok := m["events"].(map[string]any)
	if !ok {
		t.Fatalf("events should be an object: %v", m["events"])
	}
	ev2, _ := ea["properties"].(map[string]any)
	if ev2 == nil {
		t.Fatalf("events.properties missing")
	}
	if ev2Type, _ := ev2["event"].(map[string]any); ev2Type["type"] != "flat_object" {
		t.Fatalf("events.event should be flat_object (OpenSearch): %v", ev2["event"])
	}
}
