// edr/netblock/dnsparse_test.go
package netblock

import (
	"net/netip"
	"testing"
)

func TestParseDNSResults(t *testing.T) {
	in := "type: 5 e.cdn.net;type: 1 1.2.3.4;type: 28 2606:4700:4700::1111;type: 1 5.6.7.8;"
	got := ParseDNSResults(in)
	want := []string{"1.2.3.4", "2606:4700:4700::1111", "5.6.7.8"}
	if len(got) != len(want) {
		t.Fatalf("got %d addrs %v, want %d", len(got), got, len(want))
	}
	for i, w := range want {
		if got[i] != netip.MustParseAddr(w) {
			t.Errorf("addr[%d] = %v, want %v", i, got[i], w)
		}
	}
	if len(ParseDNSResults("type: 5 only.cname.com;garbage;")) != 0 {
		t.Fatal("no A/AAAA → empty")
	}
}

// TestParseDNSResultsBareFormat covers the format VM-observed on Windows 11
// (build 26200): QueryResults is a bare ";"-separated IP list ("7.7.7.7;"), with
// CNAME chains prepended as bare hostnames. The parser must extract the IPs and
// skip the names.
func TestParseDNSResultsBareFormat(t *testing.T) {
	// Bare single IP (the exact string seen on the VM for a wildcard resolve).
	if got := ParseDNSResults("7.7.7.7;"); len(got) != 1 || got[0] != netip.MustParseAddr("7.7.7.7") {
		t.Fatalf("bare IP: got %v", got)
	}
	// CNAME (bare hostname) then A record, bare format.
	got := ParseDNSResults("edge.cdn.example.net;93.184.215.14;")
	if len(got) != 1 || got[0] != netip.MustParseAddr("93.184.215.14") {
		t.Fatalf("bare cname+A: got %v", got)
	}
	// Bare IPv4 + IPv6 mix.
	if got := ParseDNSResults("1.2.3.4;2606:4700::1111;"); len(got) != 2 {
		t.Fatalf("bare v4+v6: got %v", got)
	}
}
