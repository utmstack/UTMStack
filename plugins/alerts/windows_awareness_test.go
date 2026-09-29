package main

// Windows administrator-awareness rules. The fabricated raw fixtures in
// testdata/filter-contracts/windows-awareness.json run through the modeled
// Windows filter; every Windows rule is then evaluated with the pinned go-sdk.
// A fixture must match exactly its listed rules, and each groupBy path of a
// matched rule must resolve to a scalar the way the alerts plugin resolves it.
// Parser plugins, history, indexing and notifications are not executed here.
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestWindowsAwarenessRules(t *testing.T) {
	var fixtures []Fixture
	var awareness []string
	for _, manifest := range loadFilterContracts(t) {
		if manifest.Technology == "Windows administrator awareness" {
			fixtures = append(fixtures, manifest.Fixtures...)
			awareness = append(awareness, manifest.Rules...)
		}
	}
	if len(fixtures) == 0 || len(awareness) == 0 {
		t.Fatal("no Windows awareness manifest")
	}
	paths, err := filepath.Glob("../../rules/windows/*.yml")
	if err != nil {
		t.Fatal(err)
	}
	rules := map[string]*plugins.Rule{}
	for _, p := range paths {
		b, err := utils.ReadPbYaml(p)
		if err != nil {
			t.Fatal(err)
		}
		rule := new(plugins.Rule)
		if err = protojson.Unmarshal(b, rule); err != nil {
			t.Fatal(err)
		}
		rule.Normalize()
		rules[strings.TrimPrefix(p, "../../")] = rule
	}
	cache := plugins.NewCELCache("windows-awareness-test")
	covered := map[string]bool{}
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			out, issues, err := normalize("../..", f, cache)
			if err != nil {
				t.Fatal(err)
			}
			for _, issue := range issues {
				t.Error(issue)
			}
			for path, rule := range rules {
				matched, err := cache.Eval(rule.Where, out)
				if err != nil {
					t.Fatalf("%s: %v", path, err)
				}
				if matched != f.Rules[path] {
					t.Errorf("%s matched=%v, want %v", path, matched, f.Rules[path])
				}
				if !matched {
					continue
				}
				covered[path] = true
				if len(rule.GroupBy) == 0 || len(rule.DeduplicateBy) != 0 || len(rule.Correlation) != 0 {
					t.Errorf("%s: awareness rules group repeated events and need no history", path)
				}
				event := new(plugins.Event)
				if err = utils.StringToProtoMessage(&out, event); err != nil {
					t.Fatal(err)
				}
				alert := &plugins.Alert{Name: rule.Name, DataType: event.DataType, DataSource: event.DataSource,
					Adversary: event.Origin, Target: event.Target, Events: []*plugins.Event{event}, GroupBy: rule.GroupBy}
				wire, err := utils.ProtoMessageToString(alert)
				if err != nil {
					t.Fatal(err)
				}
				for _, field := range rule.GroupBy {
					if _, ok := scalarGroupingValue(alertGroupingValue(*wire, field)); !ok {
						t.Errorf("%s: groupBy %s has no scalar value", path, field)
					}
				}
			}
		})
	}
	for _, path := range awareness {
		if !covered[path] {
			t.Errorf("no positive fixture for %s", path)
		}
	}
}
