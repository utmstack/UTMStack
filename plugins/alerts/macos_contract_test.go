package main

// Offline macOS parser and correlation contracts. JSON key sanitization,
// rename/cast/grok/add/trim/delete are modeled from the documented pipeline;
// the closed EventProcessor is not executed. CEL, final Event conversion,
// placeholder resolution, query construction and threshold decisions use go-sdk.
// History searches reach only a local mock, never a customer instance.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"

	"text/template"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

type macFixture struct {
	Drop       bool           `json:"drop"`
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	Raw        string         `json:"raw"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	Matches    []string       `json:"matches"`
}

func macPut(m map[string]any, path string, value any, remove bool) {
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
func macGet(m map[string]any, p string) (any, bool) {
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
func macConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/macos/macos.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}
func macRegex(t *testing.T, g *plugins.Grok, cfg *plugins.Config) *regexp.Regexp {
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
func macParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "macos", "dataSource": dataSource, "log": map[string]any{}}
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
					v, ok := macGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					r := macRegex(t, g, cfg)
					m := r.FindStringSubmatch(str)
					if m == nil {
						continue
					}
					for i, p := range g.Patterns {
						if p.FieldName != "" {
							macPut(draft, p.FieldName, m[r.SubexpIndex(fmt.Sprintf("f%d", i))], false)
						}
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := macGet(draft, p); ok {
							macPut(draft, s.Rename.To, v, false)
							macPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					for _, p := range s.Trim.Fields {
						if v, ok := macGet(draft, p); ok {
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
							macPut(draft, p, str, false)
						}
					}
				case "add":
					macPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						macPut(draft, p, nil, true)
					}
				case "json":
					source, ok := macGet(draft, s.Json.Source)
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
					for key, value := range macSanitizeJSON(parsed) {
						macPut(draft, "log."+key, value, false)
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := macGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								macPut(draft, field, utils.CastString(value), false)
							case "int":
								macPut(draft, field, utils.CastInt64(value), false)
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
func macRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths, e := filepath.Glob("../../rules/macos/*.yml")
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

func macSanitizeJSON(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		utils.SanitizeField(&key)
		if nested, ok := value.(map[string]any); ok {
			value = macSanitizeJSON(nested)
		}
		out[key] = value
	}
	return out
}

func macFixtures(t *testing.T) []macFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/macos_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixtures []macFixture
	if e = json.Unmarshal(b, &fixtures); e != nil {
		t.Fatal(e)
	}
	return fixtures
}

func macCombined(t *testing.T, first string, second string) *plugins.Config {
	t.Helper()
	out := new(plugins.Config)
	for _, name := range []string{first, second} {
		if name == "" {
			continue
		}
		b, e := utils.ReadPbYaml("../../filters/macos/" + name + ".yml")
		if e != nil {
			t.Fatal(e)
		}
		cfg := new(plugins.Config)
		if e = protojson.Unmarshal(b, cfg); e != nil {
			t.Fatal(e)
		}
		out.Pipeline = append(out.Pipeline, cfg.Pipeline...)
	}
	return out
}
func macMarker(name string) string {
	if name == "endpoint_security_bypass" {
		return "log.correlationCandidate.endpointSecurity"
	}
	if name == "macos_ransomware_indicators" {
		return "log.correlationCandidate.ransomware"
	}
	return ""
}
func TestMacOSRawContracts(t *testing.T) {
	rules, cache := macRules(t), plugins.NewCELCache("macos-raw")
	configs := map[string]*plugins.Config{"main": macConfig(t), "wrapper-first": macCombined(t, "macos-syslog", "macos"), "wrapper-last": macCombined(t, "macos", "macos-syslog")}
	for order, cfg := range configs {
		for _, f := range macFixtures(t) {
			t.Run(order+"/"+f.Name, func(t *testing.T) {
				out := macParse(t, cfg, f.Raw, f.DataSource, cache)
				if f.Drop {
					if out != "" {
						t.Fatal("noise event was not dropped")
					}
					return
				}
				if out == "" {
					t.Fatal("event unexpectedly dropped")
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
				for _, name := range f.Matches {
					expected[name] = true
				}
				for name, r := range rules {
					yes, e := cache.Eval(r.Where, out)
					if e != nil {
						t.Fatalf("%s CEL error: %v", name, e)
					}
					if expected[name] && !yes {
						t.Errorf("%s did not match", name)
					}
					if expected[name] && yes {
						macCheckGrouping(t, r, out)
					}
					if len(f.Matches) == 0 && yes {
						t.Errorf("negative fixture matched %s", name)
					}
					if marker := macMarker(name); marker != "" {
						present := gjson.Get(out, marker).String() == "match"
						if present != yes {
							t.Errorf("%s marker/predicate mismatch", name)
						}
					}
					if yes {
						for _, search := range r.Correlation {
							for _, term := range search.With {
								v := term.Value.GetStringValue()
								if strings.HasPrefix(v, "{{.") {
									field := strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")
									if !gjson.Get(out, field).Exists() {
										t.Errorf("%s unresolved %s", name, field)
									}
								}
							}
						}
					}
				}
			})
		}
	}
}

// Verify the source contract for grouping using the documented origin-to-
// adversary role. lastEvent is an indexed alias of the final wire event; its
// runtime resolution is covered by the shared alert-grouping tests.
func macCheckGrouping(t *testing.T, rule *plugins.Rule, eventJSON string) {
	t.Helper()
	if rule.Adversary != "origin" {
		t.Fatalf("unexpected macOS actor role %s", rule.Adversary)
	}
	event := new(plugins.Event)
	if e := utils.StringToProtoMessage(&eventJSON, event); e != nil {
		t.Fatal(e)
	}
	alert := &plugins.Alert{Adversary: event.Origin, Events: []*plugins.Event{event}}
	wire, e := utils.ProtoMessageToString(alert)
	if e != nil {
		t.Fatal(e)
	}
	hasIdentity := false
	for _, field := range rule.GroupBy {
		path := strings.Replace(field, "lastEvent.", "events.0.", 1)
		value := gjson.Get(*wire, path)
		if field == "adversary.host" || field == "lastEvent.dataSource" {
			if value.String() != event.DataSource || value.String() == "" || value.String() == "unknown" {
				t.Errorf("%s has no usable grouping identity %s", rule.Name, field)
			}
			hasIdentity = true
		}
		if field == "adversary.process" && event.GetOrigin().GetProcess() != "" && value.String() != event.GetOrigin().GetProcess() {
			t.Errorf("%s process grouping lost its actor", rule.Name)
		}
		if field == "lastEvent.log.message" && value.String() != gjson.Get(eventJSON, "log.message").String() {
			t.Errorf("%s message grouping lost its message", rule.Name)
		}
	}
	if !hasIdentity {
		t.Errorf("%s grouping lacks a source identity", rule.Name)
	}
}

func TestMacOSSDKHistory(t *testing.T) {
	// The SDK owns a process-wide OpenSearch singleton. Isolate this local mock
	// so other technology tests can initialize their own clients in this suite.
	if os.Getenv("UTM_MACOS_HISTORY_CHILD") != "1" {
		command := exec.Command(os.Args[0], "-test.run=^TestMacOSSDKHistory$")
		command.Env = append(os.Environ(), "UTM_MACOS_HISTORY_CHILD=1")
		if out, e := command.CombinedOutput(); e != nil {
			t.Fatalf("isolated history test: %v\n%s", e, out)
		}
		return
	}

	cfg, rules, cache := macConfig(t), macRules(t), plugins.NewCELCache("macos-history")
	var history []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			_, _ = w.Write([]byte(`{"v11-log-macos-test":{"mappings":{"properties":{"@timestamp":{"type":"date"},"dataSource":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"log":{"properties":{"correlationCandidate":{"properties":{"endpointSecurity":{"type":"keyword"},"ransomware":{"type":"keyword"}}}}}}}}}`))
			return
		}
		if r.URL.Path != "/v11-log-macos-*/_search" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		b, e := io.ReadAll(r.Body)
		if e != nil {
			t.Error(e)
		}
		query := string(b)
		clauses := append(gjson.Get(query, "query.bool.must").Array(), gjson.Get(query, "query.bool.filter").Array()...)
		if len(clauses) != 3 {
			t.Errorf("missing history clauses: %s", query)
		}
		hits := []map[string]any{}
		for _, doc := range history {
			yes := true
			for _, clause := range clauses {
				if term := clause.Get("term"); term.Exists() {
					for field, v := range term.Map() {
						if gjson.Get(doc, strings.TrimSuffix(field, ".keyword")).String() != v.Get("value").String() {
							yes = false
						}
					}
				} else if span := clause.Get("range"); span.Exists() {
					for field, limits := range span.Map() {
						stamp, e := time.Parse(time.RFC3339Nano, gjson.Get(doc, field).String())
						if e != nil {
							t.Error(e)
						}
						cutoff, e := time.Parse(time.RFC3339Nano, limits.Get("gte").String())
						if e != nil {
							t.Error(e)
						}
						if stamp.Before(cutoff) {
							yes = false
						}
					}
				} else {
					t.Errorf("unsupported clause %s", clause.Raw)
				}
			}
			if yes {
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-macos-test", "_source": map[string]any{}})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "hits": map[string]any{"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits}})
	}))
	defer server.Close()
	if e := sdkos.Connect([]string{server.URL}, "", ""); e != nil {
		t.Fatal(e)
	}
	mutate := func(doc, field string, value any) string {
		var m map[string]any
		if e := json.Unmarshal([]byte(doc), &m); e != nil {
			t.Fatal(e)
		}
		macPut(m, field, value, value == nil)
		b, e := json.Marshal(m)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	for _, f := range macFixtures(t) {
		if f.Name != "endpoint_security_bypass native positive" && f.Name != "macos_ransomware_indicators native positive" {
			continue
		}
		t.Run(f.Name, func(t *testing.T) {
			out := macParse(t, cfg, f.Raw, f.DataSource, cache)
			name := f.Matches[0]
			search := rules[name].Correlation[0]
			prior := mutate(out, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
			history = nil
			for i := uint64(0); i < search.Count-1; i++ {
				history = append(history, prior)
			}
			yes, _, e := search.Execute(&out)
			if e != nil || yes {
				t.Fatalf("below threshold: %v %v", yes, e)
			}
			history = append(history, prior)
			yes, _, e = search.Execute(&out)
			if e != nil || !yes {
				t.Fatalf("at threshold: %v %v", yes, e)
			}
			for _, change := range []struct {
				field string
				value any
			}{{"dataSource", "other-mac"}, {macMarker(name), nil}, {"@timestamp", time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano)}} {
				history = nil
				for i := uint64(0); i < search.Count; i++ {
					history = append(history, mutate(prior, change.field, change.value))
				}
				yes, _, e = search.Execute(&out)
				if e != nil || yes {
					t.Errorf("wrong/benign/expired %s: %v %v", change.field, yes, e)
				}
			}
		})
	}
}

// Optional bounded customer evidence stays outside the repository. Only counts
// and field names are logged; this remains a projected parser/CEL replay.
func TestMacOSPrivateEvidence(t *testing.T) {
	dir := os.Getenv("UTM_MACOS_EVIDENCE")
	if dir == "" {
		t.Skip("set UTM_MACOS_EVIDENCE to bounded private evidence directory")
	}
	paths, e := filepath.Glob(filepath.Join(dir, "macos-*.json"))
	if e != nil {
		t.Fatal(e)
	}
	cfg, rules, cache := macConfig(t), macRules(t), plugins.NewCELCache("macos-private")
	total := 0
	missing := map[string]int{}
	matches := map[string]int{}
	for _, path := range paths {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		var doc map[string]any
		if e = json.Unmarshal(b, &doc); e != nil {
			t.Fatal(e)
		}
		seen := map[string]bool{}
		var visit func(any)
		visit = func(value any) {
			switch value := value.(type) {
			case map[string]any:
				if src, ok := value["_source"].(map[string]any); ok {
					id, _ := value["_id"].(string)
					if seen[id] {
						return
					}
					seen[id] = true
					raw, _ := src["raw"].(string)
					source, _ := src["dataSource"].(string)
					out := macParse(t, cfg, raw, source, cache)
					if out == "" {
						t.Errorf("%s unexpectedly dropped", id)
						return
					}
					stored, e := json.Marshal(src)
					if e != nil {
						t.Fatal(e)
					}
					total++
					for rawField, standard := range map[string]string{"timestamp": "deviceTime", "process": "origin.process", "message": "log.message"} {
						want := gjson.Get(raw, rawField)
						if !want.Exists() {
							continue
						}
						if gjson.Get(out, standard).String() != want.String() {
							t.Errorf("%s incorrect projected %s", id, standard)
						}
						if gjson.GetBytes(stored, standard).String() != want.String() {
							missing[standard]++
						}
					}
					if gjson.Get(out, "origin.host").String() != source {
						t.Errorf("%s incorrect host", id)
					}
					if !gjson.GetBytes(stored, "origin.host").Exists() {
						missing["origin.host"]++
					}
					if !gjson.GetBytes(stored, "log.eventMessage").Exists() {
						missing["log.eventMessage"]++
					}
					for name, r := range rules {
						yes, e := cache.Eval(r.Where, out)
						if e != nil {
							t.Errorf("%s %s CEL: %v", id, name, e)
						}
						if yes {
							matches[name]++
						}
					}
					return
				}
				for _, v := range value {
					visit(v)
				}
			case []any:
				for _, v := range value {
					visit(v)
				}
			}
		}
		visit(doc)
	}
	if total == 0 {
		t.Fatal("no private evidence records")
	}
	t.Logf("private distinct records=%d stored mismatches/absences=%v projected CEL candidates=%v", total, missing, matches)
}
