package main

// Verify the SDK Event/Alert wire-field contract and rule role selection using
// synthetic raw events. This does not execute the closed correlation service or
// prove live alert creation; lastEvent is the alert plugin's indexed alias for
// events[0], supplied by the companion alert-field resolution change.
import (
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
)

func merakiAlertWire(t *testing.T, rule *plugins.Rule, out string) string {
	t.Helper()
	event := new(plugins.Event)
	if err := utils.StringToProtoMessage(&out, event); err != nil {
		t.Fatal(err)
	}
	alert := &plugins.Alert{Events: []*plugins.Event{event}}
	switch rule.Adversary {
	case "origin":
		alert.Adversary, alert.Target = event.Origin, event.Target
	case "target":
		alert.Adversary, alert.Target = event.Target, event.Origin
	default:
		t.Fatalf("unsupported actor side %q", rule.Adversary)
	}
	wire, err := utils.ProtoMessageToString(alert)
	if err != nil {
		t.Fatal(err)
	}
	return *wire
}

func merakiGroupIdentity(t *testing.T, rule *plugins.Rule, out string) string {
	t.Helper()
	wire := merakiAlertWire(t, rule, out)
	var parts []string
	fields := append(append([]string{}, rule.GroupBy...), rule.DeduplicateBy...)
	for _, field := range fields {
		path := strings.Replace(field, "lastEvent.", "events.0.", 1)
		value := gjson.Get(wire, path)
		if !value.Exists() || value.String() == "" {
			// The documented standalone legacy Air Marshal envelope lacks the
			// device header. Collector+BSSID still supplies an actual identity.
			if field == "lastEvent.log.merakiType" {
				continue
			}
			t.Errorf("positive raw event cannot resolve alert identity field %s", field)
		}
		parts = append(parts, field+"="+value.String())
	}
	if len(parts) == 0 {
		t.Fatal("positive rule has no resolved identity fields")
	}
	return strings.Join(parts, "\x00")
}

func TestMerakiAlertFieldContracts(t *testing.T) {
	cfg, rules, cache := merakiConfig(t), merakiRules(t), plugins.NewCELCache("meraki-alert-identities")
	covered := map[string]bool{}
	for _, fixture := range merakiFixtures(t) {
		if len(fixture.Matches) == 0 {
			continue
		}
		out := merakiParse(t, cfg, fixture.Raw, fixture.DataSource, cache)
		for _, name := range fixture.Matches {
			rule := rules[name]
			if rule == nil {
				t.Fatalf("unknown fixture rule %q", name)
			}
			if yes, err := cache.Eval(rule.Where, out); err != nil || !yes {
				t.Fatalf("%s must match %s before alert identity checks: %v %v", fixture.Name, name, yes, err)
			}
			merakiGroupIdentity(t, rule, out)
			covered[name] = true
		}
		if fixture.Name == "AMP blocked positive" {
			rule := rules["advanced_malware_protection_alerts"]
			if rule.Adversary != "target" || gjson.Get(out, "origin.ip").String() != "192.0.2.8" || gjson.Get(out, "target.ip").String() != "198.51.100.9" {
				t.Fatal("download must preserve physical client/server roles and choose the remote server as alert adversary")
			}
			wire := merakiAlertWire(t, rule, out)
			if gjson.Get(wire, "adversary.ip").String() != "198.51.100.9" || gjson.Get(wire, "target.ip").String() != "192.0.2.8" {
				t.Fatal("SDK alert payload lost the selected remote/client roles")
			}
			// The previous origin setting selects the downloading client as
			// adversary; keep this before/after distinction explicit.
			oldRole := &plugins.Rule{Adversary: "origin"}
			before := merakiAlertWire(t, oldRole, out)
			if gjson.Get(before, "adversary.ip").String() != "192.0.2.8" || gjson.Get(before, "target.ip").String() != "198.51.100.9" {
				t.Fatal("before-role control did not select the original network source")
			}
		}
	}
	if len(rules) != 7 || len(covered) != len(rules) {
		t.Fatalf("every Meraki rule needs a raw positive identity check: rules=%d covered=%v", len(rules), covered)
	}
}

