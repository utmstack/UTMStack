package main

// SDK history contracts for synthetic FortiGate events. Extraction uses the
// documented offline model in fortigate_contract_test.go; all search requests,
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

func TestFortiGateSDKHistory(t *testing.T) {
	// The SDK owns a process-wide OpenSearch singleton. Isolate this local mock
	// so other technology tests can initialize their own clients in this suite.
	if os.Getenv("UTM_FORTIGATE_HISTORY_CHILD") != "1" {
		command := exec.Command(os.Args[0], "-test.run=^TestFortiGateSDKHistory$")
		command.Env = append(os.Environ(), "UTM_FORTIGATE_HISTORY_CHILD=1")
		if out, e := command.CombinedOutput(); e != nil {
			t.Fatalf("isolated history test: %v\n%s", e, out)
		}
		return
	}

	cfg, rules, cache := fortiConfig(t), fortiRules(t), plugins.NewCELCache("fortigate-history")
	var history []string
	var expectedClauses int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			// Text fields exercise the SDK's .keyword mapping resolution; IP and
			// keyword fields exercise exact mappings without that suffix.
			_, _ = io.WriteString(w, `{"v11-log-firewall-fortigate-traffic-test":{"mappings":{"properties":{"@timestamp":{"type":"date"},"dataSource":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"origin":{"properties":{"ip":{"type":"ip"},"user":{"type":"text","fields":{"keyword":{"type":"keyword"}}}}},"log":{"properties":{"devid":{"type":"keyword"},"vd":{"type":"keyword"},"type":{"type":"keyword"},"subtype":{"type":"keyword"},"logid":{"type":"keyword"},"correlationCandidate":{"properties":{"vpnAuthFailure":{"type":"keyword"},"ipsCritical":{"type":"keyword"},"dlp":{"type":"keyword"},"virusOutbreak":{"type":"keyword"}}}}}}}}}`)
			return
		}
		if r.URL.Path != "/v11-log-firewall-fortigate-traffic-*/_search" {
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
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-firewall-fortigate-traffic-test", "_source": map[string]any{}})
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
		fortiPut(m, field, value, value == nil)
		b, e := json.Marshal(m)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}

	// All fixture names refer to fabricated documentation-shaped raw logs.
	// Require each named case so a missing fixture cannot silently skip a rule.
	fixtures := map[string]fortiFixture{}
	for _, fixture := range fortiFixtures(t) {
		fixtures[fixture.Name] = fixture
	}
	cases := []struct {
		rule, fixture string
		count         uint64
		within        string
	}{
		{"admin_account_compromise", "admin success positive", 5, "15m"},
		{"fortigate_vpn_brute_force", "VPN authentication positive", 10, "15m"},
		{"ips_critical_severity_events", "IPS critical blocked positive", 3, "15m"},
		{"dlp_data_exfiltration", "DLP blocked positive", 3, "1h"},
		{"antivirus_outbreak_detection", "infected-file outbreak positive", 5, "1h"},
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
			expectedClauses = len(search.With) + 1 // processing-time lower bound
			out := fortiParse(t, cfg, fixture.Raw, fixture.DataSource, cache)
			if match, e := cache.Eval(rule.Where, out); e != nil || !match {
				t.Fatalf("positive raw fixture does not match: %v %v", match, e)
			}
			parse := func(raw string) string {
				return fortiParse(t, cfg, raw, fixture.DataSource, cache)
			}
			priorRaw := fixture.Raw
			if test.rule == "admin_account_compromise" {
				priorRaw = strings.NewReplacer("0100032001", "0100032002", "Admin login successful", "Admin login failed", `status="success"`, `status="failed"`).Replace(priorRaw)
			}
			prior := mutate(parse(priorRaw), "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
			if marker := fortiMarkers[test.rule]; marker != "" && gjson.Get(prior, "log.correlationCandidate."+marker).String() != "match" {
				t.Fatal("matching raw history did not produce its candidate marker")
			}
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
			check("inside_window", mutate(prior, "@timestamp", time.Now().Add(-duration+10*time.Second).UTC().Format(time.RFC3339Nano)), search.Count, true)
			check("expired", mutate(prior, "@timestamp", time.Now().Add(-duration-10*time.Second).UTC().Format(time.RFC3339Nano)), search.Count, false)
			for _, change := range []struct{ field, value string }{
				{"origin.ip", "198.51.100.99"}, {"log.devid", "OTHER-FIREWALL"}, {"log.vd", "other-vdom"},
			} {
				check("different_"+change.field, mutate(prior, change.field, change.value), search.Count, false)
			}
			for _, term := range search.With {
				if term.Field == "dataSource" {
					check("different_ingress", mutate(prior, "dataSource", "other-relay"), search.Count, false)
				}
			}
			if test.rule == "admin_account_compromise" {
				check("different_account", mutate(prior, "origin.user", "other-user"), search.Count, false)
				check("successful_login_is_not_failed_history", mutate(parse(fixture.Raw), "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)), search.Count, false)
				check("different_event_type", mutate(prior, "log.type", "traffic"), search.Count, false)
				check("different_event_subtype", mutate(prior, "log.subtype", "vpn"), search.Count, false)
			} else {
				marker := "log.correlationCandidate." + fortiMarkers[test.rule]
				check("unmarked_history", mutate(prior, marker, nil), search.Count, false)
				// Reparse benign raw events so the filter, rather than a test-only
				// deletion, establishes why they cannot satisfy a candidate count.
				var benignRaws []string
				switch test.rule {
				case "fortigate_vpn_brute_force":
					benignRaws = []string{strings.NewReplacer(`logid="0101039426" `, "", "ssl-login-fail", "ssl-login-succ", "SSL VPN login fail", "SSL VPN login success", "SSL user failed to log in", "SSL user logged in").Replace(fixture.Raw)}
				case "ips_critical_severity_events":
					benignRaws = []string{strings.ReplaceAll(fixture.Raw, `severity="critical"`, `severity="low"`), strings.ReplaceAll(fixture.Raw, `action="dropped"`, `action="detected"`)}
				case "dlp_data_exfiltration":
					benignRaws = []string{strings.ReplaceAll(fixture.Raw, `subtype="dlp"`, `subtype="webfilter"`), strings.NewReplacer(`action="block"`, `action="allow"`, "DLP sensor blocked transfer", "ordinary transfer").Replace(fixture.Raw)}
				case "antivirus_outbreak_detection":
					benignRaws = []string{strings.ReplaceAll(fixture.Raw, `eventtype="infected"`, `eventtype="analytics"`), strings.ReplaceAll(fixture.Raw, `action="blocked"`, `action="passthrough"`)}
				}
				if len(benignRaws) == 0 {
					t.Fatal("missing non-candidate raw history")
				}
				for i, raw := range benignRaws {
					benign := parse(raw)
					if match, e := cache.Eval(rule.Where, benign); e != nil || match || gjson.Get(benign, marker).Exists() {
						t.Fatalf("benign raw event retains a candidate: %v %v", match, e)
					}
					if gjson.Get(benign, "origin.ip").String() != gjson.Get(out, "origin.ip").String() {
						t.Fatal("benign history must retain the same IP to isolate candidate semantics")
					}
					check(fmt.Sprintf("non_candidate_raw_%d", i), mutate(benign, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)), search.Count, false)
				}
			}
			identityFields := []string{"origin.ip", "log.devid", "log.vd"}
			if test.rule == "admin_account_compromise" {
				identityFields = append(identityFields, "origin.user")
			}
			for _, field := range identityFields {
				t.Run("missing_"+field, func(t *testing.T) {
					missing := mutate(out, field, nil)
					if match, e := cache.Eval(rule.Where, missing); e != nil || match {
						t.Fatalf("where must reject missing identity before history: %v %v", match, e)
					}
					// Reproduce the SDK failure that the predicate guard prevents.
					if yes, _, e := search.Execute(&missing); e == nil || yes {
						t.Error("missing identity must fail placeholder resolution")
					}
				})
			}
		})
	}
}
