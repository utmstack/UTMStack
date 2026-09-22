package main

// Offline Bitdefender extraction model, not the closed EventProcessor.
// Explicit YAML grok/rename/cast/trim/add/delete and observed KV splitting are
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

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

type bitdefFixture struct {
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	Raw        string         `json:"raw"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	Matches    []string       `json:"matches"`
}

func bitdefPut(m map[string]any, path string, value any, remove bool) {
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
func bitdefGet(m map[string]any, p string) (any, bool) {
	var v any = m
	for _, k := range strings.Split(p, ".") {
		n, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok = n[k]
		if !ok {
			return nil, false
		}
	}
	return v, true
}
func bitdefConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/antivirus/bitdefender_gz.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}
func bitdefRegex(t *testing.T, g *plugins.Grok, cfg *plugins.Config) *regexp.Regexp {
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
func bitdefParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "antivirus-bitdefender-gz", "dataSource": dataSource, "log": map[string]any{}}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dataType := range stage.DataTypes {
			if dataType == "antivirus-bitdefender-gz" {
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
					v, ok := bitdefGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					r := bitdefRegex(t, g, cfg)
					m := r.FindStringSubmatch(str)
					if m == nil {
						continue
					}
					for i, p := range g.Patterns {
						if p.FieldName != "" {
							bitdefPut(draft, p.FieldName, m[r.SubexpIndex(fmt.Sprintf("f%d", i))], false)
						}
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := bitdefGet(draft, p); ok {
							bitdefPut(draft, s.Rename.To, v, false)
							bitdefPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					for _, p := range s.Trim.Fields {
						if v, ok := bitdefGet(draft, p); ok {
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
							bitdefPut(draft, p, str, false)
						}
					}
				case "add":
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add function %s", s.Add.Function)
					}
					bitdefPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						bitdefPut(draft, p, nil, true)
					}
				case "kv":
					v, ok := bitdefGet(draft, s.Kv.Source)
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
							bitdefPut(draft, "log."+key, pair[1], false)
						}
					}
				case "dynamic":
					if s.Dynamic.Plugin != "com.utmstack.geolocation" {
						t.Fatalf("unsupported dynamic plugin %s", s.Dynamic.Plugin)
					}
					field := s.Dynamic.Params["source"].GetStringValue()
					v, ok := bitdefGet(draft, field)
					if !ok {
						t.Fatalf("missing dynamic source %s", field)
					}
					ip := net.ParseIP(fmt.Sprint(v))
					if ip == nil || ip.IsUnspecified() {
						t.Fatalf("invalid address reaches geolocation: %s", field)
					}
					// The external geolocation service is not executed.
				case "json":
					source, ok := bitdefGet(draft, s.Json.Source)
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
					for key, value := range bitdefSanitizeJSON(parsed) {
						bitdefPut(draft, "log."+key, value, false)
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := bitdefGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								bitdefPut(draft, field, utils.CastString(value), false)
							case "int":
								bitdefPut(draft, field, utils.CastInt64(value), false)
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
	if e = utils.StringToProtoMessage(&in, ev); e != nil {
		t.Fatal(e)
	}
	out, e := utils.ProtoMessageToString(ev)
	if e != nil {
		t.Fatal(e)
	}
	return *out
}
func bitdefRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths, e := filepath.Glob("../../rules/antivirus/bitdefender_gz/*.y*ml")
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

func bitdefSanitizeJSON(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		utils.SanitizeField(&key)
		if nested, ok := value.(map[string]any); ok {
			value = bitdefSanitizeJSON(nested)
		}
		out[key] = value
	}
	return out
}

func bitdefFixtures(t *testing.T) []bitdefFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/bitdefender_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []bitdefFixture
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	return cases
}

func TestBitdefenderRawContracts(t *testing.T) {
	cfg, rules, cache := bitdefConfig(t), bitdefRules(t), plugins.NewCELCache("bitdefender-raw")
	positive := map[string]int{}
	negative := map[string]int{}
	if len(rules) != 21 {
		t.Fatalf("rules: %d", len(rules))
	}
	for _, f := range bitdefFixtures(t) {
		t.Run(f.Name, func(t *testing.T) {
			out := bitdefParse(t, cfg, f.Raw, f.DataSource, cache)
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
				t.Error("raw altered")
			}
			expected := map[string]bool{}
			for _, n := range f.Matches {
				expected[n] = true
			}
			for name, r := range rules {
				yes, e := cache.Eval(r.Where, out)
				if e != nil {
					t.Fatalf("%s: %v", name, e)
				}
				if yes != expected[name] {
					t.Errorf("%s matched %v want %v", name, yes, expected[name])
				}
				if yes {
					positive[name]++
				} else {
					negative[name]++
				}
				if yes {
					for _, search := range r.Correlation {
						for _, term := range search.With {
							value := term.Value.GetStringValue()
							if strings.HasPrefix(value, "{{.") {
								field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
								if !gjson.Get(out, field).Exists() {
									t.Errorf("%s unresolved %s", name, field)
								}
							}
						}
					}
				}
				if yes {
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
					if gjson.Get(out, "target.ip").String() != gjson.Get(*wire, "target.ip").String() {
						t.Error("endpoint identity lost")
					}
					if gjson.Get(out, "origin.ip").String() != gjson.Get(*wire, "adversary.ip").String() {
						t.Error("attacker identity lost")
					}
				}
			}
		})
	}
	for name := range rules {
		if positive[name] == 0 || negative[name] == 0 {
			t.Errorf("%s missing positive/negative coverage: %d/%d", name, positive[name], negative[name])
		}
	}
}
func TestBitdefenderPrivateReplay(t *testing.T) {
	p := os.Getenv("BITDEFENDER_PRIVATE_DOCUMENTS")
	if p == "" {
		t.Skip("private live records supplied separately")
	}
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	var docs []struct {
		ID       string         `json:"id"`
		Index    string         `json:"index"`
		Instance string         `json:"instance"`
		Source   map[string]any `json:"source"`
	}
	if e = json.Unmarshal(b, &docs); e != nil {
		t.Fatal(e)
	}
	cfg, rules, cache := bitdefConfig(t), bitdefRules(t), plugins.NewCELCache("bitdefender-private")
	results := []map[string]any{}
	for _, d := range docs {
		out := bitdefParse(t, cfg, d.Source["raw"].(string), d.Source["dataSource"].(string), cache)
		matches := []string{}
		for name, r := range rules {
			yes, e := cache.Eval(r.Where, out)
			if e != nil {
				t.Fatal(e)
			}
			if yes {
				matches = append(matches, name)
			}
		}
		var parsed map[string]any
		if e = json.Unmarshal([]byte(out), &parsed); e != nil {
			t.Fatal(e)
		}
		results = append(results, map[string]any{"id": d.ID, "index": d.Index, "instance": d.Instance, "parsed": parsed, "matches": matches})
	}
	if p := os.Getenv("BITDEFENDER_PRIVATE_OUTPUT"); p != "" {
		b, e := json.MarshalIndent(results, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("replayed %d private records; predicate candidates are not observed alerts", len(docs))
}
