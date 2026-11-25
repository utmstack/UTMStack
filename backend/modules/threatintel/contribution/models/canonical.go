package models

import (
	"encoding/json"
	"sort"
	"strings"
)

// CanonicalJSON renders v as JSON that does not depend on the order of anything that has no meaning:
// object keys come out sorted, as encoding/json already does, and so do the elements of every list.
// The associations of an entity are built from a map, so they come out in a different order on every
// run; two runs over the same data must still be recognised as the same entity.
func CanonicalJSON(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return "", err
	}

	return canonical(generic), nil
}

func canonical(x any) string {
	switch t := x.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		parts := make([]string, 0, len(keys))

		for _, k := range keys {
			key, _ := json.Marshal(k)
			parts = append(parts, string(key)+":"+canonical(t[k]))
		}

		return "{" + strings.Join(parts, ",") + "}"

	case []any:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, canonical(e))
		}

		sort.Strings(parts)

		return "[" + strings.Join(parts, ",") + "]"

	default:
		b, _ := json.Marshal(t)

		return string(b)
	}
}
