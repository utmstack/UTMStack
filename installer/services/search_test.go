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
	if ev["type"] != "flattened" {
		t.Fatalf("event should be flattened: %v", m["event"])
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
	if ev, _ := m["event"].(map[string]any); ev["type"] != "flattened" {
		t.Fatalf("event should be flattened: %v", m["event"])
	}
	if ct, _ := m["controls"].(map[string]any); ct["type"] != "keyword" {
		t.Fatalf("controls should be keyword: %v", m["controls"])
	}
}
