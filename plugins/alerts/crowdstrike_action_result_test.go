package main

// Ordered-step model of the CrowdStrike sign-in address and outcome steps, run
// with this module's go-sdk CEL. It models only the renames that precede those
// steps (SourceIp/LocalIP to origin.ip, Success to log.eventSuccess and the
// float response code). The raw fixture lines are replayed through the isolated
// EventProcessor parser separately; this test does not execute raw parsing.
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

const crowdstrikeBruteForceRule = "../../rules/crowdstrike/multiple_authentication_failures_(possible_brute_force_attack).yml"

type crowdstrikeCase struct {
	Name       string         `json:"name"`
	Raw        string         `json:"raw"`
	Result     string         `json:"result"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	BruteForce *bool          `json:"bruteForce"`
}

func crowdstrikeCases(t *testing.T) []crowdstrikeCase {
	t.Helper()
	data, err := os.ReadFile("testdata/crowdstrike_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []crowdstrikeCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 15 {
		t.Fatal("missing outcome regression classes")
	}
	return cases
}

func crowdstrikeFilter(t *testing.T) *plugins.Config {
	t.Helper()
	b, err := utils.ReadPbYaml("../../filters/crowdstrike/crowdstrike.yml")
	if err != nil {
		t.Fatal(err)
	}
	cfg := new(plugins.Config)
	if err := protojson.Unmarshal(b, cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func crowdstrikeBruteForce(t *testing.T) *plugins.Rule {
	t.Helper()
	b, err := utils.ReadPbYaml(crowdstrikeBruteForceRule)
	if err != nil {
		t.Fatal(err)
	}
	r := new(plugins.Rule)
	if err := protojson.Unmarshal(b, r); err != nil {
		t.Fatal(err)
	}
	r.Normalize()
	return r
}

// crowdstrikeModel builds the event as the filter holds it just before the
// modeled steps.
func crowdstrikeModel(t *testing.T, raw string) map[string]any {
	t.Helper()
	var in struct {
		Metadata map[string]any `json:"metadata"`
		Event    map[string]any `json:"event"`
	}
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatal(err)
	}
	logFields := map[string]any{"metadataEventType": in.Metadata["eventType"]}
	event := map[string]any{"log": logFields}
	if v, ok := in.Event["Success"]; ok {
		logFields["eventSuccess"] = v
	}
	if attrs, ok := in.Event["Attributes"].(map[string]any); ok {
		if s, ok := attrs["status_code"].(string); ok && s != "" {
			code, err := strconv.ParseFloat(s, 64)
			if err != nil {
				t.Fatal(err)
			}
			event["statusCode"] = code
		}
	}
	// The LocalIP rename runs after the SourceIp rename.
	for _, key := range []string{"SourceIp", "LocalIP"} {
		if v, ok := in.Event[key].(string); ok {
			event["origin"] = map[string]any{"ip": v}
		}
	}
	if v, ok := in.Event["UserIp"]; ok {
		logFields["event"] = map[string]any{"UserIp": v}
	}
	return event
}

func crowdstrikeApply(t *testing.T, cfg *plugins.Config, cache *plugins.CELCache, event map[string]any) {
	t.Helper()
	eval := func(where string) bool {
		if where == "" {
			return true
		}
		state, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		ok, err := cache.Eval(where, string(state))
		if err != nil {
			t.Fatalf("%s: %v", where, err)
		}
		return ok
	}
	renamed, adds := false, 0
	for _, stage := range cfg.Pipeline {
		for _, step := range stage.Steps {
			if r := step.Rename; r != nil && len(r.From) == 1 && r.From[0] == "log.event.UserIp" {
				renamed = true
				if r.To != "origin.ip" {
					t.Fatalf("UserIp renamed to %q", r.To)
				}
				if eval(r.Where) {
					logFields := event["log"].(map[string]any)
					nested := logFields["event"].(map[string]any)
					event["origin"] = map[string]any{"ip": nested["UserIp"]}
					delete(nested, "UserIp")
				}
			}
			if a := step.Add; a != nil && a.Params["key"].GetStringValue() == "actionResult" {
				adds++
				if eval(a.Where) {
					event["actionResult"] = a.Params["value"].GetStringValue()
				}
			}
		}
	}
	if !renamed || adds == 0 {
		t.Fatal("missing sign-in address or result mapping")
	}
}

func TestCrowdStrikeActionResultContract(t *testing.T) {
	cfg, rule, cache := crowdstrikeFilter(t), crowdstrikeBruteForce(t), plugins.NewCELCache("crowdstrike-final-outcome")
	for _, tc := range crowdstrikeCases(t) {
		t.Run(tc.Name, func(t *testing.T) {
			event := crowdstrikeModel(t, tc.Raw)
			crowdstrikeApply(t, cfg, cache, event)
			got, exists := event["actionResult"]
			if tc.Result == "" && exists {
				t.Fatalf("unexpected actionResult %v", got)
			}
			if tc.Result != "" && got != tc.Result {
				t.Fatalf("actionResult = %v, want %q", got, tc.Result)
			}
			state, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			if want, ok := tc.Expected["origin.ip"]; ok && gjson.GetBytes(state, "origin.ip").String() != want {
				t.Fatalf("origin.ip = %s, want %v", gjson.GetBytes(state, "origin.ip").Raw, want)
			}
			for _, path := range tc.Absent {
				if gjson.GetBytes(state, path).Exists() {
					t.Fatalf("%s should be absent", path)
				}
			}
			for _, result := range []string{"success", "failure", "denied"} {
				matched, err := cache.Eval(`equals("actionResult","`+result+`")`, string(state))
				if err != nil || matched != (tc.Result == result) {
					t.Errorf("%s predicate = %v (%v)", result, matched, err)
				}
			}
			if tc.BruteForce != nil {
				matched, err := cache.Eval(rule.Where, string(state))
				if err != nil || matched != *tc.BruteForce {
					t.Errorf("brute-force predicate = %v (%v), want %v", matched, err, *tc.BruteForce)
				}
			}
		})
	}
}

// The brute-force rule's history search runs through the real SDK against an
// isolated loopback mock that evaluates only the asserted term and time clauses.
func TestCrowdStrikeBruteForceHistory(t *testing.T) {
	if os.Getenv("UTM_CROWDSTRIKE_HISTORY_CHILD") != "1" {
		c := exec.Command(os.Args[0], "-test.run=^TestCrowdStrikeBruteForceHistory$")
		c.Env = append(os.Environ(), "UTM_CROWDSTRIKE_HISTORY_CHILD=1")
		if b, err := c.CombinedOutput(); err != nil {
			t.Fatalf("isolated history: %v\n%s", err, b)
		}
		return
	}
	cfg, rule, cache := crowdstrikeFilter(t), crowdstrikeBruteForce(t), plugins.NewCELCache("crowdstrike-history")
	if len(rule.Correlation) != 1 {
		t.Fatal("missing history search")
	}
	search := rule.Correlation[0]
	if search.Count != 5 || search.Within != "15m" {
		t.Fatal("threshold or window changed")
	}
	window := 15 * time.Minute
	var history []string
	var wantTerms map[string]string
	queries := 0
	mapping := map[string]any{"properties": map[string]any{
		"origin":     map[string]any{"properties": map[string]any{"ip": map[string]any{"type": "text", "fields": map[string]any{"keyword": map[string]any{"type": "keyword"}}}}},
		"log":        map[string]any{"properties": map[string]any{"eventSuccess": map[string]any{"type": "boolean"}}},
		"@timestamp": map[string]any{"type": "date"},
	}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			_ = json.NewEncoder(w).Encode(map[string]any{"v11-log-crowdstrike-test": map[string]any{"mappings": mapping}})
			return
		}
		if r.URL.Path != "/v11-log-crowdstrike-*/_search" {
			t.Errorf("unexpected request %s", r.URL.Path)
			http.Error(w, "bad request", 400)
			return
		}
		queries++
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		q := string(body)
		got := map[string]string{}
		cutoff := time.Time{}
		for _, clause := range append(gjson.Get(q, "query.bool.filter").Array(), gjson.Get(q, "query.bool.must").Array()...) {
			if term := clause.Get("term"); term.Exists() {
				for field, value := range term.Map() {
					got[strings.TrimSuffix(field, ".keyword")] = value.Get("value").String()
				}
			} else if span := clause.Get("range"); span.Exists() {
				if cutoff, err = time.Parse(time.RFC3339Nano, span.Get("@timestamp.gte").String()); err != nil {
					t.Error(err)
				}
			} else {
				t.Errorf("unsupported clause %s", clause.Raw)
			}
		}
		if len(gjson.Get(q, "query.bool.must_not").Array()) != 0 {
			t.Error("unexpected negative clause")
		}
		if fmt.Sprint(got) != fmt.Sprint(wantTerms) {
			t.Errorf("scope mismatch: %v, want %v", got, wantTerms)
		}
		if delta := time.Since(cutoff) - window; delta < -2*time.Second || delta > 2*time.Second {
			t.Errorf("unexpected time cutoff %v", delta)
		}
		hits := []map[string]any{}
		for _, doc := range history {
			match := true
			for field, value := range got {
				if gjson.Get(doc, field).String() != value {
					match = false
				}
			}
			stamp, err := time.Parse(time.RFC3339Nano, gjson.Get(doc, "@timestamp").String())
			if err != nil || stamp.Before(cutoff) {
				match = false
			}
			if match {
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-crowdstrike-test", "_source": map[string]any{}})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "hits": map[string]any{"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits}})
	}))
	defer server.Close()
	if err := sdkos.Connect([]string{server.URL}, "", ""); err != nil {
		t.Fatal(err)
	}
	var failure crowdstrikeCase
	for _, tc := range crowdstrikeCases(t) {
		if tc.Name == "sign-in-failure-maps-client" {
			failure = tc
		}
	}
	event := crowdstrikeModel(t, failure.Raw)
	crowdstrikeApply(t, cfg, cache, event)
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	current := string(b)
	if ok, err := cache.Eval(rule.Where, current); err != nil || !ok {
		t.Fatalf("trigger predicate failed: %v %v", ok, err)
	}
	address := gjson.Get(current, "origin.ip").String()
	wantTerms = map[string]string{"origin.ip": address, "log.eventSuccess": "false"}
	prior := func(ip string, success bool, age time.Duration) string {
		doc := map[string]any{"origin": map[string]any{"ip": ip}, "log": map[string]any{"eventSuccess": success},
			"@timestamp": time.Now().Add(-age).UTC().Format(time.RFC3339Nano)}
		out, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		return string(out)
	}
	check := func(name, doc string, count int, want bool) {
		t.Run(name, func(t *testing.T) {
			history = nil
			for i := 0; i < count; i++ {
				history = append(history, doc)
			}
			ok, _, err := search.Execute(&current)
			if err != nil || ok != want {
				t.Fatalf("history %v want %v: %v", ok, want, err)
			}
		})
	}
	check("below_threshold", prior(address, false, time.Minute), 4, false)
	check("at_threshold", prior(address, false, time.Minute), 5, true)
	check("expired", prior(address, false, window+time.Minute), 5, false)
	check("other_address", prior("198.51.100.99", false, time.Minute), 5, false)
	check("successful_sign_ins", prior(address, true, time.Minute), 5, false)
	// Before this change the sign-in address was deleted, so the placeholder
	// had no value and the search could not run.
	without := gjson.Get(current, "origin").Raw
	stripped := strings.Replace(current, `"origin":`+without, `"origin":{}`, 1)
	before := queries
	if _, _, err := search.Execute(&stripped); err == nil {
		t.Error("missing origin.ip accepted")
	}
	if queries != before {
		t.Error("missing placeholder executed a query")
	}
}
