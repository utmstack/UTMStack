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

// Cisco ASA regression checks. Every raw input is FABRICATED (testdata/cisco-asa): Cisco's
// documentation site refused automated access and no Cisco ASA record was available, so the
// lines follow the filter's own patterns, use RFC 5737 and RFC 3849 documentation addresses
// and example names, and none is claimed to be a documented Cisco format. These tests use
// the pinned go-sdk (v1.1.36) for YAML decoding, CEL and Event conversion. They do not run the
// EventProcessor: asaModel mirrors its step plugins, and testdata/cisco-asa/replay.py runs
// the same lines through the public playground. See filters/audits/cisco-asa.md.

const (
	asaFilter   = "../../filters/cisco/asa.yml"
	asaRulesDir = "../../rules/cisco/asa"
	asaData     = "testdata/cisco-asa"
	asaTenant   = "00000000-0000-4000-8000-000000000001"
	asaAbsent   = "<absent>"
)

var asaEnvelope = map[string]bool{"id": true, "timestamp": true, "deviceTime": true, "dataType": true,
	"dataSource": true, "tenantId": true, "tenantName": true, "raw": true, "errors": true}

func asaPipeline(t *testing.T) *plugins.Pipeline {
	t.Helper()
	encoded, err := utils.ReadPbYaml(asaFilter)
	if err != nil {
		t.Fatal(err)
	}
	config := new(plugins.Config)
	if err := protojson.Unmarshal(encoded, config); err != nil {
		t.Fatal(err)
	}
	if len(config.Pipeline) != 1 || len(config.Pipeline[0].DataTypes) != 1 ||
		config.Pipeline[0].DataTypes[0] != "firewall-cisco-asa" {
		t.Fatalf("unexpected pipeline layout: %v", config.Pipeline)
	}
	return config.Pipeline[0]
}

