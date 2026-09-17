package main

// These are offline parser/rule contracts, not the closed EventProcessor. The
// model concatenates documented grok patterns, executes them with Go RE2, applies
// the filter's transforms, and uses the real SDK for CEL and Event conversion.
// KV is omitted deliberately: every asserted/detection field must be recovered
// by the quote-aware grok steps. Dynamic enrichment and live correlation are not run.
import (
	"bytes"
	"encoding/json"
	"fmt"
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

type fwFixture struct {
	Name     string         `json:"name"`
	Raw      string         `json:"raw"`
	Expected map[string]any `json:"expected"`
	Absent   []string       `json:"absent"`
	Matches  []string       `json:"matches"`
}

func fwPut(m map[string]any, path string, value any, remove bool) {
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
func fwGet(m map[string]any, p string) (any, bool) {
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
func fwConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/fortinet/fortiweb.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}
func fwRegex(t *testing.T, g *plugins.Grok, cfg *plugins.Config) *regexp.Regexp {
	t.Helper()
	var pattern strings.Builder
	for i, p := range g.Patterns {
		if p.FieldName != "" {
			fmt.Fprintf(&pattern, "(?P<f%d>%s)", i, p.Pattern)
		} else {
			pattern.WriteString("(?:" + p.Pattern + ")")
		}
	}
	pats := map[string]string{"greedy": ".*"}
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
func fwParse(t *testing.T, cfg *plugins.Config, raw string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "firewall-fortiweb", "log": map[string]any{}}
	for _, stage := range cfg.Pipeline {
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
					v, ok := fwGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					r := fwRegex(t, g, cfg)
					m := r.FindStringSubmatch(str)
					if m == nil {
						continue
					}
					for i, p := range g.Patterns {
						if p.FieldName != "" {
							fwPut(draft, p.FieldName, m[r.SubexpIndex(fmt.Sprintf("f%d", i))], false)
						}
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := fwGet(draft, p); ok {
							fwPut(draft, s.Rename.To, v, false)
							fwPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					for _, p := range s.Trim.Fields {
						if v, ok := fwGet(draft, p); ok {
							str := v.(string)
							switch s.Trim.Function {
							case "prefix":
								str = strings.TrimPrefix(str, s.Trim.Substring)
							case "suffix":
								str = strings.TrimSuffix(str, s.Trim.Substring)
							default:
								t.Fatalf("unsupported trim %s", s.Trim.Function)
							}
							fwPut(draft, p, str, false)
						}
					}
				case "add":
					fwPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						fwPut(draft, p, nil, true)
					}
				case "kv", "dynamic": // Intentional exclusions described above.
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
func fwRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths, e := filepath.Glob("../../rules/fortinet/fortiweb/*.yml")
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
func TestFortiWebRawContracts(t *testing.T) {
	cfg := fwConfig(t)
	rules := fwRules(t)
	cache := plugins.NewCELCache("fortiweb-raw-contract")
	b, e := os.ReadFile("testdata/fortiweb_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixtures []fwFixture
	if e = json.Unmarshal(b, &fixtures); e != nil {
		t.Fatal(e)
	}
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			out := fwParse(t, cfg, f.Raw, cache)
			for p, want := range f.Expected {
				got := gjson.Get(out, p)
				if !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("%s: got %v, want %v", p, got.Value(), want)
				}
			}
			for _, p := range f.Absent {
				if gjson.Get(out, p).Exists() {
					t.Errorf("unexpected %s", p)
				}
			}
			if gjson.Get(out, "raw").String() != f.Raw {
				t.Error("raw changed")
			}
			want := map[string]bool{}
			for _, name := range f.Matches {
				if rules[name] == nil {
					t.Fatalf("unknown expected rule %s", name)
				}
				want[name] = true
			}
			for name, r := range rules {
				got, e := cache.Eval(r.Where, out)
				if e != nil {
					t.Errorf("%s: %v", name, e)
				} else if got != want[name] {
					t.Errorf("%s trigger=%v, want %v", name, got, want[name])
				}
				if got {
					for _, search := range r.Correlation {
						for _, term := range search.With {
							if !gjson.Get(out, term.Field).Exists() {
								t.Errorf("%s: matched event lacks history field %s", name, term.Field)
							}
						}
					}
				}
			}
		})
	}
}
func TestFortiWebRuleHistoryAndGrouping(t *testing.T) {
	for name, r := range fwRules(t) {
		t.Run(name, func(t *testing.T) {
			if r.Adversary != "origin" {
				t.Error("request source must be adversary")
			}
			if len(r.GroupBy) > 0 && len(r.DeduplicateBy) > 0 {
				t.Error("grouping and deduplication are exclusive")
			}
			for _, p := range append(r.GroupBy, r.DeduplicateBy...) {
				if p != "adversary.ip" && p != "target.ip" && p != "lastEvent.log.subtype" {
					t.Errorf("unexpected alert path %s", p)
				}
			}
			for _, s := range r.Correlation {
				if s.Count <= 1 {
					t.Error("single detections must not require a history query")
				}
				fields := map[string]bool{}
				for _, x := range s.With {
					if x.Operator != "filter_term" {
						t.Error("history must use exact matches")
					}
					fields[x.Field] = true
					if x.Value.GetStringValue() != "{{."+x.Field+"}}" {
						t.Errorf("history value does not use trigger's %s", x.Field)
					}
				}
				for _, p := range []string{"origin.ip", "target.ip", "log.type", "log.subtype", "log.action"} {
					if !fields[p] {
						t.Errorf("history missing %s", p)
					}
				}
			}
		})
	}
}
