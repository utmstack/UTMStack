package main

// Offline Azure extraction model, not the closed EventProcessor.
// Explicit YAML JSON/key sanitization, grok, rename, add and delete steps are modeled.
// CEL and Event serialization use SDK v1.1.31. History requests are tested separately
// with that SDK. External geolocation is mocked only when a fixture declares it.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

type azureFixture struct {
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	Raw        string         `json:"raw"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	Matches    []string       `json:"matches"`
	Enrichment map[string]any `json:"enrichment"`
}

func azurePut(m map[string]any, path string, value any, remove bool) {
	p := azurePath(path)
	for _, k := range p[:len(p)-1] {
		n, ok := m[k].(map[string]any)
		if !ok {
			if remove {
				return
			}
			n = map[string]any{}
			m[k] = n
		}
		m = n
	}
	if remove {
		delete(m, p[len(p)-1])
	} else {
		m[p[len(p)-1]] = value
	}
}
func azureGet(m map[string]any, p string) (any, bool) {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	value := gjson.GetBytes(b, p)
	return value.Value(), value.Exists()
}
func azurePath(path string) []string {
	var out []string
	var part strings.Builder
	escaped := false
	for _, ch := range path {
		if escaped {
			part.WriteRune(ch)
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else if ch == '.' {
			out = append(out, part.String())
			part.Reset()
		} else {
			part.WriteRune(ch)
		}
	}
	return append(out, part.String())
}
func azureConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/azure/azure-eventhub.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}
func azureRegex(t *testing.T, g *plugins.Grok, cfg *plugins.Config) *regexp.Regexp {
	t.Helper()
	var pattern strings.Builder
	for i, p := range g.Patterns {
		if p.FieldName != "" {
			fmt.Fprintf(&pattern, "(?P<f%d>%s)", i, p.Pattern)
		} else {
			pattern.WriteString("(?:" + p.Pattern + ")")
		}
	}
	pats := map[string]string{"greedy": ".*", "data": ".*?", "word": "[A-Za-z0-9_-]+", "space": "\\s+"}
	for k, v := range cfg.Patterns {
		pats[k] = v
	}
	tmpl, e := template.New("grok").Option("missingkey=error").Parse(pattern.String())
	if e != nil {
		t.Fatal(e)
	}
	var b bytes.Buffer
	if e = tmpl.Execute(&b, pats); e != nil {
		t.Fatal(e)
	}
	r, e := regexp.Compile(b.String())
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func azureParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache, enrichment ...map[string]any) string {
	return azureParseMode(t, cfg, raw, dataSource, cache, false, enrichment...)
}

// Both modes model the unresolved nested-key behavior of the closed JSON step.
func azureParseMode(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache, preserveNested bool, enrichment ...map[string]any) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "azure", "dataSource": dataSource, "log": map[string]any{}}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dataType := range stage.DataTypes {
			if dataType == "azure" {
				matched = true
			}
		}
		if !matched {
			continue
		}
		for _, s := range stage.Steps {
			b, e := protojson.Marshal(s)
			if e != nil {
				t.Fatal(e)
			}
			var step map[string]map[string]any
			if e = json.Unmarshal(b, &step); e != nil {
				t.Fatal(e)
			}
			for kind, body := range step {
				if w, ok := body["where"].(string); ok && w != "" {
					snapshot, err := json.Marshal(draft)
					if err != nil {
						t.Fatal(err)
					}
					match, e := cache.Eval(w, string(snapshot))
					if e != nil {
						t.Fatal(e)
					}
					if !match {
						continue
					}
				}
				switch kind {
				case "grok":
					g := s.Grok
					src := g.Source
					if src == "" {
						src = "raw"
					}
					v, ok := azureGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					r := azureRegex(t, g, cfg)
					m := r.FindStringSubmatch(str)
					if m == nil {
						continue
					}
					for i, p := range g.Patterns {
						if p.FieldName != "" {
							azurePut(draft, p.FieldName, m[r.SubexpIndex(fmt.Sprintf("f%d", i))], false)
						}
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := azureGet(draft, p); ok {
							azurePut(draft, s.Rename.To, v, false)
							azurePut(draft, p, nil, true)
							break
						}
					}
				case "add":
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add function %s", s.Add.Function)
					}
					azurePut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						azurePut(draft, p, nil, true)
					}
				case "dynamic":
					if s.Dynamic.Plugin != "com.utmstack.geolocation" {
						t.Fatalf("unsupported dynamic plugin %s", s.Dynamic.Plugin)
					}
					field := s.Dynamic.Params["source"].GetStringValue()
					v, ok := azureGet(draft, field)
					if !ok {
						t.Fatalf("missing dynamic source %s", field)
					}
					ip := net.ParseIP(fmt.Sprint(v))
					if ip == nil || ip.IsUnspecified() {
						t.Fatalf("invalid address reaches geolocation: %s", field)
					}
					// The external geolocation service is not executed. Fixtures may explicitly supply its mocked output.
					for _, fields := range enrichment {
						for path, value := range fields {
							if !strings.HasPrefix(path, "origin.geolocation.") {
								t.Fatal("unexpected enrichment field")
							}
							azurePut(draft, path, value, false)
						}
					}
				case "json":
					source, ok := azureGet(draft, s.Json.Source)
					if !ok {
						continue
					}
					str, ok := source.(string)
					if !ok {
						t.Fatalf("JSON source is not a string")
					}
					var parsed map[string]any
					if e := json.Unmarshal([]byte(str), &parsed); e != nil {
						t.Fatal(e)
					}
					normalized := azureSanitizeJSON(parsed)
					if preserveNested {
						normalized = map[string]any{}
						for key, value := range parsed {
							utils.SanitizeField(&key)
							normalized[key] = value
						}
					}
					for key, value := range normalized {
						azurePut(draft, "log."+key, value, false)
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := azureGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								azurePut(draft, field, utils.CastString(value), false)
							case "float":
								azurePut(draft, field, utils.CastFloat64(value), false)
							case "int":
								azurePut(draft, field, utils.CastInt64(value), false)
							default:
								t.Fatalf("unsupported cast %s", s.Cast.To)
							}
						}
					}
				case "drop":
					return ""
				default:
					t.Fatalf("unsupported filter step %s", kind)
				}
			}
		}
	}
	b, e := json.Marshal(draft)
	if e != nil {
		t.Fatal(e)
	}
	in := string(b)
	ev := new(plugins.Event)
	// Check before the permissive SDK converter: no scratch or unknown standard field may survive.
	if e = protojson.Unmarshal(b, ev); e != nil {
		t.Fatalf("strict final Draft schema: %v", e)
	}
	if e = utils.StringToProtoMessage(&in, ev); e != nil {
		t.Fatal(e)
	}
	out, e := utils.ProtoMessageToString(ev)
	if e != nil {
		t.Fatal(e)
	}
	return *out
}
func azureRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths := []string{}
	e := filepath.WalkDir("../../rules/cloud/azure", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			paths = append(paths, path)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}

	out := map[string]*plugins.Rule{}
	for _, p := range paths {
		b, e := utils.ReadPbYaml(p)
		if e != nil {
			t.Fatal(e)
		}
		r := new(plugins.Rule)
		if e = protojson.Unmarshal(b, r); e != nil {
			t.Fatal(e)
		}
		r.Normalize()
		out[strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))] = r
	}
	return out
}

func azureSanitizeJSON(input map[string]any) map[string]any {
	var walk func(any) any
	walk = func(input any) any {
		switch v := input.(type) {
		case map[string]any:
			out := map[string]any{}
			for key, value := range v {
				utils.SanitizeField(&key)
				out[key] = walk(value)
			}
			return out
		case []any:
			out := make([]any, len(v))
			for i, x := range v {
				out[i] = walk(x)
			}
			return out
		default:
			return input
		}
	}
	return walk(input).(map[string]any)
}

func azureFixtures(t *testing.T) []azureFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/azure_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []azureFixture
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	return cases
}
func azureAssert(t *testing.T, f azureFixture, out string, rules map[string]*plugins.Rule, cache *plugins.CELCache) map[string]bool {
	t.Helper()
	for field, want := range f.Expected {
		got := gjson.Get(out, field)
		b, _ := json.Marshal(want)
		if !got.Exists() || got.Raw != string(b) {
			if fmt.Sprint(got.Value()) != fmt.Sprint(want) {
				t.Errorf("%s got %v want %v", field, got.Value(), want)
			}
		}
	}
	for _, field := range f.Absent {
		if gjson.Get(out, field).Exists() {
			t.Errorf("unexpected %s", field)
		}
	}
	if gjson.Get(out, "raw").String() != f.Raw {
		t.Error("raw altered")
	}
	expected := map[string]bool{}
	for _, name := range f.Matches {
		expected[name] = true
	}
	actual := map[string]bool{}
	for name, r := range rules {
		yes, e := cache.Eval(r.Where, out)
		if e != nil {
			t.Fatalf("%s: %v", name, e)
		}
		if yes {
			actual[name] = true
		}
		if yes != expected[name] {
			t.Errorf("%s matched %v want %v", name, yes, expected[name])
		}
		if yes {
			for _, search := range r.Correlation {
				for _, term := range search.With {
					v := term.Value.GetStringValue()
					if strings.HasPrefix(v, "{{.") {
						path := strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")
						if !gjson.Get(out, path).Exists() {
							t.Errorf("%s unresolved %s", name, path)
						}
					}
				}
			}
			ev := new(plugins.Event)
			if e := utils.StringToProtoMessage(&out, ev); e != nil {
				t.Fatal(e)
			}
			if r.Adversary != "origin" {
				t.Errorf("unexpected actor direction %s", r.Adversary)
			}
			alert := &plugins.Alert{Adversary: ev.Origin, Target: ev.Target, Events: []*plugins.Event{ev}}
			wire, e := utils.ProtoMessageToString(alert)
			if e != nil {
				t.Fatal(e)
			}
			if gjson.Get(*wire, "adversary.ip").String() != gjson.Get(out, "origin.ip").String() {
				t.Error("actor IP lost")
			}
		}
	}
	return actual
}
func TestAzureRawContracts(t *testing.T) {
	cfg, rules, cache := azureConfig(t), azureRules(t), plugins.NewCELCache("azure-raw")
	if len(rules) != 40 {
		t.Fatalf("rules %d", len(rules))
	}
	positive, negative := map[string]int{}, map[string]int{}
	for _, f := range azureFixtures(t) {
		for _, preserve := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/nested-preserved-%v", f.Name, preserve), func(t *testing.T) {
				out := azureParseMode(t, cfg, f.Raw, f.DataSource, cache, preserve, f.Enrichment)
				actual := azureAssert(t, f, out, rules, cache)
				for name := range rules {
					if actual[name] {
						positive[name]++
					} else {
						negative[name]++
					}
				}
			})
		}
	}
	b, e := os.ReadFile("testdata/azure_changed_rules.json")
	if e != nil {
		t.Fatal(e)
	}
	var changed []string
	if e = json.Unmarshal(b, &changed); e != nil {
		t.Fatal(e)
	}
	for _, name := range changed {
		if positive[name] == 0 || negative[name] == 0 {
			t.Errorf("%s missing positive/negative raw fixture", name)
		}
	}
}
func TestAzurePrivateRecords(t *testing.T) {
	path := os.Getenv("AZURE_PRIVATE_FIXTURES")
	if path == "" {
		t.Skip("private native record expectations not supplied")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var cases []azureFixture
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	cfg, rules, cache := azureConfig(t), azureRules(t), plugins.NewCELCache("azure-private")
	var outputs []map[string]any
	for _, f := range cases {
		t.Run(f.Name, func(t *testing.T) {
			out := azureParseMode(t, cfg, f.Raw, f.DataSource, cache, true)
			azureAssert(t, f, out, rules, cache)
			var v map[string]any
			if e = json.Unmarshal([]byte(out), &v); e != nil {
				t.Fatal(e)
			}
			outputs = append(outputs, map[string]any{"id": f.Name, "parsed": v})
		})
	}
	if path := os.Getenv("AZURE_PRIVATE_OUTPUT"); path != "" {
		b, e := json.MarshalIndent(outputs, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}
