package main

// Offline Windows parser and correlation contracts. JSON key sanitization,
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
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
	"google.golang.org/protobuf/types/known/structpb"
	"text/template"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

type winFixture struct {
	Name       string         `json:"name"`
	DataSource string         `json:"dataSource"`
	Raw        string         `json:"raw"`
	Expected   map[string]any `json:"expected"`
	Absent     []string       `json:"absent"`
	Matches    []string       `json:"matches"`
}

func winPut(m map[string]any, path string, value any, remove bool) {
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
func winGet(m map[string]any, p string) (any, bool) {
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
func winConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/windows/windows-events.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}
func winRegex(t *testing.T, g *plugins.Grok, cfg *plugins.Config) *regexp.Regexp {
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
func winParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "wineventlog", "dataSource": dataSource, "log": map[string]any{}}
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
					v, ok := winGet(draft, src)
					if !ok {
						continue
					}
					str, ok := v.(string)
					if !ok {
						t.Fatalf("non-string grok source %s", src)
					}
					r := winRegex(t, g, cfg)
					m := r.FindStringSubmatch(str)
					if m == nil {
						continue
					}
					for i, p := range g.Patterns {
						if p.FieldName != "" {
							winPut(draft, p.FieldName, m[r.SubexpIndex(fmt.Sprintf("f%d", i))], false)
						}
					}
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := winGet(draft, p); ok {
							winPut(draft, s.Rename.To, v, false)
							winPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					for _, p := range s.Trim.Fields {
						if v, ok := winGet(draft, p); ok {
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
							winPut(draft, p, str, false)
						}
					}
				case "add":
					winPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						winPut(draft, p, nil, true)
					}
				case "json":
					source, ok := winGet(draft, s.Json.Source)
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
					for key, value := range winSanitizeJSON(parsed) {
						winPut(draft, "log."+key, value, false)
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := winGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								winPut(draft, field, utils.CastString(value), false)
							case "int":
								winPut(draft, field, utils.CastInt64(value), false)
							default:
								t.Fatalf("unsupported cast %s", s.Cast.To)
							}
						}
					}
				case "drop":
					t.Fatal("fixture unexpectedly dropped")
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
func winRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths, e := filepath.Glob("../../rules/windows/*.yml")
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

func winSanitizeJSON(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		utils.SanitizeField(&key)
		if nested, ok := value.(map[string]any); ok {
			value = winSanitizeJSON(nested)
		}
		out[key] = value
	}
	return out
}

func winFixtures(t *testing.T) []winFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/windows_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixtures []winFixture
	if e = json.Unmarshal(b, &fixtures); e != nil {
		t.Fatal(e)
	}
	return fixtures
}

func TestWindowsRawContracts(t *testing.T) {
	cfg, rules, cache := winConfig(t), winRules(t), plugins.NewCELCache("windows-raw")
	for _, f := range winFixtures(t) {
		t.Run(f.Name, func(t *testing.T) {
			out := winParse(t, cfg, f.Raw, f.DataSource, cache)
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
			for name, r := range rules {
				if len(r.Correlation) == 0 || !strings.Contains(r.Where, "log.authenticationSourceDomain") {
					continue
				}
				matched, e := cache.Eval(r.Where, out)
				if e != nil {
					t.Fatal(e)
				}
				marker := gjson.Get(out, winCandidateField(name)).String() == "match"
				if marker != matched {
					t.Errorf("%s marker=%v predicate=%v", name, marker, matched)
				}
			}
			if len(f.Matches) == 0 {
				for name, r := range rules {
					matched, e := cache.Eval(r.Where, out)
					if e != nil {
						t.Fatalf("%s compile/eval: %v", name, e)
					}
					if matched {
						t.Errorf("identity-less negative fixture matched %s", name)
					}
				}
			}
			for _, name := range f.Matches {
				r := rules[name]
				if r == nil {
					t.Fatalf("unknown rule %s", name)
				}
				got, e := cache.Eval(r.Where, out)
				if e != nil || !got {
					t.Fatalf("%s where=%v error=%v", name, got, e)
				}
				for _, search := range r.Correlation {
					for _, term := range search.With {
						if value := term.Value.GetStringValue(); strings.HasPrefix(value, "{{.") {
							field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
							if !gjson.Get(out, field).Exists() {
								t.Errorf("%s unresolved %s", name, field)
							}
						}
					}
				}
			}
		})
	}
}

