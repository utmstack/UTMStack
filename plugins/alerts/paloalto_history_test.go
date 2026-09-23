package main

// SDK history contracts for synthetic PaloAlto events. Extraction uses the
// documented offline model in paloalto_contract_test.go; all search requests,
// placeholder expansion, mapping lookup and count decisions use the pinned SDK.
// The HTTP server below is loopback-only and never queries a customer instance.
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

func TestPaloAltoSDKHistory(t *testing.T) {
	// The SDK owns a process-wide OpenSearch singleton. Isolate this local mock
	// so other technology tests can initialize their own clients in this suite.
	if os.Getenv("UTM_PALOALTO_HISTORY_CHILD") != "1" {
		command := exec.Command(os.Args[0], "-test.run=^TestPaloAltoSDKHistory$")
		command.Env = append(os.Environ(), "UTM_PALOALTO_HISTORY_CHILD=1")
		if out, e := command.CombinedOutput(); e != nil {
			t.Fatalf("isolated history test: %v\n%s", e, out)
		}
		return
	}

	cfg, rules, cache := paloaltoConfig(t), paloaltoRules(t), plugins.NewCELCache("paloalto-history")
	var history []string
	var expectedClauses int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			// Text fields exercise the SDK's .keyword mapping resolution; IP and
			// keyword fields exercise exact mappings without that suffix.
			_, _ = io.WriteString(w, `{"v11-log-firewall-paloalto-test":{"mappings":{"properties":{"@timestamp":{"type":"date"},"dataSource":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"origin":{"properties":{"ip":{"type":"ip"}}},"target":{"properties":{"ip":{"type":"ip"}}},"log":{"properties":{"paScope":{"type":"keyword"},"paVsys":{"type":"keyword"},"paDirection":{"type":"keyword"},"correlationCandidate":{"properties":{"paloalto":{"properties":{"authFailure":{"type":"keyword"},"dns":{"type":"keyword"},"url":{"type":"keyword"},"vulnerability":{"type":"keyword"}}}}}}}}}}}`)
			return
		}
		if r.URL.Path != "/v11-log-firewall-paloalto-*/_search" {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unsupported request", http.StatusBadRequest)
			return
		}
		b, e := io.ReadAll(r.Body)
		if e != nil {
			t.Error(e)
			return
		}
		query := string(b)
		clauses := append(gjson.Get(query, "query.bool.must").Array(), gjson.Get(query, "query.bool.filter").Array()...)
		if len(clauses) != expectedClauses {
			t.Errorf("history clauses: got %d want %d: %s", len(clauses), expectedClauses, query)
		}
		hits := []map[string]any{}
		for _, doc := range history {
			yes := true
			for _, clause := range clauses {
				if term := clause.Get("term"); term.Exists() {
					for field, v := range term.Map() {
						value := gjson.Get(doc, strings.TrimSuffix(field, ".keyword"))
						if !value.Exists() || value.String() != v.Get("value").String() {
							yes = false
						}
					}
				} else if span := clause.Get("range"); span.Exists() {
					for field, limits := range span.Map() {
						stamp, e := time.Parse(time.RFC3339Nano, gjson.Get(doc, field).String())
						if e != nil {
							t.Error(e)
						}
						cutoff, e := time.Parse(time.RFC3339Nano, limits.Get("gte").String())
						if e != nil {
							t.Error(e)
						}
						if stamp.Before(cutoff) {
							yes = false
						}
					}
				} else {
					t.Errorf("unsupported clause %s", clause.Raw)
					yes = false
				}
			}
			if yes {
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-firewall-paloalto-test", "_source": map[string]any{}})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "hits": map[string]any{"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits}})
	}))
	defer server.Close()
	if e := sdkos.Connect([]string{server.URL}, "", ""); e != nil {
		t.Fatal(e)
	}
	mutate := func(doc, field string, value any) string {
		var m map[string]any
		if e := json.Unmarshal([]byte(doc), &m); e != nil {
			t.Fatal(e)
		}
		paloaltoPut(m, field, value, value == nil)
		b, e := json.Marshal(m)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}

	// All fixture names refer to fabricated documentation-shaped raw logs.
	// Require each named case so a missing fixture cannot silently skip a rule.
	fixtures := map[string]paloaltoFixture{}
	for _, fixture := range paloaltoFixtures(t) {
		fixtures[fixture.Name] = fixture
	}

	cases := []struct {
		name, marker, benign string
		count                uint64
		within               string
	}{
		{"panos_admin_brute_force", "authFailure", "successful-auth", 10, "15m"},
		{"panos_dns_security_alerts", "dns", "ordinary-spyware-not-dns-history-candidate", 3, "30m"},
		{"panos_url_filtering_blocks", "url", "benign-url-block-not-history-candidate", 5, "30m"},
		{"zero_day_exploit_prevention", "vulnerability", "low-vulnerability-not-history-candidate", 2, "30m"},
		{"zero_day_exploit_prevention_server_to_client", "vulnerability", "low-vulnerability-not-history-candidate", 2, "30m"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fixture, ok := fixtures["positive-"+c.name]
			if !ok {
				t.Fatal("missing positive fixture")
			}
			rule := rules[c.name]
			if rule == nil || len(rule.Correlation) != 1 {
				t.Fatal("missing history")
			}
			search := rule.Correlation[0]
			if search.Count != c.count || search.Within != c.within {
				t.Fatal("threshold changed")
			}
			expectedClauses = len(search.With) + 1
			out := paloaltoParse(t, cfg, fixture.Raw, fixture.DataSource, cache)
			marker := "log.correlationCandidate.paloalto." + c.marker
			if match, e := cache.Eval(rule.Where, out); e != nil || !match || gjson.Get(out, marker).String() != "match" {
				t.Fatal("positive predicate/marker mismatch")
			}
			prior := mutate(out, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
			check := func(name, doc string, count uint64, want bool) {
				t.Run(name, func(t *testing.T) {
					history = nil
					for i := uint64(0); i < count; i++ {
						history = append(history, doc)
					}
					yes, _, err := search.Execute(&out)
					if err != nil || yes != want {
						t.Fatalf("history got %v want %v: %v", yes, want, err)
					}
				})
			}
			check("below_threshold", prior, c.count-1, false)
			check("at_threshold", prior, c.count, true)
			duration, e := time.ParseDuration(c.within)
			if e != nil {
				t.Fatal(e)
			}
			check("expired", mutate(prior, "@timestamp", time.Now().Add(-duration-time.Minute).UTC().Format(time.RFC3339Nano)), c.count, false)
			check("unmarked", mutate(prior, marker, nil), c.count, false)
			for _, term := range search.With {
				if !strings.HasPrefix(term.Value.GetStringValue(), "{{.") {
					continue
				}
				field := term.Field
				check("different_"+field, mutate(prior, field, "203.0.113.99"), c.count, false)
				t.Run("missing_"+field, func(t *testing.T) {
					missing := mutate(out, field, nil)
					if match, e := cache.Eval(rule.Where, missing); e != nil || match {
						t.Fatal("predicate allows unresolved identity")
					}
					if yes, _, e := search.Execute(&missing); e == nil || yes {
						t.Fatal("expected SDK missing-placeholder error")
					}
				})
			}
			benign, ok := fixtures[c.benign]
			if !ok {
				t.Fatal("missing noncandidate fixture")
			}
			normalized := paloaltoParse(t, cfg, benign.Raw, fixture.DataSource, cache)
			if gjson.Get(normalized, "origin.ip").String() != gjson.Get(out, "origin.ip").String() {
				t.Fatal("benign comparison must retain source address")
			}
			if match, e := cache.Eval(rule.Where, normalized); e != nil || match || gjson.Get(normalized, marker).Exists() {
				t.Fatal("noncandidate contributes marker")
			}
			check("benign_raw_history", mutate(normalized, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)), c.count, false)
		})
	}
}
