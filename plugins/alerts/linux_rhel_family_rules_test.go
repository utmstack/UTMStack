package main

// The rebuilt RHEL-family rules read what the Linux filter writes: journal keys
// renamed to log.syslogIdentifier and log.message, and the audit collector's
// records (log.avc, log.macstatus, log.paths, action, origin.process and
// origin.command). Fabricated raw records go through the ordered filter model and
// the real SDK CEL. deduplicateBy keys are resolved the way grouping.go resolves
// them, so every branch keeps a key more specific than the collector.
import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

var linuxRHELFamilyRules = []string{
	"linux/rhel_family/openshift_security_violations.yml",
	"linux/rhel_family/rpm_database_tampering.yml",
	"linux/rhel_family/secure_boot_violations.yml",
	"linux/rhel_family/selinux_policy_violations.yml",
	"linux/rhel_family/yum_dnf_repository_attacks.yml",
}

func TestLinuxRHELFamilyRules(t *testing.T) {
	data, err := os.ReadFile("testdata/linux_rhel_family_rules.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name  string   `json:"name"`
		Rule  string   `json:"rule"`
		Raw   string   `json:"raw"`
		Match bool     `json:"match"`
		Keys  []string `json:"deduplicateKeys"`
	}
	if err = json.Unmarshal(data, &cases); err != nil || len(cases) == 0 {
		t.Fatalf("fixtures: %v", err)
	}
	rules := map[string]*plugins.Rule{}
	for _, path := range linuxRHELFamilyRules {
		blob, err := utils.ReadPbYaml("../../rules/" + path)
		if err != nil {
			t.Fatal(err)
		}
		rule := new(plugins.Rule)
		if err = protojson.Unmarshal(blob, rule); err != nil {
			t.Fatal(err)
		}
		rule.Normalize()
		// Every firing of a groupBy rule still stores a child alert; these rules
		// deduplicate instead, and need no history query.
		if len(rule.Correlation) != 0 || len(rule.GroupBy) != 0 || len(rule.DeduplicateBy) < 2 || rule.DeduplicateBy[0] != "dataSource" {
			t.Fatalf("%s: want deduplicateBy starting with dataSource, no groupBy and no history", path)
		}
		rules[path] = rule
	}
	cache := plugins.NewCELCache("linux-rhel-family-rules")
	positives, negatives := map[string]int{}, map[string]int{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			rule, ok := rules[tc.Rule]
			if !ok {
				t.Fatalf("unknown rule %s", tc.Rule)
			}
			event := linuxActionResultNormalize(t, tc.Raw)
			matched, err := cache.Eval(rule.Where, event)
			if err != nil || matched != tc.Match {
				t.Fatalf("%s predicate=%v want %v: %v\n%s", tc.Rule, matched, tc.Match, err, event)
			}
			if !matched {
				negatives[tc.Rule]++
				return
			}
			positives[tc.Rule]++
			alert := `{"name":` + strconv.Quote(rule.Name) + `,"dataSource":"synthetic-collector","events":[` + event + `]}`
			var resolved []string
			for _, field := range rule.DeduplicateBy {
				if _, ok := scalarGroupingValue(alertGroupingValue(alert, field)); ok {
					resolved = append(resolved, field)
				}
			}
			if !reflect.DeepEqual(resolved, tc.Keys) {
				t.Errorf("deduplicateBy resolves %v, want %v", resolved, tc.Keys)
			}
		})
	}
	for _, path := range linuxRHELFamilyRules {
		if positives[path] == 0 || negatives[path] == 0 {
			t.Errorf("%s: missing positive or negative cases", path)
		}
	}
}
