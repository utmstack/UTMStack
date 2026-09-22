package main

// SDK Event/Alert wire-contract tests from synthetic ESET JSON. Alert side
// selection is asserted according to the rule metadata; this is not execution
// of the closed correlation service and does not prove live alert creation.
// lastEvent is the indexed alias resolved from events[0] by the companion fix.
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
)

func esetAlertWire(t *testing.T, rule *plugins.Rule, out string) string {
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

func esetGroupIdentity(t *testing.T, rule *plugins.Rule, out string) string {
	t.Helper()
	wire := esetAlertWire(t, rule, out)
	var parts []string
	fields := append(append([]string{}, rule.GroupBy...), rule.DeduplicateBy...)
	for _, field := range fields {
		path := strings.Replace(field, "lastEvent.", "events.0.", 1)
		value := gjson.Get(wire, path)
		if !value.Exists() || value.String() == "" {
			// These vendor-specific detector details are documented optional.
			// The managed endpoint namespace/key and collector remain required.
			if field == "lastEvent.log.rulename" || field == "lastEvent.log.ruleid" || field == "lastEvent.log.threatname" || field == "lastEvent.log.target" {
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

func TestESETAlertFieldContracts(t *testing.T) {
	cfg, rules, cache := esetConfig(t), esetRules(t), plugins.NewCELCache("eset-alert-fields")
	covered := map[string]bool{}
	for _, fixture := range esetFixtures(t) {
		if fixture.ParseError || len(fixture.Matches) == 0 {
			continue
		}
		out := esetParse(t, cfg, fixture.Raw, fixture.DataSource, cache)
		for _, name := range fixture.Matches {
			rule := rules[name]
			if rule == nil {
				t.Fatalf("unknown fixture rule %q", name)
			}
			if yes, err := cache.Eval(rule.Where, out); err != nil || !yes {
				t.Fatalf("%s must match %s before alert checks: %v %v", fixture.Name, name, yes, err)
			}
			esetGroupIdentity(t, rule, out)
			covered[name] = true
		}
	}
	if len(rules) != 14 || len(covered) != len(rules) {
		t.Fatalf("all14 consumers require raw positive identity coverage: rules=%d covered=%v", len(rules), covered)
	}
}

func TestESETDirectionalActorsAndPolicyBoundaries(t *testing.T) {
	cfg, rules, cache := esetConfig(t), esetRules(t), plugins.NewCELCache("eset-consumer-boundaries")
	parse := func(m map[string]any) string {
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return esetParse(t, cfg, string(b), "eset-test-relay", cache)
	}
	matches := func(out, name string, want bool) {
		t.Helper()
		rule := rules[name]
		if rule == nil {
			t.Fatalf("rule missing: %s", name)
		}
		if yes, err := cache.Eval(rule.Where, out); err != nil || yes != want {
			t.Fatalf("%s match=%v want=%v: %v", name, yes, want, err)
		}
	}
	base := func(kind string) map[string]any {
		return map[string]any{"event_type": kind, "hostname": "HOST-A", "source_uuid": "c1c09d6e-24bd-4b45-b1c6-f3c227c9df8e", "ipv4": "192.0.2.10", "severity": "Warning"}
	}
	for _, inbound := range []bool{false, true} {
		event := base("FirewallAggregated_Event")
		event["inbound"], event["threat_name"], event["event"], event["action"] = inbound, "Botnet.CnC.Generic", "Botnet communication detected", "blocked"
		event["source_address"], event["target_address"] = "192.0.2.10", "198.51.100.9"
		event["account"], event["process_name"] = "LAB\\user", `C:\Synthetic\client.exe`
		name, other := "botnet_communication_attempts", "botnet_inbound_communication_attempts"
		if inbound {
			event["source_address"], event["target_address"] = "198.51.100.9", "192.0.2.10"
			name, other = other, name
		}
		out := parse(event)
		matches(out, name, true)
		matches(out, other, false)
		matches(out, "network_attack_detection", false)
		if gjson.Get(out, "origin.ip").String() != event["source_address"] || gjson.Get(out, "target.ip").String() != event["target_address"] {
			t.Fatal("physical endpoints were swapped")
		}
		wire := esetAlertWire(t, rules[name], out)
		if gjson.Get(wire, "adversary.ip").String() != "198.51.100.9" || gjson.Get(wire, "target.ip").String() != "192.0.2.10" || gjson.Get(wire, "target.host").String() != "HOST-A" {
			t.Fatal("botnet alert did not retain remote adversary and managed target")
		}
		if gjson.Get(wire, "adversary.host").Exists() {
			t.Fatal("managed hostname must not be copied to the remote botnet side")
		}
		if !inbound {
			before := esetAlertWire(t, &plugins.Rule{Adversary: "origin"}, out)
			if gjson.Get(before, "adversary.ip").String() != "192.0.2.10" {
				t.Fatal("old origin-role control must select the managed client")
			}
		}
		for _, invalid := range []any{nil, "false", "true", 0, 1} {
			event["inbound"] = invalid
			neutral := parse(event)
			matches(neutral, name, false)
			matches(neutral, other, false)
			matches(neutral, "network_attack_detection", false)
			if gjson.Get(neutral, "origin.host").Exists() || gjson.Get(neutral, "target.host").Exists() {
				t.Fatal("unknown direction must not assign managed hostname to a network side")
			}
		}
		delete(event, "inbound")
		neutral := parse(event)
		matches(neutral, name, false)
		matches(neutral, other, false)
	}
	policy := base("HipsAggregated_Event")
	policy["application"], policy["operation"], policy["target"], policy["action"], policy["rule_name"] = `C:\Synthetic\utility.exe`, "Read file", `C:\Data\ordinary.txt`, "blocked", "Restrictive file access policy"
	out := parse(policy)
	matches(out, "host_intrusion_prevention_triggers", false)
	matches(out, "suspicious_process_behavior", false)
	// Specific activity rules are intentionally factual policy-block alerts;
	// they do not claim that these benign examples establish malware.
	policy["application"] = `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`
	out = parse(policy)
	matches(out, "suspicious_powershell_activity_blocked", true)
	matches(out, "host_intrusion_prevention_triggers", false)
	policy["application"], policy["operation"], policy["target"] = `C:\Synthetic\utility.exe`, "Write registry value", `HKLM\Software\Synthetic`
	out = parse(policy)
	matches(out, "registry_modification_attempts_blocked", true)
	matches(out, "host_intrusion_prevention_triggers", false)
	policy["operation"], policy["rule_name"] = "Attempt to run a suspicious object", "Suspicious application launch"
	out = parse(policy)
	matches(out, "host_intrusion_prevention_triggers", true)
	inspect := base("EnterpriseInspectorAlert_Event")
	inspect["rulename"] = "Backup agent disabled"
	out = parse(inspect)
	matches(out, "eset_agent_tampering", false)
	// Incidental tokens in filenames, account names or arbitrary descriptions
	// must not reclassify an unrelated structured antivirus detection.
	threat := base("Threat_Event")
	threat["threat_name"], threat["scanner_id"], threat["action_taken"] = "Synthetic.Other", "Real-time file system protection", "Detected"
	threat["object_uri"] = "file:///C:/NewHeur-machine-learning-botnet-ransomware.encrypted"
	threat["username"] = "registry exploit quarantine failed"
	out = parse(threat)
	for name := range rules {
		matches(out, name, false)
	}
	audit := base("Audit_Event")
	audit["domain"], audit["action"], audit["target"], audit["result"], audit["user"] = "Native user", "Login attempt", "AttemptedAdmin", "Failure", "AuditingActor"
	out = parse(audit)
	matches(out, "eset_console_abuse", true)
	wire := esetAlertWire(t, rules["eset_console_abuse"], out)
	if gjson.Get(wire, "target.user").String() != "AttemptedAdmin" || gjson.Get(wire, "adversary.user").String() != "AuditingActor" {
		t.Fatal("attempted account and supplied audit actor must remain distinct")
	}
}
