package main

import (
	"regexp"
	"strconv"
	"strings"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/tidwall/gjson"
)

var groupingArrayIndex = regexp.MustCompile(`\.[0-9]+(\.|$)`)

// addAlertGroupingTerms keeps indexed query paths separate from wire value
// paths. Both grouping and deduplication use this same query construction.
// False means there is no usable key: callers must not run a name-only search.
func addAlertGroupingTerms(builder *sdkos.BoolBuilder, alertJSON string, fields []string) bool {
	added := false
	for _, field := range fields {
		field = strings.TrimSuffix(field, ".keyword")
		value, ok := scalarGroupingValue(alertGroupingValue(alertJSON, field))
		if !ok {
			continue
		}
		// OpenSearch flattens array elements under the indexed field name.
		searchField := groupingArrayIndex.ReplaceAllStringFunc(field, func(index string) string {
			if strings.HasSuffix(index, ".") {
				return "."
			}
			return ""
		})
		builder.FilterTerm(searchField, value)
		added = true
	}
	return added
}

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
