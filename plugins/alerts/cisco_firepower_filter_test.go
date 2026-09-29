package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"unicode/utf8"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Cisco Firepower Threat Defense regression checks. Every raw input is FABRICATED
// (testdata/cisco-firepower). The 4300xx lines follow the key layout of genuine device records
// that were reviewed privately and are not part of this repository; the LINA lines follow the
// filter's own patterns. They use RFC 5737 and RFC 3849 documentation addresses (RFC 1918 and
// RFC 4193 private addresses only for the private-destination near misses of the
// non-standard-port rule), example names and a made-up device UUID. These tests use the pinned
// go-sdk (v1.1.36) for YAML decoding, key sanitizing, CEL and Event conversion. They do not run
// the EventProcessor: fpModel mirrors its step plugins, and testdata/cisco-firepower/replay.py
// runs the same lines through the public playground. See filters/audits/cisco-firepower.md.

const (
	fpFilter   = "../../filters/cisco/firepower.yml"
	fpRulesDir = "../../rules/cisco/firepower"
	fpData     = "testdata/cisco-firepower"
	fpTenant   = "00000000-0000-4000-8000-000000000001"
	fpAbsent   = "<absent>"
	fpIPS      = "intrusion_prevention_high_priority_events"
	fpC2       = "c2_nonstandard_port"
)

// The 4300xx message IDs seen in genuine records; only these are split into log.<Key> fields.
const fpObserved = `oneOf("log.messageId", [430001, 430002, 430003, 430007])`

var fpEnvelope = map[string]bool{"id": true, "timestamp": true, "deviceTime": true, "dataType": true,
	"dataSource": true, "tenantId": true, "tenantName": true, "raw": true, "errors": true}

func fpPipeline(t *testing.T) *plugins.Pipeline {
	t.Helper()
	encoded, err := utils.ReadPbYaml(fpFilter)
	if err != nil {
		t.Fatal(err)
	}
	config := new(plugins.Config)
	if err := protojson.Unmarshal(encoded, config); err != nil {
		t.Fatal(err)
	}
	if len(config.Pipeline) != 1 || len(config.Pipeline[0].DataTypes) != 1 ||
		config.Pipeline[0].DataTypes[0] != "firewall-cisco-firepower" {
		t.Fatalf("unexpected pipeline layout: %v", config.Pipeline)
	}
	return config.Pipeline[0]
}

// fpWhere returns the where clause of whichever step kind is set.
func fpWhere(step *plugins.Step) string {
	where := ""
	step.ProtoReflect().Range(func(_ protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		msg := v.Message()
		if fd := msg.Descriptor().Fields().ByName("where"); fd != nil {
			where = msg.Get(fd).String()
		}
		return false
	})
	return where
}

// Before this revision, 98 clauses compared log.* directly (for example log.messageId==113032),
// and one called the undeclared function lgreaterOrEqual. These rewrite a helper clause back to
// the direct form, so both can be compared; the observed-ID list becomes a chain of ==.
var (
	fpRawLog      = regexp.MustCompile(`(^|[^"\w.])log\.[A-Za-z0-9_.]+\s*(==|!=|>=|<=|<|>)`)
	fpHelperToRaw = []struct {
		helper *regexp.Regexp
		raw    string
	}{
		{regexp.MustCompile(`equals\("log\.messageId", (\d+)\)`), `log.messageId==$1`},
		{regexp.MustCompile(`greaterOrEqual\("log\.messageId", (\d+)\)`), `log.messageId>=$1`},
		{regexp.MustCompile(`lessOrEqual\("log\.messageId", (\d+)\)`), `log.messageId<=$1`},
		{regexp.MustCompile(`equals\("log\.severity", "(\d)"\)`), `log.severity=="$1"`},
		{regexp.MustCompile(regexp.QuoteMeta(fpObserved)),
			`(log.messageId==430001 || log.messageId==430002 || log.messageId==430003 || log.messageId==430007)`},
	}
	fpHelperCall = regexp.MustCompile(`(\w+)\("([A-Za-z0-9_.]+)"(?:, ("[^"]*"|\d+|\[[\d, ]+\]))?\)`)
	fpNumber     = regexp.MustCompile(`\d+`)
)

func fpRawForm(where string) string {
	for _, r := range fpHelperToRaw {
		where = r.helper.ReplaceAllString(where, r.raw)
	}
	return where
}

// fpTruthTable builds events that all carry a log object: log.messageId as the JSON number the
// filter's cast produces (around every literal in the clause), log.severity as the one-digit
// levels 0 to 7, and every other field the clause reads absent, equal to its literal, in other
// case, containing it, or unrelated.
func fpTruthTable(where string) []string {
	ids := map[float64]bool{100000: true}
	severities := []string{"6"}
	others := map[string][]any{}
	for _, m := range fpHelperCall.FindAllStringSubmatch(where, -1) {
		field, arg := m[2], m[3]
		switch field {
		case "log.messageId":
			for _, lit := range fpNumber.FindAllString(arg, -1) {
				n, _ := strconv.Atoi(lit)
				for d := -1; d <= 1; d++ {
					ids[float64(n+d)] = true
				}
			}
		case "log.severity":
			severities = []string{"0", "1", "2", "3", "4", "5", "6", "7"}
		default:
			if others[field] == nil {
				others[field] = []any{nil, "unrelated"}
			}
			if lit, err := strconv.Unquote(arg); err == nil && lit != "" {
				others[field] = append(others[field], lit, strings.ToLower(lit), strings.ToUpper(lit), "x "+lit+" y")
			}
		}
	}
	names := make([]string, 0, len(others))
	for f := range others {
		names = append(names, f)
	}
	sort.Strings(names)
	combos := []map[string]any{{}}
	for _, f := range names {
		var next []map[string]any
		for _, c := range combos {
			for _, v := range others[f] {
				n := map[string]any{f: v}
				for k, x := range c {
					n[k] = x
				}
				next = append(next, n)
			}
		}
		combos = next
	}
	var docs []string
	for id := range ids {
		for _, sev := range severities {
			for _, c := range combos {
				doc := map[string]any{"id": "x", "dataType": "firewall-cisco-firepower", "raw": "x",
					"log": map[string]any{"messageId": id, "severity": sev}}
				for f, v := range c {
					if v != nil {
						fpSet(doc, f, v)
					}
				}
				b, _ := json.Marshal(doc)
				docs = append(docs, string(b))
			}
		}
	}
	return docs
}

