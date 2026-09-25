package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"text/template"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// SentinelOne regression checks. Every input is fabricated (see
// testdata/sentinel-one). These tests use the pinned go-sdk v1.1.33 for YAML
// decoding, CEL and Event conversion; they do not run the EventProcessor.
// testdata/sentinel-one/replay.py runs the same raw lines through the public
// playground. See filters/audits/sentinel-one.md for the evidence and limits.

const s1Filter = "../../filters/antivirus/sentinel-one.yml"
const s1Rules = "../../rules/antivirus/sentinel-one"

// kv keeps only the first word of these values; each one is re-extracted in full.
var s1MultiWord = map[string]string{
	"accountName":                     "log.accName",
	"eventDesc":                       "log.eventDescription",
	"suser":                           "log.sourceUser",
	"duser":                           "log.destinationUser",
	"endpointDeviceControlDeviceName": "log.endpointDeviceName",
	"sourceGroupName":                 "log.sourceGpName",
	"sourceIpAddresses":               "log.sourceIps",
	"sourceMacAddresses":              "log.sourceMacs",
	"siteName":                        "log.siteName",
}

func s1Config(t *testing.T) *plugins.Pipeline {
	t.Helper()
	encoded, err := utils.ReadPbYaml(s1Filter)
	if err != nil {
		t.Fatal(err)
	}
	config := new(plugins.Config)
	if err := protojson.Unmarshal(encoded, config); err != nil {
		t.Fatal(err)
	}
	if len(config.Pipeline) != 1 || len(config.Pipeline[0].DataTypes) != 1 ||
		config.Pipeline[0].DataTypes[0] != "antivirus-sentinel-one" {
		t.Fatalf("unexpected pipeline layout: %v", config.Pipeline)
	}
	return config.Pipeline[0]
}

