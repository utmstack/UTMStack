package main

// Fabricated macOS unified-log records run through the offline macOS parser
// model (macParse) and the pinned SDK CEL. They pin the de-duplication keys of
// the XProtect rule and that XProtect's own service is recognised whatever the
// letter case of its name. The EventProcessor playground separately runs the
// real parser and alert plugins.
import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

func macVolumeRaw(t *testing.T, process, subsystem, message string) string {
	t.Helper()
	event := map[string]any{
		"timestamp": "2026-09-29T10:00:00Z", "process": process, "process_identifier": 123, "thread_identifier": 456,
		"activity_identifier": 0, "class_name": "OSLogEntryLog", "level": "default", "message": message,
		"store_category": "undefined", "sender": process,
	}
	if subsystem != "" {
		event["subsystem"] = subsystem
		event["category"] = "xprotect"
	}
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestMacOSXProtectAlertVolume(t *testing.T) {
	cfg, cache := macConfig(t), plugins.NewCELCache("macos-alert-volume")
	r := macRules(t)["xprotect_evasion"]
	if r == nil {
		t.Fatal("missing xprotect_evasion")
	}
	if want := []string{"dataSource", "adversary.process"}; len(r.GroupBy) != 0 || !reflect.DeepEqual(r.DeduplicateBy, want) {
		t.Fatalf("grouping got groupBy %v deduplicateBy %v, want deduplicateBy %v", r.GroupBy, r.DeduplicateBy, want)
	}
	const rules = "Using XProtect rules location: /var/protected/xprotect/XProtect.bundle/Contents/Resources/XProtect.yara"
	for _, tc := range []struct {
		name string
		raw  string
		want bool
	}{
		// XProtect's own service loading its rules, in the spelling macOS logs and the documented one.
		{"XprotectService loads its rules", macVolumeRaw(t, "XprotectService", "com.apple.xprotect", rules), false},
		{"XProtectService loads its rules", macVolumeRaw(t, "XProtectService", "com.apple.xprotect", rules), false},
		{"another process touches the rules", macVolumeRaw(t, "bash", "", "cp /tmp/x /Library/Apple/System/Library/CoreServices/XProtect.bundle/Contents/Resources/XProtect.yara"), true},
		{"a process deletes the rules", macVolumeRaw(t, "XprotectService", "com.apple.xprotect", "delete /var/protected/xprotect/XProtect.bundle/Contents/Resources/XProtect.yara"), true},
		{"MRT terminated", macVolumeRaw(t, "MRT", "", "terminate"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := macParse(t, cfg, tc.raw, "mac-lab", cache)
			got, err := cache.Eval(r.Where, out)
			if err != nil || got != tc.want {
				t.Fatalf("where got %v (%v), want %v for %s", got, err, tc.want, out)
			}
			if !tc.want {
				return
			}
			// adversary: origin, so adversary.process is the event's origin.process.
			for key, path := range map[string]string{"dataSource": "dataSource", "adversary.process": "origin.process"} {
				if v := gjson.Get(out, path); v.Type != gjson.String || v.String() == "" {
					t.Fatalf("de-duplication key %s (%s) does not resolve in %s", key, path, out)
				}
			}
		})
	}
}
