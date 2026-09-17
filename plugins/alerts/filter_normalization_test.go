package main

// This is a normalization-stage model, NOT the closed EventProcessor parser.
// It uses the real SDK for YAML decoding, CEL evaluation and final Event conversion.
// Complex grok, CSV, KV, JSON extraction and dynamic plugins are deliberately skipped.
import (
	"encoding/json"
	"fmt"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type Fixture struct {
	Name     string          `json:"name"`
	Filter   string          `json:"filter"`
	Input    map[string]any  `json:"input"`
	Expected map[string]any  `json:"expected"`
	Absent   []string        `json:"absent"`
	Rules    map[string]bool `json:"rules"`
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
	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(b, cfg); err != nil {
		return "", nil, err
	}
	data, _ := json.Marshal(f.Input)
	draft := map[string]any{}
	json.Unmarshal(data, &draft)
	issues := []string{}
	for _, stage := range cfg.Pipeline {
		for i, step := range stage.Steps {
			sb, _ := protojson.Marshal(step)
			obj := map[string]map[string]any{}
			json.Unmarshal(sb, &obj)
			for kind, body := range obj {
				if kind == "dynamic" || kind == "json" || kind == "kv" || kind == "csv" || kind == "xml" || kind == "reformat" {
					continue
				}
				if kind == "grok" && !(step.Grok != nil && len(step.Grok.Patterns) == 1 && step.Grok.Patterns[0].Pattern == "{{.greedy}}") {
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
					put(draft, p["key"].GetStringValue(), p["value"].AsInterface(), false)
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
								}
							}
							put(draft, p, str, false)
						}
					}
				case "grok":
					g := step.Grok
					if v, ok := valueAt(draft, g.Source); ok {
						if str, ok := v.(string); ok {
							put(draft, g.Patterns[0].FieldName, str, false)
						}
					}
				case "drop":
					return "", issues, fmt.Errorf("fixture dropped")
				}
			}
		}
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

// TestFilterNormalization supplies synthetic extraction results to the documented
// normalization steps. It is not a raw-log or dynamic-plugin integration test.
func TestFilterNormalization(t *testing.T) {
	b, err := os.ReadFile("testdata/filter-normalization.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []Fixture
	if err = json.Unmarshal(b, &fixtures); err != nil {
		t.Fatal(err)
	}
	cache := plugins.NewCELCache("filter-normalization-test")
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
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
			}
		})
	}
}
