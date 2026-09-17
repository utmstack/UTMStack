package main

// SDK history contracts for synthetic FortiWeb events. Extraction uses the
// documented offline model in fortiweb_contract_test.go; all search requests,
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

func TestFortiWebSDKHistory(t *testing.T) {
	// The SDK owns a process-wide OpenSearch singleton. Isolate this local mock
	// so other technology tests can initialize their own clients in this suite.
	if os.Getenv("UTM_FORTIWEB_HISTORY_CHILD") != "1" {
		command := exec.Command(os.Args[0], "-test.run=^TestFortiWebSDKHistory$")
		command.Env = append(os.Environ(), "UTM_FORTIWEB_HISTORY_CHILD=1")
		if out, e := command.CombinedOutput(); e != nil {
			t.Fatalf("isolated history test: %v\n%s", e, out)
		}
		return
	}

	cfg, rules, cache := fwConfig(t), fwRules(t), plugins.NewCELCache("fortiweb-history")
	var history []string
	var expectedClauses int
	var searchRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			// Text fields exercise the SDK's .keyword mapping resolution; IP
			// fields exercise exact mappings without that suffix.
			_, _ = io.WriteString(w, `{"v11-log-firewall-fortiweb-test":{"mappings":{"properties":{"@timestamp":{"type":"date"},"origin":{"properties":{"ip":{"type":"ip"}}},"target":{"properties":{"ip":{"type":"ip"}}},"log":{"properties":{"type":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"subtype":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"action":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"fileUploadViolation":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"maintype":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"severitylevel":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"owasptop10":{"type":"text","fields":{"keyword":{"type":"keyword"}}}}}}}}}`)
			return
		}
		if r.URL.Path != "/v11-log-firewall-fortiweb-*/_search" {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unsupported request", http.StatusBadRequest)
			return
		}
		searchRequests++
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
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-firewall-fortiweb-test", "_source": map[string]any{}})
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
		fwPut(m, field, value, value == nil)
		b, e := json.Marshal(m)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}

	// These fixtures contain fabricated raw logs, not retained customer payloads.
	// Require each named case so a missing fixture cannot silently omit a rule.
	data, e := os.ReadFile("testdata/fortiweb_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var rawFixtures []fwFixture
	if e = json.Unmarshal(data, &rawFixtures); e != nil {
		t.Fatal(e)
	}
	fixtures := map[string]fwFixture{}
	for _, fixture := range rawFixtures {
		fixtures[fixture.Name] = fixture
	}
	cases := []struct {
		rule, fixture string
		count         uint64
		within        string
	}{
		{"web_application_attacks_detection", "Generic Attacks_Alert_Deny", 3, "15m"},
		{"file_upload_security_violations", "ordinary_php_upload", 3, "30m"},
		{"owasp_top10_violations", "HTTP_Illegal URL Parameter Value", 5, "15m"},
	}
	for _, test := range cases {
		t.Run(test.rule, func(t *testing.T) {
			fixture, ok := fixtures[test.fixture]
			if !ok {
				t.Fatalf("required raw fixture %q missing", test.fixture)
			}
			rule := rules[test.rule]
			if rule == nil || len(rule.Correlation) != 1 {
				t.Fatalf("expected exactly one history request for %s", test.rule)
			}
			search := rule.Correlation[0]
			if search.Count != test.count || search.Within != test.within {
				t.Fatalf("unexpected threshold %d/%s", search.Count, search.Within)
			}
			expectedClauses = len(search.With) + 1
			parse := func(raw string) string { return fwParse(t, cfg, raw, cache) }
			out := parse(fixture.Raw)
			if match, e := cache.Eval(rule.Where, out); e != nil || !match {
				t.Fatalf("positive raw fixture does not match: %v %v", match, e)
			}
			prior := mutate(out, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
			check := func(name, historical string, count uint64, want bool) {
				t.Run(name, func(t *testing.T) {
					history = nil
					for i := uint64(0); i < count; i++ {
						history = append(history, historical)
					}
					yes, _, e := search.Execute(&out)
					if e != nil || yes != want {
						t.Fatalf("history result got %v want %v: %v", yes, want, e)
					}
				})
			}
			check("below_threshold", prior, search.Count-1, false)
			check("at_threshold", prior, search.Count, true)
			duration, e := time.ParseDuration(search.Within)
			if e != nil {
				t.Fatal(e)
			}
			// The SDK's window is a processing-time lower bound on @timestamp,
			// not a proof of vendor event-time ordering or customer alert creation.
			check("inside_window", mutate(prior, "@timestamp", time.Now().Add(-duration+10*time.Second).UTC().Format(time.RFC3339Nano)), search.Count, true)
			check("expired", mutate(prior, "@timestamp", time.Now().Add(-duration-10*time.Second).UTC().Format(time.RFC3339Nano)), search.Count, false)
			for _, change := range []struct{ field, value string }{
				{"origin.ip", "198.51.100.99"}, {"target.ip", "192.0.2.99"},
				{"log.type", "traffic"}, {"log.subtype", "unrelated-class"}, {"log.action", "Pass"},
			} {
				check("different_"+change.field, mutate(prior, change.field, change.value), search.Count, false)
			}
			if test.rule == "owasp_top10_violations" {
				for _, change := range []struct{ field, value string }{
					{"log.severitylevel", "Low"}, {"log.maintype", "unrelated-main-class"}, {"log.owasptop10", "unrelated-category"},
				} {
					check("different_"+change.field, mutate(prior, change.field, change.value), search.Count, false)
				}
			}
			// Reparse a benign raw event; all endpoints are unchanged. For upload
			// history, its exact class/action are also unchanged, leaving only the
			// filter-derived eligibility marker to exclude the ordinary request.
			var benignRaw string
			switch test.rule {
			case "web_application_attacks_detection":
				benignRaw = strings.ReplaceAll(fixture.Raw, "Generic Attacks", "Information Disclosure")
			case "file_upload_security_violations":
				benignRaw = strings.ReplaceAll(fixture.Raw, "file upload violation for document.php", "ordinary document request")
			case "owasp_top10_violations":
				benignRaw = strings.ReplaceAll(fixture.Raw, "severity_level=Medium", "severity_level=Low")
			}
			benign := parse(benignRaw)
			if match, e := cache.Eval(rule.Where, benign); e != nil || match {
				t.Fatalf("benign raw event retains eligibility: %v %v", match, e)
			}
			for _, field := range []string{"origin.ip", "target.ip"} {
				if gjson.Get(benign, field).String() != gjson.Get(out, field).String() {
					t.Fatalf("benign history changed %s, masking candidate exclusion", field)
				}
			}
			if test.rule == "file_upload_security_violations" {
				if gjson.Get(benign, "log.fileUploadViolation").Exists() {
					t.Fatal("ordinary request acquired an upload violation marker")
				}
				for _, field := range []string{"log.type", "log.subtype", "log.action"} {
					if gjson.Get(benign, field).String() != gjson.Get(out, field).String() {
						t.Fatalf("benign upload history changed %s, masking marker exclusion", field)
					}
				}
			}
			check("benign_raw_history", mutate(benign, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)), search.Count, false)
			for _, term := range search.With {
				value := term.Value.GetStringValue()
				if !strings.HasPrefix(value, "{{.") || !strings.HasSuffix(value, "}}") {
					t.Fatalf("unexpected non-placeholder history term %s", term.Field)
				}
				field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
				t.Run("missing_"+field, func(t *testing.T) {
					missing := mutate(out, field, nil)
					if match, e := cache.Eval(rule.Where, missing); e != nil || match {
						t.Fatalf("where must reject missing history field: %v %v", match, e)
					}
					// Prove the failure avoided by the predicate guard. A missing
					// placeholder aborts before any history HTTP search is issued.
					before := searchRequests
					if yes, _, e := search.Execute(&missing); e == nil || yes || searchRequests != before {
						t.Error("missing history field did not abort before search")
					}
				})
			}
		})
	}
}
