package main

// Regression checks for the CrowdStrike review corrections in filters/crowdstrike/crowdstrike.yml
// and rules/crowdstrike.
//
// Every input in testdata/crowdstrike-review/raw.json is a fabricated Falcon event-stream record:
// documentation addresses (RFC 5737), example names and all-zero identifiers. The records carry no
// UserIp and no case asserts actionResult, because the sign-in address mapping and the outcome
// values belong to a separate change.
//
// csReviewNormalize models how EventProcessor 8a3ade7 runs this filter: the json, rename, trim,
// cast, add and delete step plugins (plugins/<step>/main.go) and drop, with every where clause
// evaluated by this module's go-sdk CEL and a failed step leaving the event unchanged, followed by
// the Event conversion the engine does before analysis. Step kinds it does not model fail the
// test instead of being skipped. The geolocation step is a separate enrichment plugin and is not
// modeled, so origin.geolocation is never asserted. Rule conditions are evaluated on the resulting
// Event as the CEL plugin does, and each alert is built as the CEL plugin's generateAlert builds
// it, so grouping and deduplication keys are resolved by this plugin's grouping.go.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	csReviewFilterFile = "../../filters/crowdstrike/crowdstrike.yml"
	csReviewRuleDir    = "../../rules/crowdstrike"
	csReviewCaseFile   = "testdata/crowdstrike-review/raw.json"
)

type csReviewCase struct {
	Name            string          `json:"name"`
	Purpose         string          `json:"purpose"`
	Raw             json.RawMessage `json:"raw"`
	Expected        map[string]any  `json:"expected"`
	Absent          []string        `json:"absent"`
	Rules           []string        `json:"rules"`
	MissingIdentity []string        `json:"missingIdentity"`
}

type csReviewEvent struct {
	csReviewCase
	rawText   string
	event     *plugins.Event
	eventJSON string
}

var (
	csReviewIPv4          = regexp.MustCompile(`\b\d{1,3}(?:\.\d{1,3}){3}\b`)
	csReviewFieldArgument = regexp.MustCompile(`\b(?:equals|equalsIgnoreCase|exists|oneOf|contains|containsAll|startsWith|endsWith|regexMatch|inCIDR|greaterThan|lessThan|greaterOrEqual|lessOrEqual)\(\s*"([^"]+)"`)
)

func csReviewCases(t *testing.T) []csReviewCase {
	t.Helper()
	data, err := os.ReadFile(csReviewCaseFile)
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Cases []csReviewCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	var documentation []*net.IPNet
	for _, cidr := range []string{"192.0.2.0/24", "198.51.100.0/24", "203.0.113.0/24"} {
		_, network, _ := net.ParseCIDR(cidr)
		documentation = append(documentation, network)
	}
	seen := map[string]bool{}
	for _, c := range file.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("missing or repeated case name %q", c.Name)
		}
		seen[c.Name] = true
		for _, address := range csReviewIPv4.FindAllString(string(c.Raw), -1) {
			ip := net.ParseIP(address)
			if ip == nil || !slices.ContainsFunc(documentation, func(n *net.IPNet) bool { return n.Contains(ip) }) {
				t.Fatalf("%s: %s is not a documentation address", c.Name, address)
			}
		}
		if gjson.GetBytes(c.Raw, "event.UserIp").Exists() {
			t.Fatalf("%s: UserIp belongs to the separate sign-in address change", c.Name)
		}
		if _, ok := c.Expected["actionResult"]; ok {
			t.Fatalf("%s: actionResult belongs to the separate outcome change", c.Name)
		}
	}
	if len(file.Cases) < 10 {
		t.Fatalf("only %d cases", len(file.Cases))
	}
	return file.Cases
}

