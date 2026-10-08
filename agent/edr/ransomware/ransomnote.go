package ransomware

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// strongNoteTokens are substrings that make a basename a ransom note on any
// extension. Deliberate choices: "ransom" is a whole distinctive word;
// "bitcoin"/"wallet"/"payment" are note vocabulary; "we_have"/"all_your"/
// "your_files"/"h2w" appear in lock-in messages. "decrypt" is NOT strong:
// it also appears in legit "decryptor"-style tools, so it stays weak and is
// gated on a text-ish extension.
var strongNoteTokens = []string{
	"ransom", "bitcoin", "wallet", "payment",
	"we_have", "all_your", "your_files", "h2w",
}

// weakNoteTokens only count when the file has a text-ish extension — a
// binary README.bin or how_to_decrypt.bmp is not a note.
var weakNoteTokens = []string{"readme", "how_to", "deleted", "decrypt", "recovery"}

// textishExts are the extensions a weak token will fire on (or no extension).
var textishExts = map[string]bool{
	".txt": true, ".md": true, ".html": true, ".htm": true, ".rtf": true, ".dat": true,
}

// prefixNoteForms are explicit prefix shapes that fire regardless of
// extension. Attackers consistently keep README*/RECOVER*/RESTORE* at the
// start of the name even when the token list drifts (e.g. RECOVER.md,
// RESTORE.YOUR.DATA.md).
var prefixNoteForms = []string{"readme", "recover", "restore"}

// IsRansomNote fuzzy-matches a basename against known ransom-note naming
// patterns (case-insensitive). Three tiers:
//  1. explicit prefix forms (readme*/recover*/restore*) fire on any extension,
//  2. strong substrings (ransom/bitcoin/payment/…) fire on any extension,
//  3. weak substrings (how_to/deleted/decrypt/…) fire only on text-ish exts.
func IsRansomNote(basename string) bool {
	name := strings.ToLower(basename)
	ext := strings.ToLower(filepath.Ext(basename))
	textish := ext == "" || textishExts[ext]

	for _, p := range prefixNoteForms {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	for _, tok := range strongNoteTokens {
		if strings.Contains(name, tok) {
			return true
		}
	}
	if textish {
		for _, tok := range weakNoteTokens {
			if strings.Contains(name, tok) {
				return true
			}
		}
	}
	return false
}

// RansomNoteSensor fires when a mutating op lands on a path whose basename
// fuzzy-matches a known ransom-note shape. One signal per (PID, path) — a
// note written in chunks must not fire repeatedly.
type RansomNoteSensor struct {
	cfg   config.EDRConfig
	now   func() time.Time
	mu    sync.Mutex
	fired map[string]bool // "pid\x00path" → already emitted
}

// NewRansomNoteSensor builds the sensor. A nil now falls back to time.Now.
func NewRansomNoteSensor(cfg config.EDRConfig, now func() time.Time) *RansomNoteSensor {
	if now == nil {
		now = time.Now
	}
	return &RansomNoteSensor{cfg: cfg, now: now, fired: map[string]bool{}}
}

// Record ingests one file op and returns evidence only when the path basename
// matches a ransom note and this (PID, path) has not fired before.
func (r *RansomNoteSensor) Record(ev FileEvent) *Evidence {
	if ev.Op != OpWrite && ev.Op != OpCreate && ev.Op != OpRename {
		return nil
	}
	base := filepath.Base(ev.Path)
	if !IsRansomNote(base) {
		return nil
	}
	key := fmt.Sprintf("%d\x00%s", ev.PID, ev.Path)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.fired[key] {
		return nil
	}
	r.fired[key] = true
	return &Evidence{
		PID:    ev.PID,
		Kind:   KindRansomNote,
		Weight: float64(r.cfg.Ransomware.FuzzyWeight("ransom_note")),
		Detail: base,
		TS:     r.now(),
	}
}
