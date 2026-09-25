package main

// Offline ESET extraction model, not the closed EventProcessor.
// Explicit YAML grok/JSON/rename/cast/add/reformat/delete are modeled. Grok
// patterns are consumed in order, with whitespace trimmed before each non-empty
// prefix match, as the EventProcessor grok plugin does. JSON key sanitization
// uses SDK utilities; malformed JSON is an explicit model error.
// CEL and Event serialization use SDK v1.1.31. Separate history tests exercise
// actual SDK query behavior. External geolocation and the closed executor are not run.
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

type esetFixture struct {
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	ID         *string        `json:"id,omitempty"`
	Raw        string         `json:"raw"`
	Ingress    map[string]any `json:"ingress,omitempty"`
	ParseError bool           `json:"parseError,omitempty"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	Matches    []string       `json:"matches"`
}

func esetPut(m map[string]any, path string, value any, remove bool) {
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
func esetGet(m map[string]any, p string) (any, bool) {
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
func esetConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/antivirus/esmc-eset.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}
func esetRegex(t *testing.T, pattern string, cfg *plugins.Config) *regexp.Regexp {
	t.Helper()
	// ESET uses explicit YAML expressions, with no approximated built-in grok aliases.
	pats := map[string]string{}
	for k, v := range cfg.Patterns {
		pats[k] = v
	}
	tmpl, e := template.New("grok").Option("missingkey=error").Parse(pattern)
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
func esetParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache) string {
	return esetParseEvent(t, cfg, raw, dataSource, "synthetic-ingress-event", cache)
}
func esetParseEvent(t *testing.T, cfg *plugins.Config, raw, dataSource, id string, cache *plugins.CELCache) string {
	return esetParseInput(t, cfg, raw, dataSource, id, nil, cache)
}
func esetParseInput(t *testing.T, cfg *plugins.Config, raw, dataSource, id string, ingress map[string]any, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "antivirus-esmc-eset", "dataSource": dataSource, "log": map[string]any{}}
	for key, value := range ingress {
		draft[key] = value
	}
	if id != "" {
		draft["id"] = id
	}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dataType := range stage.DataTypes {
			if dataType == "antivirus-esmc-eset" {
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
					v, ok := esetGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					found, matched := map[string]string{}, 0
					for _, p := range g.Patterns {
						str = strings.TrimSpace(str)
						if str == "" {
							break
						}
						m := esetRegex(t, p.Pattern, cfg).FindString(str)
						if m == "" || !strings.HasPrefix(str, m) {
							break
						}
						matched++
						if p.FieldName != "" {
							found[p.FieldName] = strings.TrimSpace(m)
						}
						str = strings.TrimPrefix(str, m)
					}
					if matched != len(g.Patterns) {
						continue
					}
					for field, value := range found {
						esetPut(draft, field, value, false)
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := esetGet(draft, p); ok {
							esetPut(draft, s.Rename.To, v, false)
							esetPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					for _, p := range s.Trim.Fields {
						if v, ok := esetGet(draft, p); ok {
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
							esetPut(draft, p, str, false)
						}
					}
				case "add":
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add function %s", s.Add.Function)
					}
					esetPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						esetPut(draft, p, nil, true)
					}
				case "dynamic":
					if s.Dynamic.Plugin != "com.utmstack.geolocation" {
						t.Fatalf("unsupported dynamic plugin %s", s.Dynamic.Plugin)
					}
					field := s.Dynamic.Params["source"].GetStringValue()
					v, ok := esetGet(draft, field)
					if !ok {
						t.Fatalf("missing dynamic source %s", field)
					}
					ip := net.ParseIP(fmt.Sprint(v))
					if ip == nil || ip.IsUnspecified() {
						t.Fatalf("invalid address reaches geolocation: %s", field)
					}
					// The external geolocation service is not executed.
				case "json":
					source, ok := esetGet(draft, s.Json.Source)
					if !ok {
						continue
					}
					str, ok := source.(string)
					if !ok {
						t.Fatalf("JSON source is not a string")
					}
					var parsed map[string]any
					if e := json.Unmarshal([]byte(str), &parsed); e != nil {
						return "MODEL_JSON_ERROR"
					}
					for key, value := range esetSanitizeJSON(parsed) {
						esetPut(draft, "log."+key, value, false)
					}
				case "reformat":
					if s.Reformat.Function != "time" {
						t.Fatalf("unsupported reformat %s", s.Reformat.Function)
					}
					for _, field := range s.Reformat.Fields {
						if value, ok := esetGet(draft, field); ok {
							str, ok := value.(string)
							if !ok {
								t.Fatalf("non-string timestamp %s", field)
							}
							parsed, err := time.Parse(s.Reformat.FromFormat, str)
							if err != nil {
								t.Fatalf("invalid guarded timestamp %s: %v", field, err)
							}
							esetPut(draft, field, parsed.Format(s.Reformat.ToFormat), false)
						}
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := esetGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								esetPut(draft, field, utils.CastString(value), false)
							case "int":
								esetPut(draft, field, utils.CastInt64(value), false)
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
func esetRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths, e := filepath.Glob("../../rules/antivirus/esmc-eset/*.yml")
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

func esetSanitizeJSON(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		utils.SanitizeField(&key)
		if nested, ok := value.(map[string]any); ok {
			value = esetSanitizeJSON(nested)
		}
		out[key] = value
	}
	return out
}

func esetFixtures(t *testing.T) []esetFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/eset_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []esetFixture
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	return cases
}

func TestESETRawContracts(t *testing.T) {
	cfg, rules, cache := esetConfig(t), esetRules(t), plugins.NewCELCache("eset-raw")
	if len(rules) != 14 {
		t.Fatalf("expected14ESETconsumers, got%d", len(rules))
	}
	positives, negatives := map[string]int{}, map[string]int{}
	for _, f := range esetFixtures(t) {
		t.Run(f.Name, func(t *testing.T) {
			id := "synthetic-ingress-event"
			if f.ID != nil {
				id = *f.ID
			}
			out := esetParseInput(t, cfg, f.Raw, f.DataSource, id, f.Ingress, cache)
			if f.ParseError {
				if out != "MODEL_JSON_ERROR" {
					t.Error("expected malformed JSON model error")
				}
				return
			}
			if out == "MODEL_JSON_ERROR" {
				t.Fatal("unexpected malformed JSON model error")
			}
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
			expected := map[string]bool{}
			for _, n := range f.Matches {
				expected[n] = true
			}
			for name, r := range rules {
				if f.Matches == nil {
					t.Fatal("fixture requires explicit predicate expectations")
				}
				yes, e := cache.Eval(r.Where, out)
				if e != nil {
					t.Fatalf("%s CEL %v", name, e)
				}
				if yes {
					positives[name]++
				} else {
					negatives[name]++
				}
				if yes != expected[name] {
					t.Errorf("%s match=%v want=%v", name, yes, expected[name])
				}
				marker := map[string]string{"advanced_heuristic_detection_triggers": "heuristicRemediation", "eset_console_abuse": "consoleAuthenticationFailure", "eset_quarantine_failures": "quarantineFailure"}[name]
				if marker != "" && (gjson.Get(out, "log.correlationCandidate."+marker).String() == "match") != yes {
					t.Errorf("%s candidate mismatch", name)
				}

				if yes {
					for _, search := range r.Correlation {
						for _, term := range search.With {
							v := term.Value.GetStringValue()
							if strings.HasPrefix(v, "{{.") {
								p := strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")
								if !gjson.Get(out, p).Exists() {
									t.Errorf("%s missing history %s", name, p)
								}
							}
						}
					}

				}
			}
		})
	}
	for name := range rules {
		if positives[name] == 0 || negatives[name] == 0 {
			t.Errorf("%s missing positive/negative raw predicate coverage", name)
		}
	}
}
