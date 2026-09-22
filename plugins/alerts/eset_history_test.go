package main

// Synthetic ESET JSON exercises the documented field contract, including
// explicitly labeled compatibility assumptions for non-exhaustive vendor values.
// Parsing uses the bounded offline model; CEL, mapping/placeholder resolution and
// threshold decisions use SDK1.1.31 against an isolated loopback OpenSearch mock.
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

func TestESETSDKHistory(t *testing.T) {
	if os.Getenv("UTM_ESET_HISTORY_CHILD") != "1" {
		command := exec.Command(os.Args[0], "-test.run=^TestESETSDKHistory$")
		command.Env = append(os.Environ(), "UTM_ESET_HISTORY_CHILD=1")
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("isolated history test: %v\n%s", err, out)
		}
		return
	}
	cfg, rules, cache := esetConfig(t), esetRules(t), plugins.NewCELCache("eset-history")
	const collector = "eset-test-relay"
	const endpoint = "20cc734b-379b-4021-8c7f-f21d1e470f29"
	var history []string
	queries := 0
	var expectedTerms map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			// Text-with-keyword and exact IP mappings both exercise SDK lookup.
			_, _ = io.WriteString(w, `{"v11-log-antivirus-esmc-eset-test":{"mappings":{"properties":{"@timestamp":{"type":"date"},"dataSource":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"target":{"properties":{"user":{"type":"text","fields":{"keyword":{"type":"keyword"}}}}},"log":{"properties":{"endpointKey":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"endpointKeyType":{"type":"keyword"},"correlationCandidate":{"properties":{"heuristicRemediation":{"type":"keyword"},"consoleAuthenticationFailure":{"type":"keyword"},"quarantineFailure":{"type":"keyword"}}}}}}}}`)
			return
		}
		if r.URL.Path != "/v11-log-antivirus-esmc-eset-*/_search" {
			t.Errorf("unexpected request %s", r.URL.Path)
			http.Error(w, "unsupported request", http.StatusBadRequest)
			return
		}
		queries++
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		query := string(body)
		clauses := append(gjson.Get(query, "query.bool.must").Array(), gjson.Get(query, "query.bool.filter").Array()...)
		if len(clauses) != len(expectedTerms)+1 {
			t.Errorf("expected exact identity/candidate scopes and time range, got %s", query)
		}
		expected := map[string]string{}
		for field, value := range expectedTerms {
			expected[field] = value
		}
		for _, clause := range clauses {
			if term := clause.Get("term"); term.Exists() {
				for field, value := range term.Map() {
					path := strings.TrimSuffix(field, ".keyword")
					want, exists := expected[path]
					if !exists || value.Get("value").String() != want {
						t.Errorf("unexpected history term %s", term.Raw)
					}
					delete(expected, path)
				}
			}
		}
		if len(expected) != 0 {
			t.Errorf("history omitted scopes: %v", expected)
		}
		hits := []map[string]any{}
		for _, doc := range history {
			match := true
			for _, clause := range clauses {
				if term := clause.Get("term"); term.Exists() {
					for field, value := range term.Map() {
						actual := gjson.Get(doc, strings.TrimSuffix(field, ".keyword"))
						if !actual.Exists() || actual.String() != value.Get("value").String() {
							match = false
						}
					}
				} else if span := clause.Get("range"); span.Exists() {
					for field, bounds := range span.Map() {
						stamp, err := time.Parse(time.RFC3339Nano, gjson.Get(doc, field).String())
						if err != nil {
							t.Error(err)
						}
						cutoff, err := time.Parse(time.RFC3339Nano, bounds.Get("gte").String())
						if err != nil {
							t.Error(err)
						}
						if field != "@timestamp" || !bounds.Get("gte").Exists() || stamp.Before(cutoff) {
							match = false
						}
					}
				} else {
					t.Errorf("unsupported history clause %s", clause.Raw)
					match = false
				}
			}
			if match {
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-antivirus-esmc-eset-test", "_source": map[string]any{}})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "hits": map[string]any{"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits}})
	}))
	defer server.Close()
	if err := sdkos.Connect([]string{server.URL}, "", ""); err != nil {
		t.Fatal(err)
	}
	mutate := func(doc, field string, value any) string {
		var m map[string]any
		if err := json.Unmarshal([]byte(doc), &m); err != nil {
			t.Fatal(err)
		}
		parts := strings.Split(field, ".")
		node := m
		for _, part := range parts[:len(parts)-1] {
			next, ok := node[part].(map[string]any)
			if !ok {
				next = map[string]any{}
				node[part] = next
			}
			node = next
		}
		if value == nil {
			delete(node, parts[len(parts)-1])
		} else {
			node[parts[len(parts)-1]] = value
		}
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	rawDocument := func(event map[string]any) string {
		b, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	base := func(eventType string) map[string]any {
		return map[string]any{"event_type": eventType, "source_uuid": endpoint, "hostname": "HOST-A", "ipv4": "192.0.2.10", "severity": "Warning"}
	}
	heuristic := base("Threat_Event")
	heuristic["threat_name"], heuristic["scanner_id"], heuristic["action_taken"], heuristic["threat_handled"] = "Synthetic/NewHeur.Test", "Real-time file system protection", "Deleted", true
	console := base("Audit_Event")
	console["domain"], console["action"], console["target"], console["user"], console["result"] = "Native user", "Login attempt", "Administrator", "", "Failure"
	quarantine := base("Threat_Event")
	quarantine["threat_name"], quarantine["action_taken"], quarantine["action_error"], quarantine["threat_handled"] = "Synthetic/Test", "Quarantine", "Quarantine failed: access denied", false
	// NewHeur, Failure and remediation strings below are compatibility test
	// values, not observations or an exhaustive ESET export vocabulary.
	cases := []struct {
		rule, marker, within string
		count                uint64
		event                map[string]any
	}{
		{"advanced_heuristic_detection_triggers", "heuristicRemediation", "30m", 3, heuristic},
		{"eset_console_abuse", "consoleAuthenticationFailure", "30m", 10, console},
		{"eset_quarantine_failures", "quarantineFailure", "1h", 5, quarantine},
	}
	for _, test := range cases {
		t.Run(test.rule, func(t *testing.T) {
			rule := rules[test.rule]
			if rule == nil || len(rule.Correlation) != 1 {
				t.Fatal("required rule/history request missing")
			}
			search := rule.Correlation[0]
			n := 4
			if test.rule == "eset_console_abuse" {
				n++
			}
			if search.Count != test.count || search.Within != test.within || len(search.With) != n {
				t.Fatal("history threshold or scope changed")
			}
			expectedTerms = map[string]string{"dataSource": collector, "log.endpointKeyType": "uuid", "log.endpointKey": endpoint, "log.correlationCandidate." + test.marker: "match"}
			if test.rule == "eset_console_abuse" {
				expectedTerms["target.user"] = "Administrator"
			}
			parse := func(raw string) string { return esetParse(t, cfg, raw, collector, cache) }
			raw := rawDocument(test.event)
			out := parse(raw)
			marker := "log.correlationCandidate." + test.marker
			if yes, err := cache.Eval(rule.Where, out); err != nil || !yes || gjson.Get(out, marker).String() != "match" {
				t.Fatalf("synthetic documented-field positive did not produce candidate: %v %v", yes, err)
			}
			stamp := time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)
			prior := mutate(out, "@timestamp", stamp)
			check := func(name, doc string, count uint64, want bool) {
				t.Run(name, func(t *testing.T) {
					history = nil
					for i := uint64(0); i < count; i++ {
						history = append(history, doc)
					}
					match, _, err := search.Execute(&out)
					if err != nil || match != want {
						t.Fatalf("history got %v want %v: %v", match, want, err)
					}
				})
			}
			check("below_threshold", prior, test.count-1, false)
			check("at_threshold", prior, test.count, true)
			duration, err := time.ParseDuration(test.within)
			if err != nil {
				t.Fatal(err)
			}
			check("inside_window", mutate(prior, "@timestamp", time.Now().Add(-duration+10*time.Second).UTC().Format(time.RFC3339Nano)), test.count, true)
			check("expired", mutate(prior, "@timestamp", time.Now().Add(-duration-10*time.Second).UTC().Format(time.RFC3339Nano)), test.count, false)
			check("different_collector", mutate(prior, "dataSource", "other-relay"), test.count, false)
			check("different_endpoint", mutate(prior, "log.endpointKey", "OTHER-ENDPOINT"), test.count, false)
			check("different_identity_namespace", mutate(prior, "log.endpointKeyType", "host"), test.count, false)
			check("unmarked_history", mutate(prior, marker, nil), test.count, false)
			if test.rule == "eset_console_abuse" {
				check("different_attempted_account", mutate(prior, "target.user", "OtherAdmin"), test.count, false)
			}
			changes := []map[string]any{}
			switch test.rule {
			case "advanced_heuristic_detection_triggers":
				changes = []map[string]any{{"threat_name": "Synthetic/Other", "object_uri": "file:///C:/NewHeur-Suspicious-behavior.exe"}, {"action_taken": "Detected"}, {"action_error": "File busy"}, {"threat_handled": false}, {"event_type": "Audit_Event"}}
			case "eset_console_abuse":
				changes = []map[string]any{{"result": "Success"}, {"action": "Policy assigned", "detail": "administrator login failed"}, {"domain": "Client task"}, {"target": nil}, {"event_type": "Threat_Event"}}
			case "eset_quarantine_failures":
				changes = []map[string]any{{"action_error": ""}, {"action_taken": "Blocked", "action_error": "File busy", "object_uri": "file:///C:/quarantine-failed.exe"}, {"event_type": "Audit_Event"}}
			}
			for i, change := range changes {
				modified := raw
				for field, value := range change {
					modified = mutate(modified, field, value)
				}
				// Attempt to supply the marker in raw JSON; the producer must
				// clear it and recompute from actual class/action/identity.
				modified = mutate(modified, "correlationCandidate."+test.marker, "match")
				benign := parse(modified)
				if yes, err := cache.Eval(rule.Where, benign); err != nil || yes || gjson.Get(benign, marker).Exists() {
					t.Fatalf("noncandidate raw %d retained candidate: %v %v", i, yes, err)
				}
				check(fmt.Sprintf("noncandidate_raw_%d", i), mutate(benign, "@timestamp", stamp), test.count, false)
			}
			if test.rule == "eset_console_abuse" {
				for _, result := range []string{"Denied", "Rejected"} {
					denied := parse(mutate(raw, "result", result))
					if yes, err := cache.Eval(rule.Where, denied); err != nil || !yes || gjson.Get(denied, marker).String() != "match" || gjson.Get(denied, "actionResult").String() != "denied" {
						t.Fatalf("explicitly rejected login must count as unsuccessful: %v %v", yes, err)
					}
					check("explicit_result_"+result, mutate(denied, "@timestamp", stamp), test.count, true)
				}
			}
			fields := []string{"dataSource", "log.endpointKeyType", "log.endpointKey"}
			if test.rule == "eset_console_abuse" {
				fields = append(fields, "target.user")
			}
			for _, field := range fields {
				t.Run("missing_"+field, func(t *testing.T) {
					missing := mutate(out, field, nil)
					if yes, err := cache.Eval(rule.Where, missing); err != nil || yes {
						t.Fatalf("missing identity remains eligible: %v %v", yes, err)
					}
					before := queries
					if _, _, err := search.Execute(&missing); err == nil || queries != before {
						t.Fatal("nil identity must abort before search")
					}
				})
			}
			for _, identity := range []struct {
				kind, key string
				remove    []string
			}{
				{"host", "HOST-A", []string{"source_uuid"}},
				{"ip", "192.0.2.10", []string{"source_uuid", "hostname"}},
			} {
				alternative := raw
				for _, field := range identity.remove {
					alternative = mutate(alternative, field, nil)
				}
				out = parse(alternative)
				if yes, err := cache.Eval(rule.Where, out); err != nil || !yes || gjson.Get(out, marker).String() != "match" {
					t.Fatalf("%s fallback must remain eligible: %v %v", identity.kind, yes, err)
				}
				expectedTerms["log.endpointKeyType"], expectedTerms["log.endpointKey"] = identity.kind, identity.key
				check(identity.kind+"_fallback_threshold", mutate(out, "@timestamp", stamp), test.count, true)
			}
			noEndpoint := raw
			for _, field := range []string{"source_uuid", "hostname", "ipv4"} {
				noEndpoint = mutate(noEndpoint, field, nil)
			}
			// A management relay in the syslog header must never rescue the
			// missing managed endpoint identity used by this correlation.
			noEndpoint = "Sep 17 12:00:00 RELAY-NOT-ENDPOINT ERAServer[123]: " + noEndpoint
			missing := parse(noEndpoint)
			if yes, err := cache.Eval(rule.Where, missing); err != nil || yes || gjson.Get(missing, marker).Exists() {
				t.Fatalf("relay cannot replace managed identity: %v %v", yes, err)
			}
		})
	}
}