// No where clause compares log.* directly, and every clause compiles and runs without an error
// on a line that no header pattern accepted (no log object). Every clause that uses equals,
// greaterOrEqual or lessOrEqual on log.messageId, equals on log.severity or the observed-ID list
// has the same truth table as the direct comparison on events with a log object, and is false
// when there is none.
func TestCiscoFirepowerWhereHelpers(t *testing.T) {
	steps := fpPipeline(t).Steps
	cache := plugins.NewCELCache("cisco-firepower-where")
	noLog := `{"id":"x","dataType":"firewall-cisco-firepower","dataSource":"fixture-ftd","tenantId":"` + fpTenant + `","raw":"x"}`
	cast := -1
	for i, step := range steps {
		if c := step.Cast; c != nil && c.To == "int" && c.Where == "" && len(c.Fields) == 1 && c.Fields[0] == "log.messageId" {
			cast = i
			break
		}
	}
	if cast < 0 {
		t.Fatal("no unconditional int cast of log.messageId")
	}
	var raw []string
	clauses, checked, rows, observed := 0, 0, 0, 0
	for i, step := range steps {
		where := fpWhere(step)
		if where == "" {
			continue
		}
		clauses++
		if _, err := cache.Eval(where, noLog); err != nil {
			t.Errorf("step %d fails without a log object: %q: %v", i, where, err)
		}
		if fpRawLog.MatchString(where) {
			raw = append(raw, where)
			continue
		}
		observed += strings.Count(where, fpObserved)
		pre := fpRawForm(where)
		if pre == where {
			continue
		}
		checked++
		if i < cast {
			t.Errorf("step %d reads log.messageId before it is cast to a number: %q", i, where)
		}
		if got, err := cache.Eval(where, noLog); err != nil || got {
			t.Errorf("step %d without a log object: %q returned %t, error %v", i, where, got, err)
		}
		for _, doc := range fpTruthTable(where) {
			rows++
			got, err := cache.Eval(where, doc)
			want, errPre := cache.Eval(pre, doc)
			if err != nil || errPre != nil || got != want {
				t.Errorf("step %d: %q=%t (%v), %q=%t (%v) on %s", i, where, got, err, pre, want, errPre, doc)
				break
			}
		}
	}
	if len(raw) > 0 {
		t.Errorf("%d where clauses compare log.* directly and fail without a log object, for example %q", len(raw), raw[0])
	}
	if observed != 14 {
		t.Errorf("%d clauses restrict the key-value split to the observed IDs, want 14", observed)
	}
	if checked < 468 {
		t.Errorf("%d helper clauses read log.messageId or log.severity, want at least 468", checked)
	}
	t.Logf("%d where clauses, %d helper clauses, %d truth-table rows", clauses, checked, rows)
}

// The geolocation plugin writes its result at the destination path, so a destination below
// an address field would replace the address with an object.
func TestCiscoFirepowerGeolocationDestinations(t *testing.T) {
	steps := fpPipeline(t).Steps
	scalars := map[string]bool{}
	for _, step := range steps {
		if g := step.Grok; g != nil {
			for _, p := range g.Patterns {
				if p.FieldName != "" {
					scalars[p.FieldName] = true
				}
			}
		}
		if a := step.Add; a != nil {
			scalars[a.Params["key"].GetStringValue()] = true
		}
		if r := step.Rename; r != nil {
			scalars[r.To] = true
		}
		if c := step.Cast; c != nil {
			for _, f := range c.Fields {
				scalars[f] = true
			}
		}
		if tr := step.Trim; tr != nil {
			for _, f := range tr.Fields {
				scalars[f] = true
			}
		}
	}
	count := 0
	for i, step := range steps {
		d := step.Dynamic
		if d == nil || d.Plugin != "com.utmstack.geolocation" {
			continue
		}
		count++
		src, dst := d.Params["source"].GetStringValue(), d.Params["destination"].GetStringValue()
		parts := strings.Split(dst, ".")
		for n := 1; n <= len(parts); n++ {
			prefix := strings.Join(parts[:n], ".")
			if scalars[prefix] || prefix == src {
				t.Errorf("step %d writes %s under %s, which holds a scalar", i, dst, prefix)
			}
		}
		want := src + "Geolocation"
		if src == "origin.ip" || src == "target.ip" {
			want = strings.TrimSuffix(src, ".ip") + ".geolocation"
		} else if !strings.HasPrefix(src, "log.") {
			t.Errorf("step %d: unexpected geolocation source %s", i, src)
		}
		if dst != want {
			t.Errorf("step %d: geolocation of %s goes to %s, want %s", i, src, dst, want)
		}
		if d.Where != `exists("`+src+`")` {
			t.Errorf("step %d: where %q, want exists(%q)", i, d.Where, src)
		}
	}
	if count != 18 {
		t.Errorf("%d geolocation steps, want 18", count)
	}
	// The sibling key survives finalization next to the address.
	draft := `{"id":"x","dataType":"firewall-cisco-firepower","raw":"x","log":{"localIp":"192.0.2.1","localIpGeolocation":{"asn":64501,"country":"Fabricated Country B"}}}`
	event := new(plugins.Event)
	if err := utils.StringToProtoMessage(&draft, event); err != nil {
		t.Fatal(err)
	}
	if event.Log["localIp"].GetStringValue() != "192.0.2.1" ||
		event.Log["localIpGeolocation"].GetStructValue().GetFields()["asn"].GetNumberValue() != 64501 {
		t.Errorf("finalized log: %v", event.Log)
	}
}

// fpModel mirrors the ordered step execution of the public EventProcessor at commit
// 8a3ade72bd9d12db21f6b273200588fb49540f14 (pkg/parsing/parsing.go and plugins/{grok,kv,trim,
// add,rename,cast,delete}/main.go): each where clause is evaluated with the SDK on the whole
// draft, a failing clause is recorded as an error and skips the step, every field name a step
// writes goes through utils.SanitizeField, and every write follows sjson.Set. The geolocation
// steps are not executed here. It is a model used to guard the filter in CI, not a substitute
// for replay.py.
type fpModel struct {
	steps []*plugins.Step
	defs  map[string]string
	cache *plugins.CELCache
	regex map[string]*regexp.Regexp
}

