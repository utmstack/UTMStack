package main

// These fabricated SonicOS lines run through a step-by-step model of the
// checked-in SonicWall filter and the pinned SDK CEL. They pin which Botnet
// Filter events the botnet rule alerts on and the de-duplication keys that keep
// it from flooding. The EventProcessor playground separately runs the real
// parser and alert plugins.
import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

// Shared grok templates as the v11 backend exports them to patterns.yaml
// (backend/src/main/resources/config/liquibase/changelog/20250616001_insert_utm_regex_pattern.xml).
var sonicVolumePatterns = map[string]string{
	"data":     `(.*?)`,
	"greedy":   `.*`,
	"integer":  `(?:[+-]?(?:[0-9]+))`,
	"ipv4":     `(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.)){3}((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))`,
	"notSpace": `\S+`,
	"space":    `\s+`,
	"word":     `\b\w+\b`,
}

var sonicVolumeRegex = map[string]*regexp.Regexp{}

func sonicVolumeCompile(t *testing.T, pattern string) *regexp.Regexp {
	t.Helper()
	if r, ok := sonicVolumeRegex[pattern]; ok {
		return r
	}
	tmpl, err := template.New("grok").Option("missingkey=error").Parse(pattern)
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if err = tmpl.Execute(&b, sonicVolumePatterns); err != nil {
		t.Fatal(err)
	}
	r, err := regexp.Compile(b.String())
	if err != nil {
		t.Fatal(err)
	}
	sonicVolumeRegex[pattern] = r
	return r
}

// sonicVolumeGrok follows the EventProcessor grok step: it trims the text before
// each pattern, requires a non-empty match at the start, consumes that match,
// and writes fields only when every pattern matched.
func sonicVolumeGrok(t *testing.T, g *plugins.Grok, str string) ([][2]string, bool) {
	t.Helper()
	var found [][2]string
	for _, p := range g.Patterns {
		str = strings.TrimSpace(str)
		if str == "" {
			return nil, false
		}
		m := sonicVolumeCompile(t, p.Pattern).FindString(str)
		if m == "" || !strings.HasPrefix(str, m) {
			return nil, false
		}
		if p.FieldName != "" {
			found = append(found, [2]string{p.FieldName, strings.TrimSpace(m)})
		}
		str = strings.TrimPrefix(str, m)
	}
	return found, true
}

