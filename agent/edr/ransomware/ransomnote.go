package ransomware

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// RansomNoteSensor fires on the appearance of a file whose name matches a
// ransom-note token. Strong tokens (exact note names) fire on any
// extension; weak tokens fire only on text-like extensions; prefix forms
// (readme*, recover*, restore*) fire regardless of extension. The token
// list is built from the defaults + cfg.Ransomware.Behavior.RansomNoteExts.
type RansomNoteSensor struct {
	cfg     config.EDRConfig
	now     func() time.Time
	strong  []string
	weak    []string
	prefix  []string
	textExt map[string]bool
	mu      sync.Mutex
	// lastFire maps PID → the last emission time, for the per-PID cooldown.
	lastFire map[int]time.Time
}

// strongNoteTokens are exact (or prefix) names that identify a ransom note
// on their own, on any extension. "bitcoin" stays strong: it is not a common
// legitimate filename token. "wallet" and "payment" are deliberately NOT
// strong — they appear in browser wallet caches, e-commerce flows, and
// crypto tools, so they fire only on text-like extensions via
// weakNoteTokens below.
var strongNoteTokens = []string{
	"howto", "h0wt0", "readme", "read_me",
	"recover", "recovery", "decrypt", "decryption",
	"restore", "ransom", "bitcoin",
	"we_have", "all_your", "your_files", "h2w",
}

// weakNoteTokens identify a note only when the file is text-like — these
// names appear in legitimate software far more often.
var weakNoteTokens = []string{
	"instructions", "warning", "notice", "message", "info",
	"please", "important", "urgent",
	"wallet", "payment",
}

// defaultTextExts are extensions a note is plausibly written in.
var defaultTextExts = map[string]bool{
	".txt": true, ".md": true, ".markdown": true, ".html": true, ".htm": true,
	".rtf": true, ".text": true, ".log": true, ".csv": true,
}

// NewRansomNoteSensor builds the sensor. A nil now falls back to time.Now.
// Behavior.RansomNoteExts extends the default token table (each entry is
// matched case-insensitively against the lowercased file name).
func NewRansomNoteSensor(cfg config.EDRConfig, now func() time.Time) *RansomNoteSensor {
	if now == nil {
		now = time.Now
	}
	extra := map[string]bool{}
	for _, t := range cfg.Ransomware.Behavior.RansomNoteExts {
		if t != "" {
			extra[strings.ToLower(t)] = true
		}
	}
	s := &RansomNoteSensor{
		cfg:      cfg,
		now:      now,
		textExt:  map[string]bool{},
		lastFire: map[int]time.Time{},
	}
	for e := range defaultTextExts {
		s.textExt[e] = true
	}
	for _, t := range strongNoteTokens {
		s.strong = append(s.strong, t)
	}
	for _, t := range weakNoteTokens {
		s.weak = append(s.weak, t)
	}
	sort.Strings(s.strong)
	sort.Strings(s.weak)
	// Strong tokens also match by prefix ("readme.txt"); caller-supplied
	// tokens ending in "*" are treated as prefixes too.
	for _, t := range []string{"readme", "recover", "restore"} {
		s.prefix = append(s.prefix, t)
	}
	for t := range extra {
		if strings.HasSuffix(t, "*") {
			s.prefix = append(s.prefix, strings.TrimSuffix(t, "*"))
		} else {
			s.strong = append(s.strong, t)
		}
	}
	sort.Strings(s.prefix)
	sort.Strings(s.strong)
	return s
}

// matchesNote returns the matched token if the path's basename looks like a
// ransom note, or "" otherwise.
func (r *RansomNoteSensor) matchesNote(path string) string {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(base))
	text := r.textExt[ext]

	for _, tok := range r.strong {
		if base == tok || strings.Contains(base, tok) {
			return tok
		}
	}
	for _, tok := range r.prefix {
		if strings.HasPrefix(base, tok) {
			return tok
		}
	}
	if text {
		for _, tok := range r.weak {
			if base == tok || strings.Contains(base, tok) {
				return tok
			}
		}
	}
	return ""
}

// Record returns evidence if the event is a write or rename that places a
// file whose name matches a note token. Cooldown: one fire per PID per
// CooldownMs. OpCreate is never considered (an open is not a modification),
// so a note dropped as new content arrives as OpWrite.
func (r *RansomNoteSensor) Record(ev FileEvent) *Evidence {
	if ev.Op != OpWrite && ev.Op != OpRename {
		return nil
	}
	tok := r.matchesNote(ev.Path)
	if tok == "" {
		return nil
	}
	b := behavior(r.cfg)
	cooldown := time.Duration(b.CooldownMs) * time.Millisecond
	now := r.now()

	r.mu.Lock()
	defer r.mu.Unlock()
	if last, ok := r.lastFire[ev.PID]; ok && now.Sub(last) < cooldown {
		return nil
	}
	r.lastFire[ev.PID] = now
	return &Evidence{
		PID:    ev.PID,
		Kind:   KindRansomNote,
		Weight: fuzzyWeight(r.cfg, KindRansomNote),
		Detail: "note token " + tok + " in " + filepath.Base(ev.Path),
		TS:     now,
	}
}
