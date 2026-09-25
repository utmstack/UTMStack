package main

// Ordered-step model of the ESXi sign-in and outcome steps, run with this
// module's go-sdk CEL. It starts from the parsed log.message and follows the
// filter from the session grok onward; the grok model mirrors the EventProcessor
// grok plugin (trim before each pattern, prefix-only matches, all or nothing).
// Header parsing of the raw fixture lines runs in the isolated parser instead.
import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

type esxiCase struct {
	Name     string         `json:"name"`
	Message  string         `json:"message"`
	Result   string         `json:"result"`
	Expected map[string]any `json:"expected"`
	Absent   []string       `json:"absent"`
}

func esxiGrok(t *testing.T, event map[string]any, g *plugins.Grok) {
	t.Helper()
	value, ok := valueAt(event, g.Source)
	if !ok {
		return
	}
	remaining, _ := value.(string)
	found := map[string]string{}
	matched := 0
	for _, p := range g.Patterns {
		remaining = strings.TrimSpace(remaining)
		if remaining == "" {
			break
		}
		re, err := regexp.Compile(p.Pattern)
		if err != nil {
			t.Fatal(err)
		}
		match := re.FindString(remaining)
		if match == "" || !strings.HasPrefix(remaining, match) {
			break
		}
		matched++
		name := p.FieldName
		utils.SanitizeField(&name)
		found[name] = strings.TrimSpace(match)
		remaining = strings.TrimPrefix(remaining, match)
	}
	if matched != len(g.Patterns) {
		return
	}
	for name, v := range found {
		put(event, name, v, false)
	}
}

func esxiRun(t *testing.T, cfg *plugins.Config, cache *plugins.CELCache, message string) string {
	t.Helper()
	event := map[string]any{"log": map[string]any{}}
	if message != "" {
		put(event, "log.message", message, false)
	}
	state := func() string {
		b, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	where := func(expr string) bool {
		if expr == "" {
			return true
		}
		ok, err := cache.Eval(expr, state())
		if err != nil {
			t.Fatalf("%s: %v", expr, err)
		}
		return ok
	}
	started := false
	for _, stage := range cfg.Pipeline {
		for _, step := range stage.Steps {
			if g := step.Grok; g != nil && len(g.Patterns) > 0 && g.Patterns[0].FieldName == "log.loginEvent" {
				started = true
			}
			if !started {
				continue
			}
			switch {
			case step.Grok != nil:
				if where(step.Grok.Where) {
					esxiGrok(t, event, step.Grok)
				}
			case step.Trim != nil:
				for _, f := range step.Trim.Fields {
					if v, ok := valueAt(event, f); ok {
						s, _ := v.(string)
						if step.Trim.Function == "suffix" {
							s = strings.TrimSuffix(s, step.Trim.Substring)
						} else {
							s = strings.TrimPrefix(s, step.Trim.Substring)
						}
						put(event, f, s, false)
					}
				}
			case step.Rename != nil:
				if where(step.Rename.Where) {
					for _, f := range step.Rename.From {
						if v, ok := valueAt(event, f); ok {
							put(event, step.Rename.To, v, false)
							put(event, f, nil, true)
						}
					}
				}
			case step.Add != nil:
				if where(step.Add.Where) {
					put(event, step.Add.Params["key"].GetStringValue(), step.Add.Params["value"].GetStringValue(), false)
				}
			case step.Delete != nil:
				for _, f := range step.Delete.Fields {
					put(event, f, nil, true)
				}
			}
		}
	}
	if !started {
		t.Fatal("missing session sign-in step")
	}
	return state()
}

func TestVMwareESXiActionResultContract(t *testing.T) {
	data, err := os.ReadFile("testdata/vmware_esxi_action_result.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []esxiCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 10 {
		t.Fatal("missing outcome regression classes")
	}
	b, err := utils.ReadPbYaml("../../filters/vmware/vmware-esxi.yml")
	if err != nil {
		t.Fatal(err)
	}
	cfg := new(plugins.Config)
	if err := protojson.Unmarshal(b, cfg); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("vmware-esxi-final-outcome")
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			state := esxiRun(t, cfg, cache, tc.Message)
			got := gjson.Get(state, "actionResult")
			if tc.Result == "" && got.Exists() {
				t.Fatalf("unexpected actionResult %s", got.Raw)
			}
			if tc.Result != "" && got.String() != tc.Result {
				t.Fatalf("actionResult = %s, want %q", got.Raw, tc.Result)
			}
			for path, want := range tc.Expected {
				if v := gjson.Get(state, path); v.String() != want {
					t.Fatalf("%s = %s, want %v", path, v.Raw, want)
				}
			}
			for _, path := range tc.Absent {
				if gjson.Get(state, path).Exists() {
					t.Fatalf("%s should be absent: %s", path, state)
				}
			}
			for _, scratch := range []string{"log.loginEvent", "log.loginUser", "log.loginAddress", "log.loginClient"} {
				if gjson.Get(state, scratch).Exists() {
					t.Fatalf("scratch field %s kept", scratch)
				}
			}
			for _, result := range []string{"success", "failure", "denied"} {
				matched, err := cache.Eval(`equals("actionResult","`+result+`")`, state)
				if err != nil || matched != (tc.Result == result) {
					t.Errorf("%s predicate = %v (%v)", result, matched, err)
				}
			}
		})
	}
}