func fpNewModel(t *testing.T) *fpModel {
	t.Helper()
	encoded, err := utils.ReadPbYaml(filepath.Join(fpData, "patterns.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Patterns map[string]string `json:"patterns"`
	}
	if err := json.Unmarshal(encoded, &file); err != nil {
		t.Fatal(err)
	}
	return &fpModel{steps: fpPipeline(t).Steps, defs: file.Patterns,
		cache: plugins.NewCELCache("cisco-firepower-model"), regex: map[string]*regexp.Regexp{}}
}

// compile expands {{.name}} like the SDK regexp cache and compiles the result.
func (m *fpModel) compile(t *testing.T, pattern string) *regexp.Regexp {
	t.Helper()
	if re, ok := m.regex[pattern]; ok {
		return re
	}
	final := pattern
	for i := 0; i < 10 && strings.Contains(final, "{{"); i++ {
		parsed, err := template.New("pattern").Option("missingkey=error").Parse(final)
		if err != nil {
			t.Fatalf("pattern %q: %v", pattern, err)
		}
		var out bytes.Buffer
		if err := parsed.Execute(&out, m.defs); err != nil {
			t.Fatalf("pattern %q: %v", pattern, err)
		}
		if out.String() == final {
			break
		}
		final = out.String()
	}
	re, err := regexp.Compile(final)
	if err != nil {
		t.Fatalf("pattern %q: %v", pattern, err)
	}
	m.regex[pattern] = re
	return re
}

func fpGet(doc map[string]any, path string) (any, bool) {
	var cur any = doc
	for _, part := range strings.Split(path, ".") {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		if cur, ok = obj[part]; !ok {
			return nil, false
		}
	}
	return cur, true
}

// fpSet follows sjson.Set for plain dotted paths: a missing or scalar parent becomes an object.
func fpSet(doc map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	cur := doc
	for _, part := range parts[:len(parts)-1] {
		next, ok := cur[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[part] = next
		}
		cur = next
	}
	cur[parts[len(parts)-1]] = value
}

func fpDelete(doc map[string]any, path string) {
	parts := strings.Split(path, ".")
	cur := doc
	for _, part := range parts[:len(parts)-1] {
		next, ok := cur[part].(map[string]any)
		if !ok {
			return
		}
		cur = next
	}
	delete(cur, parts[len(parts)-1])
}

// fpString follows gjson.Result.String for JSON-decoded values.
func fpString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	case nil:
		return ""
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

func (m *fpModel) run(t *testing.T, raw string) (map[string]any, []string) {
	t.Helper()
	doc := map[string]any{"id": "fixture", "dataType": "firewall-cisco-firepower", "dataSource": "fixture-ftd",
		"@timestamp": "2026-09-23T14:00:00Z", "tenantId": fpTenant, "raw": raw}
	var errs []string
	for i, step := range m.steps {
		if where := fpWhere(step); where != "" {
			draft, _ := json.Marshal(doc)
			ok, err := m.cache.Eval(where, string(draft))
			if err != nil {
				errs = append(errs, err.Error())
			}
			if !ok {
				continue
			}
		}
		var err error
		switch {
		case step.Grok != nil:
			err = m.grok(t, doc, step.Grok)
		case step.Kv != nil:
			err = m.kv(doc, step.Kv)
		case step.Trim != nil:
			err = m.trim(t, doc, step.Trim)
		case step.Add != nil:
			key := step.Add.Params["key"].GetStringValue()
			utils.SanitizeField(&key)
			if err = utils.ValidateReservedField(key, false); err == nil && step.Add.Function == "string" {
				fpSet(doc, key, step.Add.Params["value"].GetStringValue())
			} else if err == nil {
				err = fmt.Errorf("add function %q not modelled", step.Add.Function)
			}
		case step.Rename != nil:
			to := step.Rename.To
			utils.SanitizeField(&to)
			for _, from := range step.Rename.From {
				if v, ok := fpGet(doc, from); ok {
					switch v.(type) { // utils.GetValueOf keeps scalars and turns JSON into text
					case map[string]any, []any:
						v = fpString(v)
					case nil:
						v = ""
					}
					fpSet(doc, to, v)
					fpDelete(doc, from)
				}
			}
		case step.Cast != nil:
			if step.Cast.To != "int" {
				t.Fatalf("step %d: cast to %s not modelled", i, step.Cast.To)
			}
			for _, f := range step.Cast.Fields {
				if v, ok := fpGet(doc, f); ok {
					fpSet(doc, f, float64(utils.CastInt64(v)))
				}
			}
		case step.Delete != nil:
			for _, f := range step.Delete.Fields {
				fpDelete(doc, f)
			}
		case step.Dynamic != nil:
			// geolocation: see TestCiscoFirepowerGeolocationDestinations and replay.py
		default:
			t.Fatalf("step %d: kind not modelled", i)
		}
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	return doc, errs
}

func (m *fpModel) grok(t *testing.T, doc map[string]any, g *plugins.Grok) error {
	source := "raw"
	if g.Source != "" {
		source = g.Source
	}
	v, ok := fpGet(doc, source)
	if !ok {
		return nil
	}
	value := fpString(v)
	type capture struct{ field, value string }
	var store []capture
	size := 0
	for _, p := range g.Patterns {
		value = strings.TrimSpace(value)
		if utf8.RuneCountInString(value) == 0 {
			break
		}
		match := m.compile(t, p.Pattern).FindString(value)
		if match == "" || !strings.HasPrefix(value, match) {
			break
		}
		field := p.FieldName
		utils.SanitizeField(&field)
		if err := utils.ValidateReservedField(field, true); err != nil {
			return err
		}
		size++
		if field != "" {
			store = append(store, capture{field, strings.TrimSpace(match)})
		}
		value = strings.TrimPrefix(value, match)
	}
	if size == len(g.Patterns) {
		for _, c := range store {
			fpSet(doc, c.field, c.value)
		}
	}
	return nil
}

// kv follows plugins/kv/main.go: split on fieldSplit, cut each pair at the first valueSplit,
// sanitize the key and store the trimmed value as text under log.<key>; a later copy of a key
// replaces an earlier one.
func (m *fpModel) kv(doc map[string]any, k *plugins.Kv) error {
	source := "raw"
	if k.Source != "" {
		source = k.Source
	}
	v, ok := fpGet(doc, source)
	if !ok {
		return fmt.Errorf("kv source %s does not exist", source)
	}
	for _, pair := range strings.Split(strings.TrimSpace(fpString(v)), k.FieldSplit) {
		if key, value, found := strings.Cut(pair, k.ValueSplit); found {
			utils.SanitizeField(&key)
			fpSet(doc, "log."+key, strings.TrimSpace(value))
		}
	}
	return nil
}

func (m *fpModel) trim(t *testing.T, doc map[string]any, tr *plugins.Trim) error {
	for _, f := range tr.Fields {
		if err := utils.ValidateReservedField(f, false); err != nil {
			return err
		}
		v, ok := fpGet(doc, f)
		if !ok || fpString(v) == "" {
			continue
		}
		s := strings.TrimSpace(fpString(v))
		switch tr.Function {
		case "prefix":
			s = strings.TrimPrefix(s, tr.Substring)
		case "suffix":
			s = strings.TrimSuffix(s, tr.Substring)
		case "substring":
			s = strings.ReplaceAll(s, tr.Substring, "")
		case "regex":
			found := m.compile(t, tr.Substring).FindAllString(s, -1)
			if len(found) == 0 {
				continue
			}
			for _, x := range found {
				s = strings.ReplaceAll(s, x, "")
			}
		}
		fpSet(doc, f, strings.TrimSpace(s))
	}
	return nil
}

// fpFinalize converts the draft to the SDK Event and back to JSON the way the playground's
// event writer stores it.
func fpFinalize(t *testing.T, doc map[string]any) (*plugins.Event, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(doc)
	draft := string(b)
	event := new(plugins.Event)
	if err := utils.StringToProtoMessage(&draft, event); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var stored map[string]any
	if err := json.Unmarshal(out, &stored); err != nil {
		t.Fatal(err)
	}
	return event, stored
}

// fpFields flattens an event to dotted leaf paths, keeping empty objects as leaves.
func fpFields(event map[string]any) map[string]any {
	out := map[string]any{}
	var walk func(v any, prefix string)
	walk = func(v any, prefix string) {
		if obj, ok := v.(map[string]any); ok && (len(obj) > 0 || prefix == "") {
			for k, x := range obj {
				if prefix == "" && fpEnvelope[k] {
					continue
				}
				p := k
				if prefix != "" {
					p = prefix + "." + k
				}
				walk(x, p)
			}
			return
		}
		out[prefix] = v
	}
	walk(event, "")
	return out
}

var fpGeoPath = regexp.MustCompile(`^(origin\.geolocation|target\.geolocation|log\.[A-Za-z0-9]+Geolocation)(\.|$)`)

type fpCase struct {
	LogObject bool           `json:"logObject"`
	Fields    map[string]any `json:"fields"`
	Alerts    []string       `json:"alerts"`
}

func fpFixtures(t *testing.T) (map[string]string, map[string]fpCase) {
	t.Helper()
	var raw struct {
		Cases map[string]string `json:"cases"`
	}
	var expected struct {
		Cases map[string]fpCase `json:"cases"`
	}
	for name, target := range map[string]any{"raw.json": &raw, "expected.json": &expected} {
		data, err := os.ReadFile(filepath.Join(fpData, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if len(raw.Cases) == 0 || len(raw.Cases) != len(expected.Cases) {
		t.Fatalf("raw fixtures %d, expectations %d", len(raw.Cases), len(expected.Cases))
	}
	return raw.Cases, expected.Cases
}

// fpModelEvents runs every fabricated line through the model and finalizes it.
func fpModelEvents(t *testing.T) (map[string]*plugins.Event, map[string]map[string]any, map[string][]string) {
	t.Helper()
	model := fpNewModel(t)
	raw, _ := fpFixtures(t)
	events, stored, errs := map[string]*plugins.Event{}, map[string]map[string]any{}, map[string][]string{}
	for name, line := range raw {
		doc, e := model.run(t, line)
		events[name], stored[name] = fpFinalize(t, doc)
		errs[name] = e
	}
	return events, stored, errs
}

// Each change, with positive and near-miss lines. A nil value means the path must be absent.
var fpChangeCases = []struct {
	change, fixture, path string
	want                  any
}{
	{"F-H1", "430003-https", "log.messageId", float64(430003)},
	{"F-H1", "430003-https", "log.severity", "1"},
	{"F-H1", "430003-https", "severity", "high"},
	{"F-H1", "430003-space-after-pri", "log.messageId", float64(430003)},
	{"F-H1", "302013-real-header", "log.direction", "inbound"},
	{"F-H1", "302013-real-header", "origin.ip", "198.51.100.7"},
	{"F-H1 near miss", "unparsed-no-pri", "log", nil},
	{"F-H1 near miss", "unparsed-no-space", "log", nil},
	{"F-H1 near miss", "unparsed-asa-prefix", "log", nil},
	{"F-H1 near miss", "unparsed-glued-header", "log", nil},
	{"F-H1 near miss", "unparsed-linux-sshd", "log", nil},
	{"F-H1 older shape", "302013-header-bsd", "log.localIp", "ftd01.example.com"},
	{"F-A3", "302013-header-no-pri", "log.direction", "inbound"},
	{"F-K1", "430003-https", "origin.ip", "192.0.2.10"},
	{"F-K1", "430003-https", "target.ip", "198.51.100.20"},
	{"F-K1", "430003-https", "origin.port", float64(51000)},
	{"F-K1", "430003-https", "target.port", float64(443)},
	{"F-K1", "430003-https", "protocol", "tcp"},
	{"F-K1", "430003-https", "log.SrcIP", nil},
	{"F-K1", "430003-https", "log.DstIP", nil},
	{"F-K1", "430003-https", "log.SrcPort", nil},
	{"F-K1", "430003-https", "log.DstPort", nil},
	{"F-K1", "430003-https", "log.Protocol", nil},
	{"F-K1", "430003-https", "log.PrefilterPolicy", "Example Prefilter Policy"},
	{"F-K1", "430003-https", "log.ApplicationProtocol", "HTTPS"},
	{"F-K1", "430003-https", "log.InitiatorPackets", "6"},
	{"F-K1 underscore", "430003-dns-ttl", "log.DNS_TTL", "300"},
	{"F-K1 underscore", "430003-dns-ttl", "log.DNSTTL", nil},
	{"F-K1", "430003-dns-ttl", "protocol", "udp"},
	{"F-K1", "430003-dns-ttl", "target.port", float64(53)},
	{"F-K1", "430003-useragent", "log.UserAgent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Example/1.0"},
	{"F-K1", "430003-useragent", "log.Client", "Example Client"},
	{"F-K1", "430003-useragent", "log.ftdHead", nil},
	{"F-K1 crafted text", "430003-useragent-spoof", "origin.ip", "192.0.2.10"},
	{"F-K1 crafted text", "430003-useragent-spoof", "target.port", float64(80)},
	{"F-K1 crafted text", "430003-useragent-spoof", "log.AccessControlRuleAction", "Allow"},
	{"F-K1 crafted text", "430003-useragent-spoof", "log.User", "alice"},
	{"F-K1", "430003-icmp", "origin.port", nil},
	{"F-K1", "430003-icmp", "protocol", "icmp"},
	{"F-K1", "430003-ipv6", "target.ip", "2001:db8::20"},
	{"F-K1 near miss", "430003-bad-address", "origin.ip", nil},
	{"F-K1 near miss", "430003-bad-address", "log.SrcIP", "not-an-address"},
	{"F-K1", "430001-p2-user", "log.Priority", "2"},
	{"F-K1", "430001-p2-user", "log.Classification", "Attempted User Privilege Gain"},
	{"F-K1", "430002-block", "log.AccessControlRuleAction", "Block with reset"},
	{"F-K1", "430007-elephant-inside", "log.AccessControlRuleReason", "Elephant Flow"},
	{"F-K1 joined", "glued-text", "log.messageId", float64(430003)},
	{"F-K1 joined", "glued-text", "origin", nil},
	{"F-K1 joined", "glued-text", "log.EventPriority", nil},
	{"F-K1 joined", "glued-text-second-intrusion", "log.Priority", nil},
	{"F-K1 scope", "scope-430005", "log.messageId", float64(430005)},
	{"F-K1 scope", "scope-430005", "log.eventType", nil},
	{"F-K1 scope", "scope-430005", "origin", nil},
	{"F-K1 scope", "scope-430008", "log.EventPriority", nil},
	{"F-A2", "302013-header-device-ipv4", "log.localIp", "192.0.2.1"},
	{"F-A3", "302013-outbound", "log.direction", "outbound"},
	{"F-A3", "302013-header-bsd", "log.direction", "inbound"},
	{"F-A4", "302304-teardown", "protocol", "TCP"},
	{"F-A5", "305011-built", "action", "Built dynamic TCP translation"},
	{"F-A5", "305011-built", "protocol", "TCP"},
	{"F-A5", "305012-teardown", "action", "Teardown dynamic TCP translation"},
	{"F-A6", "302017-gre", "target.user", "erin"},
	{"F-A6", "302017-gre", "log.firewallUserTo", "dave"},
	{"F-A6", "302017-gre", "log.firewallUserFrom", "carol"},
	{"F-A7", "106102-permitted", "actionResult", "accepted"},
	{"F-A7 near miss", "106102-denied", "actionResult", "denied"},
	{"F-A8", "113009-with-equals", "origin.user", "alice"},
	{"F-A8", "113009-with-equals", "log.policy", "DfltGrpPolicy"},
	{"F-A8", "113011-with-equals", "origin.user", "alice"},
	{"F-A8", "113011-with-equals", "log.policy", "GP1"},
	{"F-A8 near miss", "113009-without-equals", "origin.user", "alice"},
	{"F-W2", "109201-uauth", "origin.user", "alice"},
	{"F-W2", "109201-uauth", "log.session", "0x1a2b"},
	{"F-G1", "302003-hostname", "origin.ip", "host-b.example.com"},
	{"F-G1", "302003-hostname", "target.ip", "198.51.100.7"},
	{"F-G1", "302003-hostname", "log.localAddress", "host-b.example.com"},
	{"F-G1 near miss", "302003-ip", "log.localAddress", "192.0.2.10"},
	{"F-G1 near miss", "302004-to", "log.localAddress", "192.0.2.10"},
	{"F-G2", "302024-mapped-no-port", "log.mappedIpFrom", "198.51.100.7"},
	{"F-G2", "302024-mapped-no-port", "log.mappedIpTo", "203.0.113.5"},
	{"F-G2", "302024-mapped-no-port", "log.mappedPortFrom", nil},
	{"F-G2 near miss", "302022-mapped-port", "log.mappedIpFrom", "198.51.100.7"},
	{"F-G2 near miss", "302022-mapped-port", "log.mappedPortFrom", "443"},
}

// Every fabricated line through the model: the named change cases, no where errors, and every
// stored field except geolocation equal to the playground result recorded in expected.json.
func TestCiscoFirepowerExtractionModel(t *testing.T) {
	// The key names below rely on go-sdk v1.1.36 keeping underscores and removing spaces.
	for in, want := range map[string]string{"DNS_TTL": "DNS_TTL", "Prefilter Policy": "PrefilterPolicy"} {
		got := in
		utils.SanitizeField(&got)
		if got != want {
			t.Fatalf("utils.SanitizeField(%q) = %q, want %q", in, got, want)
		}
	}
	_, stored, errs := fpModelEvents(t)
	_, expected := fpFixtures(t)
	for _, c := range fpChangeCases {
		var got any = fpAbsent
		if v, ok := fpGet(stored[c.fixture], c.path); ok {
			got = v
		}
		want := c.want
		if want == nil {
			want = fpAbsent
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s %s: %s = %v, want %v", c.change, c.fixture, c.path, got, want)
		}
	}
	for name, want := range expected {
		if len(errs[name]) > 0 {
			t.Errorf("%s: %d where errors, first: %.200s", name, len(errs[name]), errs[name][0])
		}
		if _, ok := stored[name]["log"]; ok != want.LogObject {
			t.Errorf("%s: log object present=%t", name, ok)
		}
		got := fpFields(stored[name])
		keys := map[string]bool{}
		for k := range got {
			keys[k] = true
		}
		for k := range want.Fields {
			keys[k] = true
		}
		for k := range keys {
			if fpGeoPath.MatchString(k) {
				continue
			}
			g, gok := got[k]
			w, wok := want.Fields[k]
			if gok != wok || !reflect.DeepEqual(g, w) {
				t.Errorf("%s: %s = %v (present %t), want %v (present %t)", name, k, g, gok, w, wok)
			}
		}
	}
}

// The grok plugin trims the remaining text before each pattern and treats an empty match as no
// match, which drops the whole step. So no pattern may prefer empty text at the start of a
// non-empty text: each pattern, expanded as the engine does, is tried alone on texts that start
// with every printable ASCII character and with one non-ASCII letter.
func TestCiscoFirepowerGrokPatternsNeverMatchEmpty(t *testing.T) {
	m := fpNewModel(t)
	probes := []string{"é x"}
	for c := '!'; c <= '~'; c++ {
		probes = append(probes, string(c)+" x")
	}
	checked := 0
	for i, step := range m.steps {
		if step.Grok == nil {
			continue
		}
		for j, p := range step.Grok.Patterns {
			re := m.compile(t, p.Pattern)
			checked++
			for _, probe := range probes {
				if loc := re.FindStringIndex(probe); loc != nil && loc[1] == 0 {
					t.Errorf("step %d pattern %d (%s) %q matches empty text at the start of %q",
						i, j, p.FieldName, p.Pattern, probe)
					break
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no grok pattern checked")
	}
}

// The where clauses of the new, reordered and guarded steps, evaluated with the SDK in filter order.
func TestCiscoFirepowerStepPredicates(t *testing.T) {
	steps := fpPipeline(t).Steps
	cache := plugins.NewCELCache("cisco-firepower-steps")
	eval := func(where, doc string) bool {
		t.Helper()
		ok, err := cache.Eval(where, doc)
		if err != nil {
			t.Fatalf("%q: %v", where, err)
		}
		return ok
	}
	// F-H1: three header patterns read raw; the third runs only while the first two set nothing.
	var headers []*plugins.Grok
	for _, step := range steps {
		if g := step.Grok; g != nil && (g.Source == "" || g.Source == "raw") {
			for _, p := range g.Patterns {
				if p.FieldName == "log.messageId" {
					headers = append(headers, g)
					break
				}
			}
		}
	}
	if len(headers) != 3 || headers[0].Where != "" || headers[1].Where != "" {
		t.Fatalf("F-H1: header groks %d, want two unconditional ones and a third", len(headers))
	}
	if !eval(headers[2].Where, `{"raw":"x"}`) || eval(headers[2].Where, `{"raw":"x","log":{"messageId":"-302013"}}`) {
		t.Errorf("F-H1: third header grok %q must run only when no messageId was set", headers[2].Where)
	}
	// F-K1: the key-value split reads log.msg of the observed IDs only, and never a joined text.
	var split *plugins.Kv
	for _, step := range steps {
		if k := step.Kv; k != nil && k.Source == "log.msg" {
			split = k
		}
	}
	if split == nil || split.FieldSplit != ", " || split.ValueSplit != ": " {
		t.Fatalf("F-K1: key-value split of log.msg: %v", split)
	}
	for id := 430000; id <= 430009; id++ {
		observed := id == 430001 || id == 430002 || id == 430003 || id == 430007
		doc := fmt.Sprintf(`{"raw":"x","log":{"messageId":%d,"msg":"SrcIP: 192.0.2.10, DstIP: 198.51.100.20"}}`, id)
		if eval(split.Where, doc) != observed {
			t.Errorf("F-K1: %d split=%t, want %t", id, !observed, observed)
		}
	}
	for _, doc := range []string{
		`{"raw":"x","log":{"messageId":430003,"msg":"SrcIP: 192.0.2.10, DstIP: 198.51.1<113>%FTD-1-430003: SrcIP: 192.0.2.11"}}`,
		`{"raw":"x","log":{"messageId":430003}}`,
		`{"raw":"x"}`,
	} {
		if eval(split.Where, doc) {
			t.Errorf("F-K1: split must not run on %s", doc)
		}
	}
	// F-A7: the two actionResult adds of 106102/106103, in filter order.
	var adds []*plugins.Add
	for _, step := range steps {
		if a := step.Add; a != nil && a.Params["key"].GetStringValue() == "actionResult" && strings.Contains(a.Where, `"log.messageId", 106102`) {
			adds = append(adds, a)
		}
	}
	if len(adds) != 2 {
		t.Fatalf("106102/106103 actionResult adds: %d", len(adds))
	}
	for _, c := range []struct {
		id              int
		captured, final string
	}{{106102, "permitted", "accepted"}, {106103, "permitted", "accepted"}, {106102, "Permitted", "accepted"},
		{106102, "denied", "denied"}, {106103, "denied", "denied"}} {
		value := c.captured
		for _, a := range adds {
			doc := fmt.Sprintf(`{"raw":"x","log":{"messageId":%d},"actionResult":%q}`, c.id, value)
			if eval(a.Where, doc) {
				value = a.Params["value"].GetStringValue()
			}
		}
		if value != c.final {
			t.Errorf("F-A7: %d with %q ends as %q, want %q", c.id, c.captured, value, c.final)
		}
	}
	// F-A8: the second 113009/113011 variant runs only while origin.user is unset.
	for _, id := range []int{113009, 113011} {
		var writers []string
		for _, step := range steps {
			g := step.Grok
			if g == nil || !strings.Contains(fpRawForm(g.Where), fmt.Sprintf("log.messageId==%d", id)) {
				continue
			}
			for _, p := range g.Patterns {
				if p.FieldName == "origin.user" {
					writers = append(writers, g.Where)
				}
			}
		}
		if len(writers) != 2 {
			t.Fatalf("F-A8: %d origin.user writers for %d, want 2", len(writers), id)
		}
		without := fmt.Sprintf(`{"raw":"x","log":{"messageId":%d}}`, id)
		with := fmt.Sprintf(`{"raw":"x","log":{"messageId":%d},"origin":{"user":"alice"}}`, id)
		if !eval(writers[0], without) || !eval(writers[1], without) {
			t.Errorf("F-A8: %d variants must run while no user is set", id)
		}
		if eval(writers[1], with) {
			t.Errorf("F-A8: %d second variant %q runs after the first set origin.user", id, writers[1])
		}
		if eval(writers[1], fmt.Sprintf(`{"raw":"x","log":{"messageId":%d},"origin":{"user":"alice"}}`, id+1)) {
			t.Errorf("F-A8: %d second variant matches another message", id)
		}
	}
	// F-W2: the 109201-109213 steps call declared functions and cover exactly that range.
	var uauth []string
	for _, step := range steps {
		if w := fpWhere(step); strings.Contains(w, "109201") {
			uauth = append(uauth, w)
		}
	}
	if len(uauth) != 3 {
		t.Fatalf("F-W2: %d steps for 109201-109213, want 3", len(uauth))
	}
	for _, w := range uauth {
		for id, want := range map[int]bool{109200: false, 109201: true, 109207: true, 109213: true, 109214: false} {
			if eval(w, fmt.Sprintf(`{"raw":"x","log":{"messageId":%d}}`, id)) != want {
				t.Errorf("F-W2: %q on %d, want %t", w, id, want)
			}
		}
	}
}

func fpLoadRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(fpRulesDir, "*.y*ml"))
	if err != nil || len(files) != 5 {
		t.Fatalf("Cisco Firepower rules: %d files, error %v", len(files), err)
	}
	out := map[string]*plugins.Rule{}
	for _, path := range files {
		encoded, err := utils.ReadPbYaml(path)
		if err != nil {
			t.Fatal(err)
		}
		rule := new(plugins.Rule)
		if err := protojson.Unmarshal(encoded, rule); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		rule.Normalize()
		out[strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))] = rule
	}
	return out
}

func fpSearches(searches []*plugins.SearchRequest) string {
	var parts []string
	for _, s := range searches {
		var with []string
		for _, e := range s.With {
			with = append(with, e.Field+" "+e.Operator+" "+e.Value.GetStringValue())
		}
		p := fmt.Sprintf("%s[%s] within %s count %d", s.IndexPattern, strings.Join(with, "; "), s.Within, s.Count)
		if len(s.Or) > 0 {
			p += " or(" + fpSearches(s.Or) + ")"
		}
		parts = append(parts, p)
	}
	return strings.Join(parts, " | ")
}

func fpPlaceholders(searches []*plugins.SearchRequest, out map[string]bool) {
	for _, s := range searches {
		for _, e := range s.With {
			if v := e.Value.GetStringValue(); strings.HasPrefix(v, "{{.") && strings.HasSuffix(v, "}}") {
				out[strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")] = true
			}
		}
		fpPlaceholders(s.Or, out)
	}
}

// Names, metadata, impact, grouping, MITRE labels and the history search stay as they were;
// only the two conditions change, and they no longer read key names the filter never writes.
func TestCiscoFirepowerRuleContract(t *testing.T) {
	const docs = "https://www.cisco.com/c/en/us/td/docs/security/"
	want := map[string]string{
		"advanced_malware_protection_alerts": "Advanced Malware Protection (AMP) Alert Detection|Initial Access|T1566 - Phishing|origin|3/3/2|lastEvent.log.sha256,adversary.ip|" +
			docs + "firepower/70/configuration/guide/fpmc-config-guide-v70/file_malware_events_and_network_file_trajectory.html,https://attack.mitre.org/techniques/T1566/|",
		fpC2: "Command and Control on Non-Standard Ports|Command and Control|T1571 - Non-Standard Port|origin|3/2/1|adversary.ip,target.ip,target.port|" +
			docs + "secure-firewall/management-center/device-config/710/management-center-device-config-71/connection-log-fields.html,https://attack.mitre.org/techniques/T1571/|" +
			"v11-log-firewall-cisco-firepower-*[origin.ip filter_term {{.origin.ip}}; target.ip filter_term {{.target.ip}}] within 1h count 5",
		fpIPS: "Intrusion Prevention System High Priority Events|Execution|T1203 - Exploitation for Client Execution|origin|3/3/3|adversary.ip,target.ip|" +
			docs + "secure-firewall/management-center/device-config/710/management-center-device-config-71/intrusion-overview.html,https://attack.mitre.org/techniques/T1203/|",
		"ioc_matches": "Firepower IOC (Indicator of Compromise) Detection|Initial Access|T1566 - Phishing|origin|3/3/2|adversary.ip|" +
			docs + "firepower/70/configuration/guide/fpmc-config-guide-v70/file_malware_events_and_network_file_trajectory.html,https://attack.mitre.org/tactics/TA0040/,https://attack.mitre.org/techniques/T1566/|",
		"threat_intelligence_director_alerts": "Threat Intelligence Director (TID) Alert Detection|Command and Control|T1071.001 - Application Layer Protocol: Web Protocols|origin|3/3/2|lastEvent.log.tidIndicator,adversary.ip|" +
			docs + "firepower/70/configuration/guide/fpmc-config-guide-v70/tid_overview.html,https://attack.mitre.org/techniques/T1071/|",
	}
	rules := fpLoadRules(t)
	for stem, rule := range rules {
		got := fmt.Sprintf("%s|%s|%s|%s|%d/%d/%d|%s|%s|%s", rule.Name, rule.Category, rule.Technique, rule.Adversary,
			rule.Impact.Confidentiality, rule.Impact.Integrity, rule.Impact.Availability, strings.Join(rule.GroupBy, ","),
			strings.Join(rule.References, ","), fpSearches(rule.Correlation))
		if got != want[stem] {
			t.Errorf("%s:\n got %s\nwant %s", stem, got, want[stem])
		}
		if len(rule.DataTypes) != 1 || rule.DataTypes[0] != "firewall-cisco-firepower" || len(rule.DeduplicateBy) != 0 {
			t.Errorf("%s: dataTypes %v deduplicateBy %v", stem, rule.DataTypes, rule.DeduplicateBy)
		}
	}
	for stem, gone := range map[string][]string{
		fpIPS: {`"log.eventType"`, `"log.priority"`, `"log.severity"`, `"log.impact"`, `"log.classification"`},
		fpC2:  {`"log.appProto"`, `"log.initiatorPackets"`, `unknown-tcp`},
	} {
		for _, name := range gone {
			if strings.Contains(rules[stem].Where, name) {
				t.Errorf("%s still reads %s", stem, name)
			}
		}
	}
}

func fpEvent(t *testing.T, body string) *plugins.Event {
	t.Helper()
	input := `{"dataType":"firewall-cisco-firepower","dataSource":"fixture-ftd","tenantId":"` + fpTenant + `",` + body + `}`
	event := new(plugins.Event)
	if err := utils.StringToProtoMessage(&input, event); err != nil {
		t.Fatalf("%s: %v", body, err)
	}
	return event
}

// Synthetic normalized events shaped as the filter stores them: key-value fields are text,
// ports are numbers.
func fpConnection(app, dst string, port int, packets string) string {
	return fmt.Sprintf(`"log":{"messageId":430003,"ApplicationProtocol":%q,"InitiatorPackets":%q},"origin":{"ip":"192.0.2.10","port":51000},"target":{"ip":%q,"port":%d},"protocol":"tcp"`,
		app, packets, dst, port)
}

func fpIntrusion(priority, classification string) string {
	return fmt.Sprintf(`"log":{"messageId":430001,"Priority":%q,"Classification":%q,"severity":"1"},"origin":{"ip":"198.51.100.7","port":51000},"target":{"ip":"192.0.2.20","port":80},"protocol":"tcp"`,
		priority, classification)
}

var fpRuleCases = []struct {
	rule, name, body string
	want             bool
}{
	{fpIPS, "priority 1", fpIntrusion("1", "Misc Attack"), true},
	{fpIPS, "priority 2, Attempted User Privilege Gain", fpIntrusion("2", "Attempted User Privilege Gain"), true},
	{fpIPS, "priority 2, Attempted Administrator Privilege Gain", fpIntrusion("2", "Attempted Administrator Privilege Gain"), true},
	{fpIPS, "priority 2, Web Application Attack", fpIntrusion("2", "Web Application Attack"), true},
	{fpIPS, "priority 3, Exploit Kit Activity Detected", fpIntrusion("3", "Exploit Kit Activity Detected"), true},
	{fpIPS, "priority 2, Misc Attack", fpIntrusion("2", "Misc Attack"), false},
	{fpIPS, "priority 3 at syslog level 1", fpIntrusion("3", "Potential Corporate Privacy Violation"), false},
	{fpIPS, "short class name attempted-user", fpIntrusion("2", "attempted-user"), false},
	{fpIPS, "connection event with Priority 1", `"log":{"messageId":430003,"Priority":"1","EventPriority":"High"},"origin":{"ip":"192.0.2.10"}`, false},
	{fpIPS, "old key names only", `"log":{"messageId":430001,"eventType":"IPS_EVENT","priority":1,"impact":"HIGH"},"origin":{"ip":"198.51.100.7"}`, false},
	{fpC2, "HTTP to an outside address on 8081", fpConnection("HTTP", "203.0.113.10", 8081, "6"), true},
	{fpC2, "SSL to an outside address on 9001", fpConnection("SSL", "198.51.100.30", 9001, "6"), true},
	{fpC2, "HTTPS to an outside address on 4443", fpConnection("HTTPS", "203.0.113.11", 4443, "6"), true},
	{fpC2, "HTTP to a global IPv6 address on 8081", fpConnection("HTTP", "2001:db8::20", 8081, "1"), true},
	{fpC2, "HTTP on listed port 8080", fpConnection("HTTP", "203.0.113.10", 8080, "6"), false},
	{fpC2, "HTTPS on 443", fpConnection("HTTPS", "203.0.113.10", 443, "6"), false},
	{fpC2, "SSL on listed port 993", fpConnection("SSL", "203.0.113.10", 993, "6"), false},
	{fpC2, "HTTP to 10.0.0.0/8", fpConnection("HTTP", "10.0.0.20", 8181, "6"), false},
	{fpC2, "SSL to 172.16.0.0/12", fpConnection("SSL", "172.31.255.254", 9001, "6"), false},
	{fpC2, "HTTP to 192.168.0.0/16", fpConnection("HTTP", "192.168.0.20", 8181, "6"), false},
	{fpC2, "HTTP to fc00::/7", fpConnection("HTTP", "fd00::20", 8081, "6"), false},
	{fpC2, "Unknown application", fpConnection("Unknown", "203.0.113.13", 4444, "6"), false},
	{fpC2, "no initiator packets", fpConnection("HTTP", "203.0.113.10", 8081, "0"), false},
	{fpC2, "no packet counter", `"log":{"messageId":430001,"ApplicationProtocol":"HTTP"},"origin":{"ip":"192.0.2.10"},"target":{"ip":"203.0.113.10","port":8081}`, false},
	{fpC2, "no destination port", `"log":{"messageId":430003,"ApplicationProtocol":"HTTP","InitiatorPackets":"6"},"origin":{"ip":"192.0.2.10"},"target":{"ip":"203.0.113.15"}`, false},
	{fpC2, "no source address", `"log":{"messageId":430003,"ApplicationProtocol":"HTTP","InitiatorPackets":"6","SrcIP":"not-an-address"},"target":{"ip":"203.0.113.10","port":8081}`, false},
	{fpC2, "old key names only", `"log":{"messageId":430003,"appProto":"HTTP","initiatorPackets":true},"origin":{"ip":"192.0.2.10"},"target":{"ip":"203.0.113.10","port":8081}`, false},
}

// go-sdk CEL on synthetic normalized events for the two changed conditions, and on the model
// output of every fabricated line against the alerts the playground raised (expected.json).
func TestCiscoFirepowerRulePredicates(t *testing.T) {
	rules := fpLoadRules(t)
	cache := plugins.NewCELCache("cisco-firepower-rules")
	for _, c := range fpRuleCases {
		t.Run(c.rule+"/"+c.name, func(t *testing.T) {
			got, err := cache.Eval(rules[c.rule].Where, fpEvent(t, c.body))
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("match=%t want=%t", got, c.want)
			}
		})
	}
	events, _, _ := fpModelEvents(t)
	_, expected := fpFixtures(t)
	stems := make([]string, 0, len(rules))
	for stem := range rules {
		stems = append(stems, stem)
	}
	sort.Strings(stems)
	fired := map[string]int{}
	for name, event := range events {
		var got []string
		for _, stem := range stems {
			match, err := cache.Eval(rules[stem].Where, event)
			if err != nil {
				t.Fatalf("%s on %s: %v", stem, name, err)
			}
			if match {
				got = append(got, stem)
				fired[stem]++
			}
		}
		if want := expected[name].Alerts; !reflect.DeepEqual(got, want) && !(len(got) == 0 && len(want) == 0) {
			t.Errorf("%s: rules %v, playground alerts %v", name, got, want)
		}
	}
	if fired[fpIPS] == 0 || fired[fpC2] == 0 {
		t.Errorf("fabricated lines must include positives of both changed rules: %v", fired)
	}
}

// Whenever a rule with a history search matches, every {{.field}} placeholder resolves; an
// unresolved one fails the search, and five failures disable the rule with a Circuit Breaker
// alert. Checked on the model output of every fabricated line and on the synthetic events.
func TestCiscoFirepowerHistoryPlaceholders(t *testing.T) {
	rules := fpLoadRules(t)
	cache := plugins.NewCELCache("cisco-firepower-history")
	events, _, _ := fpModelEvents(t)
	for _, c := range fpRuleCases {
		events["synthetic: "+c.name] = fpEvent(t, c.body)
	}
	names := make([]string, 0, len(events))
	for name := range events {
		names = append(names, name)
	}
	sort.Strings(names)
	checked := 0
	for stem, rule := range rules {
		fields := map[string]bool{}
		fpPlaceholders(rule.Correlation, fields)
		if len(fields) == 0 {
			continue
		}
		for _, name := range names {
			match, err := cache.Eval(rule.Where, events[name])
			if err != nil {
				t.Fatalf("%s on %s: %v", stem, name, err)
			}
			if !match {
				continue
			}
			checked++
			doc, err := utils.ProtoMessageToString(events[name])
			if err != nil {
				t.Fatal(err)
			}
			for field := range fields {
				if gjson.Get(*doc, field).Value() == nil {
					t.Errorf("%s matches %s without %s; its history search would fail", stem, name, field)
				}
			}
		}
	}
	if checked == 0 {
		t.Error("no positive case reached a history search")
	}
}