// Evaluate the generated OpenSearch query against a tiny local history. This is
// deliberately limited to exact terms and the timestamp range these rules use.
func winHistoryMatches(t *testing.T, query string, event string) bool {
	t.Helper()
	clauses := append(gjson.Get(query, "query.bool.filter").Array(), gjson.Get(query, "query.bool.must").Array()...)
	if len(clauses) < 5 {
		t.Errorf("missing history constraints: %s", query)
	}
	for _, term := range clauses {
		if object := term.Get("term"); object.Exists() {
			for field, value := range object.Map() {
				got := gjson.Get(event, strings.TrimSuffix(field, ".keyword"))
				if !got.Exists() || got.String() != value.Get("value").String() {
					return false
				}
			}
		} else if object := term.Get("range"); object.Exists() {
			for field, conditions := range object.Map() {
				got, e := time.Parse(time.RFC3339Nano, gjson.Get(event, field).String())
				if e != nil {
					t.Fatal(e)
				}
				for op, value := range conditions.Map() {
					bound, e := time.Parse(time.RFC3339Nano, value.String())
					if e != nil {
						t.Fatal(e)
					}
					if op != "gte" {
						t.Fatalf("unsupported range %s", op)
					}
					if got.Before(bound) {
						return false
					}
				}
			}
		} else {
			t.Fatalf("unsupported history clause %s", term.Raw)
		}
	}
	return true
}

func TestWindowsSDKHistory(t *testing.T) {
	cfg, rules, cache := winConfig(t), winRules(t), plugins.NewCELCache("windows-history")
	var history []string
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_mapping") {
			_, _ = w.Write([]byte(`{"v11-log-wineventlog-test":{"mappings":{"properties":{"@timestamp":{"type":"date"},"dataSource":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"log":{"properties":{"authenticationSource":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"authenticationSourceType":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"eventCode":{"type":"long"},"eventDataTicketEncryptionType":{"type":"long"},"eventDataPreAuthType":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"authenticationSourceDomain":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"authenticationCandidate":{"properties":{"kerberoastingDetection":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"asrepRoastingDetection":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"silverTicketDetection":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"goldenTicketDetection":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"adfsAuthenticationAnomalies":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"bruteforceAttack":{"type":"text","fields":{"keyword":{"type":"keyword"}}},"bruteforceMultipleLogonFailureFollowedBySuccess":{"type":"text","fields":{"keyword":{"type":"keyword"}}}}}}},"target":{"properties":{"user":{"type":"text","fields":{"keyword":{"type":"keyword"}}}}}}}}}`))
			return
		}
		requests++
		if r.URL.Path != "/v11-log-wineventlog-*/_search" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		body, e := io.ReadAll(r.Body)
		if e != nil {
			t.Error(e)
		}
		hits := []map[string]any{}
		for _, doc := range history {
			if winHistoryMatches(t, string(body), doc) {
				var source map[string]any
				if e := json.Unmarshal([]byte(doc), &source); e != nil {
					t.Error(e)
				}
				hits = append(hits, map[string]any{"_id": fmt.Sprint(len(hits)), "_index": "v11-log-wineventlog-test", "_source": source})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"took": 1, "timed_out": false, "_shards": map[string]any{"total": 1, "successful": 1, "failed": 0}, "hits": map[string]any{"total": map[string]any{"value": len(hits), "relation": "eq"}, "hits": hits}})
	}))
	defer server.Close()
	if e := sdkos.Connect([]string{server.URL}, "", ""); e != nil {
		t.Fatal(e)
	}
	mutate := func(event, field string, value any) string {
		var doc map[string]any
		if e := json.Unmarshal([]byte(event), &doc); e != nil {
			t.Fatal(e)
		}
		winPut(doc, field, value, value == nil)
		b, e := json.Marshal(doc)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	for _, f := range winFixtures(t) {
		if len(f.Matches) == 0 {
			continue
		}
		t.Run(f.Name, func(t *testing.T) {
			trigger := winParse(t, cfg, f.Raw, f.DataSource, cache)
			for _, name := range f.Matches {
				r := rules[name]
				for _, search := range r.Correlation {
					previous := mutate(trigger, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
					// Success-after-failures searches 4625, not the triggering 4624.
					for _, term := range search.With {
						if term.Field == "log.eventCode" && term.Value.GetStringValue() == "" {
							var failedRaw map[string]any
							if e := json.Unmarshal([]byte(f.Raw), &failedRaw); e != nil {
								t.Fatal(e)
							}
							failedRaw["eventCode"] = term.Value.AsInterface()
							b, e := json.Marshal(failedRaw)
							if e != nil {
								t.Fatal(e)
							}
							previous = winParse(t, cfg, string(b), f.DataSource, cache)
							previous = mutate(previous, "@timestamp", time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano))
						}
					}
					history = nil
					for i := uint64(0); i < search.Count-1; i++ {
						history = append(history, previous)
					}
					ok, _, e := search.Execute(&trigger)
					if e != nil || ok {
						t.Fatalf("%s below threshold result=%v error=%v", name, ok, e)
					}
					history = append(history, previous)
					ok, _, e = search.Execute(&trigger)
					if e != nil || !ok {
						t.Fatalf("%s threshold result=%v error=%v", name, ok, e)
					}
					for _, field := range []string{"dataSource", "log.authenticationSource", "log.authenticationSourceType", "log.authenticationSourceDomain", "log.authenticationCandidate", "log.eventCode"} {
						history = nil
						for i := uint64(0); i < search.Count; i++ {
							history = append(history, mutate(previous, field, "different"))
						}
						ok, _, e = search.Execute(&trigger)
						if e != nil || ok {
							t.Fatalf("%s mixed %s result=%v error=%v", name, field, ok, e)
						}
					}
					history = nil
					for i := uint64(0); i < search.Count; i++ {
						history = append(history, mutate(previous, "@timestamp", time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano)))
					}
					ok, _, e = search.Execute(&trigger)
					if e != nil || ok {
						t.Fatalf("%s expired history result=%v error=%v", name, ok, e)
					}
				}
				for _, field := range []string{"log.authenticationSource", "log.authenticationSourceType", "log.authenticationSourceDomain"} {
					missing := mutate(trigger, field, nil)
					ok, e := cache.Eval(r.Where, missing)
					if e != nil || ok {
						t.Errorf("%s accepted missing %s", name, field)
					}
				}
			}
		})
	}
	// Reproduce the original regression with the actual SDK. An `or` sibling
	// cannot rescue this: Execute returns before it reaches Or on missing IP.
	before := requests
	legacy := &plugins.SearchRequest{IndexPattern: "v11-log-wineventlog-*", Count: 3, With: []*plugins.Expression{{Field: "origin.ip.keyword", Operator: "filter_term", Value: structpb.NewStringValue("{{.origin.ip}}")}}}
	legacy.Or = []*plugins.SearchRequest{{IndexPattern: "v11-log-wineventlog-*", Count: 1, With: []*plugins.Expression{{Field: "target.user.keyword", Operator: "filter_term", Value: structpb.NewStringValue("{{.target.user}}")}}}}
	ipless := `{"target":{"user":"analyst"}}`
	ok, _, e := legacy.Execute(&ipless)
	if e == nil || ok || requests != before {
		t.Fatalf("nil placeholder unexpectedly queried: %v %v", ok, e)
	}
}

