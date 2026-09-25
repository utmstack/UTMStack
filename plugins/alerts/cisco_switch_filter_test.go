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

// Cisco switch regression checks. Every raw input is FABRICATED (testdata/cisco-switch): Cisco's
// documentation site refused automated access, so the lines copy the text shapes the filter's
// steps read, with invented values: MAC addresses in the locally administered range
// 02:00:00:xx:xx:xx written in Cisco's dotted form, RFC 5737 and RFC 3849 documentation
// addresses, and example host, user and interface names. None is claimed to be a documented
// Cisco format. These tests use the pinned go-sdk v1.1.33 for YAML decoding, CEL and Event
// conversion. They do not run the EventProcessor: cswModel mirrors its step plugins, and
// testdata/cisco-switch/replay.py runs the same lines through the public playground. See
// filters/audits/cisco-switch.md.

const (
	cswFilter   = "../../filters/cisco/cs_switch.yml"
	cswRulesDir = "../../rules/cisco/cs_switch"
	cswData     = "testdata/cisco-switch"
	cswTenant   = "00000000-0000-4000-8000-000000000001"
	cswAbsent   = "<absent>"
)

var cswEnvelope = map[string]bool{"id": true, "timestamp": true, "deviceTime": true, "dataType": true,
	"dataSource": true, "tenantId": true, "tenantName": true, "raw": true, "errors": true}

func cswPipeline(t *testing.T) *plugins.Pipeline {
	t.Helper()
	encoded, err := utils.ReadPbYaml(cswFilter)
	if err != nil {
		t.Fatal(err)
	}
	config := new(plugins.Config)
	if err := protojson.Unmarshal(encoded, config); err != nil {
		t.Fatal(err)
	}
	if len(config.Pipeline) != 1 || len(config.Pipeline[0].DataTypes) != 1 ||
		config.Pipeline[0].DataTypes[0] != "cisco-switch" {
		t.Fatalf("unexpected pipeline layout: %v", config.Pipeline)
	}
	return config.Pipeline[0]
}