func TestMerakiAMPRecordIdentityFallback(t *testing.T) {
	cfg, rules, cache := merakiConfig(t), merakiRules(t), plugins.NewCELCache("meraki-amp-fallback")
	rule := rules["advanced_malware_protection_alerts"]
	var raw, source string
	for _, fixture := range merakiFixtures(t) {
		if fixture.Name == "AMP retrospective without endpoint positive" {
			raw, source = fixture.Raw, fixture.DataSource
		}
	}
	if raw == "" {
		t.Fatal("required synthetic retrospective fixture missing")
	}
	const firstID = "cb955b9f-744d-4b8e-9b29-d077d6229c36"
	const secondID = "e0421218-60d1-4b97-bc7a-5d7cdca619b5"
	first := merakiParseEvent(t, cfg, raw, source, firstID, cache)
	second := merakiParseEvent(t, cfg, raw, source, secondID, cache)
	if merakiGroupIdentity(t, rule, first) != merakiGroupIdentity(t, rule, second) || gjson.Get(first, "log.malwareGroupingType").String() != "hash" {
		t.Fatal("same valid file hash on the same collector/device must group independently of event ID")
	}
	hash := gjson.Get(first, "log.sha256").String()
	if len(hash) != 64 || !strings.Contains(raw, "sha256="+hash) {
		t.Fatal("retrospective fixture must carry its actual synthetic SHA256")
	}
	for _, variant := range []struct{ name, raw, source string }{
		{"different hash", strings.Replace(raw, "sha256="+hash, "sha256="+strings.Repeat("b", 64), 1), source},
		{"different collector", raw, "other-meraki-relay"},
		{"different appliance", strings.Replace(raw, " LAB-MX ", " LAB-OTHER ", 1), source},
	} {
		other := merakiParseEvent(t, cfg, variant.raw, variant.source, firstID, cache)
		if merakiGroupIdentity(t, rule, first) == merakiGroupIdentity(t, rule, other) {
			t.Fatalf("%s must not share an AMP alert group", variant.name)
		}
	}
	for _, value := range []string{"", "malformed"} {
		badRaw := strings.Replace(raw, "sha256="+hash+" ", "", 1)
		if value != "" {
			badRaw = strings.Replace(raw, "sha256="+hash, "sha256="+value, 1)
		}
		one := merakiParseEvent(t, cfg, badRaw, source, firstID, cache)
		two := merakiParseEvent(t, cfg, badRaw, source, secondID, cache)
		for _, event := range []struct{ out, id string }{{one, firstID}, {two, secondID}} {
			if yes, err := cache.Eval(rule.Where, event.out); err != nil || !yes {
				t.Fatalf("explicit malicious disposition must remain eligible with actual event ID: %v %v", yes, err)
			}
			if gjson.Get(event.out, "log.malwareGroupingType").String() != "event" || gjson.Get(event.out, "log.malwareGroupingKey").String() != event.id {
				t.Fatal("hashless grouping must use the actual ingress record ID")
			}
			if gjson.Get(event.out, "origin.ip").Exists() || gjson.Get(event.out, "target.ip").Exists() {
				t.Fatal("IP-less retrospective event must not acquire an endpoint")
			}
			if value != "" && gjson.Get(event.out, "log.sha256").String() != value {
				t.Fatal("malformed vendor hash must remain available for investigation")
			}
		}
		if merakiGroupIdentity(t, rule, one) == merakiGroupIdentity(t, rule, two) {
			t.Fatal("two incomplete records must not pool under collector-only grouping")
		}
		// An externally supplied ingress ID could equal a hash string. The
		// explicit identity type prevents pooling these different meanings.
		collision := merakiParseEvent(t, cfg, badRaw, source, hash, cache)
		if merakiGroupIdentity(t, rule, first) == merakiGroupIdentity(t, rule, collision) {
			t.Fatal("file hash and record ID namespaces must remain distinct")
		}
		missing := merakiParseEvent(t, cfg, badRaw, source, "", cache)
		if yes, err := cache.Eval(rule.Where, missing); err != nil || yes || gjson.Get(missing, "log.malwareGroupingKey").Exists() {
			t.Fatalf("without a valid hash or actual ingress ID no grouping identity may be invented: %v %v", yes, err)
		}
	}
}
