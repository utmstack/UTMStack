package main

import (
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// alertGroupingValue resolves the fields that rules can use in groupBy and
// deduplicateBy. The wire Alert has events, while the indexed document exposes
// the final event as lastEvent. Resolve that alias before building the search,
// using the same event as newAlert. Keep the search field itself as lastEvent.*.
func alertGroupingValue(alertJSON, field string) gjson.Result {
	if strings.HasPrefix(field, "lastEvent.") {
		count := gjson.Get(alertJSON, "events.#").Int()
		if count == 0 {
			return gjson.Result{}
		}
		field = "events." + strconv.FormatInt(count-1, 10) + "." + strings.TrimPrefix(field, "lastEvent.")
	}
	return gjson.Get(alertJSON, field)
}

// A map or array is not an exact-match grouping term. In particular, finding only
// a non-scalar must not enable a name-only search that groups unrelated alerts.
func scalarGroupingValue(value gjson.Result) (any, bool) {
	switch value.Type {
	case gjson.String:
		return value.String(), true
	case gjson.Number:
		return value.Float(), true
	case gjson.True, gjson.False:
		return value.Bool(), true
	default:
		return nil, false
	}
}