// Opt-in replay of bounded private OpenSearch evidence. No customer log content,
// identities or requests are printed; this tests projected normalization and CEL,
// not the production parser or a historical customer query.
func TestWindowsPrivateEvidence(t *testing.T) {
	dir := os.Getenv("UTM_WINDOWS_EVIDENCE")
	if dir == "" {
		t.Skip("set UTM_WINDOWS_EVIDENCE to the private bounded evidence directory")
	}
	paths, e := filepath.Glob(filepath.Join(dir, "windows-*.json"))
	if e != nil {
		t.Fatal(e)
	}
	if len(paths) == 0 {
		t.Fatal("no evidence files")
	}
	cfg, rules, cache := winConfig(t), winRules(t), plugins.NewCELCache("windows-private-evidence")
	samples, ipless := 0, 0
	matches := map[string]int{}
	observed := map[string]int{}
	for _, path := range paths {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		seen := map[string]bool{}
		for _, bucket := range gjson.GetBytes(b, "aggregations.codes.buckets").Array() {
			hits := append(bucket.Get("representative.hits.hits").Array(), bucket.Get("missing_source.representative.hits.hits").Array()...)
			for _, hit := range hits {
				id := hit.Get("_id").String()
				if seen[id] {
					continue
				}
				seen[id] = true
				src := hit.Get("_source")
				raw := src.Get("raw").String()
				out := winParse(t, cfg, raw, src.Get("dataSource").String(), cache)
				samples++
				if !gjson.Get(out, "origin.ip").Exists() {
					ipless++
				}
				for rawField, standard := range map[string]string{"computer": "target.host", "timestamp": "deviceTime", "data.Workstation": "origin.host", "data.ProcessName": "origin.path"} {
					source := gjson.Get(raw, rawField)
					if source.Type != gjson.String || source.String() == "" || source.String() == "-" {
						continue
					}
					if gjson.Get(out, standard).String() != source.String() {
						t.Errorf("%s projected %s differs from source %s", id, standard, rawField)
					}
					if src.Get(standard).String() != source.String() {
						observed[standard]++
					}
				}
				for name, r := range rules {
					yes, e := cache.Eval(r.Where, out)
					if e != nil {
						t.Errorf("%s %s CEL error: %v", id, name, e)
					}
					if !yes {
						continue
					}
					matches[name]++
					for _, search := range r.Correlation {
						for _, term := range search.With {
							value := term.Value.GetStringValue()
							if !strings.HasPrefix(value, "{{.") {
								continue
							}
							field := strings.TrimSuffix(strings.TrimPrefix(value, "{{."), "}}")
							if !gjson.Get(out, field).Exists() {
								t.Errorf("%s %s unresolved placeholder %s", id, name, field)
							}
						}
					}
				}
			}
		}
	}
	t.Logf("private unique raw samples=%d projected IP-less=%d standard field mismatches in stored output=%v projected CEL candidates=%v", samples, ipless, observed, matches)
}

func winCandidateField(ruleName string) string {
	parts := strings.Split(ruleName, "_")
	key := parts[0]
	for _, p := range parts[1:] {
		key += strings.ToUpper(p[:1]) + p[1:]
	}
	return "log.authenticationCandidate." + key
}
