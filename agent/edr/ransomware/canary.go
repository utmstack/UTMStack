package ransomware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	store  canaryStore
	opmu   sync.Mutex // serializes descriptor ownership during load/plant/close
	mu     sync.RWMutex
	set    map[string]bool // normalized path → true
	paths  map[string]canaryFileIdentity
	ids    map[canaryFileIdentity]*os.File // live pins prevent Unix inode reuse
	closed bool
	mounts map[string]bool // optional Linux mount identities, checked on pinned directories
}

const maxActiveCanaries = 4096

func NewManager(store canaryStore) *Manager {
	return &Manager{store: store, set: map[string]bool{}, paths: map[string]canaryFileIdentity{}, ids: map[canaryFileIdentity]*os.File{}}
}

type canaryFileIdentity struct{ device, inode uint64 }

func prepareCanaryPin(f *os.File) (canaryFileIdentity, *os.File, error) {
	info, err := f.Stat()
	if err != nil {
		return canaryFileIdentity{}, nil, err
	}
	id, hasID := canaryFileID(info)
	if !hasID {
		if canaryPinRequired {
			return id, nil, errors.New("canary file identity unavailable")
		}
		return id, nil, nil
	}
	pin, err := duplicateCanaryFile(f)
	return id, pin, err
}

// remember takes ownership of pin even on failure. The caller holds opmu.
func (m *Manager) remember(path string, id canaryFileIdentity, pin *os.File) error {
	defer func() {
		if pin != nil {
			pin.Close()
		}
	}()
	key := normCanaryPath(path)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return os.ErrClosed
	}
	if !m.set[key] && len(m.set) >= maxActiveCanaries {
		return errors.New("active canary count limit exceeded")
	}
	oldID, hadID := m.paths[key]
	m.set[key] = true
	if pin != nil {
		m.paths[key] = id
		if m.ids[id] == nil {
			m.ids[id], pin = pin, nil
		}
	} else {
		delete(m.paths, key)
	}
	if hadID && oldID != id {
		used := false
		for _, current := range m.paths {
			if current == oldID {
				used = true
				break
			}
		}
		if !used {
			if oldPin := m.ids[oldID]; oldPin != nil {
				oldPin.Close()
			}
			delete(m.ids, oldID)
		}
	}
	return nil
}

// Close retires all memberships and releases owned Unix pins. The service
// calls it after joining its native event readers. Calls are idempotent.
func (m *Manager) Close() error {
	m.opmu.Lock()
	defer m.opmu.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	var failures []error
	for _, pin := range m.ids {
		failures = append(failures, pin.Close())
	}
	m.set, m.paths, m.ids = nil, nil, nil
	return errors.Join(failures...)
}

// SetMountScope confines Linux canary creation/adoption to the current identities
// of declared fanotify mounts. Every use checks the pinned directory, so a mount
// replacement between configuration and planting fails closed. It is structural.
func (m *Manager) SetMountScope(mounts []string) error {
	m.opmu.Lock()
	defer m.opmu.Unlock()
	if m.closed {
		return os.ErrClosed
	}
	// A failed configuration must not accidentally leave unrestricted planting.
	m.mu.Lock()
	m.mounts = map[string]bool{}
	m.mu.Unlock()
	if len(mounts) == 0 || len(mounts) > 16 {
		return errors.New("canary mount scope requires 1..16 declared mounts")
	}
	allowed := map[string]bool{}
	for _, path := range mounts {
		d, err := openCanaryDir(path)
		if err != nil {
			return err
		}
		key, err := d.mountKey()
		d.Close()
		if err != nil {
			return err
		}
		allowed[key] = true
	}
	m.mu.Lock()
	m.mounts = allowed
	m.mu.Unlock()
	return nil
}

func (m *Manager) openDirectory(path string) (*canaryDir, error) {
	d, err := openCanaryDir(path)
	if err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.mounts != nil {
		key, err := d.mountKey()
		if err != nil || !m.mounts[key] {
			d.Close()
			return nil, errors.New("canary directory is not on a declared monitored mount")
		}
	}
	return d, nil
}