func csReviewFilter(t *testing.T) *plugins.Config {
	t.Helper()
	b, err := utils.ReadPbYaml(csReviewFilterFile)
	if err != nil {
		t.Fatal(err)
	}
	cfg := new(plugins.Config)
	if err := protojson.Unmarshal(b, cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func csReviewRules(t *testing.T) map[string]*plugins.Rule {
	t.Helper()
	rules := map[string]*plugins.Rule{}
	for _, pattern := range []string{"*.yml", "*.yaml"} {
		paths, err := filepath.Glob(filepath.Join(csReviewRuleDir, pattern))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			b, err := utils.ReadPbYaml(path)
			if err != nil {
				t.Fatal(err)
			}
			rule := new(plugins.Rule)
			if err := protojson.Unmarshal(b, rule); err != nil {
				t.Fatalf("%s: %v", path, err)
			}
			rule.Normalize()
			if !slices.Contains(rule.DataTypes, "crowdstrike") {
				t.Fatalf("%s: not a crowdstrike rule", path)
			}
			rules[filepath.Base(path)] = rule
		}
	}
	if len(rules) == 0 {
		t.Fatal("no CrowdStrike rules")
	}
	return rules
}

func csReviewSet(m map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	for _, key := range parts[:len(parts)-1] {
		next, ok := m[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[key] = next
		}
		m = next
	}
	m[parts[len(parts)-1]] = value
}

func csReviewDelete(m map[string]any, path string) {
	parts := strings.Split(path, ".")
	for _, key := range parts[:len(parts)-1] {
		next, ok := m[key].(map[string]any)
		if !ok {
			return
		}
		m = next
	}
	delete(m, parts[len(parts)-1])
}

// csReviewNormalize returns the Event the engine would hand to analysis and the step errors it
// would store on it, or a nil Event when a drop step matched. The error result is a model limit.
func csReviewNormalize(cfg *plugins.Config, cache *plugins.CELCache, id, raw string) (*plugins.Event, []string, error) {
	envelope := &plugins.Log{Id: id, DataType: "crowdstrike", DataSource: "synthetic-crowdstrike",
		Timestamp: "2026-09-24T12:00:00Z", TenantId: "00000000-0000-4000-8000-000000000001", Raw: raw}
	start, err := utils.ProtoMessageToString(envelope)
	if err != nil {
		return nil, nil, err
	}
	var draft map[string]any
	if err := json.Unmarshal([]byte(*start), &draft); err != nil {
		return nil, nil, err
	}
	text := func() string {
		b, err := json.Marshal(draft)
		if err != nil {
			panic(err)
		}
		return string(b)
	}
	var stepErrors []string
	where := func(expression string) bool {
		if expression == "" {
			return true
		}
		current := text()
		matched, err := cache.Evaluate(&current, expression)
		if err != nil {
			stepErrors = append(stepErrors, fmt.Sprintf("where %s: %v", expression, err))
		}
		return matched
	}
	for _, stage := range cfg.Pipeline {
		if !slices.Contains(stage.DataTypes, envelope.DataType) {
			continue
		}
		for i, step := range stage.Steps {
			before := text()
			var failure error
			switch {
			case step.Json != nil:
				if !where(step.Json.Where) {
					continue
				}
				source := gjson.Get(before, step.Json.Source)
				if !source.Exists() {
					failure = errors.New("source was not found")
					break
				}
				var pairs map[string]any
				if failure = json.Unmarshal([]byte(source.String()), &pairs); failure != nil {
					break
				}
				for key, value := range pairs {
					utils.SanitizeField(&key)
					csReviewSet(draft, "log."+key, value)
				}
			case step.Rename != nil:
				if !where(step.Rename.Where) {
					continue
				}
				to := step.Rename.To
				utils.SanitizeField(&to)
				failure = utils.ValidateReservedField(to, false)
				for _, from := range step.Rename.From {
					if failure != nil {
						break
					}
					if failure = utils.ValidateReservedField(from, false); failure != nil {
						break
					}
					value := gjson.Get(text(), from)
					if !value.Exists() {
						continue
					}
					csReviewSet(draft, to, utils.GetValueOf(value))
					csReviewDelete(draft, from)
				}
			case step.Trim != nil:
				if !where(step.Trim.Where) {
					continue
				}
				for _, field := range step.Trim.Fields {
					if failure = utils.ValidateReservedField(field, false); failure != nil {
						break
					}
					value := gjson.Get(text(), field)
					if !value.Exists() || value.String() == "" {
						continue
					}
					trimmed := strings.TrimSpace(value.String())
					switch step.Trim.Function {
					case "prefix":
						trimmed = strings.TrimPrefix(trimmed, step.Trim.Substring)
					case "suffix":
						trimmed = strings.TrimSuffix(trimmed, step.Trim.Substring)
					case "substring":
						trimmed = strings.ReplaceAll(trimmed, step.Trim.Substring, "")
					case "regex":
						pattern, err := regexp.Compile(step.Trim.Substring)
						if err != nil {
							failure = err
							break
						}
						matches := pattern.FindAllString(trimmed, -1)
						if len(matches) == 0 {
							continue
						}
						for _, match := range matches {
							trimmed = strings.ReplaceAll(trimmed, match, "")
						}
					}
					if failure != nil {
						break
					}
					csReviewSet(draft, field, strings.TrimSpace(trimmed))
				}
			case step.Cast != nil:
				if !where(step.Cast.Where) {
					continue
				}
				for _, field := range step.Cast.Fields {
					if failure = utils.ValidateReservedField(field, false); failure != nil {
						break
					}
					value := gjson.Get(text(), field)
					if !value.Exists() {
						continue
					}
					switch step.Cast.To {
					case "string":
						csReviewSet(draft, field, utils.CastString(value.Value()))
					case "int":
						csReviewSet(draft, field, utils.CastInt64(value.Value()))
					case "float":
						csReviewSet(draft, field, utils.CastFloat64(value.Value()))
					case "bool":
						csReviewSet(draft, field, utils.CastBool(value.Value()))
					case "[]string":
						return nil, stepErrors, fmt.Errorf("step %d: cast to []string is not modeled", i)
					default:
						failure = fmt.Errorf("unsupported cast type %q", step.Cast.To)
					}
					if failure != nil {
						break
					}
				}
			case step.Add != nil:
				if !where(step.Add.Where) {
					continue
				}
				if step.Add.Function != "string" {
					failure = fmt.Errorf("function %q not supported", step.Add.Function)
					break
				}
				key, hasKey := step.Add.Params["key"]
				value, hasValue := step.Add.Params["value"]
				if !hasKey || !hasValue {
					failure = errors.New("missing required parameter")
					break
				}
				name := key.GetStringValue()
				utils.SanitizeField(&name)
				if failure = utils.ValidateReservedField(name, false); failure == nil {
					csReviewSet(draft, name, value.GetStringValue())
				}
			case step.Delete != nil:
				if !where(step.Delete.Where) {
					continue
				}
				for _, field := range step.Delete.Fields {
					if failure = utils.ValidateReservedField(field, false); failure != nil {
						break
					}
					if gjson.Get(text(), field).Exists() {
						csReviewDelete(draft, field)
					}
				}
			case step.Dynamic != nil:
				// Enrichment (geolocation) is a separate plugin: evaluate its condition only.
				where(step.Dynamic.Where)
				continue
			case step.Drop != nil:
				if step.Drop.Where == "" {
					stepErrors = append(stepErrors, "drop operation requires where clause")
					continue
				}
				if where(step.Drop.Where) {
					return nil, stepErrors, nil
				}
				continue
			default:
				return nil, stepErrors, fmt.Errorf("step %d: step kind not modeled", i)
			}
			if failure != nil {
				stepErrors = append(stepErrors, fmt.Sprintf("step %d: %v", i, failure))
				draft = nil
				if err := json.Unmarshal([]byte(before), &draft); err != nil {
					return nil, stepErrors, err
				}
			}
		}
	}
	final := text()
	event := new(plugins.Event)
	if err := utils.StringToProtoMessage(&final, event); err != nil {
		return nil, stepErrors, err
	}
	if event.DeviceTime == "" {
		event.DeviceTime = event.Timestamp
	}
	event.Errors = append(event.Errors, stepErrors...)
	return event, stepErrors, nil
}

func csReviewEvents(t *testing.T) []csReviewEvent {
	t.Helper()
	cfg, cache := csReviewFilter(t), plugins.NewCELCache("crowdstrike-review-filter")
	var out []csReviewEvent
	for _, c := range csReviewCases(t) {
		var compact bytes.Buffer
		if err := json.Compact(&compact, c.Raw); err != nil {
			t.Fatalf("%s: %v", c.Name, err)
		}
		event, stepErrors, err := csReviewNormalize(cfg, cache, "cs-review-"+c.Name, compact.String())
		if err != nil {
			t.Fatalf("%s: %v", c.Name, err)
		}
		if event == nil {
			t.Fatalf("%s: dropped", c.Name)
		}
		if len(stepErrors) != 0 {
			t.Fatalf("%s: step errors %v", c.Name, stepErrors)
		}
		serialized, err := utils.ProtoMessageToString(event)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, csReviewEvent{csReviewCase: c, rawText: compact.String(), event: event, eventJSON: *serialized})
	}
	return out
}

