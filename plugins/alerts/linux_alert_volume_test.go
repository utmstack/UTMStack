package main

// Fabricated journald records run through the Linux raw model and the pinned
// SDK CEL. They pin the de-duplication key that keeps "Audit or Logging Service
// Disabled" from opening a new alert for every matching record: its groupBy
// named origin.* fields, which an alert never carries. The EventProcessor
// playground separately runs the real parser and alert plugins.
import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestLinuxLoggingServiceAlertVolume(t *testing.T) {
	blob, err := utils.ReadPbYaml("../../rules/linux/debian_family/auditd_syslog_disabling.yml")
	if err != nil {
		t.Fatal(err)
	}
	r := new(plugins.Rule)
	if err = protojson.Unmarshal(blob, r); err != nil {
		t.Fatal(err)
	}
	r.Normalize()
	if want := []string{"dataSource"}; len(r.GroupBy) != 0 || !reflect.DeepEqual(r.DeduplicateBy, want) {
		t.Fatalf("grouping got groupBy %v deduplicateBy %v, want deduplicateBy %v", r.GroupBy, r.DeduplicateBy, want)
	}
	cache := plugins.NewCELCache("linux-alert-volume")
	for _, tc := range []struct {
		message string
		want    bool
	}{
		{"systemctl stop auditd", true},
		{"systemctl disable rsyslog.service", true},
		{"systemctl mask systemd-journald.service", true},
		{"systemctl restart rsyslog.service", false},
		{"Started Session 42 of User reviewer.", false},
	} {
		t.Run(tc.message, func(t *testing.T) {
			raw, err := json.Marshal(map[string]any{"MESSAGE": tc.message, "SYSLOG_IDENTIFIER": "sudo", "_HOSTNAME": "web01"})
			if err != nil {
				t.Fatal(err)
			}
			event := linuxActionResultNormalize(t, string(raw))
			got, err := cache.Eval(r.Where, event)
			if err != nil || got != tc.want {
				t.Fatalf("where got %v (%v), want %v for %s", got, err, tc.want, event)
			}
			if got && gjson.Get(event, "dataSource").String() == "" {
				t.Fatalf("de-duplication key dataSource does not resolve in %s", event)
			}
		})
	}
}
