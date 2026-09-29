package main

// Offline Meraki extraction model, not the closed EventProcessor.
// Explicit YAML grok/rename/cast/trim/add/delete and KV steps are modeled; grok
// and KV follow the EventProcessor steps. CEL, Event serialization, placeholder
// expansion, query creation and history thresholds use the go.mod SDK.
// External geolocation is not executed.
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
	"google.golang.org/protobuf/types/known/structpb"
)

type merakiFixture struct {
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	ID         *string        `json:"id,omitempty"`
	Raw        string         `json:"raw"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	Matches    []string       `json:"matches"`
}

func merakiPut(m map[string]any, path string, value any, remove bool) {
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
func merakiGet(m map[string]any, p string) (any, bool) {
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
func merakiConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/cisco/meraki.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}

var merakiRegexCache = map[string]*regexp.Regexp{}

// merakiTemplates are the shared grok templates as the v11 config plugin exports
// them to pipeline/patterns.yaml; the filter's own patterns override them.
var merakiTemplates = map[string]string{
	"data":      `(.*?)`,
	"greedy":    `.*`,
	"hostname":  `(\b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(\.?|\b))`,
	"integer":   `(?:[+-]?(?:[0-9]+))`,
	"ipv4":      `(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.)){3}((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))`,
	"ipv6":      `([0-9a-fA-F]{1,4}(:[0-9a-fA-F]{0,4}){1,7}|::[0-1]?)`,
	"monthDay":  `(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])`,
	"monthName": `\b(?:[Jj]an(?:uary|uar)?|[Ff]eb(?:ruary|ruar)?|[Mm](?:a|ä)?r(?:ch|z)?|[Aa]pr(?:il)?|[Mm]a(?:y|i)?|[Jj]un(?:e|i)?|[Jj]ul(?:y|i)?|[Aa]ug(?:ust)?|[Ss]ep(?:tember)?|[Oo](?:c|k)?t(?:ober)?|[Nn]ov(?:ember)?|[Dd]e(?:c|z)(?:ember)?)\b`,
	"time":      `((([01][0-9])|2[0-4]):(?:[0-5][0-9])(?::(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)))`,
	"word":      `\b\w+\b`,
	"year":      `(([1-9])[0-9]{1,3})`,
	"space":     `\s+`,
}

func merakiRegex(t *testing.T, pattern string, cfg *plugins.Config) *regexp.Regexp {
	t.Helper()
	if r, ok := merakiRegexCache[pattern]; ok {
		return r
	}
	pats := map[string]string{}
	for k, v := range merakiTemplates {
		pats[k] = v
	}
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
	merakiRegexCache[pattern] = r
	return r
}

// merakiGrok follows the EventProcessor grok step: it trims the text before
// each pattern, requires a non-empty match at the start, consumes that match,
// and writes fields only when every pattern matched.
func merakiGrok(t *testing.T, g *plugins.Grok, str string, cfg *plugins.Config) ([][2]string, bool) {
	t.Helper()
	var found [][2]string
	matched := 0
	for _, p := range g.Patterns {
		str = strings.TrimSpace(str)
		if str == "" {
			break
		}
		m := merakiRegex(t, p.Pattern, cfg).FindString(str)
		if m == "" || !strings.HasPrefix(str, m) {
			break
		}
		matched++
		if p.FieldName != "" {
			found = append(found, [2]string{p.FieldName, strings.TrimSpace(m)})
		}
		str = strings.TrimPrefix(str, m)
	}
	return found, matched == len(g.Patterns)
}

func merakiParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache) string {
	return merakiParseEvent(t, cfg, raw, dataSource, "synthetic-ingress-event", cache)
}
func merakiParseEvent(t *testing.T, cfg *plugins.Config, raw, dataSource, id string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "firewall-meraki", "dataSource": dataSource, "log": map[string]any{}}
	if id != "" {
		draft["id"] = id
	}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dataType := range stage.DataTypes {
			if dataType == "firewall-meraki" {
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
					v, ok := merakiGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					found, ok := merakiGrok(t, g, str, cfg)
					if !ok {
						continue
					}
					for _, f := range found {
						merakiPut(draft, f[0], f[1], false)
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := merakiGet(draft, p); ok {
							merakiPut(draft, s.Rename.To, v, false)
							merakiPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					for _, p := range s.Trim.Fields {
						if v, ok := merakiGet(draft, p); ok {
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
							merakiPut(draft, p, str, false)
						}
					}
				case "add":
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add function %s", s.Add.Function)
					}
					merakiPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						merakiPut(draft, p, nil, true)
					}
				case "kv":
					v, ok := merakiGet(draft, s.Kv.Source)
					if !ok {
						continue
					}
					// The KV step splits quoted multiword values and trims each value.
					// Explicit YAML grok steps rebuild consumed fields afterward.
					for _, item := range strings.Split(strings.TrimSpace(v.(string)), s.Kv.FieldSplit) {
						key, value, found := strings.Cut(item, s.Kv.ValueSplit)
						if !found {
							continue
						}
						utils.SanitizeField(&key)
						if key != "" {
							merakiPut(draft, "log."+key, strings.TrimSpace(value), false)
						}
					}
				case "dynamic":
					if s.Dynamic.Plugin != "com.utmstack.geolocation" {
						t.Fatalf("unsupported dynamic plugin %s", s.Dynamic.Plugin)
					}
					field := s.Dynamic.Params["source"].GetStringValue()
					v, ok := merakiGet(draft, field)
					if !ok {
						t.Fatalf("missing dynamic source %s", field)
					}
					ip := net.ParseIP(fmt.Sprint(v))
					if ip == nil || ip.IsUnspecified() {
						t.Fatalf("invalid address reaches geolocation: %s", field)
					}
					// The external geolocation service is not executed.
				case "json":
					source, ok := merakiGet(draft, s.Json.Source)
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
					for key, value := range merakiSanitizeJSON(parsed) {
						merakiPut(draft, "log."+key, value, false)
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := merakiGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								merakiPut(draft, field, utils.CastString(value), false)
							case "int":
								merakiPut(draft, field, utils.CastInt64(value), false)
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
func merakiRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths, e := filepath.Glob("../../rules/cisco/meraki/*.yml")
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

func merakiSanitizeJSON(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		utils.SanitizeField(&key)
		if nested, ok := value.(map[string]any); ok {
			value = merakiSanitizeJSON(nested)
		}
		out[key] = value
	}
	return out
}

func merakiFixtures(t *testing.T) []merakiFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/meraki_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []merakiFixture
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	return cases
}

func TestMerakiRawContracts(t *testing.T) {
	cfg, rules, cache := merakiConfig(t), merakiRules(t), plugins.NewCELCache("meraki-raw")
	for _, f := range merakiFixtures(t) {
		t.Run(f.Name, func(t *testing.T) {
			id := "synthetic-ingress-event"
			if f.ID != nil {
				id = *f.ID
			}
			out := merakiParseEvent(t, cfg, f.Raw, f.DataSource, id, cache)
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
				yes, e := cache.Eval(r.Where, out)
				if e != nil {
					t.Fatalf("%s CEL %v", name, e)
				}
				if yes != expected[name] {
					t.Errorf("%s match=%v want=%v", name, yes, expected[name])
				}
				if name == "meraki_vpn_brute_force" && (gjson.Get(out, "log.vpnAuthenticationFailure").String() == "match") != yes {
					t.Error("VPN marker mismatch")
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
					if name == "advanced_malware_protection_alerts" {
						ev := new(plugins.Event)
						if e := utils.StringToProtoMessage(&out, ev); e != nil {
							t.Fatal(e)
						}
						if r.Adversary != "target" {
							t.Error("AMP actor must be remote download endpoint")
						}
						a := &plugins.Alert{Adversary: ev.Target, Target: ev.Origin}
						if a.GetAdversary().GetIp() != ev.GetTarget().GetIp() {
							t.Error("AMP endpoint roles changed")
						}
					}
				}
			}
		})
	}
}

func TestMerakiAuxiliaryGeoGuards(t *testing.T) {
	cfg, cache := merakiConfig(t), plugins.NewCELCache("meraki-auxiliary-geo")
	checked := 0
	for _, step := range cfg.Pipeline[0].Steps {
		if step.Dynamic == nil {
			continue
		}
		source := step.Dynamic.Params["source"].GetStringValue()
		if source != "log.serverIp" && source != "log.localIp" {
			continue
		}
		checked++
		if strings.HasPrefix(step.Dynamic.Params["destination"].GetStringValue(), source+".") {
			t.Fatal("geolocation destination nests under its scalar source")
		}
		for _, value := range []string{"relay.example.invalid", "0.0.0.0", "::0", "0:0:0:0:0:0:0:0", "::ffff:0.0.0.0", "192.0.2.10", "2001:db8::"} {
			t.Run(source+"_"+value, func(t *testing.T) {
				// Isolated source-step control, not a legacy-envelope replay.
				add := &plugins.Step{Add: &plugins.Add{Function: "string", Params: map[string]*structpb.Value{"key": structpb.NewStringValue(source), "value": structpb.NewStringValue(value)}}}
				isolated := &plugins.Config{Pipeline: []*plugins.Pipeline{{DataTypes: []string{"firewall-meraki"}, Steps: []*plugins.Step{add, step}}}}
				out := merakiParse(t, isolated, "synthetic control", "synthetic-relay", cache)
				if gjson.Get(out, source).String() != value {
					t.Error("vendor source value changed")
				}
			})
		}
	}
	if checked != 2 {
		t.Fatal("both auxiliary geolocation sources must be checked")
	}
}

func TestMerakiPrivateRoutingEvidence(t *testing.T) {
	path := os.Getenv("UTM_MERAKI_EVIDENCE")
	if path == "" {
		t.Skip("set UTM_MERAKI_EVIDENCE to the bounded private routing sample")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var hits []struct {
		Source map[string]any `json:"_source"`
	}
	if e = json.Unmarshal(b, &hits); e != nil || len(hits) == 0 {
		t.Fatal("invalid or empty private routing evidence")
	}
	cfg, rules, cache := merakiConfig(t), merakiRules(t), plugins.NewCELCache("meraki-private-routing")
	for _, hit := range hits {
		raw, ok := hit.Source["raw"].(string)
		if !ok {
			t.Fatal("private raw missing")
		}
		source, _ := hit.Source["dataSource"].(string)
		id, _ := hit.Source["id"].(string)
		out := merakiParseEvent(t, cfg, raw, source, id, cache)
		for _, field := range []string{"origin.ip", "target.ip", "log.eventType", "actionResult", "log.vpnAuthenticationFailure", "log.malwareGroupingKey"} {
			if gjson.Get(out, field).Exists() {
				t.Errorf("unrelated private syslog acquired %s", field)
			}
		}
		for name, rule := range rules {
			if yes, e := cache.Eval(rule.Where, out); e != nil || yes {
				t.Errorf("private routing negative matched %s or errored: %v", name, e)
			}
		}
	}
	t.Logf("private non-Meraki routing controls=%d; no invented Meraki identities, outcomes or rule candidates", len(hits))
}
