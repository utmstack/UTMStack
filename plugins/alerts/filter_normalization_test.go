package main

// This is a normalization-stage model, NOT the closed EventProcessor parser.
// It uses the real SDK for YAML decoding, CEL evaluation and final Event conversion.
// Synthetic-input fixtures skip complex grok, CSV, KV, JSON and dynamic plugins.
// Opt-in raw fixtures model JSON decoding/key sanitization and reject unsupported
// executed steps. Only one-field {{.greedy}} and (.*) copy captures are modeled.
import (
	"encoding/json"
	"fmt"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type Fixture struct {
	Name                     string          `json:"name"`
	Filter                   string          `json:"filter"`
	Input                    map[string]any  `json:"input"`
	Raw                      *string         `json:"raw"`
	DataType                 string          `json:"dataType"`
	DataSource               string          `json:"dataSource"`
	Expected                 map[string]any  `json:"expected"`
	Absent                   []string        `json:"absent"`
	Rules                    map[string]bool `json:"rules"`
	CheckHistoryPlaceholders bool            `json:"checkHistoryPlaceholders"`
}

func valueAt(m map[string]any, p string) (any, bool) {
	var v any = m
	for _, k := range strings.Split(p, ".") {
		switch n := v.(type) {
		case map[string]any:
			var ok bool
			v, ok = n[k]
			if !ok {
				return nil, false
			}
		case []any:
			i, e := strconv.Atoi(k)
			if e != nil || i < 0 || i >= len(n) {
				return nil, false
			}
			v = n[i]
		default:
			return nil, false
		}
	}
	return v, true
}
func put(m map[string]any, p string, v any, remove bool) {
	parts := strings.Split(p, ".")
	for _, k := range parts[:len(parts)-1] {
		next, ok := m[k].(map[string]any)
		if !ok {
			if remove {
				return
			}
			next = map[string]any{}
			m[k] = next
		}
		m = next
	}
	if remove {
		delete(m, parts[len(parts)-1])
	} else {
		m[parts[len(parts)-1]] = v
	}
}
func normalize(root string, f Fixture, cache *plugins.CELCache) (string, []string, error) {
	b, err := utils.ReadPbYaml(filepath.Join(root, "filters", f.Filter))
	if err != nil {
		return "", nil, err
	}
	cfg := new(plugins.Config)
	if err = protojson.Unmarshal(b, cfg); err != nil {
		return "", nil, err
	}
	if (f.Raw == nil) == (f.Input == nil) {
		return "", nil, fmt.Errorf("fixture requires exactly one of input or raw")
	}
	data, _ := json.Marshal(f.Input)
	draft := map[string]any{}
	json.Unmarshal(data, &draft)
	if f.Raw != nil {
		if f.DataType == "" || f.DataSource == "" {
			return "", nil, fmt.Errorf("raw fixture requires ingress dataType and dataSource")
		}
		draft = map[string]any{"raw": *f.Raw, "dataType": f.DataType, "dataSource": f.DataSource}
	}
	issues := []string{}
	matchedStage := false
	for _, stage := range cfg.Pipeline {
		if f.Raw != nil && len(stage.DataTypes) != 0 {
			matched := false
			for _, dataType := range stage.DataTypes {
				matched = matched || dataType == f.DataType
			}
			if !matched {
				continue
			}
		}
		matchedStage = true
		for i, step := range stage.Steps {
			sb, _ := protojson.Marshal(step)
			obj := map[string]map[string]any{}
			json.Unmarshal(sb, &obj)
			for kind, body := range obj {
				if f.Raw == nil && (kind == "dynamic" || kind == "json" || kind == "kv" || kind == "csv" || kind == "xml" || kind == "reformat") {
					continue
				}
				simpleGrok := step.Grok != nil && len(step.Grok.Patterns) == 1 && (step.Grok.Patterns[0].Pattern == "{{.greedy}}" || step.Grok.Patterns[0].Pattern == "(.*)")
				if f.Raw == nil && kind == "grok" && !simpleGrok {
					continue
				}
				if w, ok := body["where"].(string); ok && w != "" {
					snapshot, _ := json.Marshal(draft)
					match, e := cache.Eval(w, string(snapshot))
					if e != nil {
						issues = append(issues, fmt.Sprintf("step %d %s: %v", i, kind, e))
						continue
					}
					if !match {
						continue
					}
				}
				switch kind {
				case "json":
					source := step.Json.Source
					if source == "" {
						source = "raw"
					}
					value, ok := valueAt(draft, source)
					if !ok {
						continue
					}
					str, ok := value.(string)
					if !ok {
						return "", issues, fmt.Errorf("JSON source %s is not a string", source)
					}
					parsed, err := fixtureJSON(str)
					if err != nil {
						return "", issues, err
					}
					for key, value := range parsed {
						put(draft, "log."+key, value, false)
					}
				case "rename":
					for _, src := range step.Rename.From {
						if v, ok := valueAt(draft, src); ok {
							put(draft, step.Rename.To, v, false)
							put(draft, src, nil, true)
							break
						}
					}
				case "add":
					p := step.Add.Params
					if f.Raw != nil {
						if step.Add.Function != "string" {
							return "", issues, fmt.Errorf("raw model does not support add function %q", step.Add.Function)
						}
						if p["key"].GetStringValue() == "" || p["value"] == nil {
							return "", issues, fmt.Errorf("raw model add requires key and value")
						}
						put(draft, p["key"].GetStringValue(), utils.CastString(p["value"].AsInterface()), false)
					} else {
						put(draft, p["key"].GetStringValue(), p["value"].AsInterface(), false)
					}
				case "delete":
					for _, p := range step.Delete.Fields {
						put(draft, p, nil, true)
					}
				case "cast":
					for _, p := range step.Cast.Fields {
						if v, ok := valueAt(draft, p); ok {
							switch step.Cast.To {
							case "int":
								put(draft, p, utils.CastInt64(v), false)
							case "float":
								put(draft, p, utils.CastFloat64(v), false)
							case "string":
								put(draft, p, utils.CastString(v), false)
							default:
								if f.Raw != nil {
									return "", issues, fmt.Errorf("raw model does not support cast %s", step.Cast.To)
								}
							}
						}
					}
				case "trim":
					for _, p := range step.Trim.Fields {
						if v, ok := valueAt(draft, p); ok {
							str, isString := v.(string)
							if !isString {
								continue
							}
							switch step.Trim.Function {
							case "prefix":
								str = strings.TrimPrefix(str, step.Trim.Substring)
							case "suffix":
								str = strings.TrimSuffix(str, step.Trim.Substring)
							case "substring":
								str = strings.ReplaceAll(str, step.Trim.Substring, "")
							case "regex":
								r, e := regexp.Compile(step.Trim.Substring)
								if e == nil {
									str = r.ReplaceAllString(str, "")
								} else if f.Raw != nil {
									return "", issues, e
								}
							default:
								if f.Raw != nil {
									return "", issues, fmt.Errorf("raw model does not support trim %s", step.Trim.Function)
								}
							}
							put(draft, p, str, false)
						}
					}
				case "grok":
					if !simpleGrok {
						return "", issues, fmt.Errorf("raw model does not support complex grok")
					}
					g := step.Grok
					source := g.Source
					if f.Raw != nil && source == "" {
						source = "raw"
					}
					if v, ok := valueAt(draft, source); ok {
						if str, ok := v.(string); ok {
							if f.Raw != nil && strings.ContainsAny(str, "\r\n") {
								return "", issues, fmt.Errorf("raw model does not support multiline copy grok")
							}
							if f.Raw != nil && g.Patterns[0].Pattern == "{{.greedy}}" && cfg.Patterns["greedy"] != "" && cfg.Patterns["greedy"] != ".*" {
								return "", issues, fmt.Errorf("raw model does not support custom greedy pattern")
							}
							put(draft, g.Patterns[0].FieldName, str, false)
						}
					}
				case "drop":
					return "", issues, fmt.Errorf("fixture dropped")
				default:
					if f.Raw != nil {
						return "", issues, fmt.Errorf("raw model does not support step %s", kind)
					}
				}
			}
		}
	}
	if f.Raw != nil && !matchedStage {
		return "", issues, fmt.Errorf("no pipeline stage matches raw fixture dataType %s", f.DataType)
	}
	data, _ = json.Marshal(draft)
	str := string(data)
	event := new(plugins.Event)
	if err = utils.StringToProtoMessage(&str, event); err != nil {
		return str, issues, err
	}
	out, err := utils.ProtoMessageToString(event)
	if err != nil {
		return "", issues, err
	}
	return *out, issues, nil
}

// TestFilterNormalization accepts synthetic extraction results or opt-in raw
// JSON model fixtures. Neither mode executes the closed EventProcessor.
func TestFilterNormalization(t *testing.T) {
	var fixtures []Fixture
	for _, manifest := range loadFilterContracts(t) {
		fixtures = append(fixtures, manifest.Fixtures...)
	}
	cache := plugins.NewCELCache("filter-normalization-test")
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			if f.Raw == nil {
				t.Log("synthetic normalization input: raw extraction is not tested")
			} else {
				t.Log("raw JSON model: closed EventProcessor execution is not tested")
			}
			if len(f.Rules) == 0 {
				t.Log("rule validation: not requested by this fixture")
			}
			out, issues, err := normalize("../..", f, cache)
			if err != nil {
				t.Fatal(err)
			}
			for _, issue := range issues {
				t.Error(issue)
			}
			for p, want := range f.Expected {
				got := gjson.Get(out, p)
				if !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("%s: got %v, want %v", p, got.Value(), want)
				}
			}
			for _, p := range f.Absent {
				if gjson.Get(out, p).Exists() {
					t.Errorf("unexpected %s in %s", p, out)
				}
			}
			for path, want := range f.Rules {
				b, err := utils.ReadPbYaml(filepath.Join("../..", path))
				if err != nil {
					t.Fatal(err)
				}
				got, err := cache.Eval(gjson.GetBytes(b, "where").String(), out)
				if err != nil || got != want {
					t.Errorf("%s: got %v (%v), want %v", path, got, err, want)
				}
				if got && want && f.CheckHistoryPlaceholders {
					rule := new(plugins.Rule)
					if err := protojson.Unmarshal(b, rule); err != nil {
						t.Fatal(err)
					}
					rule.Normalize()
					for _, issue := range fixtureHistoryPlaceholders(rule.Correlation, out) {
						t.Errorf("%s: %s", path, issue)
					}
					t.Log("all-branch history placeholder preflight only: queries, counts and alerts are not executed")
				}
			}
		})
	}
}