// Command lines are stored as sent, apart from surrounding spaces. The two double-quote trims
// removed one quote from each end, which left quoted Windows paths and arguments unbalanced.
// The other fields check that the three removed renames were repeats: their values still arrive
// once, under the same names.
func TestCrowdStrikeReviewNormalizedFields(t *testing.T) {
	commandLines := 0
	for _, e := range csReviewEvents(t) {
		t.Run(e.Name, func(t *testing.T) {
			for path, want := range e.Expected {
				got := gjson.Get(e.eventJSON, path)
				if !got.Exists() || !reflect.DeepEqual(got.Value(), want) {
					t.Errorf("%s = %s, want %v", path, got.Raw, want)
				}
			}
			for _, path := range e.Absent {
				if gjson.Get(e.eventJSON, path).Exists() {
					t.Errorf("%s should be absent", path)
				}
			}
			sent := gjson.Get(e.rawText, "event.CommandLine")
			if sent.Type != gjson.String {
				return
			}
			commandLines++
			stored := gjson.Get(e.eventJSON, "log.eventCommandLine").String()
			if want := strings.TrimSpace(sent.String()); stored != want {
				t.Errorf("command line %q, sent as %q", stored, want)
			}
		})
	}
	if commandLines < 10 {
		t.Fatalf("only %d text command lines", commandLines)
	}
}