// cswWhere returns the where clause of whichever step kind is set.
func cswWhere(step *plugins.Step) string {
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

// Before this revision the 'medium' severity step read log.severity=="4".
var (
	cswRawLog      = regexp.MustCompile(`(^|[^"\w.])log\.[A-Za-z0-9_.]+\s*(==|!=|>=|<=|<|>)`)
	cswOldMedium   = `log.severity=="4"`
	cswNoLog       = `{"id":"x","dataType":"cisco-switch","dataSource":"fixture-switch","tenantId":"` + cswTenant + `","raw":"x"}`
	cswNoSeverity  = `{"id":"x","dataType":"cisco-switch","dataSource":"fixture-switch","tenantId":"` + cswTenant + `","raw":"x","log":{"msg":"on /var"}}`
	cswParsedDraft = `{"id":"x","dataType":"cisco-switch","dataSource":"fixture-switch","tenantId":"` + cswTenant + `","raw":"x",` +
		`"log":{"msg":"SW_MATM-4-MACFLAP_NOTIF: Host 0200.0000.0101 in vlan 910 is flapping between port Gi9/0/41 and port Gi9/0/42",` +
		`"facility":"SW_MATM","severity":"4","facilityMnemonic":"MACFLAP_NOTIF",` +
		`"ciscoMsg":"Host 0200.0000.0101 in vlan 910 is flapping between port Gi9/0/41 and port Gi9/0/42"}}`
)

// No where clause compares a log.* field directly. Such a clause fails, and the engine stores the
// error on the event, when the draft has no log object (a line without any '%') or no
// log.severity (a '%' but no FACILITY-SEVERITY-MNEMONIC header). Every clause evaluates without
// an error on both kinds of draft and on a parsed one.
func TestCiscoSwitchWhereClauses(t *testing.T) {
	steps := cswPipeline(t).Steps
	cache := plugins.NewCELCache("cisco-switch-where")
	drafts := []struct{ name, doc string }{
		{"no log object", cswNoLog}, {"log without severity", cswNoSeverity}, {"parsed flap", cswParsedDraft},
	}
	var raw []string
	clauses := 0
	for i, step := range steps {
		where := cswWhere(step)
		if where == "" {
			continue
		}
		clauses++
		if cswRawLog.MatchString(where) {
			raw = append(raw, where)
		}
		for _, d := range drafts {
			if _, err := cache.Eval(where, d.doc); err != nil {
				t.Errorf("step %d on a draft with %s: %q: %.200v", i, d.name, where, err)
			}
		}
	}
	if len(raw) > 0 {
		t.Errorf("%d where clauses compare log.* directly and fail without a log object, for example %q", len(raw), raw[0])
	}
	if clauses < 23 {
		t.Errorf("%d where clauses, want at least 23", clauses)
	}
	// The raw form fails on exactly those drafts, which is what stored the errors.
	for _, d := range drafts[:2] {
		if _, err := cache.Eval(cswOldMedium, d.doc); err == nil {
			t.Errorf("%s on a draft with %s: expected an error", cswOldMedium, d.name)
		}
	}
}

func cswSeveritySteps(t *testing.T) []*plugins.Add {
	t.Helper()
	var adds []*plugins.Add
	for _, step := range cswPipeline(t).Steps {
		if a := step.Add; a != nil && a.Params["key"].GetStringValue() == "severity" {
			adds = append(adds, a)
		}
	}
	if len(adds) != 3 {
		t.Fatalf("%d severity steps, want 3", len(adds))
	}
	return adds
}

// cswSeverity applies the severity steps in filter order; a later match overwrites an earlier one.
func cswSeverity(t *testing.T, cache *plugins.CELCache, adds []*plugins.Add, doc string, medium string) string {
	t.Helper()
	result := cswAbsent
	for _, a := range adds {
		where := a.Where
		if a.Params["value"].GetStringValue() == "medium" && medium != "" {
			where = medium
		}
		ok, err := cache.Eval(where, doc)
		if err != nil {
			return "error"
		}
		if ok {
			result = a.Params["value"].GetStringValue()
		}
	}
	return result
}

// The three severity steps give the same severity as before on every one-digit level and on text
// that is not a number, and no severity and no error without a level. equals and oneOf compare
// numbers, so a level written 04 or +4 now counts as 4, as 03 already counted as 3 and 05 as 5;
// before this revision such a level 4 got no severity. No sampled record has such a level.
func TestCiscoSwitchSeverityClauses(t *testing.T) {
	adds := cswSeveritySteps(t)
	if got := adds[1].Where; got != `equals("log.severity", "4")` {
		t.Fatalf("medium step: %q", got)
	}
	cache := plugins.NewCELCache("cisco-switch-severity")
	doc := func(level string) string {
		return fmt.Sprintf(`{"raw":"x","log":{"facility":"FAC","severity":%q,"facilityMnemonic":"MNEM"}}`, level)
	}
	want := map[string]string{"0": "high", "1": "high", "2": "high", "3": "high", "4": "medium", "5": "low",
		"6": "low", "7": "low", "8": cswAbsent, "SP": cswAbsent, "DFC4": cswAbsent, "": cswAbsent, "44": cswAbsent, "4 ": cswAbsent}
	for level, w := range want {
		got := cswSeverity(t, cache, adds, doc(level), "")
		old := cswSeverity(t, cache, adds, doc(level), cswOldMedium)
		if got != w || old != w {
			t.Errorf("level %q: severity %s, with the old clause %s, want %s", level, got, old, w)
		}
	}
	for level, w := range map[string]string{"03": "high", "04": "medium", "+4": "medium", "05": "low"} {
		if got := cswSeverity(t, cache, adds, doc(level), ""); got != w {
			t.Errorf("level %q: severity %s, want %s", level, got, w)
		}
	}
	if old := cswSeverity(t, cache, adds, doc("04"), cswOldMedium); old != cswAbsent {
		t.Errorf("level \"04\" with the old clause: %s, want no severity", old)
	}
	for _, d := range []string{cswNoLog, cswNoSeverity} {
		if got := cswSeverity(t, cache, adds, d, ""); got != cswAbsent {
			t.Errorf("%s: severity %s, want none and no error", d, got)
		}
		if old := cswSeverity(t, cache, adds, d, cswOldMedium); old != "error" {
			t.Errorf("%s: the old clause gave %s, want an error", d, old)
		}
	}
}

// cswModel mirrors the ordered step execution of the public EventProcessor at commit
// 497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1 (pkg/parsing/parsing.go and plugins/{grok,trim,add,
// cast,delete}/main.go): each where clause is evaluated with the SDK on the whole draft, a failing
// clause is recorded as an error and skips the step, a grok writes nothing unless every pattern
// matched at the start of the remaining text, and every write follows sjson.Set. It is a model
// used to guard the filter in CI, not a substitute for replay.py.
type cswModel struct {
	steps []*plugins.Step
	defs  map[string]string
	cache *plugins.CELCache
	regex map[string]*regexp.Regexp
}

func cswNewModel(t *testing.T) *cswModel {
	t.Helper()
	encoded, err := utils.ReadPbYaml(filepath.Join(cswData, "patterns.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Patterns map[string]string `json:"patterns"`
	}
	if err := json.Unmarshal(encoded, &file); err != nil {
		t.Fatal(err)
	}
	return &cswModel{steps: cswPipeline(t).Steps, defs: file.Patterns,
		cache: plugins.NewCELCache("cisco-switch-model"), regex: map[string]*regexp.Regexp{}}
}

// compile expands {{.name}} like the SDK regexp cache and compiles the result.
func (m *cswModel) compile(t *testing.T, pattern string) *regexp.Regexp {
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

func cswGet(doc map[string]any, path string) (any, bool) {
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

// cswSet follows sjson.Set for plain dotted paths: a missing or scalar parent becomes an object.
func cswSet(doc map[string]any, path string, value any) {
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

func cswDelete(doc map[string]any, path string) {
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

// cswString follows gjson.Result.String for JSON-decoded values.
func cswString(v any) string {
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

func (m *cswModel) run(t *testing.T, raw string) (map[string]any, []string) {
	t.Helper()
	doc := map[string]any{"id": "fixture", "dataType": "cisco-switch", "dataSource": "fixture-switch",
		"@timestamp": "2026-09-24T14:00:00Z", "tenantId": cswTenant, "raw": raw}
	var errs []string
	for i, step := range m.steps {
		if where := cswWhere(step); where != "" {
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
				cswSet(doc, key, step.Add.Params["value"].GetStringValue())
			} else if err == nil {
				err = fmt.Errorf("add function %q not modelled", step.Add.Function)
			}
		case step.Cast != nil:
			if step.Cast.To != "int" {
				t.Fatalf("step %d: cast to %s not modelled", i, step.Cast.To)
			}
			for _, f := range step.Cast.Fields {
				if v, ok := cswGet(doc, f); ok {
					cswSet(doc, f, float64(utils.CastInt64(v)))
				}
			}
		case step.Delete != nil:
			for _, f := range step.Delete.Fields {
				cswDelete(doc, f)
			}
		default:
			t.Fatalf("step %d: kind not modelled", i)
		}
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	return doc, errs
}

func (m *cswModel) grok(t *testing.T, doc map[string]any, g *plugins.Grok) error {
	source := "raw"
	if g.Source != "" {
		source = g.Source
	}
	v, ok := cswGet(doc, source)
	if !ok {
		return nil
	}
	value := cswString(v)
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
			cswSet(doc, c.field, c.value)
		}
	}
	return nil
}

func (m *cswModel) trim(t *testing.T, doc map[string]any, tr *plugins.Trim) error {
	for _, f := range tr.Fields {
		if err := utils.ValidateReservedField(f, false); err != nil {
			return err
		}
		v, ok := cswGet(doc, f)
		if !ok || cswString(v) == "" {
			continue
		}
		s := strings.TrimSpace(cswString(v))
		switch tr.Function {
		case "prefix":
			s = strings.TrimPrefix(s, tr.Substring)
		case "suffix":
			s = strings.TrimSuffix(s, tr.Substring)
		case "substring":
			s = strings.ReplaceAll(s, tr.Substring, "")
		default:
			t.Fatalf("trim function %s not modelled", tr.Function)
		}
		cswSet(doc, f, strings.TrimSpace(s))
	}
	return nil
}

// cswFinalize converts the draft to the SDK Event and back to JSON the way the playground's event
// writer stores it.
func cswFinalize(t *testing.T, doc map[string]any) (*plugins.Event, map[string]any) {
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

// cswFields flattens an event to dotted leaf paths, keeping empty objects as leaves.
func cswFields(event map[string]any) map[string]any {
	out := map[string]any{}
	var walk func(v any, prefix string)
	walk = func(v any, prefix string) {
		if obj, ok := v.(map[string]any); ok && (len(obj) > 0 || prefix == "") {
			for k, x := range obj {
				if prefix == "" && cswEnvelope[k] {
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

type cswCase struct {
	LogObject bool           `json:"logObject"`
	Fields    map[string]any `json:"fields"`
	Alerts    []string       `json:"alerts"`
}

func cswFixtures(t *testing.T) (map[string]string, map[string]cswCase) {
	t.Helper()
	var raw struct {
		Cases map[string]string `json:"cases"`
	}
	var expected struct {
		Cases map[string]cswCase `json:"cases"`
	}
	for name, target := range map[string]any{"raw.json": &raw, "expected.json": &expected} {
		data, err := os.ReadFile(filepath.Join(cswData, name))
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

// cswModelEvents runs every fabricated line through the model and finalizes it.
func cswModelEvents(t *testing.T) (map[string]*plugins.Event, map[string]map[string]any, map[string][]string) {
	t.Helper()
	model := cswNewModel(t)
	raw, _ := cswFixtures(t)
	events, stored, errs := map[string]*plugins.Event{}, map[string]map[string]any{}, map[string][]string{}
	for name, line := range raw {
		doc, e := model.run(t, line)
		events[name], stored[name] = cswFinalize(t, doc)
		errs[name] = e
	}
	return events, stored, errs
}

// Each change, with positive and near-miss lines. A nil value means the path must be absent.
var cswChangeCases = []struct {
	change, fixture, path string
	want                  any
}{
	{"F-1", "unparsed-no-percent", "log", nil},
	{"F-1", "unparsed-no-percent", "severity", nil},
	{"F-1", "unparsed-percent-without-header", "log.msg", "on /var"},
	{"F-1", "unparsed-percent-without-header", "log.severity", nil},
	{"F-1", "unparsed-percent-without-header", "severity", nil},
	{"F-1", "header-seq-ms", "severity", "medium"},
	{"F-1 control", "severity-3-link", "severity", "high"},
	{"F-1 control", "severity-5-lineproto", "severity", "low"},
	{"F-1 control", "severity-5-subfacility", "log.subFacility", "SP"},
	{"F-1 control", "severity-5-subfacility", "severity", "low"},
	{"F-1 control", "misrouted-firepower-shape", "log.facility", "FTD"},
	{"F-1 control", "misrouted-firepower-shape", "severity", "high"},
	{"F-2", "header-seq-ms", "origin.mac", "0200.0000.0101"},
	{"F-2", "header-seq-ms", "log.vlan", "910"},
	{"F-2", "header-seq-ms", "log.firstPort", "Gi9/0/41"},
	{"F-2", "header-seq-ms", "log.secondPort", "Gi9/0/42"},
	{"F-2", "header-star-year", "origin.mac", "0200.0000.0102"},
	{"F-2", "header-star-year", "log.firstPort", "Te9/1/2"},
	{"F-2", "header-star-year", "log.secondPort", "Te9/1/1"},
	{"F-2", "header-dot", "origin.mac", "0200.0000.0103"},
	{"F-2", "flap-upper-hex-port-channel", "origin.mac", "0200.00AB.CD01"},
	{"F-2", "flap-upper-hex-port-channel", "log.vlan", "930"},
	{"F-2", "flap-upper-hex-port-channel", "log.firstPort", "Po9"},
	{"F-2", "flap-upper-hex-port-channel", "log.secondPort", "Gi9/0/43"},
	{"F-2", "flap-trailing-text", "log.secondPort", "Gi9/0/42"},
	{"F-2", "flap-slot-branch", "origin.mac", "0200.0000.0105"},
	{"F-2", "flap-slot-branch", "log.slot", "SLOT3"},
	{"F-2 near miss", "flap-colon-mac", "origin.mac", nil},
	{"F-2 near miss", "flap-colon-mac", "log.vlan", nil},
	{"F-2 near miss", "flap-colon-mac", "log.firstPort", nil},
	{"F-2 near miss", "flap-thirteen-hex", "origin.mac", nil},
	{"F-2 near miss", "flap-no-vlan", "origin.mac", nil},
	{"F-2 near miss", "flap-truncated", "origin.mac", nil},
	{"F-2 near miss", "flap-truncated", "log.firstPort", nil},
	{"F-2 near miss", "flap-under-sw-vlan", "origin.mac", nil},
	{"F-3", "sisf-excess-arp", "origin.mac", "0200.0000.0201"},
	{"F-3", "sisf-no-prefix", "origin.mac", "0200.0000.0202"},
	{"F-3 near miss", "sisf-colon-mac", "origin.mac", nil},
	{"F-4", "ssh2-unexpected", "origin.ip", "198.51.100.21"},
	{"F-4", "ssh-close", "origin.ip", "198.51.100.23"},
	{"F-4 near miss", "ssh2-unexpected-ipv6", "origin.ip", nil},
	{"F-4 near miss", "ssh2-unexpected-bad-octet", "origin.ip", nil},
	{"F-4 near miss", "ssh2-unexpected-trailing", "origin.ip", nil},
	{"F-5", "dhcpd-ping-conflict", "target.ip", "192.0.2.31"},
	{"F-5 near miss", "dhcpd-no-period", "target.ip", nil},
	{"F-5 near miss", "dhcpd-bad-octet", "target.ip", nil},
	{"F-6", "logginghost-fail", "target.ip", "192.0.2.41"},
	{"F-6", "logginghost-fail", "target.port", float64(514)},
	{"F-6", "logginghost-started", "target.ip", "192.0.2.42"},
	{"F-6", "logginghost-started", "target.port", float64(6514)},
	{"F-6 near miss", "logginghost-stopped", "target.ip", nil},
	{"F-6 near miss", "logginghost-host-name", "target.ip", nil},
	{"F-6 near miss", "logginghost-port-text", "target.ip", nil},
	{"F-6 near miss", "logginghost-port-text", "target.port", nil},
	{"F-6 near miss", "logginghost-port-eleven-digits", "target.port", nil},
	{"unchanged", "dai-invalid-arp", "actionResult", "blocked"},
	{"unchanged", "dai-invalid-arp", "origin.mac", nil},
	{"unchanged", "dai-invalid-arp", "origin.ip", nil},
	{"unchanged", "dai-dhcp-snooping-deny", "actionResult", "blocked"},
	{"unchanged", "ip-dupaddr", "origin.ip", nil},
	{"unchanged", "mac-duplicate-text", "origin.mac", nil},
}

// Every fabricated line through the model: the named change cases, no where errors, and every
// stored field equal to the playground result recorded in expected.json.
func TestCiscoSwitchExtractionModel(t *testing.T) {
	_, stored, errs := cswModelEvents(t)
	_, expected := cswFixtures(t)
	for _, c := range cswChangeCases {
		var got any = cswAbsent
		if v, ok := cswGet(stored[c.fixture], c.path); ok {
			got = v
		}
		want := c.want
		if want == nil {
			want = cswAbsent
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
		got := cswFields(stored[name])
		keys := map[string]bool{}
		for k := range got {
			keys[k] = true
		}
		for k := range want.Fields {
			keys[k] = true
		}
		for k := range keys {
			g, gok := got[k]
			w, wok := want.Fields[k]
			if gok != wok || !reflect.DeepEqual(g, w) {
				t.Errorf("%s: %s = %v (present %t), want %v (present %t)", name, k, g, gok, w, wok)
			}
		}
	}
}

// Interface names are text. origin.port and target.port are whole numbers (uint32) in the SDK
// Event, and an interface name there fails the conversion of the whole event, so the flap
// interfaces go to log.firstPort and log.secondPort. Only the logging-host step writes a port,
// with a digits-only pattern, and a cast to int under the same condition follows it.
func TestCiscoSwitchPortFields(t *testing.T) {
	steps := cswPipeline(t).Steps
	writers := 0
	for i, step := range steps {
		g := step.Grok
		if g == nil {
			continue
		}
		for _, p := range g.Patterns {
			switch p.FieldName {
			case "origin.port":
				t.Errorf("step %d writes origin.port", i)
			case "target.port":
				writers++
				if p.Pattern != "[0-9]{1,5}" {
					t.Errorf("step %d writes target.port from %q, want digits only", i, p.Pattern)
				}
				next := &plugins.Cast{}
				if i+1 < len(steps) && steps[i+1].Cast != nil {
					next = steps[i+1].Cast
				}
				if next.To != "int" || len(next.Fields) != 1 || next.Fields[0] != "target.port" ||
					next.Where != g.Where+` && exists("target.port")` {
					t.Errorf("step %d: target.port is not cast to int under the same condition: %v", i, next)
				}
			case "log.firstPort", "log.secondPort":
				if p.Pattern != "{{.notSpace}}" {
					t.Errorf("step %d: %s from %q", i, p.FieldName, p.Pattern)
				}
			}
		}
	}
	if writers != 1 {
		t.Errorf("%d steps write target.port, want 1", writers)
	}
	_, stored, _ := cswModelEvents(t)
	flaps := 0
	for name, e := range stored {
		if v, ok := cswGet(e, "origin.port"); ok {
			t.Errorf("%s: origin.port = %v", name, v)
		}
		if v, ok := cswGet(e, "target.port"); ok {
			if _, number := v.(float64); !number {
				t.Errorf("%s: target.port = %v (%T), want a number", name, v, v)
			}
		}
		if _, ok := cswGet(e, "log.firstPort"); ok {
			flaps++
			if _, ok := cswGet(e, "log.secondPort"); !ok {
				t.Errorf("%s: log.firstPort without log.secondPort", name)
			}
		}
	}
	if flaps != 6 {
		t.Errorf("%d fabricated lines carry the flap interfaces, want 6", flaps)
	}
	for _, bad := range []string{`{"origin":{"port":"Gi9/0/41"}}`, `{"target":{"port":"Po9"}}`} {
		draft := `{"dataType":"cisco-switch","raw":"x",` + strings.TrimPrefix(bad, "{")
		if err := utils.StringToProtoMessage(&draft, new(plugins.Event)); err == nil {
			t.Errorf("%s converted without an error", bad)
		}
	}
}

func cswLoadRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(cswRulesDir, "*.y*ml"))
	if err != nil || len(files) != 3 {
		t.Fatalf("Cisco switch rules: %d files, error %v", len(files), err)
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

func cswSearches(searches []*plugins.SearchRequest) string {
	var parts []string
	for _, s := range searches {
		var with []string
		for _, e := range s.With {
			with = append(with, e.Field+" "+e.Operator+" "+e.Value.GetStringValue())
		}
		p := fmt.Sprintf("%s[%s] within %s count %d", s.IndexPattern, strings.Join(with, "; "), s.Within, s.Count)
		if len(s.Or) > 0 {
			p += " or(" + cswSearches(s.Or) + ")"
		}
		parts = append(parts, p)
	}
	return strings.Join(parts, " | ")
}

func cswPlaceholders(searches []*plugins.SearchRequest, out map[string]bool) {
	for _, s := range searches {
		for _, e := range s.With {
			if v := e.Value.GetStringValue(); strings.HasPrefix(v, "{{.") && strings.HasSuffix(v, "}}") {
				out[strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}")] = true
			}
		}
		cswPlaceholders(s.Or, out)
	}
}

// The VLAN hopping condition of v11 4a000bc4, which this revision leaves as it is.
const cswVlanWhere = `(equals("log.facility", "SW_VLAN") && oneOf("log.facilityMnemonic", ["VLAN_INCONSISTENCY", "MACFLAP_NOTIF", "TRUNK_MODE_CHANGE"]))
|| (equals("log.facility", "DTP") && oneOf("log.facilityMnemonic", ["NONTRUNKPORTON", "DOMAINMISMATCH", "TRUNKPORTON"]))
|| regexMatch("log.message", "(?i)(received 802.1Q BPDU on non trunk|native vlan mismatch|inconsistent vlan|double tag)")
|| (lessOrEqual("log.severity", 4) && regexMatch("log.message", "(?i)(vlan.*tag.*tag|switch.*spoofing|dtp.*negotiation)"))`

// Names, metadata, impact, adversary side and history searches stay as they were. The MAC rule
// deduplicates by adversary.mac instead of grouping by it; the VLAN rule is unchanged, condition
// included.
func TestCiscoSwitchRuleContract(t *testing.T) {
	const index = "v11-log-cisco-switch-*"
	const cisco3750 = "https://www.cisco.com/c/en/us/support/docs/switches/catalyst-3750-series-switches/72846-layer2-secftrs-catl3fixed.html"
	want := map[string]string{
		"arp_poisoning_detection": "ARP Poisoning Attack Detection|Credential Access, Collection|T1557.002 - Adversary-in-the-Middle: ARP Cache Poisoning|origin|3/3/2|" +
			"groupBy=adversary.ip,adversary.mac|deduplicateBy=|" +
			"https://www.cisco.com/c/en/us/td/docs/switches/lan/catalyst4500/12-2/25ew/configuration/guide/conf/dynarp.html,https://attack.mitre.org/techniques/T1557/002/|" +
			index + "[origin.ip filter_term {{.origin.ip}}] within 10m count 5",
		"mac_address_spoofing": "MAC Address Spoofing Detection|Initial Access|MAC Spoofing|origin|2/3/1|" +
			"groupBy=|deduplicateBy=adversary.mac|" + cisco3750 + ",https://attack.mitre.org/techniques/T1200/|" +
			index + "[origin.mac filter_term {{.origin.mac}}] within 10m count 3",
		"vlan_hopping_attempts": "VLAN Hopping Attack Detection|Defense Evasion|T1599 - Network Boundary Bridging|origin|3/3/2|" +
			"groupBy=adversary.ip,adversary.mac|deduplicateBy=|" + cisco3750 + ",https://attack.mitre.org/techniques/T1599/|",
	}
	rules := cswLoadRules(t)
	for stem, rule := range rules {
		got := fmt.Sprintf("%s|%s|%s|%s|%d/%d/%d|groupBy=%s|deduplicateBy=%s|%s|%s", rule.Name, rule.Category, rule.Technique,
			rule.Adversary, rule.Impact.Confidentiality, rule.Impact.Integrity, rule.Impact.Availability,
			strings.Join(rule.GroupBy, ","), strings.Join(rule.DeduplicateBy, ","), strings.Join(rule.References, ","),
			cswSearches(rule.Correlation))
		if got != want[stem] {
			t.Errorf("%s:\n got %s\nwant %s", stem, got, want[stem])
		}
		if len(rule.DataTypes) != 1 || rule.DataTypes[0] != "cisco-switch" {
			t.Errorf("%s: dataTypes %v", stem, rule.DataTypes)
		}
	}
	if got := strings.TrimSpace(rules["vlan_hopping_attempts"].Where); got != cswVlanWhere {
		t.Errorf("VLAN hopping condition changed:\n%s", got)
	}
	mac := strings.TrimSpace(rules["mac_address_spoofing"].Where)
	if !strings.HasPrefix(mac, `exists("origin.mac") && !regexMatch("log.msg", "(?i)(mac.*flap|is flapping between port)") && (`) ||
		!strings.HasSuffix(mac, ")") || strings.Contains(mac, "log.message") || strings.Contains(mac, "SW_MATM") {
		t.Errorf("MAC condition must require origin.mac, leave out flaps and read log.msg: %s", mac)
	}
	arp := strings.TrimSpace(rules["arp_poisoning_detection"].Where)
	if !strings.HasPrefix(arp, `exists("origin.ip") && (`) || !strings.HasSuffix(arp, ")") || strings.Contains(arp, "log.message") {
		t.Errorf("ARP condition must require origin.ip and read log.msg: %s", arp)
	}
}

func cswEvent(t *testing.T, body string) *plugins.Event {
	t.Helper()
	input := `{"dataType":"cisco-switch","dataSource":"fixture-switch","tenantId":"` + cswTenant + `",` + body + `}`
	event := new(plugins.Event)
	if err := utils.StringToProtoMessage(&input, event); err != nil {
		t.Fatalf("%s: %v", body, err)
	}
	return event
}

const (
	cswFlapText = "Host 0200.0000.0101 in vlan 910 is flapping between port Gi9/0/41 and port Gi9/0/42"
	cswFlapBody = `"log":{"facility":"SW_MATM","facilityMnemonic":"MACFLAP_NOTIF","severity":"4",` +
		`"msg":"SW_MATM-4-MACFLAP_NOTIF: ` + cswFlapText + `","ciscoMsg":"` + cswFlapText + `",` +
		`"vlan":"910","firstPort":"Gi9/0/41","secondPort":"Gi9/0/42"},"origin":{"mac":"0200.0000.0101"}`
	cswDaiBody = `"log":{"facility":"SW_DAI","facilityMnemonic":"INVALID_ARP","severity":"4","msg":"SW_DAI-4-INVALID_ARP: 1 Invalid ARPs (Req) on Gi9/0/44, vlan 910."},"actionResult":"blocked"`
)

// Synthetic normalized events. The wording of every text that is not a flap, SISF or SSH message is
// invented: the MAC and ARP rules' positive branches need addresses this filter does not map yet.
var cswRuleCases = []struct {
	rule, name, body string
	want             bool
}{
	{"mac_address_spoofing", "flap as this filter stores it", cswFlapBody, false},
	{"mac_address_spoofing", "flap wording under another mnemonic, with an address", `"log":{"facility":"EXAMPLE","facilityMnemonic":"MAC_FLAP","severity":"2","msg":"EXAMPLE-2-MAC_FLAP: duplicate mac 0200.0000.0401 is flapping between port Gi9/0/41 and port Gi9/0/42"},"origin":{"mac":"0200.0000.0401"}`, false},
	{"mac_address_spoofing", "duplicate MAC text with the address", `"log":{"facility":"EXAMPLE","facilityMnemonic":"DUP_MAC","severity":"4","msg":"EXAMPLE-4-DUP_MAC: Duplicate MAC address 0200.0000.0402 detected"},"origin":{"mac":"0200.0000.0402"}`, true},
	{"mac_address_spoofing", "duplicate MAC text without an address", `"log":{"facility":"EXAMPLE","facilityMnemonic":"DUP_MAC","severity":"4","msg":"EXAMPLE-4-DUP_MAC: Duplicate MAC address 0200.0000.0402 detected"}`, false},
	{"mac_address_spoofing", "duplicate MAC text only in log.message", `"log":{"facility":"EXAMPLE","facilityMnemonic":"DUP_MAC","severity":"4","message":"Duplicate MAC address 0200.0000.0402 detected"},"origin":{"mac":"0200.0000.0402"}`, false},
	{"mac_address_spoofing", "MAC move text with the address", `"log":{"facility":"EXAMPLE","facilityMnemonic":"MAC_MOVE","severity":"5","msg":"EXAMPLE-5-MAC_MOVE: MAC 0200.0000.0403 moved between port Gi9/0/41 and port Gi9/0/42"},"origin":{"mac":"0200.0000.0403"}`, true},
	{"mac_address_spoofing", "MAC conflict text at level 3 with the address", `"log":{"facility":"EXAMPLE","facilityMnemonic":"CONFLICT","severity":"3","msg":"EXAMPLE-3-CONFLICT: MAC address conflict for 0200.0000.0404"},"origin":{"mac":"0200.0000.0404"}`, true},
	{"mac_address_spoofing", "MAC conflict text at level 5 with the address", `"log":{"facility":"EXAMPLE","facilityMnemonic":"CONFLICT","severity":"5","msg":"EXAMPLE-5-CONFLICT: MAC address conflict for 0200.0000.0404"},"origin":{"mac":"0200.0000.0404"}`, false},
	{"mac_address_spoofing", "SW_DAI with an address (needs a future mapping, D-3)", cswDaiBody + `,"origin":{"mac":"0200.0000.0405"}`, true},
	{"mac_address_spoofing", "SW_DAI as this filter stores it", cswDaiBody, false},
	{"mac_address_spoofing", "SISF as this filter stores it", `"log":{"facility":"SISF","facilityMnemonic":"EXCESS_ARP_ACTIVITY","severity":"4","msg":"SISF-4-EXCESS_ARP_ACTIVITY: Excessive ARP activity detected for the client 0200.0000.0201. client is brought down and added to the exclusion list"},"origin":{"mac":"0200.0000.0201"}`, false},
	{"arp_poisoning_detection", "SW_DAI with a source address (needs a future mapping, D-3)", cswDaiBody + `,"origin":{"ip":"192.0.2.51"}`, true},
	{"arp_poisoning_detection", "SW_DAI as this filter stores it", cswDaiBody, false},
	{"arp_poisoning_detection", "IP DUPADDR with an address", `"log":{"facility":"IP","facilityMnemonic":"DUPADDR","severity":"4","msg":"IP-4-DUPADDR: Duplicate address 192.0.2.54 on Vlan910, sourced by 0200.0000.0304"},"origin":{"ip":"192.0.2.54"}`, true},
	{"arp_poisoning_detection", "IP SOURCEGUARD as this filter stores it", `"log":{"facility":"IP","facilityMnemonic":"SOURCEGUARD","severity":"4","msg":"IP-4-SOURCEGUARD: IP source guard deny on Gi9/0/44 vlan 910 for 192.0.2.55"}`, false},
	{"arp_poisoning_detection", "ARP phrase in log.msg with an address", `"log":{"facility":"EXAMPLE","facilityMnemonic":"ARP","severity":"4","msg":"EXAMPLE-4-ARP: gratuitous arp received from 192.0.2.56"},"origin":{"ip":"192.0.2.56"}`, true},
	{"arp_poisoning_detection", "ARP phrase only in log.message", `"log":{"facility":"EXAMPLE","facilityMnemonic":"ARP","severity":"4","message":"gratuitous arp received from 192.0.2.56"},"origin":{"ip":"192.0.2.56"}`, false},
	{"arp_poisoning_detection", "ARP phrase with capitals (contains is case-sensitive, D-3)", `"log":{"facility":"EXAMPLE","facilityMnemonic":"ARP","severity":"4","msg":"EXAMPLE-4-ARP: Gratuitous ARP received from 192.0.2.56"},"origin":{"ip":"192.0.2.56"}`, false},
	{"arp_poisoning_detection", "spoofing phrase at level 3", `"log":{"facility":"EXAMPLE","facilityMnemonic":"ARP","severity":"3","msg":"EXAMPLE-3-ARP: possible arp spoofing from 192.0.2.57"},"origin":{"ip":"192.0.2.57"}`, true},
	{"arp_poisoning_detection", "spoofing phrase at level 5", `"log":{"facility":"EXAMPLE","facilityMnemonic":"ARP","severity":"5","msg":"EXAMPLE-5-ARP: possible arp spoofing from 192.0.2.57"},"origin":{"ip":"192.0.2.57"}`, false},
	{"arp_poisoning_detection", "SSH as this filter stores it", `"log":{"facility":"SSH","facilityMnemonic":"SSH2_UNEXPECTED_MSG","severity":"4","msg":"SSH-4-SSH2_UNEXPECTED_MSG: Unexpected message type has arrived. Terminating the connection from 198.51.100.21"},"origin":{"ip":"198.51.100.21"}`, false},
	{"vlan_hopping_attempts", "SW_VLAN VLAN_INCONSISTENCY", `"log":{"facility":"SW_VLAN","facilityMnemonic":"VLAN_INCONSISTENCY","severity":"4"}`, true},
	{"vlan_hopping_attempts", "DTP TRUNKPORTON", `"log":{"facility":"DTP","facilityMnemonic":"TRUNKPORTON","severity":"5"}`, true},
	{"vlan_hopping_attempts", "flap text under SW_VLAN", `"log":{"facility":"SW_VLAN","facilityMnemonic":"MACFLAP_NOTIF","severity":"4"}`, true},
	{"vlan_hopping_attempts", "flap as this filter stores it", cswFlapBody, false},
	{"vlan_hopping_attempts", "SW_VLAN mnemonic not listed", `"log":{"facility":"SW_VLAN","facilityMnemonic":"VTPMODECHANGE","severity":"6"}`, false},
	{"vlan_hopping_attempts", "text branch in log.message (unchanged; nothing writes it, D-4)", `"log":{"facility":"CDP","facilityMnemonic":"NATIVE_VLAN_MISMATCH","severity":"4","message":"Native VLAN mismatch discovered"}`, true},
	{"vlan_hopping_attempts", "same text in log.msg", `"log":{"facility":"CDP","facilityMnemonic":"NATIVE_VLAN_MISMATCH","severity":"4","msg":"CDP-4-NATIVE_VLAN_MISMATCH: Native VLAN mismatch discovered"}`, false},
}

// SDK v1.1.33 CEL on synthetic normalized events for the three rule conditions.
func TestCiscoSwitchRulePredicates(t *testing.T) {
	rules := cswLoadRules(t)
	cache := plugins.NewCELCache("cisco-switch-rules")
	for _, c := range cswRuleCases {
		t.Run(c.rule+"/"+c.name, func(t *testing.T) {
			got, err := cache.Eval(rules[c.rule].Where, cswEvent(t, c.body))
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("match=%t want=%t", got, c.want)
			}
		})
	}
}

// The rule conditions on the model output of every fabricated line: the MAC and ARP rules match
// none (no line reaches a history search), and the VLAN rule matches exactly the lines whose
// playground alerts expected.json records.
func TestCiscoSwitchRulesOnFabricatedLines(t *testing.T) {
	rules := cswLoadRules(t)
	cache := plugins.NewCELCache("cisco-switch-lines")
	events, _, _ := cswModelEvents(t)
	_, expected := cswFixtures(t)
	for name, event := range events {
		var matched []string
		for stem, rule := range rules {
			ok, err := cache.Eval(rule.Where, event)
			if err != nil {
				t.Fatalf("%s on %s: %v", stem, name, err)
			}
			if ok {
				matched = append(matched, stem)
			}
		}
		sort.Strings(matched)
		want := append([]string{}, expected[name].Alerts...)
		sort.Strings(want)
		if strings.Join(matched, ",") != strings.Join(want, ",") {
			t.Errorf("%s: conditions match %v, playground alerts %v", name, matched, want)
		}
	}
	vlan := 0
	for _, c := range expected {
		for _, a := range c.Alerts {
			if a != "vlan_hopping_attempts" {
				t.Errorf("unexpected expected alert %s", a)
			}
			vlan++
		}
	}
	if vlan != 6 {
		t.Errorf("%d VLAN hopping alerts expected, want 6", vlan)
	}
}

// Whenever a rule with a history search matches, every {{.field}} placeholder resolves; an
// unresolved one fails the search, and five failures disable the rule with a Circuit Breaker
// alert. Checked on the model output of every fabricated line and on the synthetic events.
func TestCiscoSwitchHistoryPlaceholders(t *testing.T) {
	rules := cswLoadRules(t)
	cache := plugins.NewCELCache("cisco-switch-history")
	events, _, _ := cswModelEvents(t)
	for _, c := range cswRuleCases {
		events["synthetic: "+c.name] = cswEvent(t, c.body)
	}
	names := make([]string, 0, len(events))
	for name := range events {
		names = append(names, name)
	}
	sort.Strings(names)
	checked := 0
	for stem, rule := range rules {
		fields := map[string]bool{}
		cswPlaceholders(rule.Correlation, fields)
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
	if checked < 8 {
		t.Errorf("%d positive cases reached a history search, want at least 8", checked)
	}
}
