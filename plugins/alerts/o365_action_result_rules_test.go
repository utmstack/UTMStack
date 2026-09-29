package main

// These tests use the pinned SDK CEL and history implementation. Raw normalization
// uses the explicit offline filter model in o365ActionResultNormalize, not a live
// EventProcessor. The history transport is a loopback mock, not customer storage.
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var o365OutcomeRuleFiles = []string{
	"possible_succesfull_password_guessing_o365",
	"credential_access_microsoft_365_potential_password_spraying_attack",
	"dlp_policy_violations", "safe_links_click_patterns", "insider_risk_indicators", "information_barriers_violations",
}

func o365OutcomeRule(t *testing.T, name string) *plugins.Rule {
	t.Helper()
	b, err := utils.ReadPbYaml("../../rules/office365/" + name + ".yml")
	if err != nil {
		t.Fatal(err)
	}
	r := new(plugins.Rule)
	if err = protojson.Unmarshal(b, r); err != nil {
		t.Fatal(err)
	}
	r.Normalize()
	return r
}

func o365OutcomeRaw(t *testing.T, name string) string {
	t.Helper()
	event := map[string]any{"Workload": "Exchange", "ClientIP": "198.51.100.10", "UserId": "reviewer@example.test", "ResultStatus": "Blocked"}
	switch name {
	case "possible_succesfull_password_guessing_o365", "credential_access_microsoft_365_potential_password_spraying_attack":
		event["Operation"] = "UserLoginFailed"
		event["Workload"] = "AzureActiveDirectory"
		event["RecordType"] = 15
		event["ResultStatus"] = "Succeeded"
	case "dlp_policy_violations":
		event["Operation"] = "DLPRuleMatch"
	case "safe_links_click_patterns":
		event["Operation"] = "ClickedSafeLink"
	case "insider_risk_indicators":
		event["Operation"] = "PolicyEvaluation"
		event["PolicyName"] = "InsiderRiskPolicy"
	case "information_barriers_violations":
		event["Operation"] = "PolicyEvaluation"
		event["PolicyType"] = "InformationBarrier"
	default:
		t.Fatal("unknown case", name)
	}
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func o365OutcomeSet(t *testing.T, event, path string, value any) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(event), &m); err != nil {
		t.Fatal(err)
	}
	keys := strings.Split(path, ".")
	node := m
	for _, key := range keys[:len(keys)-1] {
		next, ok := node[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			node[key] = next
		}
		node = next
	}
	key := keys[len(keys)-1]
	if value == nil {
		delete(node, key)
	} else {
		node[key] = value
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func o365OutcomeAssert(t *testing.T, cache *plugins.CELCache, r *plugins.Rule, event string, want bool) {
	t.Helper()
	got, err := cache.Eval(r.Where, event)
	if err != nil || got != want {
		t.Fatalf("predicate got %v want %v: %v", got, want, err)
	}
}

func TestO365ActionResultRuleCompatibility(t *testing.T) {
	cache := plugins.NewCELCache("o365-outcome-rules")
	for _, name := range o365OutcomeRuleFiles {
		t.Run(name, func(t *testing.T) {
			r := o365OutcomeRule(t, name)
			event := o365ActionResultNormalize(t, o365OutcomeRaw(t, name))
			o365OutcomeAssert(t, cache, r, event, true)
			if gaps := fixtureHistoryPlaceholders(r.Correlation, event); len(gaps) != 0 {
				t.Fatal(gaps)
			}
			failureRule := strings.Contains(name, "password_")
			// Exercise extraction and normalization for a negative record as well
			// as the positive raw fixture before testing legacy stored outcomes.
			negativeRaw := o365OutcomeSet(t, o365OutcomeRaw(t, name), "ResultStatus", "Succeeded")
			if failureRule {
				negativeRaw = o365OutcomeSet(t, negativeRaw, "Operation", "UserLoggedIn")
				negativeRaw = o365OutcomeSet(t, negativeRaw, "ErrorNumber", 0)
			}
			if name == "dlp_policy_violations" {
				negativeRaw = o365OutcomeSet(t, negativeRaw, "ResultStatus", "Failed")
			}
			o365OutcomeAssert(t, cache, r, o365ActionResultNormalize(t, negativeRaw), false)
			for _, outcome := range []string{"success", "failure", "failed", "denied", "blocked", "unknown", ""} {
				want := outcome == "denied" || outcome == "blocked"
				if failureRule {
					want = outcome == "failure" || outcome == "failed"
				}
				if name == "dlp_policy_violations" {
					want = outcome != "failure" && outcome != "failed"
				}
				t.Run(outcome, func(t *testing.T) {
					o365OutcomeAssert(t, cache, r, o365OutcomeSet(t, event, "actionResult", outcome), want)
				})
			}
			// Preserve DLP's existing unknown-outcome detection scope; absence is not
			// described as success. Other migrated outcome branches need an outcome.
			o365OutcomeAssert(t, cache, r, o365OutcomeSet(t, event, "actionResult", nil), name == "dlp_policy_violations")
			unrelated := o365OutcomeSet(t, event, "action", "UnrelatedOperation")
			unrelated = o365OutcomeSet(t, unrelated, "log.PolicyName", nil)
			unrelated = o365OutcomeSet(t, unrelated, "log.PolicyType", nil)
			o365OutcomeAssert(t, cache, r, unrelated, false)
			if failureRule || name == "safe_links_click_patterns" {
				o365OutcomeAssert(t, cache, r, o365OutcomeSet(t, event, "origin.ip", nil), name == "safe_links_click_patterns")
			}
			if name == "possible_succesfull_password_guessing_o365" || name == "safe_links_click_patterns" {
				o365OutcomeAssert(t, cache, r, o365OutcomeSet(t, event, "origin.user", nil), false)
			}
		})
	}
}

func TestO365ActionResultPreservesIndependentBranches(t *testing.T) {
	cache := plugins.NewCELCache("o365-independent-branches")
	for _, tc := range []struct {
		name, event string
		want        bool
	}{
		{"insider_risk_indicators", `{"action":"InsiderRiskAlert","actionResult":"failure"}`, true},
		{"insider_risk_indicators", `{"log":{"RiskLevel":"High","AlertSource":"InsiderRiskManagement"}}`, true},
		{"insider_risk_indicators", `{"log":{"RiskLevel":"Low","AlertSource":"InsiderRiskManagement"}}`, false},
		{"information_barriers_violations", `{"action":"InformationBarrierPolicyViolation","origin":{"user":"reviewer@example.test"}}`, true},
		{"information_barriers_violations", `{"action":"CommunicationBlocked","log":{"ViolationType":"InformationBarrier"},"origin":{"user":"reviewer@example.test"}}`, true},
		{"information_barriers_violations", `{"action":"CommunicationBlocked","log":{"ViolationType":"Other"},"origin":{"user":"reviewer@example.test"}}`, false},
	} {
		t.Run(tc.name+fmt.Sprint(tc.want), func(t *testing.T) { o365OutcomeAssert(t, cache, o365OutcomeRule(t, tc.name), tc.event, tc.want) })
	}
}

func TestO365ActionResultSDKHistory(t *testing.T) {
	// Isolate the SDK's process-global OpenSearch connection and field mapper.
	if os.Getenv("UTM_O365_OUTCOME_HISTORY_CHILD") != "1" {
		c := exec.Command(os.Args[0], "-test.run=^TestO365ActionResultSDKHistory$")
		c.Env = append(os.Environ(), "UTM_O365_OUTCOME_HISTORY_CHILD=1")
		if b, err := c.CombinedOutput(); err != nil {
			t.Fatalf("isolated history: %v\n%s", err, b)
		}
		return
	}
	var history []string
	var expectedTerms map[string]string
	var window time.Duration
	queries := 0
	mapping := map[string]any{"properties": map[string]any{
		"@timestamp": map[string]any{"type": "date"}, "action": map[string]any{"type": "keyword"},
		"origin": map[string]any{"properties": map[string]any{"user": map[string]any{"type": "keyword"}, "ip": map[string]any{"type": "ip"}}},
		"log":    map[string]any{"properties": map[string]any{"PolicyType": map[string]any{"type": "keyword"}}},
	}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			_ = json.NewEncoder(w).Encode(map[string]any{"v11-log-o365-test": map[string]any{"mappings": mapping}})
			return
		}
		if r.URL.Path != "/v11-log-o365-*/_search" {
			t.Errorf("unexpected request %s", r.URL.Path)
			http.Error(w, "bad request", 400)
			return
		}
		queries++
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		q := string(b)
		terms := map[string]string{}
		cutoff := time.Time{}
		clauses := append(gjson.Get(q, "query.bool.filter").Array(), gjson.Get(q, "query.bool.must").Array()...)
		if gjson.Get(q, "query.bool.must_not").Exists() {
			t.Error("unexpected negative history clauses")
		}
		for _, clause := range clauses {
			if term := clause.Get("term"); term.Exists() {
				for field, value := range term.Map() {
					terms[strings.TrimSuffix(field, ".keyword")] = value.Get("value").String()
				}
			} else if span := clause.Get("range"); span.Exists() {
				cutoff, err = time.Parse(time.RFC3339Nano, span.Get("@timestamp.gte").String())
				if err != nil {
					t.Error(err)
				}
			} else {
				t.Errorf("unsupported history clause %s", clause.Raw)
			}
		}
		if !reflect.DeepEqual(terms, expectedTerms) {
			t.Errorf("history scope got %v want %v", terms, expectedTerms)
		}
		if delta := time.Since(cutoff) - window; delta < -2*time.Second || delta > 2*time.Second {
			t.Errorf("wrong cutoff %v", delta)
		}
		hits := []map[string]any{}
		for _, doc := range history {
			match := true
			for field, value := range terms {
				if gjson.Get(doc, field).String() != value {
					match = false
				}
			}
			stamp, err := time.Parse(time.RFC3339Nano, gjson.Get(doc, "@timestamp").String())
			if err != nil || stamp.Before(cutoff) {
				match = false
			}
			if match {
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-o365-test", "_source": map[string]any{}})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "hits": map[string]any{"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits}})
	}))
	defer server.Close()
	if err := sdkos.Connect([]string{server.URL}, "", ""); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("o365-history")
	for _, tc := range []struct {
		name, within string
		count        uint64
		terms        map[string]string
	}{
		{"possible_succesfull_password_guessing_o365", "1m", 10, map[string]string{"action": "UserLoginFailed", "origin.user": "reviewer@example.test", "origin.ip": "198.51.100.10"}},
		{"credential_access_microsoft_365_potential_password_spraying_attack", "60s", 5, map[string]string{"origin.ip": "198.51.100.10"}},
		{"safe_links_click_patterns", "30m", 5, map[string]string{"origin.user": "reviewer@example.test", "action": "ClickedSafeLink"}},
		{"information_barriers_violations", "12h", 3, map[string]string{"origin.user": "reviewer@example.test", "log.PolicyType": "InformationBarrier"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := o365OutcomeRule(t, tc.name)
			if len(r.Correlation) != 1 {
				t.Fatal("unexpected history count")
			}
			search := r.Correlation[0]
			if search.Count != tc.count || search.Within != tc.within {
				t.Fatal("history threshold/window changed")
			}
			expectedTerms = tc.terms
			window, _ = time.ParseDuration(tc.within)
			event := o365ActionResultNormalize(t, o365OutcomeRaw(t, tc.name))
			o365OutcomeAssert(t, cache, r, event, true)
			if tc.name == "possible_succesfull_password_guessing_o365" {
				if gjson.Get(event, "log.clientIP").Exists() {
					t.Fatal("test needs normalized IP without obsolete alias")
				}
				old := proto.Clone(search).(*plugins.SearchRequest)
				for _, e := range old.With {
					if e.Field == "origin.ip" {
						e.Value = structpb.NewStringValue("{{.log.clientIP}}")
					}
				}
				before := queries
				if _, _, err := old.Execute(&event); err == nil {
					t.Fatal("obsolete placeholder unexpectedly resolved")
				}
				if queries != before {
					t.Fatal("unresolved placeholder issued query")
				}
			}
			prior := o365OutcomeSet(t, event, "@timestamp", time.Now().Add(-window/2).UTC().Format(time.RFC3339Nano))
			history = nil
			for i := uint64(0); i < tc.count-1; i++ {
				history = append(history, prior)
			}
			if yes, _, err := search.Execute(&event); err != nil || yes {
				t.Fatalf("below threshold: %v %v", yes, err)
			}
			history = append(history, prior)
			if yes, _, err := search.Execute(&event); err != nil || !yes {
				t.Fatalf("at threshold: %v %v", yes, err)
			}
			expired := o365OutcomeSet(t, prior, "@timestamp", time.Now().Add(-window-time.Minute).UTC().Format(time.RFC3339Nano))
			history[len(history)-1] = expired
			if yes, _, err := search.Execute(&event); err != nil || yes {
				t.Fatalf("expired history counted: %v %v", yes, err)
			}
			for field := range expectedTerms {
				history[len(history)-1] = o365OutcomeSet(t, prior, field, "different-value")
				if yes, _, err := search.Execute(&event); err != nil || yes {
					t.Fatalf("different %s counted: %v %v", field, yes, err)
				}
			}
			for _, expression := range search.With {
				value := expression.Value.GetStringValue()
				if strings.HasPrefix(value, "{{.") {
					field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
					missing := o365OutcomeSet(t, event, field, nil)
					before := queries
					if _, _, err := search.Execute(&missing); err == nil {
						t.Fatalf("missing %s accepted", field)
					}
					if queries != before {
						t.Fatal("unresolved placeholder issued query")
					}
				}
			}
		})
	}
}
