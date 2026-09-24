package main

// These are offline parser/rule contracts, not the closed EventProcessor. The
// model concatenates documented grok patterns, executes them with Go RE2, applies
// the filter's transforms, and uses the real SDK for CEL and Event conversion.
// KV models observed space splitting so quoted payload tokens can contaminate
// intermediate fields; authoritative recovery must remove them. External
// geolocation is not run, but its input addresses are checked before enrichment.
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
	return fwParseSource(t, cfg, raw, "synthetic-fortiweb", cache)
}
func fwParseSource(t *testing.T, cfg *plugins.Config, raw, dataSource string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "firewall-fortiweb", "dataSource": dataSource, "log": map[string]any{}}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dataType := range stage.DataTypes {
			if dataType == "firewall-fortiweb" {
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
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add %s", s.Add.Function)
					}
					fwPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						fwPut(draft, p, nil, true)
					}
				case "kv":
					if s.Kv.FieldSplit != " " || s.Kv.ValueSplit != "=" {
						t.Fatal("unsupported KV separators")
					}
					value, ok := fwGet(draft, s.Kv.Source)
					if !ok {
						continue
					}
					for _, token := range strings.Split(value.(string), " ") {
						pair := strings.SplitN(token, "=", 2)
						if len(pair) == 2 {
							key := pair[0]
							utils.SanitizeField(&key)
							fwPut(draft, "log."+key, pair[1], false)
						}
					}
				case "dynamic":
					if s.Dynamic.Plugin != "com.utmstack.geolocation" {
						t.Fatal("unsupported dynamic plugin")
					}
					field := s.Dynamic.Params["source"].GetStringValue()
					value, ok := fwGet(draft, field)
					if !ok || net.ParseIP(fmt.Sprint(value)) == nil || net.ParseIP(fmt.Sprint(value)).IsUnspecified() {
						t.Errorf("invalid address reached geolocation at %s", field)
					}
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

// Private source paths are opt-in and may be a bounded glob. No customer raw
// strings, IDs or addresses are logged or included in committed fixtures.
func TestFortiWebPrivateEvidence(t *testing.T) {
	pattern := os.Getenv("UTM_FORTIWEB_EVIDENCE")
	if pattern == "" {
		t.Skip("set UTM_FORTIWEB_EVIDENCE to private sampled-hit JSON paths")
	}
	paths, e := filepath.Glob(pattern)
	if e != nil || len(paths) == 0 {
		t.Fatal("no private evidence paths")
	}
	cfg, rules, cache := fwConfig(t), fwRules(t), plugins.NewCELCache("fortiweb-private")
	populated, matched := map[string]int{}, map[string]int{}
	checkedAgent := 0
	count := 0
	for _, path := range paths {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		var hits []struct {
			Source          map[string]any `json:"_source"`
			ExpectedNetwork map[string]any `json:"expectedNetwork"`
		}
		if e = json.Unmarshal(b, &hits); e != nil {
			t.Fatal(e)
		}
		for _, hit := range hits {
			raw, ok := hit.Source["raw"].(string)
			if !ok {
				t.Fatal("raw missing from private evidence")
			}
			source, _ := hit.Source["dataSource"].(string)
			out := fwParseSource(t, cfg, raw, source, cache)
			stored, e := json.Marshal(hit.Source)
			if e != nil {
				t.Fatal(e)
			}
			for _, field := range []string{"dataSource", "origin.ip", "target.ip", "origin.port", "target.port"} {
				before, after := gjson.GetBytes(stored, field), gjson.Get(out, field)
				// Explicit private expectations document raw-authoritative fixes
				// when the indexed value itself was contaminated by payload text.
				if expected, ok := hit.ExpectedNetwork[field]; ok {
					if !after.Exists() || after.String() != fmt.Sprint(expected) {
						t.Errorf("private raw-authoritative network field differs: %s", field)
					}
				} else if before.Exists() && before.String() != after.String() {
					t.Errorf("existing private network field changed: %s", field)
				}
			}
			// Independent direct comparison for the demonstrated quoted-string
			// truncation; field contents remain private even on a failure.
			if match := regexp.MustCompile(`(?i)(?:^|\s)HTTP_agent=("(?:\\.|[^"\\])*"|[^\s]+)`).FindStringSubmatch(raw); len(match) > 1 {
				want := strings.TrimSuffix(strings.TrimPrefix(match[1], `"`), `"`)
				if gjson.Get(out, "log.httpagent").String() != want {
					t.Error("private user-agent differs from complete raw value")
				}
				checkedAgent++
			}
			for _, field := range []string{"origin.ip", "target.ip", "actionResult", "protocol", "severity", "target.url", "target.path", "log.msg", "log.subtype", "log.httpagent", "log.fileUploadViolation"} {
				if gjson.Get(out, field).Exists() {
					populated[field]++
				}
			}
			for name, rule := range rules {
				yes, e := cache.Eval(rule.Where, out)
				if e != nil {
					t.Fatalf("%s CEL: %v", name, e)
				}
				if yes {
					matched[name]++
					for _, search := range rule.Correlation {
						for _, term := range search.With {
							value := term.Value.GetStringValue()
							if strings.HasPrefix(value, "{{.") {
								field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
								if !gjson.Get(out, field).Exists() {
									t.Errorf("%s missing private history placeholder %s", name, field)
								}
							}
						}
					}
				}
			}
			count++
		}
	}
	if count == 0 {
		t.Fatal("empty private evidence")
	}
	t.Logf("private documents=%d complete user-agents compared=%d populated=%v predicates=%v", count, checkedAgent, populated, matched)
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
							value := term.Value.GetStringValue()
							if strings.HasPrefix(value, "{{.") {
								field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
								if !gjson.Get(out, field).Exists() {
									t.Errorf("%s: matched event lacks history placeholder %s", name, field)
								}
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