func s1Patterns(t *testing.T) map[string]string {
	t.Helper()
	encoded, err := utils.ReadPbYaml(filepath.Join("testdata", "sentinel-one", "patterns.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Patterns map[string]string `json:"patterns"`
	}
	if err := json.Unmarshal(encoded, &file); err != nil {
		t.Fatal(err)
	}
	return file.Patterns
}

// s1Expand fills {{.name}} the way the SDK regexp cache does for grok patterns.
func s1Expand(t *testing.T, pattern string, defs map[string]string) string {
	t.Helper()
	parsed, err := template.New("pattern").Option("missingkey=error").Parse(pattern)
	if err != nil {
		t.Fatalf("pattern %q: %v", pattern, err)
	}
	var out bytes.Buffer
	if err := parsed.Execute(&out, defs); err != nil {
		t.Fatalf("pattern %q: %v", pattern, err)
	}
	return out.String()
}

func s1Writes(step *plugins.Step) []string {
	var out []string
	if g := step.Grok; g != nil {
		for _, p := range g.Patterns {
			out = append(out, p.FieldName)
		}
	}
	if r := step.Rename; r != nil {
		out = append(out, r.To)
	}
	if a := step.Add; a != nil {
		out = append(out, a.Params["key"].GetStringValue())
	}
	return out
}

func s1Contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

func TestSentinelOneFilterContract(t *testing.T) {
	steps := s1Config(t).Steps
	header, kv := -1, -1
	for i, step := range steps {
		for _, field := range s1Writes(step) {
			if field == "log.syslogHost" {
				t.Errorf("step %d writes log.syslogHost; the CEF version slot is not a host", i)
			}
			if field == "severity" {
				t.Errorf("step %d writes severity; the vendor CEF severity scale is not established", i)
			}
		}
		if trim := step.Trim; trim != nil && s1Contains(trim.Fields, "log.syslogHost") {
			t.Errorf("step %d still trims log.syslogHost", i)
		}
		if g := step.Grok; g != nil && g.Source == "raw" && g.Where == `contains("raw", "CEF:")` {
			header = i
			var named []string
			for _, p := range g.Patterns {
				if !strings.Contains(p.FieldName, "trash") {
					named = append(named, p.FieldName)
				}
			}
			want := []string{"log.cefDeviceVendor", "log.cefDeviceProduct", "log.cefDeviceVersion",
				"log.cefSignatureId", "log.eventDescription", "log.cefSeverity", "log.restData"}
			if strings.Join(named, ",") != strings.Join(want, ",") {
				t.Errorf("CEF header fields %v, want %v", named, want)
			}
			if !strings.HasPrefix(g.Patterns[0].Pattern, "{{.data}}CEF:") {
				t.Errorf("header parse is not anchored on CEF: %q", g.Patterns[0].Pattern)
			}
		}
		if step.Kv != nil {
			kv = i
			if step.Kv.Source != "log.restData" || step.Kv.Where != `exists("log.restData")` {
				t.Errorf("kv must read log.restData only when it exists, got %q / %q", step.Kv.Source, step.Kv.Where)
			}
		}
	}
	if header < 0 || kv < 0 || header > kv {
		t.Fatalf("CEF header parse (step %d) must run before kv (step %d)", header, kv)
	}

	// Complete values, including the last key, under the existing output names.
	outputs := map[string]bool{}
	for key, out := range s1MultiWord {
		found := false
		for _, step := range steps {
			g := step.Grok
			if g == nil || g.Source != "log.restData" || len(g.Patterns) != 2 {
				continue
			}
			if g.Patterns[0].Pattern != `{{.data}}(?:^|[\s|])`+key+`=` {
				continue
			}
			found = true
			if g.Patterns[1].FieldName != out || !strings.HasSuffix(g.Patterns[1].Pattern, "|.+)") {
				t.Errorf("%s: want a last-key-safe capture into %s, got %v", key, out, g.Patterns[1])
			}
			if !strings.Contains(g.Where, key+`=\\S`) {
				t.Errorf("%s: an empty value must not be captured, where=%q", key, g.Where)
			}
		}
		if !found {
			t.Errorf("no complete-value extraction for %s", key)
		}
		outputs[out] = false
	}
	for _, step := range steps {
		if trim := step.Trim; trim != nil && trim.Function == "regex" && trim.Substring == `\s+[A-Za-z0-9_.]+=$` {
			for _, field := range trim.Fields {
				outputs[field] = true
			}
		}
	}
	for field, trimmed := range outputs {
		if !trimmed {
			t.Errorf("%s is not stripped of the following key", field)
		}
	}

	// Full rt value and a guarded deviceTime.
	var rtGrok, reformat, deviceTime bool
	for _, step := range steps {
		if g := step.Grok; g != nil && g.Source == "log.restData" && len(g.Patterns) == 2 &&
			g.Patterns[0].Pattern == `{{.data}}(?:^|[\s|])rt=#arcsightDate\(` && g.Patterns[1].FieldName == "log.rt" {
			rtGrok = true
		}
		if r := step.Reformat; r != nil && s1Contains(r.Fields, "log.deviceTimeCandidate") {
			reformat = r.Function == "time" && r.FromFormat == "Mon, 2 Jan 2006, 15:04:05 MST" &&
				r.ToFormat == "2006-01-02T15:04:05Z07:00" && strings.Contains(r.Where, "UTC$")
		}
		if r := step.Rename; r != nil && r.To == "deviceTime" {
			deviceTime = strings.Join(r.From, ",") == "log.deviceTimeCandidate" && strings.Contains(r.Where, "T[0-9]{2}")
		}
	}
	if !rtGrok || !reformat || !deviceTime {
		t.Errorf("rt/deviceTime steps: grok=%t reformat=%t rename=%t", rtGrok, reformat, deviceTime)
	}

	// The console actor is copied, not moved, and only for management events.
	writers := 0
	for i, step := range steps {
		if !s1Contains(s1Writes(step), "origin.user") {
			continue
		}
		writers++
		g := step.Grok
		if g == nil || g.Source != "log.sourceUser" || len(g.Patterns) != 1 ||
			!strings.HasPrefix(g.Where, `equals("log.cat", "SystemEvent") && `) {
			t.Errorf("step %d: origin.user must be a copy of log.sourceUser for SystemEvent only", i)
		}
	}
	if writers != 1 {
		t.Errorf("origin.user writers: %d, want 1", writers)
	}

	cleanup := steps[len(steps)-1].Delete
	if cleanup == nil || cleanup.Where != "" {
		t.Fatal("the filter must end with an unconditional cleanup")
	}
	for _, kept := range []string{"log.sourceUser", "log.eventDescription", "log.cefSeverity",
		"log.cefSignatureId", "log.ruleTime", "log.siteName", "origin.user"} {
		if s1Contains(cleanup.Fields, kept) {
			t.Errorf("cleanup deletes %s", kept)
		}
		for _, step := range steps {
			if r := step.Rename; r != nil && s1Contains(r.From, kept) {
				t.Errorf("%s is renamed away", kept)
			}
		}
	}
	for _, scratch := range []string{"log.restData", "log.deviceTimeCandidate", "log.3trash", "log.suser", "log.accountName"} {
		if !s1Contains(cleanup.Fields, scratch) {
			t.Errorf("cleanup keeps %s", scratch)
		}
	}
}

var s1RegexCall = regexp.MustCompile(`regexMatch\("[^"]*",\s*("(?:[^"\\]|\\.)*")\)`)

// The SDK regexMatch returns false for an invalid pattern, so a typo would
// silently disable a rule branch. Every filter pattern must compile as well.
func TestSentinelOnePatternsCompile(t *testing.T) {
	defs := s1Patterns(t)
	cache := plugins.NewCELCache("sentinel-one-compile")
	sample := `{"dataType":"antivirus-sentinel-one","raw":"CEF:0","log":{"eventDescription":"x","restData":"x","cat":"x"}}`
	checkWhere := func(where string) {
		if where == "" {
			return
		}
		if _, err := cache.Eval(where, sample); err != nil && !strings.Contains(err.Error(), "failed to evaluate program") {
			t.Errorf("CEL %q: %v", where, err)
		}
		for _, m := range s1RegexCall.FindAllStringSubmatch(where, -1) {
			pattern, err := strconv.Unquote(m[1])
			if err == nil {
				_, err = regexp.Compile(pattern)
			}
			if err != nil {
				t.Errorf("regexMatch pattern %s: %v", m[1], err)
			}
		}
	}
	for _, step := range s1Config(t).Steps {
		if g := step.Grok; g != nil {
			for _, p := range g.Patterns {
				if _, err := regexp.Compile(s1Expand(t, p.Pattern, defs)); err != nil {
					t.Errorf("grok %q: %v", p.Pattern, err)
				}
			}
			checkWhere(g.Where)
		}
		if tr := step.Trim; tr != nil && tr.Function == "regex" {
			if _, err := regexp.Compile(tr.Substring); err != nil {
				t.Errorf("trim %q: %v", tr.Substring, err)
			}
		}
		for _, w := range []string{step.GetKv().GetWhere(), step.GetTrim().GetWhere(), step.GetRename().GetWhere(),
			step.GetReformat().GetWhere(), step.GetAdd().GetWhere(), step.GetDelete().GetWhere()} {
			checkWhere(w)
		}
	}
	for _, rule := range s1LoadRules(t) {
		checkWhere(rule.Where)
	}
}

// s1Grok mirrors the parse loop of the public grok plugin at EventProcessor
// 497bf53dbd1ae096f7b2dbc7bce77a6bf9f22ce1 (plugins/grok/main.go): each pattern
// must match at the start of the trimmed remainder, an empty remainder stops
// the loop, and nothing is written unless every pattern matched. It is a model
// used to guard these patterns in CI, not a substitute for replay.py.
func s1Grok(t *testing.T, value string, g *plugins.Grok, defs map[string]string) (map[string]string, bool) {
	out, size := map[string]string{}, 0
	for _, p := range g.Patterns {
		value = strings.TrimSpace(value)
		if value == "" {
			break
		}
		match := regexp.MustCompile(s1Expand(t, p.Pattern, defs)).FindString(value)
		if match == "" || !strings.HasPrefix(value, match) {
			break
		}
		size++
		if p.FieldName != "" {
			out[p.FieldName] = strings.TrimSpace(match)
		}
		value = strings.TrimPrefix(value, match)
	}
	return out, size == len(g.Patterns)
}

// s1TrimRegex mirrors the regex branch of plugins/trim/main.go at the same commit.
func s1TrimRegex(value, pattern string) string {
	s := strings.TrimSpace(value)
	for _, m := range regexp.MustCompile(pattern).FindAllString(s, -1) {
		s = strings.ReplaceAll(s, m, "")
	}
	return strings.TrimSpace(s)
}

type s1Case struct {
	Log    map[string]string `json:"log"`
	Alerts []string          `json:"alerts"`
}

func s1Fixtures(t *testing.T) (map[string]string, map[string]s1Case) {
	t.Helper()
	var raw map[string]string
	var expected struct {
		Cases map[string]s1Case `json:"cases"`
	}
	for name, target := range map[string]any{"raw.json": &raw, "expected.json": &expected} {
		data, err := os.ReadFile(filepath.Join("testdata", "sentinel-one", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if len(raw) == 0 || len(raw) != len(expected.Cases) {
		t.Fatalf("raw fixtures %d, expectations %d", len(raw), len(expected.Cases))
	}
	return raw, expected.Cases
}

// Header, multi-word and rt captures on every fabricated CEF line, through the
// filter's own patterns. kv, rename, reformat and deletes are not modelled here.
func TestSentinelOneExtractionPatternsModel(t *testing.T) {
	defs := s1Patterns(t)
	steps := s1Config(t).Steps
	cache := plugins.NewCELCache("sentinel-one-model")
	var header, rt *plugins.Grok
	valueGroks := []*plugins.Grok{}
	var trim *plugins.Trim
	for _, step := range steps {
		switch g := step.Grok; {
		case g != nil && g.Where == `contains("raw", "CEF:")`:
			header = g
		case g != nil && g.Source == "log.restData" && len(g.Patterns) == 2 && g.Patterns[1].FieldName == "log.rt":
			rt = g
		case g != nil && g.Source == "log.restData" && len(g.Patterns) == 2 && strings.HasSuffix(g.Patterns[1].Pattern, "|.+)"):
			valueGroks = append(valueGroks, g)
		}
		if tr := step.Trim; tr != nil && tr.Function == "regex" {
			trim = tr
		}
	}
	if header == nil || rt == nil || trim == nil || len(valueGroks) != len(s1MultiWord) {
		t.Fatalf("steps not found: header=%t rt=%t trim=%t values=%d", header != nil, rt != nil, trim != nil, len(valueGroks))
	}
	raw, cases := s1Fixtures(t)
	checked := 0
	for name, line := range raw {
		if !strings.Contains(line, "CEF:") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			want := cases[name].Log
			got, ok := s1Grok(t, line, header, defs)
			if !ok {
				t.Fatal("CEF header not parsed")
			}
			rest := got["log.restData"]
			draft, _ := json.Marshal(map[string]any{"log": map[string]string{"restData": rest}})
			for _, g := range valueGroks {
				match, err := cache.Eval(g.Where, string(draft))
				if err != nil {
					t.Fatal(err)
				}
				if !match {
					continue
				}
				if values, ok := s1Grok(t, rest, g, defs); ok {
					got[g.Patterns[1].FieldName] = s1TrimRegex(values[g.Patterns[1].FieldName], trim.Substring)
				}
			}
			if values, ok := s1Grok(t, rest, rt, defs); ok {
				got["log.rt"] = values["log.rt"]
			}
			for _, key := range []string{"cefDeviceVendor", "cefDeviceProduct", "cefDeviceVersion",
				"cefSignatureId", "eventDescription", "cefSeverity"} {
				if got["log."+key] != want[key] {
					t.Errorf("log.%s = %q, want %q", key, got["log."+key], want[key])
				}
			}
			for _, out := range s1MultiWord {
				key := strings.TrimPrefix(out, "log.")
				if key == "eventDescription" {
					continue
				}
				if value, expected := want[key]; got[out] != value || (!expected && got[out] != "") {
					t.Errorf("%s = %q, want %q", out, got[out], value)
				}
			}
			if strings.Contains(line, "rt=#arcsightDate(") && got["log.rt"] != want["ruleTime"] {
				t.Errorf("log.rt = %q, want %q", got["log.rt"], want["ruleTime"])
			}
			if strings.Contains(rest, "CEF:") || strings.HasPrefix(rest, "|") {
				t.Errorf("header text left in log.restData: %q", rest)
			}
		})
		checked++
	}
	if checked < 40 {
		t.Fatalf("only %d CEF fixtures checked", checked)
	}
}

func s1LoadRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(s1Rules, "*.y*ml"))
	if err != nil || len(files) != 19 {
		t.Fatalf("SentinelOne rules: %d files, error %v", len(files), err)
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

func TestSentinelOneRuleConsumers(t *testing.T) {
	alertPaths, eventPaths := map[string]bool{}, map[string]bool{}
	contractPaths(new(plugins.Alert).ProtoReflect().Descriptor(), "", alertPaths)
	contractPaths(new(plugins.Event).ProtoReflect().Descriptor(), "", eventPaths)
	for stem, rule := range s1LoadRules(t) {
		if rule.Adversary != "origin" || len(rule.DataTypes) != 1 || rule.DataTypes[0] != "antivirus-sentinel-one" {
			t.Errorf("%s: adversary %q dataTypes %v", stem, rule.Adversary, rule.DataTypes)
		}
		for _, stale := range []string{"log.syslogHost", "log.eventDescToParse", `greaterOrEqual("log.confidencelevel"`} {
			if strings.Contains(rule.Where, stale) {
				t.Errorf("%s: condition still reads %s", stem, stale)
			}
		}
		for _, field := range rule.GroupBy {
			valid := alertPaths[field]
			if strings.HasPrefix(field, "lastEvent.") {
				inner := strings.TrimPrefix(field, "lastEvent.")
				valid = eventPaths[inner] || strings.HasPrefix(inner, "log.")
			}
			if !valid || strings.Contains(field, "syslogHost") {
				t.Errorf("%s: grouping path %q", stem, field)
			}
		}
	}
	rules := s1LoadRules(t)
	for stem, want := range map[string]string{
		"s1_exclusion_abuse":         "lastEvent.log.activityType,adversary.user",
		"s1_policy_downgrade":        "lastEvent.log.activityType,adversary.user",
		"memory_injection_detection": "target.host,adversary.user",
	} {
		if got := strings.Join(rules[stem].GroupBy, ","); got != want {
			t.Errorf("%s groups by %s, want %s", stem, got, want)
		}
	}
}

// SDK v1.1.33 CEL on synthetic normalized events. These check the shipped
// predicates; raw extraction for the same wording is exercised by replay.py.
func TestSentinelOneRulePredicates(t *testing.T) {
	rules := s1LoadRules(t)
	cache := plugins.NewCELCache("sentinel-one-rules")
	event := func(desc string, extra string) string {
		body := `{"dataType":"antivirus-sentinel-one","log":{"eventDescription":` + strconv.Quote(desc)
		if extra != "" {
			body += "," + extra
		}
		return body + "}}"
	}
	withHost := func(e string) string {
		return strings.TrimSuffix(e, "}") + `,"target":{"host":"endpoint-01.example.com"}}`
	}
	admin := func(name string) string {
		return event("Administrative information - New user '"+name+"' added", `"cat":"SystemEvent"`)
	}
	cases := []struct {
		rule, name, input string
		want              bool
	}{
		{"memory_injection_detection", "injection text", event("Memory injection detected in example.exe", ""), true},
		{"memory_injection_detection", "removed scratch-field branch", `{"dataType":"antivirus-sentinel-one","log":{"eventDescToParse":"Memory injection detected"}}`, false},
		{"memory_injection_detection", "near miss", event("Memory scan completed on example.exe", ""), false},

		{"behavioral_threat_detection", "endpoint", withHost(event("Behavioral anomaly detected in example.exe", "")), true},
		{"behavioral_threat_detection", "no endpoint", event("Behavioral anomaly detected in example.exe", ""), false},
		{"behavioral_threat_detection", "version slot as host", event("Behavioral anomaly detected in example.exe", `"syslogHost":"192.0.2.50"`), false},
		{"custom_detection_rule_triggers", "endpoint", withHost(event("STAR custom rule Example Watchlist triggered", "")), true},
		{"custom_detection_rule_triggers", "no endpoint", event("STAR custom rule Example Watchlist triggered", `"syslogHost":"build 23"`), false},
		{"deep_visibility_threat_indicators", "endpoint", withHost(event("Threat alert: ransomware detected in example.exe", "")), true},
		{"deep_visibility_threat_indicators", "no endpoint", event("Threat alert: ransomware detected in example.exe", ""), false},
		{"endpoint_detection_response_alerts", "endpoint", withHost(event("EDR alert: critical endpoint threat", "")), true},
		{"endpoint_detection_response_alerts", "no endpoint", event("EDR alert: critical endpoint threat", ""), false},
		{"suspicious_process_tree", "endpoint", withHost(event("Suspicious process chain started by example.exe", "")), true},
		{"suspicious_process_tree", "no endpoint", event("Suspicious process chain started by example.exe", ""), false},

		{"kernel_level_threat", "endpoint", withHost(event("Kernel exploit blocked on example.exe", "")), true},
		{"kernel_level_threat", "console text", admin("Kernel Team (kernel.team@detect.example.com)"), false},
		{"threat_intelligence_matches", "vendor MALICIOUS", event("Example file flagged", `"filecontenthash":"0123456789abcdef","confidencelevel":"MALICIOUS"`), true},
		{"threat_intelligence_matches", "lower-case malicious", event("Example file flagged", `"filecontenthash":"0123456789abcdef","confidencelevel":"malicious"`), true},
		{"threat_intelligence_matches", "vendor SUSPICIOUS", event("Example file flagged", `"filecontenthash":"0123456789abcdef","confidencelevel":"SUSPICIOUS"`), false},
		{"threat_intelligence_matches", "0-100 score", event("Example file flagged", `"filecontenthash":"0123456789abcdef","confidencelevel":"95"`), false},
		{"threat_intelligence_matches", "reputation on an endpoint", withHost(event("File reputation lookup flagged example.exe", "")), true},
		{"threat_intelligence_matches", "reputation in console text", admin("Reputation Desk (reputation.desk@example.com)"), false},

		{"s1_policy_downgrade", "protect to detect", event("Site policy mode changed from Protect to Detect", `"cat":"SystemEvent"`), true},
		{"s1_policy_downgrade", "downgraded", event("Policy downgraded for Example Site Alpha", ""), true},
		{"s1_policy_downgrade", "pair, malicious mode lowered", event("Site policy changed from Protect/Detect to Detect/Detect", ""), true},
		{"s1_policy_downgrade", "pair, suspicious mode lowered", event("Site policy changed from Protect/Protect to Protect/Detect", ""), true},
		{"s1_policy_downgrade", "upgrade", event("Site policy mode changed from Detect to Protect", ""), false},
		{"s1_policy_downgrade", "pair upgrade", event("Site policy changed from Detect/Detect to Protect/Detect", ""), false},
		{"s1_policy_downgrade", "pair upgrade, suspicious", event("Site policy changed from Protect/Detect to Protect/Protect", ""), false},
		{"s1_policy_downgrade", "role naming both modes", event(`Example Admin assigned role "Protect and Detect Reviewers" to user Policy Team (policy.team@example.com)`, ""), false},
		{"s1_policy_downgrade", "no policy wording", event("Mode changed from Protect to Detect", ""), false},

		{"agent_tampering_attempts", "agent disabled", event("Example Admin disabled the agent on endpoint-01", ""), true},
		{"agent_tampering_attempts", "agents uninstalled", event("2 agents uninstalled from Example Site Alpha", ""), true},
		{"agent_tampering_attempts", "in-word stop and agent", admin("Nonstop Example (nonstop.example@agentur.example.com)"), false},
		{"agent_tampering_attempts", "stopwatch", event("agent stopwatch sync", ""), false},
		{"threat_mitigation_failures", "mitigation failed", event("Threat mitigation failed on endpoint-01", ""), true},
		{"threat_mitigation_failures", "remediation failures", event("remediation failures on endpoint-01", ""), true},
		{"threat_mitigation_failures", "failover in a name", event("Administrative information - User 'Mitigation Failover (mitigation.failover@example.com)' Deleted", ""), false},
		{"iot_device_compromise_indicators", "firmware backdoor", event("Malicious firmware backdoor detected", `"endpointDeviceName":"Example Camera 01"`), true},
		{"iot_device_compromise_indicators", "ics inside analytics", event("Example backdoor file detected in analytics service", `"endpointDeviceName":"Example Server 01"`), false},
		{"iot_device_compromise_indicators", "iot inside patriot", event("patriot backdoor detected", `"endpointDeviceName":"Example Server 01"`), false},
	}
	for _, tc := range cases {
		t.Run(tc.rule+"/"+tc.name, func(t *testing.T) {
			rule := rules[tc.rule]
			if rule == nil {
				t.Fatalf("rule %s not found", tc.rule)
			}
			input := tc.input
			ev := new(plugins.Event)
			if err := utils.StringToProtoMessage(&input, ev); err != nil {
				t.Fatal(err)
			}
			output, err := utils.ProtoMessageToString(ev)
			if err != nil {
				t.Fatal(err)
			}
			match, err := cache.Eval(rule.Where, *output)
			if err != nil {
				t.Fatal(err)
			}
			if match != tc.want {
				t.Fatalf("match=%t want=%t on %s", match, tc.want, *output)
			}
		})
	}
}
