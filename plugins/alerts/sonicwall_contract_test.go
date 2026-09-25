package main

// Offline SonicWall extraction model, not the closed EventProcessor.
// Explicit YAML grok/kv/rename/trim/cast/add/delete steps are modeled; grok
// follows the EventProcessor step (a pattern list compiles to one regex and
// writes its named fields only when every group matches a non-empty value).
// CEL, the KV split and cast use the go.mod SDK. External geolocation is not
// executed.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

const sonicDataType = "firewall-sonicwall"

type sonicFixture struct {
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	Raw        string         `json:"raw"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	MatchRule  string         `json:"matchRule"` // rule whose where: must evaluate true
}

func sonicPut(m map[string]any, path string, value any, remove bool) {
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

func sonicGet(m map[string]any, p string) (any, bool) {
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

func sonicConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/sonicwall/sonic_wall.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}

func sonicRegex(t *testing.T, g *plugins.Grok, cfg *plugins.Config) *regexp.Regexp {
	t.Helper()
	var pattern strings.Builder
	for i, p := range g.Patterns {
		if p.FieldName != "" {
			fmt.Fprintf(&pattern, "(?P<f%d>%s)", i, p.Pattern)
		} else {
			pattern.WriteString("(?:" + p.Pattern + ")")
		}
	}
	tmpl, e := template.New("grok").Option("missingkey=error").Parse(pattern.String())
	if e != nil {
		t.Fatal(e)
	}
	// grok named captures reference the filter's local/built-in patterns.
	pats := map[string]string{
		"greedy":  ".*",
		"data":    ".*?",
		"word":    `[A-Za-z0-9_-]+`,
		"space":   `\s+`,
		"integer": `[0-9]+`,
		"ipv4":    `(?:[0-9]{1,3}\.){3}[0-9]{1,3}`,
	}
	for k, v := range cfg.Patterns {
		pats[k] = v
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

// swGrok mirrors the EventProcessor: the pattern list is compiled to one
// regex; every named field and every anonymous group must match a non-empty
// value for the step to write anything.
func sonicGrokMatches(r *regexp.Regexp, s string, g *plugins.Grok) (map[int]string, bool) {
	m := r.FindStringSubmatch(s)
	if m == nil {
		return nil, false
	}
	out := map[int]string{}
	for i, p := range g.Patterns {
		if p.FieldName == "" {
			continue
		}
		gi := r.SubexpIndex(fmt.Sprintf("f%d", i))
		if gi < 0 {
			return nil, false
		}
		if m[gi] == "" {
			return nil, false
		}
		out[i] = m[gi]
	}
	return out, true
}

func sonicParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": sonicDataType, "dataSource": dataSource, "log": map[string]any{}}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dt := range stage.DataTypes {
			if dt == sonicDataType {
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
					v, ok := sonicGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					r := sonicRegex(t, g, cfg)
					vals, ok := sonicGrokMatches(r, str, g)
					if !ok {
						continue
					}
					for i, p := range g.Patterns {
						if p.FieldName != "" {
							sonicPut(draft, p.FieldName, vals[i], false)
						}
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := sonicGet(draft, p); ok {
							sonicPut(draft, s.Rename.To, v, false)
							sonicPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					for _, p := range s.Trim.Fields {
						if v, ok := sonicGet(draft, p); ok {
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
							sonicPut(draft, p, str, false)
						}
					}
				case "add":
					sonicPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						sonicPut(draft, p, nil, true)
					}
				case "kv":
					v, ok := sonicGet(draft, s.Kv.Source)
					if !ok {
						continue
					}
					for _, item := range strings.Split(v.(string), s.Kv.FieldSplit) {
						pair := strings.SplitN(item, s.Kv.ValueSplit, 2)
						if len(pair) != 2 {
							continue
						}
						key := pair[0]
						utils.SanitizeField(&key)
						if key != "" {
							sonicPut(draft, "log."+key, pair[1], false)
						}
					}
				case "dynamic":
					// external geolocation service is not executed
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := sonicGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								sonicPut(draft, field, utils.CastString(value), false)
							case "int":
								sonicPut(draft, field, utils.CastInt64(value), false)
							case "float":
								sonicPut(draft, field, utils.CastFloat64(value), false)
							default:
								t.Fatalf("unsupported cast %s", s.Cast.To)
							}
						}
					}
				default:
					t.Fatalf("unmodeled step kind %q", kind)
				}
			}
		}
	}
	out, e := json.Marshal(draft)
	if e != nil {
		t.Fatal(e)
	}
	return string(out)
}

func TestSonicWallRawContracts(t *testing.T) {
	data, err := os.ReadFile("testdata/sonicwall_admin_auth_raw.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []sonicFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	cfg := sonicConfig(t)
	cache := plugins.NewCELCache("sonicwall-contract")

	// Load the shipped admin-auth rule so we assert the real where: clause.
	rule := loadSonicRule(t, "../../rules/sonicwall/sonicwall_firewall/sonicwall_admin_auth_failures.yml")

	for _, fx := range fixtures {
		t.Run(fx.Name, func(t *testing.T) {
			normalized := sonicParse(t, cfg, fx.Raw, fx.DataSource, cache)
			for path, want := range fx.Expected {
				got, ok := sonicGetMust(normalized, path)
				if !ok {
					t.Fatalf("expected %s to be present", path)
				}
				if fmt.Sprint(got) != fmt.Sprint(want) {
					t.Fatalf("%s = %v, want %v", path, got, want)
				}
			}
			for _, path := range fx.Absent {
				if got, ok := sonicGetMust(normalized, path); ok {
					t.Fatalf("%s should be absent, got %v", path, got)
				}
			}
			if fx.MatchRule != "" {
				if fx.MatchRule != "admin_auth_failures" {
					t.Fatalf("unknown rule %q", fx.MatchRule)
				}
				m, e := cache.Eval(rule.Where, normalized)
				if e != nil {
					t.Fatalf("rule CEL: %v", e)
				}
				if !m {
					t.Fatalf("admin_auth_failures where: did not match: %s", normalized)
				}
			}
		})
	}
}

func sonicGetMust(normalized, path string) (any, bool) {
	var draft map[string]any
	if e := json.Unmarshal([]byte(normalized), &draft); e != nil {
		panic(e)
	}
	return sonicGet(draft, path)
}

func loadSonicRule(t *testing.T, path string) *plugins.Rule {
	t.Helper()
	b, e := utils.ReadPbYaml(path)
	if e != nil {
		t.Fatal(e)
	}
	r := new(plugins.Rule)
	if e = protojson.Unmarshal(b, r); e != nil {
		t.Fatal(e)
	}
	r.Normalize()
	return r
}
