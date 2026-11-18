package netblock

import (
	"net/netip"
	"testing"
)

func TestSystemNetIdentityIsSetBasedAndRoleAware(t *testing.T) {
	a, b := netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("2001:db8::1")
	first := SystemNets{ServerHosts: []netip.Addr{a, b}, Resolvers: []netip.Addr{a}}
	equivalent := SystemNets{ServerHosts: []netip.Addr{b, netip.MustParseAddr("::ffff:192.0.2.1"), a}, Resolvers: []netip.Addr{a, a}, Gateways: []netip.Addr{}}
	if systemNetsIdentity(first) != systemNetsIdentity(equivalent) {
		t.Fatal("address ordering/duplicates/mapped IPv4 changed safety identity")
	}
	if systemNetsIdentity(first) == systemNetsIdentity(SystemNets{ServerHosts: []netip.Addr{a}, Resolvers: []netip.Addr{a, b}}) {
		t.Fatal("management and resolver roles collapsed")
	}
	if systemNetsIdentity(first) == systemNetsIdentity(SystemNets{ServerHosts: []netip.Addr{a, b}, Resolvers: []netip.Addr{a}, Gateways: []netip.Addr{b}}) {
		t.Fatal("gateway exemption missing from identity")
	}
}