// An unconditional rename with the same source and target as an earlier one can never find
// its source again: the earlier step already moved it.
func TestCrowdStrikeReviewRenamesNotRepeated(t *testing.T) {
	seen := map[string]int{}
	for _, stage := range csReviewFilter(t).Pipeline {
		for i, step := range stage.Steps {
			if step.Rename == nil || step.Rename.Where != "" {
				continue
			}
			key := strings.Join(step.Rename.From, ",") + " -> " + step.Rename.To
			if first, ok := seen[key]; ok {
				t.Errorf("step %d repeats step %d: %s", i, first, key)
				continue
			}
			seen[key] = i
		}
	}
}

// Rule conditions on the normalized fixtures give exactly the listed rules. Every field a rule
// reads, in its condition or its history search, must be one the filter writes: the CrowdStrike
// plugin sends only the event-stream keys metadata and event, so a name such as
// log.event_simpleName (Falcon Data Replicator) can never arrive.
func TestCrowdStrikeReviewRuleConditions(t *testing.T) {
	cfg, rules := csReviewFilter(t), csReviewRules(t)
	cache := plugins.NewCELCache("crowdstrike-review-rules")
	written := map[string]bool{"dataType": true, "dataSource": true, "tenantId": true}
	for _, stage := range cfg.Pipeline {
		for _, step := range stage.Steps {
			if s := step.Rename; s != nil {
				to := s.To
				utils.SanitizeField(&to)
				written[to] = true
			}
			if s := step.Add; s != nil {
				key := s.Params["key"].GetStringValue()
				utils.SanitizeField(&key)
				written[key] = true
			}
		}
	}
	for name, rule := range rules {
		var read []string
		for _, m := range csReviewFieldArgument.FindAllStringSubmatch(rule.Where, -1) {
			read = append(read, m[1])
		}
		var searches func([]*plugins.SearchRequest)
		searches = func(list []*plugins.SearchRequest) {
			for _, s := range list {
				for _, x := range s.With {
					read = append(read, x.Field)
					if v := x.Value.GetStringValue(); strings.HasPrefix(v, "{{.") && strings.HasSuffix(v, "}}") {
						read = append(read, strings.TrimSuffix(strings.TrimPrefix(v, "{{."), "}}"))
					}
				}
				searches(s.Or)
			}
		}
		searches(rule.Correlation)
		if len(read) == 0 {
			t.Errorf("%s: no field found in its condition", name)
		}
		for _, field := range read {
			if !written[strings.TrimSuffix(field, ".keyword")] {
				t.Errorf("%s reads %s, which the filter never writes", name, field)
			}
		}
	}
	for _, e := range csReviewEvents(t) {
		t.Run(e.Name, func(t *testing.T) {
			var matched []string
			for name, rule := range rules {
				ok, err := cache.Eval(rule.Where, e.event)
				if err != nil {
					t.Errorf("%s: %v", name, err)
				}
				if ok {
					matched = append(matched, name)
				}
			}
			sort.Strings(matched)
			want := append([]string{}, e.Rules...)
			sort.Strings(want)
			if !reflect.DeepEqual(matched, want) && !(len(matched) == 0 && len(want) == 0) {
				t.Errorf("matched %v, want %v", matched, want)
			}
		})
	}
}

