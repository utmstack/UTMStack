package main

// Fabricated FortiGate SSL-VPN failure lines run through the step-by-step
// FortiGate filter model (fortiParse) and the pinned SDK CEL. They pin the
// de-duplication keys that keep the VPN brute force rule from opening a new
// alert for every user name one source address tries. The rule's history
// threshold is pinned separately by TestFortiGateSDKHistory.
import (
	"reflect"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

func TestFortiGateVPNBruteForceAlertVolume(t *testing.T) {
	cfg, cache := fortiConfig(t), plugins.NewCELCache("fortigate-alert-volume")
	r := fortiRules(t)["fortigate_vpn_brute_force"]
	if r == nil {
		t.Fatal("missing fortigate_vpn_brute_force")
	}
	want := []string{"dataSource", "lastEvent.event.devid", "lastEvent.event.vd", "adversary.ip"}
	if len(r.GroupBy) != 0 || !reflect.DeepEqual(r.DeduplicateBy, want) {
		t.Fatalf("grouping got groupBy %v deduplicateBy %v, want deduplicateBy %v", r.GroupBy, r.DeduplicateBy, want)
	}
	// Two user names from one address: the keys must be the same, so the second
	// failure after the threshold is a duplicate rather than a new alert.
	var keys []string
	for _, user := range []string{"vpn-user", "another-user"} {
		raw := `<189>date=2026-09-29 time=10:00:00 devid="FGT-LAB" vd="root" type="event" subtype="vpn" logid="0101039426" action="ssl-login-fail" logdesc="SSL VPN login fail" remip=203.0.113.10 user="` + user + `" group="Remote Users" msg="SSL user failed to log in"`
		out := fortiParse(t, cfg, raw, "firewall-lab", cache)
		if got, err := cache.Eval(r.Where, out); err != nil || !got {
			t.Fatalf("where got %v (%v) for %s", got, err, out)
		}
		key := ""
		for key2, path := range map[string]string{"dataSource": "dataSource", "lastEvent.event.devid": "event.devid", "lastEvent.event.vd": "event.vd", "adversary.ip": "origin.ip"} {
			v := gjson.Get(out, path)
			if v.Type != gjson.String || v.String() == "" {
				t.Fatalf("de-duplication key %s (%s) does not resolve in %s", key2, path, out)
			}
		}
		for _, path := range []string{"dataSource", "event.devid", "event.vd", "origin.ip"} {
			key += gjson.Get(out, path).String() + "|"
		}
		if gjson.Get(out, "origin.user").String() != user {
			t.Fatalf("user %s not parsed in %s", user, out)
		}
		keys = append(keys, key)
	}
	if keys[0] != keys[1] {
		t.Fatalf("keys differ by user name: %v", keys)
	}
}
