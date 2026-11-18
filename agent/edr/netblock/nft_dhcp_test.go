package netblock

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// These exceptions are for an isolated DHCP client, not a DHCP server. Port
// numbers alone must not exempt the same pair in the other address family.
func TestIsolationDHCPClientDirectionsAndFamilies(t *testing.T) {
	p, err := isolationPlan("test", testIsolation(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"input4":  "meta nfproto ipv4 udp sport 67 udp dport 68 accept",
		"output4": "meta nfproto ipv4 udp sport 68 udp dport 67 accept",
		"input6":  "meta nfproto ipv6 udp sport 547 udp dport 546 accept",
		"output6": "meta nfproto ipv6 udp sport 546 udp dport 547 accept",
	}
	seen := map[string]bool{}
	for _, r := range p.Rules {
		if !strings.HasPrefix(r.Text, "udp sport ") && !strings.HasPrefix(r.Text, "meta nfproto ") {
			continue
		}
		matched := false
		for key, text := range want {
			if r.Text != text || !strings.HasPrefix(key, r.Chain) {
				continue
			}
			family, src, dst := "ipv4", 67, 68
			if strings.HasSuffix(key, "6") {
				family, src, dst = "ipv6", 547, 546
			}
			if r.Chain == "output" {
				src, dst = dst, src
			}
			expr := []any{matchExpr(metaExpr("nfproto"), family), matchExpr(payloadExpr("udp", "sport"), src), matchExpr(payloadExpr("udp", "dport"), dst), verdictExpr("accept")}
			if !reflect.DeepEqual(r.Expr, expr) || seen[key] {
				t.Errorf("unexpected or duplicate DHCP exception: %+v", r)
			}
			seen[key], matched = true, true
		}
		if !matched {
			t.Errorf("broad or reversed DHCP exception: chain=%s rule=%s", r.Chain, r.Text)
		}
	}
	for key := range want {
		if !seen[key] {
			t.Errorf("missing isolated-client exception %s: %s", key, want[key])
		}
	}
}
