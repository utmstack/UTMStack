package main

import (
	"path/filepath"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// Both rules used the undefined function oneof and failed to compile on every Google event.
func TestGoogleAuditChangeRulesCompile(t *testing.T) {
	cache := plugins.NewCELCache("google-audit-change-rules")
	load := func(name string) *plugins.Rule {
		b, err := utils.ReadPbYaml(filepath.Join("../..", "rules/cloud/google", name+".yml"))
		if err != nil {
			t.Fatal(err)
		}
		rule := new(plugins.Rule)
		if err := protojson.Unmarshal(b, rule); err != nil {
			t.Fatal(err)
		}
		return rule
	}
	sink, iam := load("gcp_logging_sink_modified"), load("gcp_iam_policy_changed")
	event := func(service, method string) string {
		return `{"dataType":"google","origin":{"user":"admin@example.test"},"log":{` +
			`"protoPayloadServiceName":"` + service + `","protoPayloadMethodName":"` + method + `",` +
			`"logName":"projects/example/logs/cloudaudit.googleapis.com%2Factivity",` +
			`"protoPayload":{"request":{"policy":{"bindings":[{"role":"roles/owner"}]}}}}}`
	}
	cases := []struct {
		name  string
		rule  *plugins.Rule
		event string
		want  bool
	}{
		{"sink-created", sink, event("logging.googleapis.com", "google.logging.v2.ConfigServiceV2.CreateSink"), true},
		{"sink-listed", sink, event("logging.googleapis.com", "google.logging.v2.ConfigServiceV2.ListSinks"), false},
		{"iam-policy-set", iam, event("cloudresourcemanager.googleapis.com", "SetIamPolicy"), true},
		{"iam-policy-read", iam, event("cloudresourcemanager.googleapis.com", "GetIamPolicy"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cache.Eval(tc.rule.Where, tc.event)
			if err != nil || got != tc.want {
				t.Errorf("%s = %v (%v); want %v", tc.rule.Name, got, err, tc.want)
			}
		})
	}
}
