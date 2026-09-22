package main

// Offline Sophos Central JSON extraction model, not the closed EventProcessor.
// Explicit YAML JSON, grok, rename, add and delete steps are
// modeled. CEL, Event serialization, placeholder expansion, query creation and
// history thresholds use SDK v1.1.31. External geolocation is not executed.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

type sophosCentralFixture struct {
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	Raw        string         `json:"raw"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	Matches    []string       `json:"matches"`
}

func sophosCentralPut(m map[string]any, path string, value any, remove bool) {
	p := strings.Split(path, ".")
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
func sophosCentralGet(m map[string]any, p string) (any, bool) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, false
	}
	v := gjson.GetBytes(b, p)
	return v.Value(), v.Exists()
}
func sophosCentralConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/sophos/sophos_central.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}
func sophosCentralRegex(t *testing.T, g *plugins.Grok, cfg *plugins.Config) *regexp.Regexp {
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
func sophosCentralParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "sophos-central", "dataSource": dataSource, "log": map[string]any{}}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dataType := range stage.DataTypes {
			if dataType == "sophos-central" {
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
					v, ok := sophosCentralGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					r := sophosCentralRegex(t, g, cfg)
					m := r.FindStringSubmatch(str)
					if m == nil {
						continue
					}
					for i, p := range g.Patterns {
						if p.FieldName != "" {
							sophosCentralPut(draft, p.FieldName, m[r.SubexpIndex(fmt.Sprintf("f%d", i))], false)
						}
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := sophosCentralGet(draft, p); ok {
							sophosCentralPut(draft, s.Rename.To, v, false)
							sophosCentralPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					for _, p := range s.Trim.Fields {
						if v, ok := sophosCentralGet(draft, p); ok {
							str, ok := v.(string)
							if !ok {
								continue
							}
							switch s.Trim.Function {
							case "prefix":
								str = strings.TrimPrefix(str, s.Trim.Substring)
							case "suffix":
								str = strings.TrimSuffix(str, s.Trim.Substring)
							default:
								t.Fatalf("unsupported trim %s", s.Trim.Function)
							}
							sophosCentralPut(draft, p, str, false)
						}
					}
				case "add":
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add function %s", s.Add.Function)
					}
					sophosCentralPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						sophosCentralPut(draft, p, nil, true)
					}
				case "kv":
					v, ok := sophosCentralGet(draft, s.Kv.Source)
					if !ok {
						continue
					}
					// Observed KV output splits quoted multiword values.
					// Explicit YAML grok steps rebuild consumed fields afterward.
					for _, item := range strings.Split(v.(string), s.Kv.FieldSplit) {
						pair := strings.SplitN(item, s.Kv.ValueSplit, 2)
						if len(pair) != 2 {
							continue
						}
						key := pair[0]
						utils.SanitizeField(&key)
						if key != "" {
							sophosCentralPut(draft, "log."+key, pair[1], false)
						}
					}
				case "dynamic":
					if s.Dynamic.Plugin != "com.utmstack.geolocation" {
						t.Fatalf("unsupported dynamic plugin %s", s.Dynamic.Plugin)
					}
					field := s.Dynamic.Params["source"].GetStringValue()
					v, ok := sophosCentralGet(draft, field)
					if !ok {
						t.Fatalf("missing dynamic source %s", field)
					}
					ip := net.ParseIP(fmt.Sprint(v))
					if ip == nil || ip.IsUnspecified() {
						t.Fatalf("invalid address reaches geolocation: %s", field)
					}
					// The external geolocation service is not executed.
				case "json":
					source, ok := sophosCentralGet(draft, s.Json.Source)
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
					for key, value := range sophosCentralSanitizeJSON(parsed) {
						sophosCentralPut(draft, "log."+key, value, false)
					}
				case "reformat":
					for _, field := range s.Reformat.Fields {
						value, ok := sophosCentralGet(draft, field)
						if !ok {
							continue
						}
						stamp, err := time.Parse(s.Reformat.FromFormat, fmt.Sprint(value))
						if err != nil {
							t.Fatalf("time conversion %s: %v", field, err)
						}
						sophosCentralPut(draft, field, stamp.Format(s.Reformat.ToFormat), false)
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := sophosCentralGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								sophosCentralPut(draft, field, utils.CastString(value), false)
							case "float":
								sophosCentralPut(draft, field, utils.CastFloat64(value), false)
							case "int":
								sophosCentralPut(draft, field, utils.CastInt64(value), false)
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
	if e = protojson.Unmarshal([]byte(in), ev); e != nil {
		t.Fatal(e)
	}
	out, e := utils.ProtoMessageToString(ev)
	if e != nil {
		t.Fatal(e)
	}
	return *out
}
func sophosCentralRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths, e := filepath.Glob("../../rules/sophos/sophos_central/*.yml")
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
		out[strings.TrimSuffix(filepath.Base(p), ".yml")] = r
	}
	return out
}

func sophosCentralSanitizeJSON(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		utils.SanitizeField(&key)
		if nested, ok := value.(map[string]any); ok {
			value = sophosCentralSanitizeJSON(nested)
		}
		out[key] = value
	}
	return out
}

func sophosCentralFixtures(t *testing.T) []sophosCentralFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/sophos_central_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []sophosCentralFixture
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	return cases
}

func sophosCentralCheck(t *testing.T, fixtures []sophosCentralFixture) {
	cfg, rules, cache := sophosCentralConfig(t), sophosCentralRules(t), plugins.NewCELCache("sophos_central")
	if len(rules) != 20 {
		t.Fatalf("rules: %d", len(rules))
	}
	coverage := map[string]int{}
	for _, f := range fixtures {
		for _, name := range f.Matches {
			coverage[name]++
		}
	}
	if len(fixtures) > 45 {
		for name := range rules {
			if coverage[name] == 0 || coverage[name] == len(fixtures) {
				t.Fatalf("missing positive/negative coverage for %s", name)
			}
		}
	}
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			out := sophosCentralParse(t, cfg, f.Raw, f.DataSource, cache)
			for field, want := range f.Expected {
				got := gjson.Get(out, field)
				if !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("%s got %v want %v", field, got.Value(), want)
				}
			}
			for _, field := range f.Absent {
				if gjson.Get(out, field).Exists() {
					t.Errorf("unexpected %s", field)
				}
			}
			if gjson.Get(out, "raw").String() != f.Raw {
				t.Error("raw changed")
			}
			expect := map[string]bool{}
			for _, n := range f.Matches {
				expect[n] = true
			}
			for name, r := range rules {
				matched, err := cache.Eval(r.Where, out)
				if err != nil {
					t.Fatal(err)
				}
				if matched != expect[name] {
					t.Errorf("%s matched %v want %v", name, matched, expect[name])
				}
				if !matched {
					continue
				}
				event := new(plugins.Event)
				if err := protojson.Unmarshal([]byte(out), event); err != nil {
					t.Fatal(err)
				}
				alert := &plugins.Alert{Events: []*plugins.Event{event}}
				switch r.Adversary {
				case "origin":
					alert.Adversary, alert.Target = event.Origin, event.Target
				case "target":
					alert.Adversary, alert.Target = event.Target, event.Origin
				default:
					t.Fatalf("unsupported adversary %s", r.Adversary)
				}
				if event.Target != nil && (alert.Target == nil || alert.Target.Host != event.Target.Host || alert.Target.Ip != event.Target.Ip) {
					t.Error("managed endpoint must remain alert target")
				}
				if event.Origin != nil && (alert.Adversary == nil || alert.Adversary.Ip != event.Origin.Ip) {
					t.Error("explicit outbound CryptoGuard initiator must remain adversary")
				}
				wire, err := utils.ProtoMessageToString(alert)
				if err != nil {
					t.Fatal(err)
				}
				for _, field := range r.GroupBy {
					path := strings.Replace(field, "lastEvent.", "events.0.", 1)
					if (field == "lastEvent.dataSource") && !gjson.Get(*wire, path).Exists() {
						t.Errorf("grouping scope missing: %s", field)
					}
				}
				for _, search := range r.Correlation {
					for _, term := range search.With {
						v := term.Value.GetStringValue()
						if strings.HasPrefix(v, "{{.") {
							field := strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")
							if !gjson.Get(out, field).Exists() {
								t.Errorf("unresolved %s", field)
							}
						}
					}
				}
			}
		})
	}
}
func TestSophosCentralRawContracts(t *testing.T) { sophosCentralCheck(t, sophosCentralFixtures(t)) }
func TestSophosCentralPrivateContracts(t *testing.T) {
	path := os.Getenv("SOPHOS_CENTRAL_PRIVATE_FIXTURES")
	if path == "" {
		t.Skip("private fixtures not supplied")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var f []sophosCentralFixture
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	sophosCentralCheck(t, f)
}
