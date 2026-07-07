// edr/netblock/indicator.go
package netblock

import (
	"fmt"
	"net/netip"
	"strings"
)

// Indicator is one normalized blocklist entry.
type Indicator struct {
	Value string // canonical string form
	Type  string // ip | cidr | domain | hostname
	Level int
}

// ParseIndicator validates and canonicalizes a raw feed value. Plan 1 handles
// ip and cidr; domain/hostname are accepted (lowercased, trimmed) for Plan 2.
func ParseIndicator(value, typ string, level int) (Indicator, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return Indicator{}, fmt.Errorf("empty indicator")
	}
	switch typ {
	case "ip":
		a, err := netip.ParseAddr(v)
		if err != nil {
			return Indicator{}, fmt.Errorf("bad ip %q: %w", v, err)
		}
		return Indicator{Value: a.String(), Type: "ip", Level: level}, nil
	case "cidr":
		p, err := netip.ParsePrefix(v)
		if err != nil {
			return Indicator{}, fmt.Errorf("bad cidr %q: %w", v, err)
		}
		return Indicator{Value: p.Masked().String(), Type: "cidr", Level: level}, nil
	case "domain", "hostname":
		return Indicator{Value: strings.ToLower(strings.TrimSuffix(v, ".")), Type: typ, Level: level}, nil
	default:
		return Indicator{}, fmt.Errorf("unknown indicator type %q", typ)
	}
}
