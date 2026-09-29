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

func TestAzureSDKHistory(t *testing.T) {
	if os.Getenv("UTM_Azure_HISTORY_CHILD") != "1" {
		c := exec.Command(os.Args[0], "-test.run=^TestAzureSDKHistory$")
		c.Env = append(os.Environ(), "UTM_Azure_HISTORY_CHILD=1")
		if b, e := c.CombinedOutput(); e != nil {
			t.Fatalf("isolated history: %v\n%s", e, b)
		}
		return
	}
	cfg, rules, cache := azureConfig(t), azureRules(t), plugins.NewCELCache("azure-history")
	var history []string
	var terms, notTerms map[string]string
	var window time.Duration
	queries := 0
	mapping := map[string]any{"properties": map[string]any{}}
	props := mapping["properties"].(map[string]any)
	paths := []string{"dataSource", "log.azureScopeType", "log.azureScope", "log.azureActorType", "log.azureActor", "origin.ip"}
	for name, r := range rules {
		if len(r.Correlation) > 0 {
			if azureHistoryMarkers[name] == "" {
				t.Fatalf("history rule %s has no correlation marker", name)
			}
			paths = append(paths, "log.correlationCandidate."+azureHistoryMarkers[name])
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
			_ = json.NewEncoder(w).Encode(map[string]any{"v11-log-azure-test": map[string]any{"mappings": mapping}})
			return
		}
		if r.URL.Path != "/v11-log-azure-*/_search" {
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
		expectedTerms := map[string]string{}
		for k, v := range terms {
			expectedTerms[k] = v
		}

		if !same(gotTerms, expectedTerms) || !same(gotNot, notTerms) {
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
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-azure-test", "_source": map[string]any{}})
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
		azurePut(m, path, value, value == nil)
		b, e := json.Marshal(m)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}

	cases := []struct {
		rule, fixture, within string
		count                 uint64
		ip                    bool
	}{
		{"aks_security_threats", "activity-success-0", "30m", 10, false},
		{"app_registration_abuse", "audit-success-0", "1h", 3, false},
		{"application_gateway_waf_alerts", "waf-blocked", "10m", 5, true},
		{"azure_ad_password_spray", "signin-invalid-password", "15m", 15, true},
		{"azure_bulk_role_changes", "audit-success-7", "30m", 10, false},
		{"azure_kubernetes_secret_access", "k8s-secrets", "30m", 5, false},
		{"azure_laps_credential_dump", "audit-success-10", "1h", 3, false},
		{"azure_ropc_authentication", "signin-ropc", "1h", 5, false},
		{"key_vault_access_spikes", "vault-SecretGet", "10m", 20, false},
		{"pim_role_activation_abuse", "audit-success-9", "4h", 3, false},
	}
	fixtures := map[string]azureFixture{}
	for _, f := range azureFixtures(t) {
		fixtures[f.Name] = f
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			r := rules[tc.rule]
			if r == nil || len(r.Correlation) != 1 {
				t.Fatal("missing or extra history")
			}
			search := r.Correlation[0]
			if search.Count != tc.count || search.Within != tc.within || len(search.Or) != 0 {
				t.Fatal("threshold/window changed")
			}
			var e error
			window, e = time.ParseDuration(tc.within)
			if e != nil {
				t.Fatal(e)
			}
			f := fixtures[tc.fixture]
			out := azureParse(t, cfg, f.Raw, f.DataSource, cache)
			if yes, e := cache.Eval(r.Where, out); e != nil || !yes {
				t.Fatalf("raw trigger failed: %v %v", yes, e)
			}
			// The filter stores this marker and the rule counts it under the same name.
			marker := "log.correlationCandidate." + azureHistoryMarkers[tc.rule]
			terms = map[string]string{"dataSource": "collector-test", "log.azureScopeType": "directory", "log.azureScope": "directory-test", marker: "true"}
			notTerms = map[string]string{}
			if tc.rule == "azure_kubernetes_secret_access" || tc.rule == "application_gateway_waf_alerts" {
				terms["log.azureScopeType"] = "resource"
				terms["log.azureScope"] = "/subscriptions/sub-test/providers/Microsoft.Example/resources/test"
			}
			if tc.rule == "key_vault_access_spikes" {
				terms["log.azureScopeType"] = "resource"
				terms["log.azureScope"] = "/subscriptions/sub-test/providers/Microsoft.KeyVault/vaults/vault-test"
			}
			if tc.ip {
				terms["origin.ip"] = "198.51.100.4"
			} else {
				terms["log.azureActorType"] = "user"
				terms["log.azureActor"] = "admin@example.test"
				switch tc.rule {
				case "aks_security_threats", "key_vault_access_spikes":
					terms["log.azureActorType"] = "ip"
					terms["log.azureActor"] = "198.51.100.4"
				case "azure_kubernetes_secret_access":
					terms["log.azureActor"] = "actor-test"
				case "azure_ropc_authentication":
					terms["log.azureActor"] = "actor@example.test"
				}
			}
			prior := mutate(out, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
			check := func(name, doc string, count uint64, want bool) {
				t.Run(name, func(t *testing.T) {
					history = nil
					for i := uint64(0); i < count; i++ {
						history = append(history, doc)
					}
					yes, _, e := search.Execute(&out)
					if e != nil || yes != want {
						t.Fatalf("history %v want %v: %v", yes, want, e)
					}
				})
			}
			check("below_threshold", prior, tc.count-1, false)
			check("at_threshold", prior, tc.count, true)
			check("expired", mutate(prior, "@timestamp", time.Now().Add(-window-time.Minute).UTC().Format(time.RFC3339Nano)), tc.count, false)
			check("inside_window", mutate(prior, "@timestamp", time.Now().Add(-window+time.Minute).UTC().Format(time.RFC3339Nano)), tc.count, true)
			for field := range terms {
				if field == marker {
					continue
				}
				check("different_"+field, mutate(prior, field, "other"), tc.count, false)
				without := mutate(out, field, nil)
				if yes, e := cache.Eval(r.Where, without); e != nil || yes {
					t.Errorf("predicate accepts missing identity %s: %v", field, e)
				}
				before := queries
				if _, _, e := search.Execute(&without); e == nil {
					t.Errorf("missing placeholder %s accepted", field)
				}
				if queries != before {
					t.Error("missing placeholder executes query")
				}
			}
			check("unrelated_population", mutate(prior, marker, nil), tc.count, false)
			// Source-derived marker must not be inherited from raw input, even when all identities match.
			var raw map[string]any
			if e := json.Unmarshal([]byte(f.Raw), &raw); e != nil {
				t.Fatal(e)
			}
			raw["correlationCandidate"] = map[string]any{azureHistoryMarkers[tc.rule]: "true"}
			raw["category"] = "AppServiceConsoleLogs"
			raw["operationName"] = "Microsoft.Web/sites/log"
			delete(raw, "properties")
			bytes, _ := json.Marshal(raw)
			unrelated := azureParse(t, cfg, string(bytes), f.DataSource, cache)
			if gjson.Get(unrelated, marker).Exists() {
				t.Error("forged history marker survived")
			}
			check("unrelated_raw_activity", mutate(unrelated, "@timestamp", time.Now().UTC().Format(time.RFC3339Nano)), tc.count, false)
		})
	}
	if len(cases) != 10 {
		t.Fatal("history coverage")
	}
}
