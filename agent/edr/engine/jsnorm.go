package engine

import "github.com/utmstack/UTMStack/agent/edr/jsnorm"

// normalizeJS returns the reference-compatible view of HTML script blocks.
// The bytecode VM uses the same tokenizer for its streaming jsnorm host API.
func normalizeJS(scripts [][][]byte) ([]byte, bool) {
	return jsnorm.NormalizeBlocks(scripts)
}
