// Package jsonflex reads numbers the way the event processor writes them.
//
// The event processor serialises 64-bit integers as JSON strings ("asn": "6849"), the way protojson
// does, so a field the backend declares as a number arrives as text. Declaring it as a number then fails
// the whole document that carries it. Int64 accepts both forms.
package jsonflex

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// Int64 reads a JSON number, or a string holding one. null, an empty string and text that is not a
// number read as 0: a field that cannot be used must not make the rest of the document unreadable.
func Int64(raw json.RawMessage) int64 {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return 0
	}

	if raw[0] == '"' {
		var s string
		if json.Unmarshal(raw, &s) != nil {
			return 0
		}

		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0
		}

		return n
	}

	if n, err := strconv.ParseInt(string(raw), 10, 64); err == nil {
		return n
	}

	// A number written with a fraction or an exponent, such as 6849.0 or 6.849e3.
	var f float64
	if json.Unmarshal(raw, &f) != nil {
		return 0
	}

	return int64(f)
}