// groupBy and deduplicateBy name alert fields. An alert has no origin: the CEL plugin copies the
// event's origin into adversary (adversary: origin), and grouping.go skips keys it cannot resolve,
// so origin.* keys never group and never deduplicate.
func TestCrowdStrikeReviewAlertKeys(t *testing.T) {
	rules := csReviewRules(t)
	eventPaths, alertPaths := map[string]bool{}, map[string]bool{}
	contractPaths(new(plugins.Event).ProtoReflect().Descriptor(), "", eventPaths)
	contractPaths(new(plugins.Alert).ProtoReflect().Descriptor(), "", alertPaths)
	alertPath := func(p string) bool {
		p = groupingArrayIndex.ReplaceAllString(strings.TrimSuffix(p, ".keyword"), "$1")
		if rest, ok := strings.CutPrefix(p, "lastEvent."); ok {
			return eventPaths[rest] || strings.HasPrefix(rest, "log.")
		}
		return alertPaths[p]
	}
	needed := map[string]bool{}
	for name, rule := range rules {
		for _, field := range append(append([]string{}, rule.GroupBy...), rule.DeduplicateBy...) {
			if !alertPath(field) {
				t.Errorf("%s: %s is not an alert field", name, field)
			}
			if strings.HasPrefix(field, "adversary.") {
				needed[name] = true
			}
		}
	}
	exercised := map[string]bool{}
	for _, e := range csReviewEvents(t) {
		for _, name := range e.Rules {
			rule, ok := rules[name]
			if !ok {
				t.Fatalf("%s: unknown rule %s", e.Name, name)
			}
			fields := rule.GroupBy
			if len(rule.DeduplicateBy) > 0 {
				fields = rule.DeduplicateBy
			}
			if len(fields) == 0 {
				continue
			}
			t.Run(e.Name+"/"+name, func(t *testing.T) {
				adversary, target := e.event.Origin, e.event.Target
				if strings.ToLower(rule.Adversary) == "target" {
					adversary, target = e.event.Target, e.event.Origin
				}
				alert := &plugins.Alert{Name: rule.Name, TenantId: e.event.TenantId, DataSource: e.event.DataSource,
					DataType: e.event.DataType, Category: rule.Category, Technique: rule.Technique,
					Description: rule.Description, Impact: rule.Impact, References: rule.References,
					Adversary: adversary, Target: target, Events: []*plugins.Event{e.event},
					DeduplicateBy: rule.DeduplicateBy, GroupBy: rule.GroupBy}
				wire, err := utils.ProtoMessageToString(alert)
				if err != nil {
					t.Fatal(err)
				}
				builder := sdkos.NewBoolBuilder(context.Background(), nil, "crowdstrike-review-keys")
				builder.FilterTerm("name", rule.Name)
				if !addAlertGroupingTerms(builder, *wire, fields) {
					t.Fatalf("no usable key in %v", fields)
				}
				query, buildErrors := builder.BuildWithErrors()
				if len(buildErrors) != 0 {
					t.Fatal(buildErrors)
				}
				terms := map[string]any{}
				for _, clause := range query.Bool.Filter {
					for field, term := range clause.Term {
						terms[field] = term["value"]
					}
				}
				for _, field := range fields {
					field = strings.TrimSuffix(field, ".keyword")
					value, resolved := scalarGroupingValue(alertGroupingValue(*wire, field))
					if slices.Contains(e.MissingIdentity, field) {
						if resolved {
							t.Errorf("%s resolved to %v on a record without it", field, value)
						}
						continue
					}
					if !resolved {
						t.Errorf("%s does not resolve", field)
						continue
					}
					if side, ok := strings.CutPrefix(field, "adversary."); ok {
						if want := gjson.Get(e.eventJSON, "origin."+side).Value(); value != want {
							t.Errorf("%s = %v, want the event's origin.%s %v", field, value, side, want)
						}
					}
					if terms[field] != value {
						t.Errorf("search term %s = %v, want %v", field, terms[field], value)
					}
				}
				if name := terms["name"]; name != rule.Name {
					t.Errorf("name term %v", name)
				}
			})
			exercised[name] = true
		}
	}
	for name := range needed {
		if !exercised[name] {
			t.Errorf("%s: no fixture exercises its adversary keys", name)
		}
	}
}