// asaWhere returns the where clause of whichever step kind is set.
func asaWhere(step *plugins.Step) string {
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

// Before this revision, 519 clauses compared log.* directly (for example log.messageId==106001).
// These rewrite a helper clause back to that form, so both can be compared.
var (
	asaRawLog      = regexp.MustCompile(`(^|[^"\w.])log\.[A-Za-z0-9_.]+\s*(==|!=|>=|<=|<|>)`)
	asaHelperToRaw = []struct {
		helper *regexp.Regexp
		raw    string
	}{
		{regexp.MustCompile(`equals\("log\.messageId", (\d+)\)`), `log.messageId==$1`},
		{regexp.MustCompile(`greaterOrEqual\("log\.messageId", (\d+)\)`), `log.messageId>=$1`},
		{regexp.MustCompile(`lessOrEqual\("log\.messageId", (\d+)\)`), `log.messageId<=$1`},
		{regexp.MustCompile(`equals\("log\.severity", "4"\)`), `log.severity=="4"`},
	}
	asaHelperCall = regexp.MustCompile(`(\w+)\("([A-Za-z0-9_.]+)"(?:, ("[^"]*"|\d+))?\)`)
)

func asaRawForm(where string) string {
	for _, r := range asaHelperToRaw {
		where = r.helper.ReplaceAllString(where, r.raw)
	}
	return where
}

// asaTruthTable builds events that all carry a log object: log.messageId as the JSON number
// the filter's cast produces (around every literal in the clause), log.severity as the
// one-digit levels 0 to 7, and every other field the clause reads absent, equal to its
// literal, in other case, containing it, or unrelated. Severity text such as 04 or +4 is left
// out on purpose: equals compares it as the number 4, the old clause did not (see the audit).
func asaTruthTable(where string) []string {
	ids := map[float64]bool{100000: true}
	severities := []string{"6"}
	others := map[string][]any{}
	for _, m := range asaHelperCall.FindAllStringSubmatch(where, -1) {
		field, arg := m[2], m[3]
		switch field {
		case "log.messageId":
			n, _ := strconv.Atoi(arg)
			for d := -1; d <= 1; d++ {
				ids[float64(n+d)] = true
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
				doc := map[string]any{"id": "x", "dataType": "firewall-cisco-asa", "raw": "x",
					"log": map[string]any{"messageId": id, "severity": sev}}
				for f, v := range c {
					if v != nil {
						asaSet(doc, f, v)
					}
				}
				b, _ := json.Marshal(doc)
				docs = append(docs, string(b))
			}
		}
	}
	return docs
}

// No where clause compares log.* directly. Every clause that uses equals, greaterOrEqual or
// lessOrEqual on log.messageId, or equals on log.severity, compiles, has the same truth table
// as the direct comparison on events with a log object, and is false without an error when no
// header pattern produced a log object.
func TestCiscoASAWhereHelpers(t *testing.T) {
	steps := asaPipeline(t).Steps
	cache := plugins.NewCELCache("cisco-asa-where")
	noLog := `{"id":"x","dataType":"firewall-cisco-asa","dataSource":"fixture-asa","tenantId":"` + asaTenant + `","raw":"x"}`
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
	checked, rows := 0, 0
	for i, step := range steps {
		where := asaWhere(step)
		if where == "" {
			continue
		}
		if asaRawLog.MatchString(where) {
			raw = append(raw, where)
			continue
		}
		pre := asaRawForm(where)
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
		for _, doc := range asaTruthTable(where) {
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
	if checked < 520 {
		t.Errorf("%d helper clauses read log.messageId or log.severity, want at least 520", checked)
	}
	t.Logf("%d helper clauses, %d truth-table rows", checked, rows)
}

// The geolocation plugin writes its result at the destination path, so a destination below
// an address field would replace the address with an object.
func TestCiscoASAGeolocationDestinations(t *testing.T) {
	steps := asaPipeline(t).Steps
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
	draft := `{"id":"x","dataType":"firewall-cisco-asa","raw":"x","log":{"localIp":"192.0.2.1","localIpGeolocation":{"asn":64501,"country":"Fabricated Country B"}}}`
	event := new(plugins.Event)
	if err := utils.StringToProtoMessage(&draft, event); err != nil {
		t.Fatal(err)
	}
	if event.Log["localIp"].GetStringValue() != "192.0.2.1" ||
		event.Log["localIpGeolocation"].GetStructValue().GetFields()["asn"].GetNumberValue() != 64501 {
		t.Errorf("finalized log: %v", event.Log)
	}
}

// asaModel mirrors the ordered step execution of the public EventProcessor at commit
// 497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1 (pkg/parsing/parsing.go and plugins/{grok,trim,
// add,rename,cast,delete}/main.go): each where clause is evaluated with the SDK on the whole
// draft, a failing clause is recorded as an error and skips the step, and every write follows
// sjson.Set. The geolocation steps are not executed here. It is a model used to guard the
// filter in CI, not a substitute for replay.py.
type asaModel struct {
	steps []*plugins.Step
	defs  map[string]string
	cache *plugins.CELCache
	regex map[string]*regexp.Regexp
}

func asaNewModel(t *testing.T) *asaModel {
	t.Helper()
	encoded, err := utils.ReadPbYaml(filepath.Join(asaData, "patterns.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Patterns map[string]string `json:"patterns"`
	}
	if err := json.Unmarshal(encoded, &file); err != nil {
		t.Fatal(err)
	}
	return &asaModel{steps: asaPipeline(t).Steps, defs: file.Patterns,
		cache: plugins.NewCELCache("cisco-asa-model"), regex: map[string]*regexp.Regexp{}}
}

// compile expands {{.name}} like the SDK regexp cache and compiles the result.
func (m *asaModel) compile(t *testing.T, pattern string) *regexp.Regexp {
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

func asaGet(doc map[string]any, path string) (any, bool) {
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

// asaSet follows sjson.Set for plain dotted paths: a missing or scalar parent becomes an object.
func asaSet(doc map[string]any, path string, value any) {
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

func asaDelete(doc map[string]any, path string) {
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

// asaString follows gjson.Result.String for JSON-decoded values.
func asaString(v any) string {
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

func (m *asaModel) run(t *testing.T, raw string) (map[string]any, []string) {
	t.Helper()
	doc := map[string]any{"id": "fixture", "dataType": "firewall-cisco-asa", "dataSource": "fixture-asa",
		"@timestamp": "2026-09-23T14:00:00Z", "tenantId": asaTenant, "raw": raw}
	var errs []string
	for i, step := range m.steps {
		if where := asaWhere(step); where != "" {
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
		case step.Trim != nil:
			err = m.trim(t, doc, step.Trim)
		case step.Add != nil:
			key := step.Add.Params["key"].GetStringValue()
			utils.SanitizeField(&key)
			if err = utils.ValidateReservedField(key, false); err == nil && step.Add.Function == "string" {
				asaSet(doc, key, step.Add.Params["value"].GetStringValue())
			} else if err == nil {
				err = fmt.Errorf("add function %q not modelled", step.Add.Function)
			}
		case step.Rename != nil:
			to := step.Rename.To
			utils.SanitizeField(&to)
			for _, from := range step.Rename.From {
				if v, ok := asaGet(doc, from); ok {
					switch v.(type) { // utils.GetValueOf keeps scalars and turns JSON into text
					case map[string]any, []any:
						v = asaString(v)
					case nil:
						v = ""
					}
					asaSet(doc, to, v)
					asaDelete(doc, from)
				}
			}
		case step.Cast != nil:
			if step.Cast.To != "int" {
				t.Fatalf("step %d: cast to %s not modelled", i, step.Cast.To)
			}
			for _, f := range step.Cast.Fields {
				if v, ok := asaGet(doc, f); ok {
					asaSet(doc, f, float64(utils.CastInt64(v)))
				}
			}
		case step.Delete != nil:
			for _, f := range step.Delete.Fields {
				asaDelete(doc, f)
			}
		case step.Dynamic != nil:
			// geolocation: see TestCiscoASAGeolocationDestinations and replay.py
		default:
			t.Fatalf("step %d: kind not modelled", i)
		}
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	return doc, errs
}

func (m *asaModel) grok(t *testing.T, doc map[string]any, g *plugins.Grok) error {
	source := "raw"
	if g.Source != "" {
		source = g.Source
	}
	v, ok := asaGet(doc, source)
	if !ok {
		return nil
	}
	value := asaString(v)
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
			asaSet(doc, c.field, c.value)
		}
	}
	return nil
}

func (m *asaModel) trim(t *testing.T, doc map[string]any, tr *plugins.Trim) error {
	for _, f := range tr.Fields {
		if err := utils.ValidateReservedField(f, false); err != nil {
			return err
		}
		v, ok := asaGet(doc, f)
		if !ok || asaString(v) == "" {
			continue
		}
		s := strings.TrimSpace(asaString(v))
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
		asaSet(doc, f, strings.TrimSpace(s))
	}
	return nil
}

// asaFinalize converts the draft to the SDK Event and back to JSON the way the playground's
// event writer stores it.
func asaFinalize(t *testing.T, doc map[string]any) (*plugins.Event, map[string]any) {
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

// asaFields flattens an event to dotted leaf paths, keeping empty objects as leaves.
func asaFields(event map[string]any) map[string]any {
	out := map[string]any{}
	var walk func(v any, prefix string)
	walk = func(v any, prefix string) {
		if obj, ok := v.(map[string]any); ok && (len(obj) > 0 || prefix == "") {
			for k, x := range obj {
				if prefix == "" && asaEnvelope[k] {
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

var asaGeoPath = regexp.MustCompile(`^(origin\.geolocation|target\.geolocation|log\.[A-Za-z0-9]+Geolocation)(\.|$)`)

type asaCase struct {
	LogObject bool           `json:"logObject"`
	Fields    map[string]any `json:"fields"`
	Alerts    []string       `json:"alerts"`
}

func asaFixtures(t *testing.T) (map[string]string, map[string]asaCase) {
	t.Helper()
	var raw struct {
		Cases map[string]string `json:"cases"`
	}
	var expected struct {
		Cases map[string]asaCase `json:"cases"`
	}
	for name, target := range map[string]any{"raw.json": &raw, "expected.json": &expected} {
		data, err := os.ReadFile(filepath.Join(asaData, name))
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

// asaModelEvents runs every fabricated line through the model and finalizes it.
func asaModelEvents(t *testing.T) (map[string]*plugins.Event, map[string]map[string]any, map[string][]string) {
	t.Helper()
	model := asaNewModel(t)
	raw, _ := asaFixtures(t)
	events, stored, errs := map[string]*plugins.Event{}, map[string]map[string]any{}, map[string][]string{}
	for name, line := range raw {
		doc, e := model.run(t, line)
		events[name], stored[name] = asaFinalize(t, doc)
		errs[name] = e
	}
	return events, stored, errs
}

// Each change, with positive and near-miss lines. A nil value means the path must be absent.
var asaChangeCases = []struct {
	change, fixture, path string
	want                  any
}{
	{"F-C1", "unparsed-no-timestamp", "log", nil},
	{"F-C1", "unparsed-rfc5424", "log", nil},
	{"F-C1", "unparsed-linux-sshd", "log", nil},
	{"F-C1", "botnet-338001", "severity", "medium"},
	{"F-C3", "302013-outbound", "log.direction", "outbound"},
	{"F-C3", "header-bsd", "log.direction", "inbound"},
	{"F-C3 near miss", "302013-probe", "log.direction", "inbound"},
	{"F-C3 near miss", "302015-outbound", "log.direction", "outbound"},
	{"F-C4", "302304-teardown", "protocol", "TCP"},
	{"F-C4 near miss", "302303-built", "protocol", "TCP"},
	{"F-C5", "305011-built", "action", "Built dynamic TCP translation"},
	{"F-C5", "305011-built", "protocol", "TCP"},
	{"F-C5", "305012-teardown", "action", "Teardown dynamic TCP translation"},
	{"F-C5 near miss", "305012-one-digit-hour", "action", nil},
	{"F-C6", "302017-gre", "target.user", "erin"},
	{"F-C6", "302017-gre", "log.firewallUserTo", "dave"},
	{"F-C6", "302017-gre", "log.firewallUserFrom", "carol"},
	{"F-C6 near miss", "302018-gre", "origin.user", "erin"},
	{"F-C7", "106102-permitted", "actionResult", "accepted"},
	{"F-C7", "106102-permitted-arrow", "actionResult", "accepted"},
	{"F-C7", "106103-permitted", "actionResult", "accepted"},
	{"F-C7 near miss", "106102-denied", "actionResult", "denied"},
	{"F-C7 near miss", "106102-denied-arrow", "actionResult", "denied"},
	{"F-C8", "113009-with-equals", "origin.user", "alice"},
	{"F-C8", "113009-with-equals", "log.policy", "DfltGrpPolicy"},
	{"F-C8", "113011-with-equals", "origin.user", "alice"},
	{"F-C8", "113011-with-equals", "log.policy", "GP1"},
	{"F-C8 near miss", "113009-without-equals", "origin.user", "alice"},
	{"F-C9", "302003-hostname", "origin.ip", "host-b.example.com"},
	{"F-C9", "302003-hostname", "target.ip", "198.51.100.7"},
	{"F-C9", "302003-hostname", "log.localAddress", "host-b.example.com"},
	{"F-C9 near miss", "302003-ip", "log.localAddress", "192.0.2.10"},
	{"F-C9 near miss", "302004-to", "log.localAddress", "192.0.2.10"},
	{"F-C10", "302024-mapped-no-port", "log.mappedIpFrom", "198.51.100.7"},
	{"F-C10", "302024-mapped-no-port", "log.mappedIpTo", "203.0.113.5"},
	{"F-C10", "302024-mapped-no-port", "log.mappedPortFrom", nil},
	{"F-C10 near miss", "302022-mapped-port", "log.mappedIpFrom", "198.51.100.7"},
	{"F-C10 near miss", "302022-mapped-port", "log.mappedPortFrom", "443"},
}

// Every fabricated line through the model: the named change cases, no where errors, and every
// stored field except geolocation equal to the playground result recorded in expected.json.
func TestCiscoASAExtractionModel(t *testing.T) {
	_, stored, errs := asaModelEvents(t)
	_, expected := asaFixtures(t)
	for _, c := range asaChangeCases {
		var got any = asaAbsent
		if v, ok := asaGet(stored[c.fixture], c.path); ok {
			got = v
		}
		want := c.want
		if want == nil {
			want = asaAbsent
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
		got := asaFields(stored[name])
		keys := map[string]bool{}
		for k := range got {
			keys[k] = true
		}
		for k := range want.Fields {
			keys[k] = true
		}
		for k := range keys {
			if asaGeoPath.MatchString(k) {
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
func TestCiscoASAGrokPatternsNeverMatchEmpty(t *testing.T) {
	m := asaNewModel(t)
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

// The where clauses of the reordered and guarded steps, evaluated with the SDK in filter order.
func TestCiscoASAStepPredicates(t *testing.T) {
	steps := asaPipeline(t).Steps
	cache := plugins.NewCELCache("cisco-asa-steps")
	eval := func(where, doc string) bool {
		t.Helper()
		ok, err := cache.Eval(where, doc)
		if err != nil {
			t.Fatalf("%q: %v", where, err)
		}
		return ok
	}
	// F-C7: the two actionResult adds of 106102/106103, in filter order.
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
			t.Errorf("F-C7: %d with %q ends as %q, want %q", c.id, c.captured, value, c.final)
		}
	}
	// F-C8: the second 113009/113011 variant runs only while origin.user is unset.
	for _, id := range []int{113009, 113011} {
		var writers []string
		for _, step := range steps {
			g := step.Grok
			if g == nil || !strings.Contains(asaRawForm(g.Where), fmt.Sprintf("log.messageId==%d", id)) {
				continue
			}
			for _, p := range g.Patterns {
				if p.FieldName == "origin.user" {
					writers = append(writers, g.Where)
				}
			}
		}
		if len(writers) != 2 {
			t.Fatalf("F-C8: %d origin.user writers for %d, want 2", len(writers), id)
		}
		without := fmt.Sprintf(`{"raw":"x","log":{"messageId":%d}}`, id)
		with := fmt.Sprintf(`{"raw":"x","log":{"messageId":%d},"origin":{"user":"alice"}}`, id)
		if !eval(writers[0], without) || !eval(writers[1], without) {
			t.Errorf("F-C8: %d variants must run while no user is set", id)
		}
		if eval(writers[1], with) {
			t.Errorf("F-C8: %d second variant %q runs after the first set origin.user", id, writers[1])
		}
		if eval(writers[1], fmt.Sprintf(`{"raw":"x","log":{"messageId":%d},"origin":{"user":"alice"}}`, id+1)) {
			t.Errorf("F-C8: %d second variant matches another message", id)
		}
	}
}

func asaLoadRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(asaRulesDir, "*.y*ml"))
	if err != nil || len(files) != 3 {
		t.Fatalf("Cisco ASA rules: %d files, error %v", len(files), err)
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

func asaSearches(searches []*plugins.SearchRequest) string {
	var parts []string
	for _, s := range searches {
		var with []string
		for _, e := range s.With {
			with = append(with, e.Field+" "+e.Operator+" "+e.Value.GetStringValue())
		}
		p := fmt.Sprintf("%s[%s] within %s count %d", s.IndexPattern, strings.Join(with, "; "), s.Within, s.Count)
		if len(s.Or) > 0 {
			p += " or(" + asaSearches(s.Or) + ")"
		}
		parts = append(parts, p)
	}
	return strings.Join(parts, " | ")
}

func asaPlaceholders(searches []*plugins.SearchRequest, out map[string]bool) {
	for _, s := range searches {
		for _, e := range s.With {
			if v := e.Value.GetStringValue(); strings.HasPrefix(v, "{{.") && strings.HasSuffix(v, "}}") {
				out[strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")] = true
			}
		}
		asaPlaceholders(s.Or, out)
	}
}

// Names, metadata, impact, grouping and history searches stay as they were; only the two
// conditions change.
func TestCiscoASARuleContract(t *testing.T) {
	const index = "v11-log-firewall-cisco-asa-*"
	vpnBranch := func(id string) string {
		return index + "[origin.ip filter_term {{.origin.ip}}; log.messageId filter_term " + id + "] within 15m count 10"
	}
	want := map[string]string{
		"botnet_traffic_detection": "Botnet Command and Control Traffic Detected|Command and Control|T1071 - Application Layer Protocol|origin|3/2/1|adversary.ip,target.ip|" +
			"https://www.cisco.com/c/en/us/td/docs/security/asa/special/botnet/asa-botnet.pdf,https://attack.mitre.org/techniques/T1071/|",
		"ips_signature_matches": "IPS Signature Match - Malicious Pattern Detected|Initial Access|T1190 - Exploit Public-Facing Application|origin|3/3/2|adversary.ip,target.ip|" +
			"https://www.cisco.com/c/en/us/td/docs/security/asa/syslog/b_syslog.html,https://attack.mitre.org/techniques/T1190/|" +
			index + "[origin.ip filter_term {{.origin.ip}}] within 15m count 3",
		"multiple_failed_vpn_attempts": "Multiple Failed VPN Authentication Attempts|Credential Access|T1110 - Brute Force|origin|3/2/1|adversary.ip,adversary.user|" +
			"https://attack.mitre.org/techniques/T1110/,https://www.cisco.com/c/en/us/td/docs/security/asa/syslog/b_syslog/syslogs1.html|" +
			vpnBranch("113015") + " or(" + vpnBranch("113021") + " | " + vpnBranch("109034") + " | " + vpnBranch("611102") + ")",
	}
	rules := asaLoadRules(t)
	for stem, rule := range rules {
		got := fmt.Sprintf("%s|%s|%s|%s|%d/%d/%d|%s|%s|%s", rule.Name, rule.Category, rule.Technique, rule.Adversary,
			rule.Impact.Confidentiality, rule.Impact.Integrity, rule.Impact.Availability, strings.Join(rule.GroupBy, ","),
			strings.Join(rule.References, ","), asaSearches(rule.Correlation))
		if got != want[stem] {
			t.Errorf("%s:\n got %s\nwant %s", stem, got, want[stem])
		}
		if len(rule.DataTypes) != 1 || rule.DataTypes[0] != "firewall-cisco-asa" || len(rule.DeduplicateBy) != 0 {
			t.Errorf("%s: dataTypes %v deduplicateBy %v", stem, rule.DataTypes, rule.DeduplicateBy)
		}
	}
	ips := strings.TrimSpace(rules["ips_signature_matches"].Where)
	if !strings.HasPrefix(ips, `exists("origin.ip") && (`) || !strings.HasSuffix(ips, ")") || strings.Contains(ips, "log.action") {
		t.Errorf("IPS condition must be exists(\"origin.ip\") && (...) without log.action: %s", ips)
	}
	vpn := rules["multiple_failed_vpn_attempts"].Where
	if !strings.Contains(vpn, `regexMatch("log.msg", `) || strings.Contains(vpn, `"log.message"`) {
		t.Errorf("VPN condition must read log.msg: %s", vpn)
	}
}

func asaEvent(t *testing.T, body string) *plugins.Event {
	t.Helper()
	input := `{"dataType":"firewall-cisco-asa","dataSource":"fixture-asa","tenantId":"` + asaTenant + `",` + body + `}`
	event := new(plugins.Event)
	if err := utils.StringToProtoMessage(&input, event); err != nil {
		t.Fatalf("%s: %v", body, err)
	}
	return event
}

var asaRuleCases = []struct {
	rule, name, body string
	want             bool
}{
	{"ips_signature_matches", "108003 with a source address (a future mapping, D02)", `"log":{"messageId":108003,"msg":"Terminating ESMTP/SMTP connection; malicious pattern detected"},"origin":{"ip":"198.51.100.7"}`, true},
	{"ips_signature_matches", "108003 as a text id with a source address", `"log":{"messageId":"108003"},"origin":{"ip":"198.51.100.7"}`, true},
	{"ips_signature_matches", "108003 as this filter stores it, no source address", `"log":{"messageId":108003,"msg":"Terminating ESMTP/SMTP connection; malicious pattern detected"}`, false},
	{"ips_signature_matches", "log.action value that no step writes", `"log":{"messageId":420997,"action":"ips_alert"},"origin":{"ip":"198.51.100.7"}`, false},
	{"ips_signature_matches", "unrelated connection with an address", `"log":{"messageId":302013},"origin":{"ip":"198.51.100.7"}`, false},
	{"multiple_failed_vpn_attempts", "113015 with address and failing reason", `"log":{"messageId":113015,"reason":"Invalid password"},"origin":{"ip":"198.51.100.7","user":"alice"}`, true},
	{"multiple_failed_vpn_attempts", "109034 failure text in log.msg", `"log":{"messageId":109034,"msg":"Authentication failed for network user alice from 198.51.100.7/51234 to 192.0.2.10/443"},"origin":{"ip":"198.51.100.7"}`, true},
	{"multiple_failed_vpn_attempts", "611102 failure text in log.msg", `"log":{"messageId":611102,"msg":"User authentication failed: IP address: 198.51.100.7, Uname: alice"},"origin":{"ip":"198.51.100.7"}`, true},
	{"multiple_failed_vpn_attempts", "109034 failure text only in log.message", `"log":{"messageId":109034,"message":"Authentication failed for network user alice"},"origin":{"ip":"198.51.100.7"}`, false},
	{"multiple_failed_vpn_attempts", "113015 as this filter stores it, address in target.ip (D01)", `"log":{"messageId":113015,"reason":"Invalid password"},"target":{"ip":"198.51.100.7"}`, false},
	{"multiple_failed_vpn_attempts", "failure text on an unlisted message", `"log":{"messageId":113005,"msg":"authentication failed"},"origin":{"ip":"198.51.100.7"}`, false},
	{"multiple_failed_vpn_attempts", "109034 success text", `"log":{"messageId":109034,"msg":"Authentication succeeded for network user alice"},"origin":{"ip":"198.51.100.7"}`, false},
	{"botnet_traffic_detection", "338001 by message id", `"log":{"messageId":338001}`, true},
	{"botnet_traffic_detection", "unlisted 338003", `"log":{"messageId":338003}`, false},
}

// SDK CEL (v1.1.36) on synthetic normalized events for the changed rule conditions.
func TestCiscoASARulePredicates(t *testing.T) {
	rules := asaLoadRules(t)
	cache := plugins.NewCELCache("cisco-asa-rules")
	for _, c := range asaRuleCases {
		t.Run(c.rule+"/"+c.name, func(t *testing.T) {
			got, err := cache.Eval(rules[c.rule].Where, asaEvent(t, c.body))
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("match=%t want=%t", got, c.want)
			}
		})
	}
}

// Whenever a rule with a history search matches, every {{.field}} placeholder resolves; an
// unresolved one fails the search, and five failures disable the rule with a Circuit Breaker
// alert. Checked on the model output of every fabricated line and on the synthetic events.
func TestCiscoASAHistoryPlaceholders(t *testing.T) {
	rules := asaLoadRules(t)
	cache := plugins.NewCELCache("cisco-asa-history")
	events, _, _ := asaModelEvents(t)
	for _, c := range asaRuleCases {
		events["synthetic: "+c.name] = asaEvent(t, c.body)
	}
	names := make([]string, 0, len(events))
	for name := range events {
		names = append(names, name)
	}
	sort.Strings(names)
	checked := 0
	for stem, rule := range rules {
		fields := map[string]bool{}
		asaPlaceholders(rule.Correlation, fields)
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