func (m *Manager) directoryStillNamed(d *canaryDir, path string) error {
	current, err := m.openDirectory(path)
	if err != nil {
		return err
	}
	defer current.Close()
	before, err := d.file.Stat()
	if err != nil {
		return err
	}
	after, err := current.file.Stat()
	if err != nil || !os.SameFile(before, after) {
		return errors.New("canary directory pathname changed during operation")
	}
	return nil
}

// Load rehydrates the in-memory membership set from persisted records.
func (m *Manager) Load() error {
	return m.load(nil)
}

// LoadCovered revalidates only recorded canaries in the current explicit
// planting directories. Old, missing, modified and out-of-scope records never
// inflate the active count. An empty scope deliberately loads nothing.
func (m *Manager) LoadCovered(dirs []string) error {
	allowed := make(map[string]bool, len(dirs))
	for _, dir := range dirs {
		allowed[normCanaryPath(filepath.Clean(dir))] = true
	}
	return m.load(allowed)
}

func (m *Manager) load(allowed map[string]bool) error {
	m.opmu.Lock()
	defer m.opmu.Unlock()
	if m.closed {
		return os.ErrClosed
	}
	recs, err := m.store.ListCanaries()
	if err != nil {
		return err
	}
	next := map[string]bool{}
	nextPaths := map[string]canaryFileIdentity{}
	nextIDs := map[canaryFileIdentity]*os.File{}
	// Reuse existing pins for unchanged identities. A failed refresh closes
	// only its new pins and leaves the published membership usable.
	committed := false
	defer func() {
		if !committed {
			for id, pin := range nextIDs {
				if m.ids[id] != pin {
					pin.Close()
				}
			}
		}
	}()
	for _, r := range recs {
		dir := filepath.Dir(r.Path)
		if allowed != nil && !allowed[normCanaryPath(dir)] {
			continue
		}
		key := normCanaryPath(r.Path)
		if next[key] {
			continue // one admitted identity per normalized persisted path
		}
		d, err := m.openDirectory(dir)
		if err != nil {
			continue
		}
		f, err := d.open(filepath.Base(r.Path), false)
		if err == nil {
			if verifyCanary(f, r.SHA256) == nil {
				if m.directoryStillNamed(d, dir) == nil {
					if info, statErr := f.Stat(); statErr == nil {
						if !next[key] && len(next) >= maxActiveCanaries {
							f.Close()
							d.Close()
							return errors.New("active canary count limit exceeded")
						}
						next[key] = true
						if id, ok := canaryFileID(info); ok {
							nextPaths[key] = id
							if nextIDs[id] == nil {
								pin := m.ids[id]
								if pin == nil {
									pin, err = duplicateCanaryFile(f)
									if err != nil {
										f.Close()
										d.Close()
										return fmt.Errorf("pinning recorded canary: %w", err)
									}
								}
								nextIDs[id] = pin
							}
						} else if canaryPinRequired {
							f.Close()
							d.Close()
							return errors.New("recorded canary file identity unavailable")
						}
					}
				}
			}
			f.Close()
		}
		d.Close()
	}
	m.mu.Lock()
	oldIDs := m.ids
	m.set, m.paths, m.ids = next, nextPaths, nextIDs
	committed = true
	m.mu.Unlock()
	for id, pin := range oldIDs {
		if nextIDs[id] != pin {
			pin.Close()
		}
	}
	return nil
}

