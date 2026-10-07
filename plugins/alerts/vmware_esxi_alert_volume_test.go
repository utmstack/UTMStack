package main

// ESXi records after header parsing (process and message, as the ESXi filter
// stores them) evaluated with this module's go-sdk CEL. They pin the
// de-duplication keys that keep three ESXi rules from alerting on every record
// of routine activity, and check that the keys resolve. Header parsing of raw
// lines runs in the EventProcessor playground instead.
import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/encoding/protojson"
)

func esxiVolumeRule(t *testing.T, name string) *plugins.Rule {
	t.Helper()
	b, err := utils.ReadPbYaml("../../rules/vmware/vmware-esxi/" + name + ".yml")
	if err != nil {
		t.Fatal(err)
	}
	r := new(plugins.Rule)
	if err = protojson.Unmarshal(b, r); err != nil {
		t.Fatal(err)
	}
	r.Normalize()
	return r
}

func esxiVolumeEvent(t *testing.T, process, message string) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"dataType": "vmware-esxi", "dataSource": "192.0.2.21",
		"event": map[string]any{"process": process, "message": message},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestESXiAlertVolume(t *testing.T) {
	cache := plugins.NewCELCache("esxi-alert-volume")
	for _, tc := range []struct {
		file               string
		dedup              []string
		positive, negative [][2]string
	}{
		{"vm_escape_detection", []string{"dataSource", "lastEvent.event.process"},
			[][2]string{{"Vpxa", "[VpxLRO] -- ERROR lro-1001 -- 5f1c2d3e-0000-4000-8000-000000000001 -- guestOperationsFileManager -- vim.vm.guest.FileManager.listFiles: :vim.fault.FileNotFound"}},
			[][2]string{{"Vpxa", "[VpxLRO] -- BEGIN lro-1002 -- guestOperationsFileManager -- vim.vm.guest.FileManager.deleteFile -- 5f1c2d3e-0000-4000-8000-000000000002"}}},
		{"esxi_syslog_disruption", []string{"dataSource"},
			[][2]string{{"vobd", `[UserLevelCorrelator] 12480247076375us: [vob.user.vmsyslogd.remote.failure] The host "192.0.2.50:7002" has become unreachable. Remote logging to this host has stopped.`}},
			[][2]string{{"Hostd", "[Originator@6876 sub=Statssvc.StatsCollector] Calculated read I/O size 648481 for scsi0:8 is out of range"}}},
		{"esxi_firewall_modification", []string{"dataSource"},
			[][2]string{{"shell", "[shell[1234]]: [root]: esxcli network firewall ruleset set -r sshServer -e true"}},
			[][2]string{{"Hostd", "[Originator@6876 sub=Statssvc.StatsCollector] Calculated read I/O size 648481 for scsi0:8 is out of range"}}},
	} {
		t.Run(tc.file, func(t *testing.T) {
			r := esxiVolumeRule(t, tc.file)
			if len(r.GroupBy) != 0 || !reflect.DeepEqual(r.DeduplicateBy, tc.dedup) {
				t.Fatalf("grouping got groupBy %v deduplicateBy %v, want deduplicateBy %v", r.GroupBy, r.DeduplicateBy, tc.dedup)
			}
			for i, rec := range append(tc.positive, tc.negative...) {
				want := i < len(tc.positive)
				event := esxiVolumeEvent(t, rec[0], rec[1])
				if got, err := cache.Eval(r.Where, event); err != nil || got != want {
					t.Fatalf("record %d: where got %v (%v), want %v for %s", i, got, err, want, event)
				}
				if !want {
					continue
				}
				for _, key := range tc.dedup {
					path := map[string]string{"dataSource": "dataSource", "lastEvent.event.process": "event.process"}[key]
					if v := gjson.Get(event, path); v.Type != gjson.String || v.String() == "" {
						t.Fatalf("de-duplication key %s (%s) does not resolve in %s", key, path, event)
					}
				}
			}
		})
	}
}
