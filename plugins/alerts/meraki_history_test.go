package main

// These events are fabricated from the documented AnyConnect syslog format.
// The source-specific offline parser is bounded; CEL, placeholder resolution,
// mapping lookup and the history count decision execute the pinned SDK. The
// OpenSearch endpoint is a local mock, never a customer instance.
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

func TestMerakiSDKHistory(t *testing.T) {
	// The SDK client is a process-wide singleton; isolate this mock from other
	// technologies' tests when the draft PRs are tested together.
	if os.Getenv("UTM_MERAKI_HISTORY_CHILD") != "1" {
		command := exec.Command(os.Args[0], "-test.run=^TestMerakiSDKHistory$")
		command.Env = append(os.Environ(), "UTM_MERAKI_HISTORY_CHILD=1")
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("isolated history test: %v\n%s", err, out)
		}
		return
	}

	cfg, rules, cache := merakiConfig(t), merakiRules(t), plugins.NewCELCache("meraki-history")
	rule := rules["meraki_vpn_brute_force"]
	if rule == nil || len(rule.Correlation) != 1 {
		t.Fatal("VPN rule must have exactly one history request")
	}
	search := rule.Correlation[0]
	if search.Count != 10 || search.Within != "15m" || len(search.With) != 4 {
		t.Fatal("expected ten failures in fifteen minutes with collector/device/IP/class scope")
	}
	const raw = "1700000000.123456789 MX-LAB events type=anyconnect_vpn_auth_failure msg= 'Peer IP=198.51.100.10Peer port[8748] AAA[8]: AAA authenticate failed retval=7 - Authentication failure '"
	const collector = "meraki-test-relay"
	const marker = "log.vpnAuthenticationFailure"
	parse := func(raw string) string { return merakiParse(t, cfg, raw, collector, cache) }
	out := parse(raw)
	if yes, err := cache.Eval(rule.Where, out); err != nil || !yes || gjson.Get(out, marker).String() != "match" {
		t.Fatalf("documented failure must produce a matching predicate and candidate: %v %v", yes, err)
	}

	var history []string
	queries := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			// Text-with-keyword and exact IP mappings both exercise SDK lookup.
			_, _ = io.WriteString(w, `{"v11-log-firewall-meraki-test":{"mappings":{"properties":{"@timestamp":{"type":"date"},"dataSource":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"origin":{"properties":{"ip":{"type":"ip"}}},"log":{"properties":{"merakiType":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"vpnAuthenticationFailure":{"type":"keyword"}}}}}}`)
			return
		}
		if r.URL.Path != "/v11-log-firewall-meraki-*/_search" {
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
		if len(clauses) != 5 {
			t.Errorf("expected four exact scopes and time range, got %s", query)
		}
		expected := map[string]string{"dataSource": collector, "log.merakiType": "MX-LAB", "origin.ip": "198.51.100.10", marker: "match"}
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
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-firewall-meraki-test", "_source": map[string]any{}})
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
	stamp := time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)
	prior := mutate(out, "@timestamp", stamp)
	check := func(name, document string, count int, want bool) {
		t.Run(name, func(t *testing.T) {
			history = nil
			for i := 0; i < count; i++ {
				history = append(history, document)
			}
			match, _, err := search.Execute(&out)
			if err != nil || match != want {
				t.Fatalf("history result %v, expected %v: %v", match, want, err)
			}
		})
	}
	check("below_threshold", prior, 9, false)
	check("at_threshold", prior, 10, true)
	check("inside_window", mutate(prior, "@timestamp", time.Now().Add(-15*time.Minute+10*time.Second).UTC().Format(time.RFC3339Nano)), 10, true)
	check("expired", mutate(prior, "@timestamp", time.Now().Add(-15*time.Minute-10*time.Second).UTC().Format(time.RFC3339Nano)), 10, false)
	check("different_collector", mutate(prior, "dataSource", "other-relay"), 10, false)
	check("different_appliance", mutate(prior, "log.merakiType", "MX-OTHER"), 10, false)
	check("different_source", mutate(prior, "origin.ip", "198.51.100.99"), 10, false)
	check("unmarked_history", mutate(prior, marker, nil), 10, false)
	for _, eventType := range []string{"anyconnect_vpn_auth_success", "anyconnect_vpn_general", "anyconnect_vpn_session_manager"} {
		// Preserve the same Peer IP and even failure words: the explicit event
		// class, not arbitrary message text or other traffic, defines the count.
		benign := parse(strings.Replace(raw, "anyconnect_vpn_auth_failure", eventType, 1))
		if yes, err := cache.Eval(rule.Where, benign); err != nil || yes || gjson.Get(benign, marker).Exists() {
			t.Fatalf("%s retained an authentication failure: %v %v", eventType, yes, err)
		}
		check(eventType+"_is_not_failure_history", mutate(benign, "@timestamp", stamp), 10, false)
	}
	// Nested Air Marshal retains the outer events group; candidate production
	// must use the final eventType, exactly as the trigger does.
	nested := parse(strings.Replace(raw, "events type=", "events airmarshal_events type=", 1))
	if yes, err := cache.Eval(rule.Where, nested); err != nil || yes || gjson.Get(nested, marker).Exists() {
		t.Fatalf("nested non-VPN class retained a candidate: %v %v", yes, err)
	}
	check("nested_airmarshal_is_not_failure_history", mutate(nested, "@timestamp", stamp), 10, false)
	for _, field := range []string{"origin.ip", "log.merakiType", "dataSource"} {
		t.Run("missing_"+field+"_preflight", func(t *testing.T) {
			missing := mutate(out, field, nil)
			if yes, err := cache.Eval(rule.Where, missing); err != nil || yes {
				t.Fatalf("missing identity must reject the trigger: %v %v", yes, err)
			}
			before := queries
			if _, _, err := search.Execute(&missing); err == nil || queries != before {
				t.Fatal("SDK must abort unresolved identity before a history search")
			}
		})
	}
	for _, field := range []string{"dataSource", "log.merakiType"} {
		for _, invalid := range []string{"", "unknown", "-"} {
			missing := mutate(out, field, invalid)
			if yes, err := cache.Eval(rule.Where, missing); err != nil || yes {
				t.Fatalf("placeholder identity %s=%q must reject the trigger: %v %v", field, invalid, yes, err)
			}
		}
	}
}