// Plant creates new canaries exclusively through pinned, non-symlink directories.
// Existing names are never overwritten: only a recorded canary whose actual
// content still proves ownership can be adopted. The returned count includes
// safely adopted canaries. Partial failures are returned without deleting files
// whose path another process could have replaced.
func (m *Manager) Plant(dirs []string, perDir int) (int, error) {
	m.opmu.Lock()
	defer m.opmu.Unlock()
	if m.closed {
		return 0, os.ErrClosed
	}
	if perDir < 1 {
		perDir = 1
	}
	if perDir > 30 || len(dirs) > 1024 || len(dirs)*perDir > maxActiveCanaries {
		return 0, errors.New("canary planting exceeds 30 per directory or 4096 total")
	}
	records, err := m.store.ListCanaries()
	if err != nil {
		return 0, err
	}
	owned := map[string]string{}
	for _, rec := range records {
		owned[normCanaryPath(rec.Path)] = rec.SHA256
	}
	planted := 0
	var failures []error
	for _, dir := range dirs {
		d, err := m.openDirectory(dir)
		if err != nil {
			failures = append(failures, fmt.Errorf("canary directory %q: %w", dir, err))
			continue
		}
		for _, spec := range PlanCanaries([]string{dir}, perDir) {
			full := filepath.Join(dir, spec.Name)
			m.mu.RLock()
			atLimit := !m.set[normCanaryPath(full)] && len(m.set) >= maxActiveCanaries
			m.mu.RUnlock()
			if atLimit {
				failures = append(failures, errors.New("active canary count limit exceeded"))
				continue
			}
			f, err := d.open(spec.Name, true)
			if errors.Is(err, os.ErrExist) {
				if expected := owned[normCanaryPath(full)]; expected != "" {
					f, err = d.open(spec.Name, false)
					if err == nil {
						err = verifyCanary(f, expected)
						var id canaryFileIdentity
						var pin *os.File
						if err == nil {
							id, pin, err = prepareCanaryPin(f)
						}
						f.Close()
						if err == nil {
							err = m.directoryStillNamed(d, dir)
						}
						if err == nil {
							err = m.remember(full, id, pin)
						} else if pin != nil {
							pin.Close()
						}
					}
				}
				if err == nil {
					planted++
					continue
				}
			} else if err == nil {
				_, err = io.WriteString(f, canaryBody)
				if err == nil {
					err = f.Sync()
				}
				var id canaryFileIdentity
				var pin *os.File
				if err == nil {
					id, pin, err = prepareCanaryPin(f)
				}
				closeErr := f.Close()
				if err == nil {
					err = closeErr
				}
				if err == nil {
					err = m.directoryStillNamed(d, dir)
				}
				if err == nil {
					rec := cache.CanaryRecord{Path: full, SHA256: canaryHash(), Volume: filepath.VolumeName(full)}
					err = m.store.StoreCanary(rec)
				}
				if err == nil {
					err = m.remember(full, id, pin)
					if err == nil {
						planted++
						continue
					}
				} else if pin != nil {
					pin.Close()
				}
			}
			failures = append(failures, fmt.Errorf("canary %q unchanged or newly created but not registered: %w", full, err))
		}
		d.Close()
	}
	return planted, errors.Join(failures...)
}

func canaryHash() string {
	sum := sha256.Sum256([]byte(canaryBody))
	return hex.EncodeToString(sum[:])
}

func verifyCanary(f *os.File, recorded string) error {
	if recorded != canaryHash() {
		return errors.New("recorded canary ownership hash is not recognized")
	}
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Size() != int64(len(canaryBody)) {
		return errors.New("recorded canary is missing or has changed type or size")
	}
	b, err := io.ReadAll(io.LimitReader(f, int64(len(canaryBody)+1)))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != recorded {
		return errors.New("recorded canary content has changed")
	}
	return nil
}

func (m *Manager) Contains(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.set[normCanaryPath(path)]
}

// ContainsFileIdentity checks the file pinned by a Linux fanotify event,
// including a hard-link or renamed path. A new file at the old name does not
// inherit the decoy's high-confidence response authorization.
func (m *Manager) ContainsFileIdentity(device, inode uint64) bool {
	if device == 0 || inode == 0 {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ids[canaryFileIdentity{device, inode}] != nil
}

func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.set)
}
