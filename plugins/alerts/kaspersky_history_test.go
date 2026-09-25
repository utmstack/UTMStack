package main

// Offline history requests use the real SDK and an isolated loopback mock.
// The mock evaluates only the term/not-term/time clauses asserted below.
import (
	"encoding/json"
	"fmt"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestKasperskySDKHistory(t *testing.T) {
	if os.Getenv("UTM_KASPERSKY_HISTORY_CHILD") != "1" {
		c := exec.Command(os.Args[0], "-test.run=^TestKasperskySDKHistory$")
		c.Env = append(os.Environ(), "UTM_KASPERSKY_HISTORY_CHILD=1")
		if b, e := c.CombinedOutput(); e != nil {
			t.Fatalf("isolated history: %v\n%s", e, b)
		}
		return
	}
	cfg, rules, cache := kaspConfig(t), kaspRules(t), plugins.NewCELCache("kasp-history")
	var history []string
	var terms, notTerms map[string]string
	var window time.Duration
	queries := 0
	mapping := map[string]any{"properties": map[string]any{}}
	props := mapping["properties"].(map[string]any)
	paths := []string{"dataSource", "log.endpointKeyType", "log.endpointKey", "origin.ip", "target.ip", "log.cat"}
	for name, r := range rules {
		if len(r.Correlation) > 0 {
			if kasperskyHistoryMarkers[name] == "" {
				t.Fatalf("history rule %s has no correlation marker", name)
			}
			paths = append(paths, "log.correlationCandidate."+kasperskyHistoryMarkers[name])
		}
	}
	for _, path := range paths {
		node := props
		parts := strings.Split(path, ".")
		for _, part := range parts[:len(parts)-1] {
			if node[part] == nil {
				node[part] = map[string]any{"properties": map[string]any{}}
			}
			node = node[part].(map[string]any)["properties"].(map[string]any)
		}
		node[parts[len(parts)-1]] = map[string]any{"type": "text", "fields": map[string]any{"keyword": map[string]any{"type": "keyword"}}}
	}
	props["@timestamp"] = map[string]any{"type": "date"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			_ = json.NewEncoder(w).Encode(map[string]any{"v11-log-antivirus-kaspersky-test": map[string]any{"mappings": mapping}})
			return
		}
		if r.URL.Path != "/v11-log-antivirus-kaspersky-*/_search" {
			t.Errorf("unexpected request %s", r.URL.Path)
			http.Error(w, "bad request", 400)
			return
		}
		queries++
		body, e := io.ReadAll(r.Body)
		if e != nil {
			t.Error(e)
			return
		}
		q := string(body)
		clauses := append(gjson.Get(q, "query.bool.filter").Array(), gjson.Get(q, "query.bool.must").Array()...)
		negatives := gjson.Get(q, "query.bool.must_not").Array()
		gotTerms := map[string]string{}
		gotNot := map[string]string{}
		cutoff := time.Time{}
		for _, clause := range clauses {
			if term := clause.Get("term"); term.Exists() {
				for field, value := range term.Map() {
					gotTerms[strings.TrimSuffix(field, ".keyword")] = value.Get("value").String()
				}
			} else if span := clause.Get("range"); span.Exists() {
				cutoff, e = time.Parse(time.RFC3339Nano, span.Get("@timestamp.gte").String())
				if e != nil {
					t.Error(e)
				}
			} else {
				t.Errorf("unsupported clause %s", clause.Raw)
			}
		}
		for _, clause := range negatives {
			if nested := clause.Get("bool.must"); nested.Exists() {
				if len(nested.Array()) != 1 {
					t.Error("unexpected negative bool")
				}
				clause = nested.Array()[0]
			}
			if term := clause.Get("term"); term.Exists() {
				for field, value := range term.Map() {
					gotNot[strings.TrimSuffix(field, ".keyword")] = value.Get("value").String()
				}
			} else {
				t.Errorf("unsupported negative %s", clause.Raw)
			}
		}
		same := func(a, b map[string]string) bool {
			if len(a) != len(b) {
				return false
			}
			for k, v := range a {
				if b[k] != v {
					return false
				}
			}
			return true
		}
		if !same(gotTerms, terms) || !same(gotNot, notTerms) {
			t.Errorf("scope mismatch: terms=%v negatives=%v", gotTerms, gotNot)
		}
		if delta := time.Since(cutoff) - window; delta < -2*time.Second || delta > 2*time.Second {
			t.Errorf("unexpected time cutoff %v", delta)
		}
		hits := []map[string]any{}
		for _, doc := range history {
			match := true
			for f, v := range gotTerms {
				if !gjson.Get(doc, f).Exists() || gjson.Get(doc, f).String() != v {
					match = false
				}
			}
			for f, v := range gotNot {
				if gjson.Get(doc, f).String() == v {
					match = false
				}
			}
			stamp, e := time.Parse(time.RFC3339Nano, gjson.Get(doc, "@timestamp").String())
			if e != nil || stamp.Before(cutoff) {
				match = false
			}
			if match {
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-antivirus-kaspersky-test", "_source": map[string]any{}})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "hits": map[string]any{"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits}})
	}))
	defer server.Close()
	if e := sdkos.Connect([]string{server.URL}, "", ""); e != nil {
		t.Fatal(e)
	}
	mutate := func(doc, path string, value any) string {
		var m map[string]any
		if e := json.Unmarshal([]byte(doc), &m); e != nil {
			t.Fatal(e)
		}
		kaspPut(m, path, value, value == nil)
		b, e := json.Marshal(m)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	cases := []struct {
		rule, fixture, within string
		count                 uint64
		cross                 bool
		extra                 map[string]string
	}{
		{"lateral_movement_indicators", "lateral_movement_indicators", "2h", 3, false, nil},
		{"suspicious_network_activity", "suspicious_network_activity", "30m", 5, false, map[string]string{"target.ip": "192.0.2.8"}},
		{"kaspersky_ransomware_behavior", "kaspersky_ransomware_behavior", "10m", 3, false, nil},
		{"data_exfiltration_attempts", "data_exfiltration_attempts", "30m", 5, false, map[string]string{"log.cat": "NetworkThreat"}},
	}
	fixtures := map[string]kaspFixture{}
	for _, f := range kaspFixtures(t) {
		fixtures[f.Name] = f
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			r := rules[tc.rule]
			if r == nil || len(r.Correlation) != 1 {
				t.Fatal("missing history")
			}
			search := r.Correlation[0]
			if search.Count != tc.count || search.Within != tc.within {
				t.Fatal("threshold/window changed")
			}
			var e error
			window, e = time.ParseDuration(tc.within)
			if e != nil {
				t.Fatal(e)
			}
			f := fixtures[tc.fixture]
			out := kaspParse(t, cfg, f.Raw, f.DataSource, cache)
			if ok, e := cache.Eval(r.Where, out); e != nil || !ok {
				t.Fatalf("raw trigger failed: %v %v", ok, e)
			}
			// The filter stores this marker and the rule counts it under the same name.
			marker := "log.correlationCandidate." + kasperskyHistoryMarkers[tc.rule]
			terms = map[string]string{"dataSource": "collector-test", "log.endpointKeyType": "ip", marker: "match"}
			notTerms = map[string]string{}
			if tc.cross {
				notTerms["log.endpointKey"] = "198.51.100.9"
			} else {
				terms["log.endpointKey"] = "198.51.100.9"
			}
			for k, v := range tc.extra {
				terms[k] = v
			}
			prior := mutate(out, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
			if tc.cross {
				prior = mutate(prior, "log.endpointKey", "other-endpoint")
			}
			check := func(name, doc string, count uint64, want bool) {
				t.Run(name, func(t *testing.T) {
					history = nil
					for i := uint64(0); i < count; i++ {
						history = append(history, doc)
					}
					ok, _, e := search.Execute(&out)
					if e != nil || ok != want {
						t.Fatalf("history %v want %v: %v", ok, want, e)
					}
				})
			}
			check("below_threshold", prior, tc.count-1, false)
			check("at_threshold", prior, tc.count, true)
			check("expired", mutate(prior, "@timestamp", time.Now().Add(-window-time.Minute).UTC().Format(time.RFC3339Nano)), tc.count, false)
			check("inside_window", mutate(prior, "@timestamp", time.Now().Add(-window+time.Minute).UTC().Format(time.RFC3339Nano)), tc.count, true)
			for _, field := range []string{"dataSource", "log.endpointKeyType"} {
				check("different_"+field, mutate(prior, field, "other"), tc.count, false)
			}
			if tc.cross {
				check("same_endpoint", mutate(prior, "log.endpointKey", "198.51.100.9"), tc.count, false)
			} else {
				check("different_endpoint", mutate(prior, "log.endpointKey", "other"), tc.count, false)
			}
			check("unrelated_population", mutate(prior, marker, nil), tc.count, false)
			for field := range tc.extra {
				check("different_"+field, mutate(prior, field, "other"), tc.count, false)
			}
			without := mutate(out, "log.endpointKey", nil)
			before := queries
			if _, _, e := search.Execute(&without); e == nil {
				t.Error("missing required placeholder accepted")
			}
			if queries != before {
				t.Error("missing placeholder executed query")
			}
		})
	}
}
