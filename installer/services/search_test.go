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

// TestLogIndexMappingsCanonical pins the standardized-field mapping contract:
// every canonical field is explicitly typed (no per-index inference drift),
// Side objects are dynamic:false, ip fields carry a .keyword sub-field, and
// only utm.* (dynamic:true) + the event.* flat_object bag are left to inference.
func TestLogIndexMappingsCanonical(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(logIndexMappings), &m); err != nil {
		t.Fatalf("logIndexMappings is not valid JSON: %v", err)
	}
	getObj := func(path ...string) (map[string]any, bool) {
		var cur any = m
		for _, p := range path {
			cm, ok := cur.(map[string]any)
			if !ok {
				return nil, false
			}
			cur = cm[p]
		}
		cm, ok := cur.(map[string]any)
		return cm, ok
	}
	typ := func(v any) string {
		cm, _ := v.(map[string]any)
		s, _ := cm["type"].(string)
		return s
	}
	// Side objects must be pinned (no dynamic inference). Fields nest under "properties".
	for _, side := range []string{"origin", "target"} {
		o, ok := getObj(side)
		if !ok {
			t.Fatalf("%s missing from log mapping", side)
		}
		if o["dynamic"] != false {
			t.Fatalf("%s must be dynamic:false (no inference), got %v", side, o["dynamic"])
		}
		props, _ := o["properties"].(map[string]any)
		if props == nil {
			t.Fatalf("%s.properties missing", side)
		}
		if typ(props["ip"]) != "ip" {
			t.Fatalf("%s.ip must be type ip, got %v", side, typ(props["ip"]))
		}
		// ip must carry a .keyword sub-field (coordinate-map terms agg).
		ipf, _ := props["ip"].(map[string]any)
		fields, _ := ipf["fields"].(map[string]any)
		if fields == nil {
			t.Fatalf("%s.ip must have a .keyword sub-field (maps bucket on it)", side)
		}
		if _, ok := fields["keyword"]; !ok {
			t.Fatalf("%s.ip.keyword sub-field missing", side)
		}
		if _, ok := props["geolocation"].(map[string]any); !ok {
			t.Fatalf("%s.geolocation missing", side)
		}
	}
	// geolocation types pinned.
	geoObj, _ := getObj("origin")
	geo, _ := geoObj["properties"].(map[string]any)["geolocation"].(map[string]any)
	geoProps, _ := geo["properties"].(map[string]any)
	if geo["dynamic"] != false {
		t.Fatalf("origin.geolocation must be dynamic:false, got %v", geo["dynamic"])
	}
	if typ(geoProps["latitude"]) != "double" || typ(geoProps["longitude"]) != "double" {
		t.Fatalf("origin.geolocation.lat/long must be double")
	}
	if typ(geoProps["asn"]) != "long" {
		t.Fatalf("origin.geolocation.asn must be long (canonical), got %v", typ(geoProps["asn"]))
	}
	// event bag stays flat_object.
	if typ(m["event"]) != "flat_object" {
		t.Fatalf("event must be flat_object, got %v", typ(m["event"]))
	}
	// utm.* is the ONLY bag whose leaves are not standardized Event fields, so
	// the installer pins them explicitly: object + dynamic:false + known keys
	// keyword. A bare {"type":"object"} inherits the template root's
	// dynamic:false (the installer wraps logIndexMappings in
	// "mappings":{"dynamic":false,"properties":...}) and the leaves are then
	// never indexed — terms aggs return empty and rule correlations on utm.*
	// match nothing. Pinning keeps utm.X keyword (and therefore aggregatable
	// on the bare path) on every index generation.
	utm, ok := getObj("utm")
	if !ok {
		t.Fatalf("utm missing from log mapping")
	}
	if typ(utm) != "object" {
		t.Fatalf("utm must be an object, got %v", typ(utm))
	}
	if dyn, hasDyn := utm["dynamic"]; !hasDyn || dyn != false {
		t.Fatalf("utm must be dynamic:false (closed pin), got %v", utm)
	}
	utmProps, ok := utm["properties"].(map[string]any)
	if !ok || len(utmProps) < 100 {
		t.Fatalf("utm expected to pin the known keys (>=100), got: %v", utmProps)
	}
	// chart-referenced leaves must be keyword (aggregatable on the bare path).
	for _, leaf := range []string{"eventCode", "providerName", "Workload", "task", "userDomain"} {
		leafNode, ok := utmProps[leaf]
		if !ok {
			t.Fatalf("utm.%s missing from pin", leaf)
		}
		if typ(leafNode) != "keyword" {
			t.Fatalf("utm.%s must be keyword, got %v", leaf, typ(leafNode))
		}
	}
	// the one non-keyword leaf: an ISO-8601 timestamp pinned date (aggregatable as-is).
	expProps := utmProps["responseElements"].(map[string]any)["properties"].(map[string]any)
	expCreds := expProps["credentials"].(map[string]any)["properties"].(map[string]any)
	if typ(expCreds["expiration"]) != "date" {
		t.Fatalf("utm.responseElements.credentials.expiration must be date, got %v", typ(expCreds["expiration"]))
	}
	// log severity is text (string) — distinct from alert severity (long).
	if typ(m["severity"]) != "text" {
		t.Fatalf("log severity must be text, got %v", typ(m["severity"]))
	}
}