func sonicVolumeParse(t *testing.T, cfg *plugins.Config, raw string, cache *plugins.CELCache) string {
	t.Helper()
	draft := map[string]any{"raw": raw, "dataType": sonicDataType, "dataSource": "192.0.2.1", "event": map[string]any{}}
	for _, stage := range cfg.Pipeline {
		if len(stage.DataTypes) != 1 || stage.DataTypes[0] != sonicDataType {
			t.Fatalf("unexpected pipeline stage %v", stage.DataTypes)
		}
		for _, s := range stage.Steps {
			b, err := protojson.Marshal(s)
			if err != nil {
				t.Fatal(err)
			}
			var step map[string]map[string]any
			if err = json.Unmarshal(b, &step); err != nil {
				t.Fatal(err)
			}
			for kind, body := range step {
				if w, ok := body["where"].(string); ok && w != "" {
					snapshot, err := json.Marshal(draft)
					if err != nil {
						t.Fatal(err)
					}
					match, err := cache.Eval(w, string(snapshot))
					if err != nil {
						t.Fatal(err)
					}
					if !match {
						continue
					}
				}
				switch kind {
				case "grok":
					src := s.Grok.Source
					if src == "" {
						src = "raw"
					}
					v, ok := sonicGet(draft, src)
					if !ok {
						continue
					}
					found, ok := sonicVolumeGrok(t, s.Grok, v.(string))
					if !ok {
						continue
					}
					for _, f := range found {
						sonicPut(draft, f[0], f[1], false)
					}
				case "kv":
					v, ok := sonicGet(draft, s.Kv.Source)
					if !ok {
						t.Fatal("the KV step fails the event when its source is missing")
					}
					for _, item := range strings.Split(strings.TrimSpace(v.(string)), s.Kv.FieldSplit) {
						key, value, found := strings.Cut(item, s.Kv.ValueSplit)
						if !found {
							continue
						}
						utils.SanitizeField(&key)
						if key != "" {
							sonicPut(draft, "event."+key, strings.TrimSpace(value), false)
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
					if s.Add.Function != "string" {
						t.Fatalf("unsupported add function %s", s.Add.Function)
					}
					sonicPut(draft, s.Add.Params["key"].GetStringValue(), s.Add.Params["value"].AsInterface(), false)
				case "delete":
					for _, p := range s.Delete.Fields {
						sonicPut(draft, p, nil, true)
					}
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
				case "dynamic":
					// The external geolocation service is not executed.
				default:
					t.Fatalf("unmodeled step kind %q", kind)
				}
			}
		}
	}
	b, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	in := string(b)
	event := new(plugins.Event)
	if err = utils.StringToProtoMessage(&in, event); err != nil {
		t.Fatal(err)
	}
	out, err := utils.ProtoMessageToString(event)
	if err != nil {
		t.Fatal(err)
	}
	return *out
}

// sonicVolumeLine builds a SonicOS 7 syslog line in the shape the Botnet Filter
// writes; addresses are documentation or private ranges.
func sonicVolumeLine(code, src, dst, proto, msg string, action bool) string {
	line := `<129>  id=firewall sn=0017C5000001 time="2026-09-29 10:00:00" fw=192.0.2.1 pri=1 c=0 gcat=3 m=` + code +
		` srcMac=00:17:c5:00:00:01 src=` + src + ` srcZone=LAN natSrc=192.0.2.1:53000 dstMac=00:17:c5:00:00:02 dst=` + dst +
		` dstZone=WAN proto=` + proto + ` rcvd=76 sess="Auto" rule="Allow LAN to WAN" msg="` + msg + `" n=1001`
	if action {
		line += ` fw_action="drop"`
	}
	return line
}

func TestSonicWallBotnetAlertVolume(t *testing.T) {
	rule := loadSonicRule(t, "../../rules/sonicwall/sonicwall_firewall/botnet_detection.yml")
	if len(rule.GroupBy) != 0 || len(rule.Correlation) != 0 || !reflect.DeepEqual(rule.DeduplicateBy, []string{"dataSource", "target.ip"}) {
		t.Fatalf("grouping got groupBy %v deduplicateBy %v with %d history searches, want deduplicateBy [dataSource target.ip] and none",
			rule.GroupBy, rule.DeduplicateBy, len(rule.Correlation))
	}
	cfg := sonicConfig(t)
	cache := plugins.NewCELCache("sonicwall-alert-volume")
	for _, tc := range []struct {
		name, raw, responder string
		want                 bool
	}{
		{"database responder", sonicVolumeLine("1201", "10.1.1.45:51000:X0", "203.0.113.50:443:X1", "tcp/https",
			"Suspected Botnet responder blocked: Responder IP:203.0.113.50", true), "203.0.113.50", true},
		{"dynamic list responder", sonicVolumeLine("1519", "10.1.1.46:123:X16", "198.51.100.7:123:X1", "udp/ntp",
			"Suspected Botnet responder blocked: Responder IP:198.51.100.7, Source: Dynamic List", true), "198.51.100.7", true},
		{"database responder over ICMP", sonicVolumeLine("1201", "10.1.1.47::X16", "203.0.113.51::X1", "icmp",
			"Suspected Botnet responder blocked: Responder IP:203.0.113.51", true), "203.0.113.51", true},
		{"custom list responder", sonicVolumeLine("1477", "10.1.1.48:52000:X0", "198.51.100.9:443:X1", "tcp/https",
			"Suspected Botnet responder blocked: Responder IP:198.51.100.9, Source: Custom List", true), "", false},
		{"perimeter initiator", sonicVolumeLine("1200", "203.0.113.60:40000:X1", "10.1.1.10:23:X0", "tcp/telnet",
			"Suspected Botnet initiator blocked: Initiator IP:203.0.113.60", true), "", false},
		{"custom list initiator", sonicVolumeLine("1476", "203.0.113.61:40001:X1", "10.1.1.10:22:X0", "tcp/ssh",
			"Suspected Botnet initiator blocked: Initiator IP:203.0.113.61, Source: Custom List", true), "", false},
		{"dynamic list initiator", sonicVolumeLine("1518", "203.0.113.62:40002:X1", "10.1.1.10:4443:X0", "tcp/4443",
			"Suspected Botnet initiator blocked: Initiator IP:203.0.113.62, Source: Dynamic List", true), "", false},
		{"TCP flood notice", sonicVolumeLine("1370", "203.0.113.63:443:X1", "10.1.1.11:48197:X0", "tcp/https",
			"Possible TCP Flood on IF X2 - src: 203.0.113.63:443 dst: 10.1.1.11:48197 - rate: 324/sec continues", false), "", false},
		{"license notice naming Botnet", sonicVolumeLine("670", "10.1.1.1:0:X0", "10.1.1.2:0:X0", "tcp/0",
			"License of HA pair doesn't match: CFS AntiSpam Botnet", false), "", false},
		{"responder without a firewall decision", sonicVolumeLine("1519", "10.1.1.49:123:X16", "198.51.100.8:123:X1", "udp/ntp",
			"Suspected Botnet responder blocked: Responder IP:198.51.100.8, Source: Dynamic List", false), "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			event := sonicVolumeParse(t, cfg, tc.raw, cache)
			got, err := cache.Eval(rule.Where, event)
			if err != nil || got != tc.want {
				t.Fatalf("where got %v (%v), want %v for %s", got, err, tc.want, event)
			}
			if !tc.want {
				return
			}
			// adversary: origin, so the alert's target is the event's target: the listed address.
			if ip := gjson.Get(event, "target.ip").String(); ip != tc.responder || !strings.Contains(gjson.Get(event, "event.message").String(), tc.responder) {
				t.Fatalf("de-duplication key target.ip %q, want the listed address %s in %s", ip, tc.responder, event)
			}
			if gjson.Get(event, "dataSource").String() == "" || !strings.HasPrefix(gjson.Get(event, "origin.ip").String(), "10.") {
				t.Fatalf("missing dataSource or internal origin in %s", event)
			}
		})
	}
}
