package ransomware

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/utmstack/UTMStack/agent/edr/cache"
)

// canaryExts and canaryWords generate decoy names that (a) look like real
// user documents so an encryptor targets them, and (b) sort early in a
// directory listing (lead with digits) so an alphabetical walk hits them
// before real data. We deliberately avoid the "~$" Office lock-file prefix,
// which many encryptors skip as a temp file.
var canaryExts = []string{"xlsx", "docx", "pdf"}
var canaryWords = []string{"accounts", "passwords", "backup_keys", "payroll", "clients", "contracts"}

// canaryBody is inert, low-entropy filler so a canary reads as an ordinary
// document (not itself flagged as high-entropy).
const canaryBody = "CONFIDENTIAL - internal records. Do not distribute. " +
	"This document is retained for archival purposes only.\n"

type CanarySpec struct {
	Dir  string
	Name string
}

// PlanCanaries produces perDir deterministic decoy specs for each dir.
func PlanCanaries(dirs []string, perDir int) []CanarySpec {
	if perDir < 1 {
		perDir = 1
	}
	var out []CanarySpec
	for _, d := range dirs {
		for i := 0; i < perDir; i++ {
			word := canaryWords[i%len(canaryWords)]
			ext := canaryExts[i%len(canaryExts)]
			// e.g. "00__accounts.xlsx", "01__passwords.docx"
			name := "0" + string(rune('0'+i%10)) + "__" + word + "." + ext
			out = append(out, CanarySpec{Dir: d, Name: name})
		}
	}
	return out
}

type canaryStore interface {
	StoreCanary(cache.CanaryRecord) error
	ListCanaries() ([]cache.CanaryRecord, error)
}

// Manager plants and tracks canary files and answers membership queries.
type Manager struct {
	store canaryStore
	mu    sync.RWMutex
	set   map[string]bool // normalized path → true
}

func NewManager(store canaryStore) *Manager {
	return &Manager{store: store, set: map[string]bool{}}
}

func normCanaryPath(p string) string {
	return strings.ToLower(filepath.ToSlash(strings.TrimRight(p, `/\`)))
}

// Load rehydrates the in-memory membership set from persisted records.
func (m *Manager) Load() error {
	recs, err := m.store.ListCanaries()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range recs {
		m.set[normCanaryPath(r.Path)] = true
	}
	return nil
}

// Plant writes perDir canaries into each dir, hides them, persists them, and
// registers membership. Returns the number successfully planted. Directories
// that don't exist are skipped (best-effort; a volume may lack the path).
func (m *Manager) Plant(dirs []string, perDir int) (int, error) {
	planted := 0
	for _, spec := range PlanCanaries(dirs, perDir) {
		if _, err := os.Stat(spec.Dir); err != nil {
			continue
		}
		full := filepath.Join(spec.Dir, spec.Name)
		if err := os.WriteFile(full, []byte(canaryBody), 0o644); err != nil {
			continue
		}
		_ = hideFile(full)
		sum := sha256.Sum256([]byte(canaryBody))
		rec := cache.CanaryRecord{Path: full, SHA256: hex.EncodeToString(sum[:]), Volume: filepath.VolumeName(full)}
		if err := m.store.StoreCanary(rec); err != nil {
			return planted, err
		}
		m.mu.Lock()
		m.set[normCanaryPath(full)] = true
		m.mu.Unlock()
		planted++
	}
	return planted, nil
}

func (m *Manager) Contains(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.set[normCanaryPath(path)]
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.set)
}
