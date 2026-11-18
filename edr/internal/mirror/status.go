// edr/internal/mirror/status.go
package mirror

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type SigStatus struct {
	LastSuccess time.Time         `json:"last_success"`
	LastAttempt time.Time         `json:"last_attempt"`
	LastError   string            `json:"last_error"`
	Databases   map[string]string `json:"databases,omitempty"`
}

type FeedStatus struct {
	LastSuccess time.Time `json:"last_success"`
	LastAttempt time.Time `json:"last_attempt"`
	Indicators  int       `json:"indicators"`
	SHA256      string    `json:"sha256"`
	LastError   string    `json:"last_error"`
}

type status struct {
	Signatures SigStatus             `json:"signatures"`
	Feeds      map[string]FeedStatus `json:"feeds"`
	Updated    time.Time             `json:"updated"`
}

// Recorder maintains status.json at the mirror root — the platform's single
// point to monitor mirror freshness. Safe for concurrent use by the signature
// and feed sync loops.
type Recorder struct {
	mu   sync.Mutex
	root string
	st   status
}

// NewRecorder loads any existing root/status.json so prior success fields
// survive a process restart — a failure recorded after restart must not zero a
// feed's (or signatures') last_success/indicators/sha256/databases. A missing
// or unreadable file falls back to empty state; Feeds is always non-nil.
func NewRecorder(root string) *Recorder {
	var st status
	if b, err := os.ReadFile(filepath.Join(root, "status.json")); err != nil || json.Unmarshal(b, &st) != nil {
		st = status{}
	}
	if st.Feeds == nil {
		st.Feeds = map[string]FeedStatus{}
	}
	return &Recorder{root: root, st: st}
}

func (r *Recorder) SigSuccess(dbs map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	r.st.Signatures.LastSuccess = now
	r.st.Signatures.LastAttempt = now
	r.st.Signatures.LastError = ""
	r.st.Signatures.Databases = dbs
	r.write()
}

func (r *Recorder) SigFailure(errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.st.Signatures.LastAttempt = time.Now().UTC()
	r.st.Signatures.LastError = errMsg
	r.write()
}

func (r *Recorder) FeedSuccess(key string, indicators int, sha string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	r.st.Feeds[key] = FeedStatus{LastSuccess: now, LastAttempt: now, Indicators: indicators, SHA256: sha}
	r.write()
}

// FeedFailure records the attempt + error but preserves the previous success
// fields — the mirror content itself is likewise never removed on failure.
func (r *Recorder) FeedFailure(key, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.st.Feeds[key]
	f.LastAttempt = time.Now().UTC()
	f.LastError = errMsg
	r.st.Feeds[key] = f
	r.write()
}

func (r *Recorder) write() {
	r.st.Updated = time.Now().UTC()
	b, err := json.MarshalIndent(r.st, "", "  ")
	if err != nil {
		return
	}
	_ = atomicWrite(filepath.Join(r.root, "status.json"), b)
}
