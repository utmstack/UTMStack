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

// NormalizeInts returns the document with the named keys holding numbers, wherever they appear in it: a
// value written as a string that is a whole number ("641") becomes the number. The keys are the ones a
// schema declares as 64-bit integers; anything else is left as it is, so a text field that happens to
// look like a number is never changed. A document that is not valid JSON is returned unchanged.
func NormalizeInts(raw json.RawMessage, keys ...string) json.RawMessage {
	var doc any
	if json.Unmarshal(raw, &doc) != nil {
		return raw
	}

	named := make(map[string]bool, len(keys))
	for _, k := range keys {
		named[k] = true
	}

	var walk func(v any)

	walk = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			for k, val := range t {
				if s, ok := val.(string); ok && named[k] {
					if n, err := strconv.ParseInt(s, 10, 64); err == nil {
						t[k] = n

						continue
					}
				}

				walk(val)
			}
		case []any:
			for _, e := range t {
				walk(e)
			}
		}
	}

	walk(doc)

	out, err := json.Marshal(doc)
	if err != nil {
		return raw
	}

	return out
}
