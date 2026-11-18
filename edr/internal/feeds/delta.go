// edr/internal/feeds/delta.go
// Daily deltas: the mirror pulls the full accumulative for the whole fleet, so
// it computes per-feed daily deltas server-side by diffing the previous vs new
// accumulative value list. This yields real add AND del deltas in the exact
// wire format the agent already knows how to apply.
package feeds

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

// Diff returns the set difference of two value lists: adds are values present
// in next but not prev; dels are values present in prev but not next. Both are
// returned sorted so the rendered daily is deterministic regardless of input
// order.
func Diff(prev, next []string) (adds, dels []string) {
	prevSet := make(map[string]struct{}, len(prev))
	for _, v := range prev {
		prevSet[v] = struct{}{}
	}
	nextSet := make(map[string]struct{}, len(next))
	for _, v := range next {
		nextSet[v] = struct{}{}
	}
	for v := range nextSet {
		if _, ok := prevSet[v]; !ok {
			adds = append(adds, v)
		}
	}
	for v := range prevSet {
		if _, ok := nextSet[v]; !ok {
			dels = append(dels, v)
		}
	}
	sort.Strings(adds)
	sort.Strings(dels)
	return adds, dels
}

// dailyRecord is one line of the agent's daily contract. Field order (value,
// type, op) matches the wire format the agent consumes.
type dailyRecord struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Op    string `json:"op"`
}

// typeFor derives the indicator type for a value in feed name. For "ip" a value
// containing '/' is a "cidr", otherwise an "ip"; for any other feed the type is
// the feed name itself (e.g. "domain", "hostname").
func typeFor(name, v string) string {
	if name == "ip" {
		if strings.Contains(v, "/") {
			return "cidr"
		}
		return "ip"
	}
	return name
}

// RenderDaily emits the plain (uncompressed) ndjson daily artifact: one JSON
// object per line, adds first (sorted) then dels (sorted). Each record is
// marshaled with encoding/json so string escaping is correct.
func RenderDaily(adds, dels []string, name string) []byte {
	adds = append([]string(nil), adds...)
	dels = append([]string(nil), dels...)
	sort.Strings(adds)
	sort.Strings(dels)

	var buf bytes.Buffer
	write := func(values []string, op string) {
		for _, v := range values {
			line, err := json.Marshal(dailyRecord{Value: v, Type: typeFor(name, v), Op: op})
			if err != nil {
				// A struct of plain strings cannot fail to marshal.
				continue
			}
			buf.Write(line)
			buf.WriteByte('\n')
		}
	}
	write(adds, "add")
	write(dels, "del")
	return buf.Bytes()
}
