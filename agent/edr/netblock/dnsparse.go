// edr/netblock/dnsparse.go
package netblock

import (
	"net/netip"
	"strings"
)

// ParseDNSResults extracts A/AAAA addresses from a Microsoft-Windows-DNS-Client
// event-3008 QueryResults string. The format is ";"-separated records; on
// Windows 11 (build 26200, VM-confirmed) each record is a BARE address ("5.5.5.5;"),
// while some builds/schemas prefix a "type: <N> " token ("type: 1 5.5.5.5;") and
// intersperse CNAME records ("type: 5 host.cdn.net;" or a bare hostname). We take
// the LAST whitespace-separated field of each record and keep it only if it parses
// as an IP — this handles both formats and skips CNAMEs/unparseable records.
func ParseDNSResults(qr string) []netip.Addr {
	var out []netip.Addr
	for _, rec := range strings.Split(qr, ";") {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		fields := strings.Fields(rec)
		last := fields[len(fields)-1] // "5.5.5.5" or the address after "type: N"
		if a, err := netip.ParseAddr(last); err == nil {
			out = append(out, a.Unmap())
		}
	}
	return out
}
