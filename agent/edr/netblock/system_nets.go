package netblock

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/netip"
	"sort"
)

// Address order, duplicates and IPv4-mapped spellings do not change a safety
// policy. Preserve the distinct management/resolver/gateway roles in its key.
func systemNetsIdentity(sys SystemNets) string {
	var lists [3][]string
	for i, addresses := range [][]netip.Addr{sys.ServerHosts, sys.Resolvers, sys.Gateways} {
		seen := make(map[string]bool, len(addresses))
		lists[i] = make([]string, 0, len(addresses))
		for _, address := range addresses {
			name := address.Unmap().String()
			if !seen[name] {
				seen[name] = true
				lists[i] = append(lists[i], name)
			}
		}
		sort.Strings(lists[i])
	}
	data, _ := json.Marshal(lists)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
