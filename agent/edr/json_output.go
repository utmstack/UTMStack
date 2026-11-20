package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// This file implements the `--json` mode for the EDR management CLI. When the
// flag is present, every verb emits a single-line envelope to stdout in place
// of its human text. The shape is stable (the agent parses it in Y1.3):
//
//	{"ok":true,"data":{...}}
//	{"ok":false,"error":"...","data":{...}}
//
// `error` is omitted on success (omitempty) and `data` is always an object
// (never null). In JSON mode a failure prints the envelope to stdout AND exits
// non-zero; the human path is untouched.

// jsonMode is set in main() by parseJSONFlag before any verb runs.
var jsonMode bool

// parseJSONFlag reports whether args contains --json and returns a copy of args
// with every occurrence removed, so the flag can appear at any position without
// polluting verb argument slices (e.g. `config set key --json value` must not
// turn the value into "--json value").
func parseJSONFlag(args []string) (bool, []string) {
	found := false
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--json" {
			found = true
			continue
		}
		out = append(out, a)
	}
	return found, out
}

// jsonEnvelope is the stable output body for --json mode.
type jsonEnvelope struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data"`
}

// emitJSON prints the envelope when --json is active and does nothing
// otherwise, so a verb handler can call it unconditionally after producing its
// data.
func emitJSON(ok bool, errStr string, data interface{}) {
	if !jsonMode {
		return
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	b, _ := json.Marshal(jsonEnvelope{OK: ok, Error: errStr, Data: data})
	fmt.Println(string(b))
}

// failJSON emits an error envelope and exits non-zero; used only on the --json
// path (the human path keeps its own exitOn handling).
func failJSON(errStr string) {
	emitJSON(false, errStr, nil)
	os.Exit(1)
}
