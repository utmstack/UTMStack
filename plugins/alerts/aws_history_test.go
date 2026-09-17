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

func TestAWSSDKHistory(t *testing.T) {
	if os.Getenv("UTM_AWS_HISTORY_CHILD") != "1" {
		c := exec.Command(os.Args[0], "-test.run=^TestAWSSDKHistory$")
		c.Env = append(os.Environ(), "UTM_AWS_HISTORY_CHILD=1")
		if b, e := c.CombinedOutput(); e != nil {
			t.Fatalf("isolated history: %v\n%s", e, b)
		}
		return
	}
	cfg, rules, cache := awsConfig(t), awsRules(t), plugins.NewCELCache("aws-history")
	var history []string
	var terms, notTerms map[string]string
	var window time.Duration
	var ruleName string
	queries := 0
	mapping := map[string]any{"properties": map[string]any{}}
	props := mapping["properties"].(map[string]any)
	paths := []string{"dataSource", "log.awsAccountKeyType", "log.awsAccountKey", "log.awsActorKeyType", "log.awsActorKey", "origin.ip", "origin.geolocation.countryCode", "log.eventName", "log.correlationCandidate.saml_provider_change"}
	for name, r := range rules {
		if len(r.Correlation) > 0 {
			paths = append(paths, "log.correlationCandidate."+name)
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
			_ = json.NewEncoder(w).Encode(map[string]any{"v11-log-aws-test": map[string]any{"mappings": mapping}})
			return
		}
		if r.URL.Path != "/v11-log-aws-*/_search" {
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
		if ruleName == "secrets_manager_suspicious_access" {
			value := gotTerms["log.eventName"]
			if value != "GetSecretValue" && value != "BatchGetSecretValue" {
				t.Error("unexpected secret query")
			}
			expectedTerms["log.eventName"] = value
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
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-aws-test", "_source": map[string]any{}})
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
		awsPut(m, path, value, value == nil)
		b, e := json.Marshal(m)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}

	cases := []struct {
		rule, fixture, within, mode string
		count                       uint64
	}{
		{"aws_ecs_credential_theft", "ecs-assumed-role", "30m", "ip", 5},
		{"aws_golden_saml_attack", "sts-AssumeRoleWithSAML", "24h", "account", 1},
		{"aws_securityhub_finding_evasion", "securityhub-BatchUpdateFindings", "30m", "actor", 5},
		{"aws_ssm_sendcommand_abuse", "ssm-SendCommand", "30m", "actor", 5},
		{"aws_sso_suspicious_activities", "sso-CreatePermissionSet", "30m", "ip", 10},
		{"cloudformation_stack_deletion", "cloudformation-DeleteStack", "30m", "actor", 5},
		{"console_login_impossible_travel", "console-country-correlation", "30m", "actor", 1},
		{"cross_account_access_anomalies", "cross-account-role", "15m", "ip", 15},
		{"iam_backdoor_creation_attempts", "iam-CreateUser", "30m", "ip", 3},
		{"iam_privilege_escalation_paths", "iam-AddUserToGroup", "30m", "actor", 3},
		{"lambda_privilege_escalation", "iam-AttachRolePolicy", "1h", "actor", 2},
		{"mass_resource_deletion", "backup-DeleteBackupVault", "10m", "actor", 15},
		{"route53_dns_hijacking", "route53-ChangeResourceRecordSets", "30m", "actor", 10},
		{"s3_bulk_data_exfiltration", "s3-GetObject", "15m", "actor", 100},
		{"secrets_manager_suspicious_access", "secretsmanager-GetSecretValue", "10m", "actor", 10},
		{"security_group_modifications", "ec2-AuthorizeSecurityGroupIngress", "30m", "ip", 3},
		{"ssm_session_abuse", "ssm-SendCommand", "30m", "actor", 5},
		{"sts_token_abuse", "sts-explicit-no-mfa", "15m", "ip", 20},
		{"unusual_api_call_patterns", "ec2-DescribeNetworkAcls", "10m", "ip", 50},
		{"vpc_flow_log_anomalies", "ec2-DeleteFlowLogs", "24h", "ip", 2},
		{"credential_access_aws_iam_assume_role_brute_force", "malformed-trust-policy", "15m", "actor", 5},
		{"credential_access_root_console_failure_brute_force", "root-Failure-Yes", "15m", "ip", 5},
	}
	fixtures := map[string]awsFixture{}
	for _, f := range awsFixtures(t) {
		fixtures[f.Name] = f
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			ruleName = tc.rule
			r := rules[tc.rule]
			if r == nil || len(r.Correlation) != 1 {
				t.Fatal("missing or extra history")
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
			out := awsParse(t, cfg, f.Raw, f.DataSource, cache, f.Enrichment)
			if yes, e := cache.Eval(r.Where, out); e != nil || !yes {
				t.Fatalf("raw trigger failed: %v %v", yes, e)
			}
			marker := "log.correlationCandidate." + tc.rule
			if tc.rule == "aws_golden_saml_attack" {
				marker = "log.correlationCandidate.saml_provider_change"
			}
			terms = map[string]string{"dataSource": "collector-test", "log.awsAccountKeyType": "recipient", "log.awsAccountKey": "123456789012", marker: "match"}
			notTerms = map[string]string{}
			if tc.mode == "ip" {
				terms["origin.ip"] = "198.51.100.10"
			}
			if tc.mode == "actor" {
				terms["log.awsActorKeyType"] = "arn"
				terms["log.awsActorKey"] = "arn:aws:iam::123456789012:user/reviewer"
			}
			prior := out
			if tc.rule == "aws_golden_saml_attack" {
				var raw map[string]any
				if e := json.Unmarshal([]byte(f.Raw), &raw); e != nil {
					t.Fatal(e)
				}
				raw["eventSource"] = "iam.amazonaws.com"
				raw["eventName"] = "UpdateSAMLProvider"
				raw["userIdentity"] = map[string]any{"type": "IAMUser", "accountId": "123456789012", "arn": "arn:aws:iam::123456789012:user/other-admin"}
				encoded, _ := json.Marshal(raw)
				prior = awsParse(t, cfg, string(encoded), f.DataSource, cache)
				if !gjson.Get(prior, marker).Exists() {
					t.Fatal("provider change does not produce sequence candidate")
				}
				if gjson.Get(out, marker).Exists() {
					t.Fatal("SAML login counts as its own provider change")
				}
			}
			if tc.rule == "console_login_impossible_travel" {
				notTerms["origin.geolocation.countryCode"] = "US"
				prior = mutate(prior, "origin.geolocation.countryCode", "GB")
			}
			prior = mutate(prior, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
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
					t.Errorf("predicate accepts missing history identity %s: %v", field, e)
				}
				before := queries
				if _, _, e := search.Execute(&without); e == nil {
					t.Errorf("missing placeholder %s accepted", field)
				}
				if queries != before {
					t.Error("missing placeholder executed query")
				}
			}
			check("unrelated_population", mutate(prior, marker, nil), tc.count, false)
			if tc.rule == "console_login_impossible_travel" {
				check("same_country", mutate(prior, "origin.geolocation.countryCode", "US"), tc.count, false)
			}
			if tc.rule == "aws_golden_saml_attack" {
				check("unrelated_same_account_activity", mutate(out, "@timestamp", time.Now().UTC().Format(time.RFC3339Nano)), 1, false)
			}
			if tc.rule == "secrets_manager_suspicious_access" {
				if len(search.Or) != 1 || search.Or[0].Count != 5 || search.Or[0].Within != "10m" {
					t.Fatal("batch secret OR threshold changed")
				}
				batch := mutate(prior, "log.eventName", "BatchGetSecretValue")
				check("batch_below_threshold", batch, 4, false)
				check("batch_or_threshold", batch, 5, true)
				// Neither population independently reaches its threshold.
				history = []string{}
				for i := 0; i < 9; i++ {
					history = append(history, prior)
				}
				for i := 0; i < 4; i++ {
					history = append(history, batch)
				}
				if yes, _, e := search.Execute(&out); e != nil || yes {
					t.Fatalf("mixed subthreshold history accepted: %v %v", yes, e)
				}
			}
		})
	}
	if len(cases) != 22 {
		t.Fatalf("history coverage %d", len(cases))
	}
}