// TestAlertIndexMappingsCanonical pins the alert root mapping: numeric
// severity/impact, Side objects pinned, and embedded event copies (lastEvent,
// events[*]) reusing the flat_object event bag + typed Side objects.
func TestAlertIndexMappingsCanonical(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(alertIndexMappings), &m); err != nil {
		t.Fatalf("alertIndexMappings is not valid JSON: %v", err)
	}
	typ := func(v any) string {
		cm, _ := v.(map[string]any)
		s, _ := cm["type"].(string)
		return s
	}
	if typ(m["severity"]) != "long" {
		t.Fatalf("alert severity must be long (int), got %v", typ(m["severity"]))
	}
	for _, side := range []string{"adversary", "target"} {
		o, ok := m[side].(map[string]any)
		if !ok {
			t.Fatalf("alert %s missing", side)
		}
		if o["dynamic"] != false {
			t.Fatalf("alert %s must be dynamic:false, got %v", side, o["dynamic"])
		}
		props, _ := o["properties"].(map[string]any)
		if props == nil {
			t.Fatalf("alert %s.properties missing", side)
		}
		if typ(props["ip"]) != "ip" {
			t.Fatalf("alert %s.ip must be ip, got %v", side, typ(props["ip"]))
		}
	}
	// impact scores numeric.
	impObj, ok := m["impact"].(map[string]any)
	if !ok {
		t.Fatalf("alert impact missing")
	}
	imp, _ := impObj["properties"].(map[string]any)
	if typ(imp["confidentiality"]) != "long" {
		t.Fatalf("impact.confidentiality must be long, got %v", typ(imp["confidentiality"]))
	}
	// embedded event copies carry the flat_object event bag.
	for _, nested := range []string{"lastEvent", "events"} {
		obj, ok := m[nested].(map[string]any)
		if !ok {
			t.Fatalf("alert %s missing", nested)
		}
		props, _ := obj["properties"].(map[string]any)
		if props == nil {
			t.Fatalf("alert %s.properties missing", nested)
		}
		if typ(props["event"]) != "flat_object" {
			t.Fatalf("alert %s.event must be flat_object, got %v", nested, typ(props["event"]))
		}
	}
}

// TestNewBagRetroIsAdditive pins that the existing-index retro is conflict-safe:
// it adds the ip .keyword sub-field + event/controls bags, and never retypes an
// existing leaf (no base "type" on a Side object that would conflict).
func TestNewBagRetroIsAdditive(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(newBagMappings), &m); err != nil {
		t.Fatalf("newBagMappings is not valid JSON: %v", err)
	}
	// origin/target only add the ip sub-field (no dynamic flag flip, no retype).
	for _, side := range []string{"origin", "target"} {
		o, ok := m[side].(map[string]any)
		if !ok {
			t.Fatalf("retro %s missing", side)
		}
		props, _ := o["properties"].(map[string]any)
		ip, ok := props["ip"].(map[string]any)
		if !ok {
			t.Fatalf("retro %s.ip missing", side)
		}
		if typ := func() string { s, _ := ip["type"].(string); return s }(); typ != "ip" {
			t.Fatalf("retro %s.ip must be ip, got %v", side, typ)
		}
		if _, ok := ip["fields"].(map[string]any); !ok {
			t.Fatalf("retro %s.ip must add the .keyword sub-field", side)
		}
	}
	if ev, _ := m["event"].(map[string]any); ev["type"] != "flat_object" {
		t.Fatalf("retro event must be flat_object")
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
