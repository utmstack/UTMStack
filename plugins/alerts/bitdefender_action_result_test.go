package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

// Fabricated CEF payloads exercise the ordered filter model and the v11 SDK's
// CEL predicates: final outcomes, full multiword values, empty values and the
// history markers. The isolated parser is replayed separately.
func TestBitdefenderActionResultRaw(t *testing.T) {
	var cases []struct {
		Name       string            `json:"name"`
		Raw        string            `json:"raw"`
		DataSource string            `json:"dataSource"`
		Result     string            `json:"result"`
		Expected   map[string]string `json:"expected"`
		Absent     []string          `json:"absent"`
		Rule       string            `json:"rule"`
		Match      bool              `json:"match"`
	}
	data, err := os.ReadFile("testdata/bitdefender_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 12 {
		t.Fatal("task outcome and contradiction classes are missing")
	}
	config := bitdefConfig(t)
	rules := bitdefRules(t)
	cache := plugins.NewCELCache("bitdefender-task-outcome")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			out := bitdefParse(t, config, tc.Raw, tc.DataSource, cache)
			if got := gjson.Get(out, "actionResult").String(); got != tc.Result {
				t.Errorf("actionResult = %q; want %q", got, tc.Result)
			}
			for path, want := range tc.Expected {
				if got := gjson.Get(out, path); !got.Exists() || got.String() != want {
					t.Errorf("%s = %q; want %q", path, got.String(), want)
				}
			}
			for _, path := range tc.Absent {
				if gjson.Get(out, path).Exists() {
					t.Errorf("unexpected %s", path)
				}
			}
			for _, value := range []string{"success", "failure", "denied"} {
				got, err := cache.Eval(`equals("actionResult","`+value+`")`, out)
				if err != nil || got != (tc.Result == value) {
					t.Errorf("outcome predicate %s = %v (%v)", value, got, err)
				}
			}
			if tc.Rule != "" {
				name := strings.TrimSuffix(filepath.Base(tc.Rule), filepath.Ext(tc.Rule))
				rule := rules[name]
				if rule == nil {
					t.Fatalf("shipped rule %s missing", tc.Rule)
				}
				got, err := cache.Eval(rule.Where, out)
				if err != nil || got != tc.Match {
					t.Errorf("rule %s = %v (%v); want %v", tc.Rule, got, err, tc.Match)
				}
			}
		})
	}
}

// The parser plugins keep only letters, digits and dots in the names they
// write, while conditions, history terms, placeholders and grouping look names
// up exactly as written. Every name the filter and its rules use must
// therefore come through utils.SanitizeField unchanged.
func TestBitdefenderFieldNamesSurviveSanitizing(t *testing.T) {
	celField := regexp.MustCompile(`\b(?:equals|equalsIgnoreCase|oneOf|contains|containsAll|startsWith|endsWith|regexMatch|inCIDR|greaterThan|lessThan|greaterOrEqual|lessOrEqual|exists|safe)\(\s*"((?:[^"\\]|\\.)+)"`)
	placeholder := regexp.MustCompile(`\{\{\s*\.([^}\s]+)\s*\}\}`)
	check := func(where, name string) {
		t.Helper()
		cleaned := strings.TrimSuffix(name, ".keyword")
		utils.SanitizeField(&cleaned)
		if cleaned != strings.TrimSuffix(name, ".keyword") {
			t.Errorf("%s: %q is stored or looked up as %q", where, name, cleaned)
		}
	}
	checkWhere := func(where, expression string) {
		t.Helper()
		for _, m := range celField.FindAllStringSubmatch(expression, -1) {
			check(where, m[1])
		}
	}
	var walk func(where string, node any)
	walk = func(where string, node any) {
		switch n := node.(type) {
		case map[string]any:
			for k, v := range n {
				switch k {
				case "where":
					checkWhere(where, v.(string))
				case "fieldName", "key", "to", "source", "destination":
					if s, ok := v.(string); ok && s != "" {
						check(where+"."+k, s)
					}
				case "from", "fields":
					for _, s := range v.([]any) {
						check(where+"."+k, s.(string))
					}
				case "pattern", "value", "substring", "function", "plugin":
				default:
					walk(where+"."+k, v)
				}
			}
		case []any:
			for _, v := range n {
				walk(where, v)
			}
		}
	}
	for i, stage := range bitdefConfig(t).Pipeline {
		for j, step := range stage.Steps {
			b, err := protojson.Marshal(step)
			if err != nil {
				t.Fatal(err)
			}
			var doc any
			if err = json.Unmarshal(b, &doc); err != nil {
				t.Fatal(err)
			}
			walk(fmt.Sprintf("pipeline[%d].steps[%d]", i, j), doc)
		}
	}
	for name, rule := range bitdefRules(t) {
		checkWhere(name+".where", rule.Where)
		for _, search := range rule.Correlation {
			for _, term := range search.With {
				check(name+".with", term.Field)
				for _, m := range placeholder.FindAllStringSubmatch(term.Value.GetStringValue(), -1) {
					check(name+".placeholder", m[1])
				}
			}
		}
		for _, field := range append(append([]string{}, rule.GroupBy...), rule.DeduplicateBy...) {
			check(name+".grouping", field)
		}
	}
}
