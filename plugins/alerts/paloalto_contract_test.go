package main

// Offline Palo Alto extraction model, not the EventProcessor itself. The grok,
// csv, add, cast, delete and trim steps copy the behaviour of the parser plugins
// of EventProcessor 8a3ade72bd9d12db21f6b273200588fb49540f14, including their
// failures: grok matches each pattern on its own at the start of the remaining
// text and writes nothing unless every pattern matched; csv fails when a line has
// fewer columns than the step has headers. As in the engine, a failed step or
// where clause records an error on the event and the pipeline continues without
// that step's output. Every field name a step writes is cleaned like the plugins
// clean it. CEL, Event serialization, placeholder expansion, query creation and
// history thresholds use the linked SDK. External geolocation is not executed.
import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

type paloaltoFixture struct {
	Name       string              `json:"name"`
	DataSource string              `json:"dataSource"`
	Raw        string              `json:"raw"`
	Expected   map[string]any      `json:"expected"`
	Absent     []string            `json:"absent"`
	Matches    []string            `json:"matches"`
	AlertSides map[string][]string `json:"alertSides"`
}

func paloaltoPut(m map[string]any, path string, value any, remove bool) {
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
func paloaltoGet(m map[string]any, p string) (any, bool) {
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

// paloaltoWriteName returns the name the EventProcessor parser plugins actually
// write: grok, add, rename, csv, kv and json pass it through go-sdk
// utils.SanitizeField and reject a reserved name. Since go-sdk v1.1.35 the
// sanitizer keeps letters, digits, dots and underscores and removes every other
// character (earlier versions also removed underscores). Conditions and rules look
// names up as written, so a filter that writes "log.pa-type" and tests
// "log.pa-type" never matches in production and must not match here either.
func paloaltoWriteName(t *testing.T, name string, allowEmpty bool) string {
	t.Helper()
	utils.SanitizeField(&name)
	if e := utils.ValidateReservedField(name, allowEmpty); e != nil {
		t.Fatal(e)
	}
	return name
}
func paloaltoConfig(t *testing.T) *plugins.Config {
	t.Helper()
	b, e := utils.ReadPbYaml("../../filters/paloalto/pa_firewall.yml")
	if e != nil {
		t.Fatal(e)
	}
	c := new(plugins.Config)
	if e = protojson.Unmarshal(b, c); e != nil {
		t.Fatal(e)
	}
	return c
}

// paloaltoCompile compiles one pattern the way the SDK RegexpCache does: "{{...}}"
// templates are expanded from the configured shared patterns, then the pattern is
// compiled unanchored.
func paloaltoCompile(cfg *plugins.Config, pattern string) (*regexp.Regexp, error) {
	for i := 0; i < 10 && strings.Contains(pattern, "{{") && len(cfg.Patterns) > 0; i++ {
		tmpl, err := template.New("pattern").Parse(pattern)
		if err != nil {
			break
		}
		var b bytes.Buffer
		if err = tmpl.Execute(&b, cfg.Patterns); err != nil || b.String() == pattern {
			break
		}
		pattern = b.String()
	}
	return regexp.Compile(pattern)
}

// paloaltoText renders a draft value as gjson's Result.String() does for the plugins.
func paloaltoText(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

type paloaltoWrite struct {
	name  string
	value any
}

// paloaltoGrok follows the grok plugin: before each pattern the remaining text is
// trimmed; the pattern's leftmost match must be non-empty and must start the
// remaining text; the trimmed match is kept and removed from the text. Nothing is
// written unless every pattern matched.
func paloaltoGrok(cfg *plugins.Config, g *plugins.Grok, draft map[string]any) ([]paloaltoWrite, error) {
	src := g.Source
	if src == "" {
		src = "raw"
	}
	v, ok := paloaltoGet(draft, src)
	if !ok {
		return nil, nil
	}
	text := paloaltoText(v)
	var out []paloaltoWrite
	matched := 0
	for _, p := range g.Patterns {
		text = strings.TrimSpace(text)
		if utf8.RuneCountInString(text) == 0 {
			break
		}
		r, err := paloaltoCompile(cfg, p.Pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regexp: %w", err)
		}
		match := r.FindString(text)
		if match == "" || !strings.HasPrefix(text, match) {
			break
		}
		name := p.FieldName
		utils.SanitizeField(&name)
		if err = utils.ValidateReservedField(name, true); err != nil {
			return nil, err
		}
		matched++
		if name != "" {
			out = append(out, paloaltoWrite{name, strings.TrimSpace(match)})
		}
		text = strings.TrimPrefix(text, match)
	}
	if matched != len(g.Patterns) {
		return nil, nil
	}
	return out, nil
}

// paloaltoCSV follows the csv plugin: the source is read with Go's csv reader
// (strict quotes), and a record with fewer columns than there are headers fails
// the whole step, which then writes nothing. Values are trimmed.
func paloaltoCSV(c *plugins.Csv, draft map[string]any) ([]paloaltoWrite, error) {
	text := ""
	if v, ok := paloaltoGet(draft, c.Source); ok {
		text = paloaltoText(v)
	}
	reader := csv.NewReader(strings.NewReader(text))
	separator := c.Separator
	if utf8.RuneCountInString(separator) > 1 {
		return nil, errors.New("separator should be a single character")
	}
	if separator == "" {
		separator = ","
	}
	for _, r := range separator {
		reader.Comma = r
		break
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse log: %w", err)
	}
	var out []paloaltoWrite
	for _, record := range records {
		for i, header := range c.Headers {
			if i >= len(record) {
				return nil, errors.New("column index out of range, number of headers should match the number of resulting columns")
			}
			utils.SanitizeField(&header)
			if header == "" {
				continue
			}
			if err = utils.ValidateReservedField(header, false); err != nil {
				return nil, err
			}
			out = append(out, paloaltoWrite{header, strings.TrimSpace(record[i])})
		}
	}
	return out, nil
}
func paloaltoParse(t *testing.T, cfg *plugins.Config, raw string, dataSource string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": "firewall-paloalto", "dataSource": dataSource, "log": map[string]any{}}
	var stepErrors []string
	// A failed step records its error and the pipeline continues without its output.
	apply := func(kind string, writes []paloaltoWrite, err error) {
		if err != nil {
			stepErrors = append(stepErrors, fmt.Sprintf("failed to parse log during pipeline execution (%s): %v", kind, err))
			return
		}
		for _, w := range writes {
			paloaltoPut(draft, w.name, w.value, false)
		}
	}
	for _, stage := range cfg.Pipeline {
		matched := false
		for _, dataType := range stage.DataTypes {
			if dataType == "firewall-paloalto" {
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
					text := string(snapshot)
					match, e := cache.Evaluate(&text, w)
					if e != nil {
						stepErrors = append(stepErrors, "failed to evaluate where clause using CEL: "+e.Error())
						continue
					}
					if !match {
						continue
					}
				}
				switch kind {
				case "grok":
					writes, err := paloaltoGrok(cfg, s.Grok, draft)
					apply(kind, writes, err)
				case "csv":
					writes, err := paloaltoCSV(s.Csv, draft)
					apply(kind, writes, err)
				case "rename":
					for _, p := range s.Rename.From {
						if v, ok := paloaltoGet(draft, p); ok {
							paloaltoPut(draft, paloaltoWriteName(t, s.Rename.To, false), v, false)
							paloaltoPut(draft, p, nil, true)
							break
						}
					}
				case "trim":
					// The trim plugin skips absent or empty values, trims spaces and stores text.
					for _, p := range s.Trim.Fields {
						v, ok := paloaltoGet(draft, p)
						if !ok || paloaltoText(v) == "" {
							continue
						}
						str := strings.TrimSpace(paloaltoText(v))
						switch s.Trim.Function {
						case "prefix":
							str = strings.TrimPrefix(str, s.Trim.Substring)
						case "suffix":
							str = strings.TrimSuffix(str, s.Trim.Substring)
						case "substring":
							str = strings.ReplaceAll(str, s.Trim.Substring, "")
						case "regex":
							r, err := paloaltoCompile(cfg, s.Trim.Substring)
							if err != nil {
								t.Fatalf("trim regex %s: %v", s.Trim.Substring, err)
							}
							found := r.FindAllString(str, -1)
							if len(found) == 0 {
								continue
							}
							for _, m := range found {
								str = strings.ReplaceAll(str, m, "")
							}
						default:
							t.Fatalf("unsupported trim %s", s.Trim.Function)
						}
						paloaltoPut(draft, p, strings.TrimSpace(str), false)
					}
				case "add":
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add function %s", s.Add.Function)
					}
					// The add plugin writes the parameter's string value.
					paloaltoPut(draft, paloaltoWriteName(t, s.Add.Params["key"].GetStringValue(), false), s.Add.Params["value"].GetStringValue(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						paloaltoPut(draft, p, nil, true)
					}
				case "kv":
					v, ok := paloaltoGet(draft, s.Kv.Source)
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
							paloaltoPut(draft, "log."+key, pair[1], false)
						}
					}
				case "dynamic":
					if s.Dynamic.Plugin != "com.utmstack.geolocation" {
						t.Fatalf("unsupported dynamic plugin %s", s.Dynamic.Plugin)
					}
					field := s.Dynamic.Params["source"].GetStringValue()
					v, ok := paloaltoGet(draft, field)
					if !ok {
						t.Fatalf("missing dynamic source %s", field)
					}
					ip := net.ParseIP(fmt.Sprint(v))
					if ip == nil || ip.IsUnspecified() {
						t.Fatalf("invalid address reaches geolocation: %s", field)
					}
					// The external geolocation service is not executed.
				case "json":
					source, ok := paloaltoGet(draft, s.Json.Source)
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
					for key, value := range paloaltoSanitizeJSON(parsed) {
						paloaltoPut(draft, "log."+key, value, false)
					}
				case "reformat":
					for _, field := range s.Reformat.Fields {
						value, ok := paloaltoGet(draft, field)
						if !ok {
							continue
						}
						stamp, err := time.Parse(s.Reformat.FromFormat, fmt.Sprint(value))
						if err != nil {
							t.Fatalf("time conversion %s: %v", field, err)
						}
						paloaltoPut(draft, field, stamp.Format(s.Reformat.ToFormat), false)
					}
				case "cast":
					for _, field := range s.Cast.Fields {
						if value, ok := paloaltoGet(draft, field); ok {
							switch s.Cast.To {
							case "string":
								paloaltoPut(draft, field, utils.CastString(value), false)
							case "float":
								paloaltoPut(draft, field, utils.CastFloat64(value), false)
							case "int":
								paloaltoPut(draft, field, utils.CastInt64(value), false)
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
	if len(stepErrors) > 0 {
		draft["errors"] = stepErrors
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
func paloaltoRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	paths, e := filepath.Glob("../../rules/paloalto/pa_firewall/*.yml")
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

// paloaltoSanitizeJSON cleans the top-level keys only, as the EventProcessor
// JSON plugin does; keys inside nested objects keep their characters.
func paloaltoSanitizeJSON(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		utils.SanitizeField(&key)
		out[key] = value
	}
	return out
}

func paloaltoFixtures(t *testing.T) []paloaltoFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/paloalto_raw.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []paloaltoFixture
	if e = json.Unmarshal(b, &cases); e != nil {
		t.Fatal(e)
	}
	return cases
}

func paloaltoCheck(t *testing.T, fixtures []paloaltoFixture) {
	cfg, rules, cache := paloaltoConfig(t), paloaltoRules(t), plugins.NewCELCache("paloalto")
	if len(rules) != 11 {
		t.Fatalf("rules: %d", len(rules))
	}
	coverage := map[string]int{}
	for _, f := range fixtures {
		for _, name := range f.Matches {
			coverage[name]++
		}
	}
	if len(fixtures) > 60 {
		for name := range rules {
			if coverage[name] == 0 || coverage[name] == len(fixtures) {
				t.Fatalf("missing positive/negative coverage for %s", name)
			}
		}
	}
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			out := paloaltoParse(t, cfg, f.Raw, f.DataSource, cache)
			if errs := gjson.Get(out, "errors"); errs.Exists() {
				t.Errorf("parser errors: %s", errs.Raw)
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
			expect := map[string]bool{}
			for _, n := range f.Matches {
				expect[n] = true
			}
			for name, r := range rules {
				matched, err := cache.Eval(r.Where, out)
				if err != nil {
					t.Fatal(err)
				}
				if matched != expect[name] {
					t.Errorf("%s matched %v want %v", name, matched, expect[name])
				}
				if !matched {
					continue
				}
				event := new(plugins.Event)
				if err := protojson.Unmarshal([]byte(out), event); err != nil {
					t.Fatal(err)
				}
				alert := &plugins.Alert{Events: []*plugins.Event{event}}
				switch r.Adversary {
				case "origin":
					alert.Adversary, alert.Target = event.Origin, event.Target
				case "target":
					alert.Adversary, alert.Target = event.Target, event.Origin
				default:
					t.Fatalf("unsupported adversary %s", r.Adversary)
				}
				if sides, ok := f.AlertSides[name]; ok {
					if len(sides) != 2 || alert.GetAdversary().GetIp() != sides[0] || alert.GetTarget().GetIp() != sides[1] {
						t.Errorf("%s alert endpoints do not match independent expectation %v", name, sides)
					}
				} else if strings.HasPrefix(f.DataSource, "synthetic-") {
					t.Errorf("missing independent alert sides for %s", name)
				}
				wire, err := utils.ProtoMessageToString(alert)
				if err != nil {
					t.Fatal(err)
				}
				for _, field := range r.GroupBy {
					path := strings.Replace(field, "lastEvent.", "events.0.", 1)
					if (field == "lastEvent.dataSource" || field == "lastEvent.log.paScope" || field == "lastEvent.log.paVsys") && !gjson.Get(*wire, path).Exists() {
						t.Errorf("grouping scope missing: %s", field)
					}
				}
				for _, search := range r.Correlation {
					for _, term := range search.With {
						v := term.Value.GetStringValue()
						if strings.HasPrefix(v, "{{.") {
							field := strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")
							if !gjson.Get(out, field).Exists() {
								t.Errorf("unresolved %s", field)
							}
						}
					}
				}
			}
		})
	}
}
func TestPaloAltoRawContracts(t *testing.T) { paloaltoCheck(t, paloaltoFixtures(t)) }
func TestPaloAltoPrivateContracts(t *testing.T) {
	path := os.Getenv("PALOALTO_PRIVATE_FIXTURES")
	if path == "" {
		t.Skip("private fixtures not supplied")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var f []paloaltoFixture
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	paloaltoCheck(t, f)
}

// The model must fail where the parser plugins fail, or it hides real-engine defects.
func TestPaloAltoModelFollowsPlugins(t *testing.T) {
	short := &plugins.Csv{Source: "log.line", Separator: ",", Headers: []string{"log.a", "log.b", "log.c"}}
	for line, want := range map[string]int{"1,2": -1, "1,2,3": 3, `"1,5",2,3`: 3, "1,2,3,4": 3} {
		writes, err := paloaltoCSV(short, map[string]any{"log": map[string]any{"line": line}})
		if (want < 0) != (err != nil) || (want >= 0 && len(writes) != want) {
			t.Errorf("csv %q: %d writes, error %v; want %d", line, len(writes), err, want)
		}
	}
	cfg := &plugins.Config{}
	for _, tc := range []struct {
		text     string
		patterns []string
		writes   int
	}{
		{"x1", []string{"^x", "[0-9]+"}, 2},
		{"xz", []string{"^x", "y*"}, 0},          // a pattern that matches only empty text fails the step
		{"1.2.", []string{"[0-9.]+", "\\.$"}, 0}, // the first pattern swallows the final period
		{"1.2.", []string{"[0-9.]*[0-9]", "\\.$"}, 2},
		{"a b", []string{"b"}, 0}, // a match must start the remaining text
	} {
		g := &plugins.Grok{Source: "log.line"}
		for i, p := range tc.patterns {
			g.Patterns = append(g.Patterns, &plugins.Pattern{FieldName: "log.f" + strconv.Itoa(i), Pattern: p})
		}
		writes, err := paloaltoGrok(cfg, g, map[string]any{"log": map[string]any{"line": tc.text}})
		if err != nil || len(writes) != tc.writes {
			t.Errorf("grok %q %q: %d writes, error %v; want %d", tc.text, tc.patterns, len(writes), err, tc.writes)
		}
	}
}

// Each longer CSV tier must run only when the line has that many columns, counted as the
// csv plugin counts them, quoted commas included; otherwise the plugin fails the step.
func TestPaloAltoCSVTierCondition(t *testing.T) {
	cache := plugins.NewCELCache("paloalto-tiers")
	count := regexp.MustCompile(`regexMatch\("log\.paPayload", "\^\(\?:\(\?:.*\{([0-9]+)\}"\)`)
	tiers := 0
	for _, s := range paloaltoConfig(t).Pipeline[0].Steps {
		if s.Csv == nil {
			continue
		}
		m := count.FindStringSubmatch(s.Csv.Where)
		if m == nil {
			continue
		}
		need, _ := strconv.Atoi(m[1])
		need++
		if need != len(s.Csv.Headers) {
			t.Errorf("tier needs %d columns but has %d headers: %s", need, len(s.Csv.Headers), s.Csv.Where)
		}
		tiers++
		for _, have := range []int{need - 1, need, need + 1} {
			cols := make([]string, have)
			for i := range cols {
				cols[i] = []string{"1", `"a, b"`, "", `"say ""x"", y"`}[i%4]
			}
			doc, err := json.Marshal(map[string]any{"log": map[string]any{"paPayload": strings.Join(cols, ",")}})
			if err != nil {
				t.Fatal(err)
			}
			text := string(doc)
			got, err := cache.Evaluate(&text, m[0])
			if err != nil {
				t.Fatal(err)
			}
			if got != (have >= need) {
				t.Errorf("%d columns, tier of %d: condition %v", have, need, got)
			}
		}
	}
	if tiers < 15 {
		t.Fatalf("only %d conditioned CSV tiers found", tiers)
	}
}
