package feed

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/shared/logger"
)

// Signature update states, reported as signature_update_state in status.json.
const (
	UpdatePending       = "pending"        // no refresh has finished yet
	UpdateCurrent       = "current"        // the latest refresh succeeded
	UpdateRefreshFailed = "refresh_failed" // the latest refresh failed; a verified generation still serves
	UpdateStale         = "stale"          // no successful refresh for three update intervals
	UpdateUnavailable   = "unavailable"    // the refresh failed and no verified generation serves
)

// signatureFeedState keys the last successful refresh in edr.db, so the
// staleness baseline survives a service restart.
const signatureFeedState = "signatures"

// SignatureSource names the effective signature-update source for status and
// telemetry. Branded: never names the engine or the public database vendor.
func SignatureSource(cfg config.EDRConfig) string {
	if _, err := NativeMirrorURL(cfg); err == nil {
		return "utmstack-mirror"
	}
	return "unconfigured"
}

func (f *Feed) SetSpool(sp *event.Spool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.spool = sp
}

func (f *Feed) setLastSuccess(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastSuccess = t
}

// LastSuccess returns the time of the last successful signature refresh: this
// run's, or the baseline RestoreHealth recovered from before the restart (zero
// if none is known). Safe for cross-goroutine status reads.
func (f *Feed) LastSuccess() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastSuccess
}

func (f *Feed) LastError() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastError
}

// Stale reports whether the last success is older than max. A feed that has
// never succeeded has no baseline and is not reported stale.
func (f *Feed) Stale(now time.Time, max time.Duration) bool {
	ls := f.LastSuccess()
	if ls.IsZero() {
		return false
	}
	return now.Sub(ls) > max
}

// UpdateState reports the signature update state (one of the Update*
// constants) for status.json.
func (f *Feed) UpdateState() string {
	active := f.generationActive()
	f.mu.Lock()
	defer f.mu.Unlock()
	return classifyUpdate(f.clock(), f.lastSuccess, f.lastError, active, updateInterval(f.cfg.SigUpdateHours))
}

// classifyUpdate derives the update state. last is the latest known successful
// refresh, lastErr the error of the latest attempt, and active whether a
// verified signed generation is serving scans.
func classifyUpdate(now, last time.Time, lastErr string, active bool, interval time.Duration) string {
	switch {
	case lastErr != "" && !active:
		return UpdateUnavailable
	case !last.IsZero() && now.Sub(last) > 3*interval:
		return UpdateStale
	case lastErr != "":
		return UpdateRefreshFailed
	case last.IsZero():
		return UpdatePending
	}
	return UpdateCurrent
}

// updateInterval is Run's refresh interval for the configured hours.
func updateInterval(hours int) time.Duration {
	if hours <= 0 {
		return 4 * time.Hour
	}
	return time.Duration(hours) * time.Hour
}

func (f *Feed) clock() time.Time {
	if f.now == nil {
		return time.Now()
	}
	return f.now()
}

// generationActive reports whether the engine serves a verified signed
// generation (loose administrator files alone have no feed sequence).
func (f *Feed) generationActive() bool {
	if f.engine == nil {
		return false
	}
	sequence, _ := f.engine.FeedState()
	return sequence > 0
}

// RestoreHealth seeds the refresh baseline that a restart would otherwise
// lose: the last successful refresh recorded in edr.db or, when later, the
// activation time of the verified generation that is serving (an activation is
// itself a successful refresh). Without it the first failed refresh after a
// restart reported updates as unavailable while that generation kept serving.
// dbDir is the engine's signature directory. Call it before Run.
func (f *Feed) RestoreHealth(dbDir string) {
	var last time.Time
	if f.cache != nil {
		if st, err := f.cache.GetFeedState(signatureFeedState); err == nil {
			last = st.LastFullSync
		}
	}
	if f.engine != nil {
		sequence, hash := f.engine.FeedState()
		if t := GenerationActivatedAt(dbDir, sequence, hash); t.After(last) {
			last = t
		}
	}
	// A clock that stepped back must not put the baseline in the future.
	if now := f.clock(); last.After(now) {
		last = now
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if last.After(f.lastSuccess) {
		f.lastSuccess = last
	}
}

// GenerationActivatedAt returns when the verified generation that the engine
// reports as serving (its FeedState sequence and manifest hash) was activated
// in dbDir: the modification time of the CURRENT pointer, which the engine
// replaces only when it activates a generation, provided the metadata CURRENT
// names still describes that generation. Zero when none matches.
func GenerationActivatedAt(dbDir string, sequence uint64, manifestHash string) time.Time {
	if sequence == 0 {
		return time.Time{}
	}
	pointer := filepath.Join(dbDir, "CURRENT")
	info, err := os.Stat(pointer)
	if err != nil {
		return time.Time{}
	}
	id, err := readSmallFile(pointer, 256)
	if err != nil {
		return time.Time{}
	}
	name := strings.TrimSpace(string(id))
	if b, err := hex.DecodeString(name); err != nil || len(b) != 32 {
		return time.Time{}
	}
	b, err := readSmallFile(filepath.Join(dbDir, "generations", name, "metadata.json"), 64<<10)
	if err != nil {
		return time.Time{}
	}
	var meta struct {
		Sequence     uint64 `json:"sequence"`
		ManifestHash string `json:"manifest_hash"`
	}
	if json.Unmarshal(b, &meta) != nil || meta.Sequence != sequence || meta.ManifestHash != manifestHash {
		return time.Time{}
	}
	return info.ModTime()
}

func readSmallFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("file exceeds size limit")
	}
	return b, nil
}

// noteHealth records a successful refresh in edr.db and emits one branded
// health event when updates become stale, or unavailable with no verified
// generation serving; the next success re-arms it. A failed refresh while a
// verified generation serves within the staleness bound is reported only in
// status.json (refresh_failed), so a restart before the network is up raises
// no alert.
func (f *Feed) noteHealth(interval time.Duration) {
	active := f.generationActive()
	f.mu.Lock()
	state := classifyUpdate(f.clock(), f.lastSuccess, f.lastError, active, interval)
	succeeded, last, sp := f.lastError == "", f.lastSuccess, f.spool
	notify := sp != nil && !f.staleNotified && (state == UpdateStale || state == UpdateUnavailable)
	if notify {
		f.staleNotified = true
	}
	f.mu.Unlock()
	if succeeded && !last.IsZero() && f.cache != nil {
		if err := f.cache.SaveFeedState(cache.FeedState{Feed: signatureFeedState, LastFullSync: last}); err != nil {
			logger.Error("UTMStack EDR: recording the signature refresh time: %v", err)
		}
	}
	if !notify {
		return
	}
	ev := event.Event{
		Source:    event.SourceEngine,
		Action:    event.ActionHealth,
		Signature: "signature_updates_stale",
		Severity:  "warning",
	}
	if state == UpdateUnavailable {
		ev.Signature = "signature_updates_unavailable"
	}
	if js, err := ev.ToJSON(); err == nil {
		_ = sp.Append(js)
	}
}

// Winsock error numbers (winerror.h). Windows reports socket failures with
// these, and Go does not map them to the portable syscall.E* constants.
const (
	wsaeNetDown     syscall.Errno = 10050
	wsaeNetUnreach  syscall.Errno = 10051
	wsaeConnAborted syscall.Errno = 10053
	wsaeConnReset   syscall.Errno = 10054
	wsaeTimedOut    syscall.Errno = 10060
	wsaeConnRefused syscall.Errno = 10061
	wsaeHostUnreach syscall.Errno = 10065
)

// requestFailureCause names why a signature request got no HTTP response as a
// short fixed category, never the raw error text (which can carry host names
// or vendor strings), so an operator knows whether to trust the mirror
// certificate, start the mirror or fix the network path.
func requestFailureCause(err error) string {
	var (
		unknownAuthority x509.UnknownAuthorityError
		hostname         x509.HostnameError
		invalid          x509.CertificateInvalidError
		verification     *tls.CertificateVerificationError
		recordHeader     tls.RecordHeaderError
		opErr            *net.OpError
		dnsErr           *net.DNSError
		errno            syscall.Errno
		netErr           net.Error
	)
	switch {
	case errors.As(err, &unknownAuthority):
		return "untrusted certificate"
	case errors.As(err, &hostname):
		return "certificate name mismatch"
	case errors.As(err, &invalid):
		if invalid.Reason == x509.Expired {
			return "certificate expired or not yet valid"
		}
		return "certificate not valid for this server"
	case errors.As(err, &verification):
		return "certificate verification failed"
	case errors.Is(err, http.ErrSchemeMismatch), errors.As(err, &recordHeader):
		return "server does not use TLS"
	case errors.As(err, &opErr) && opErr.Op == "remote error":
		return "TLS handshake rejected"
	case errors.As(err, &dnsErr):
		return "name resolution failed"
	case errors.Is(err, context.Canceled):
		return "canceled"
	}
	if errors.As(err, &errno) {
		switch errno {
		case syscall.ECONNREFUSED, wsaeConnRefused:
			return "connection refused"
		case syscall.ECONNRESET, syscall.ECONNABORTED, wsaeConnReset, wsaeConnAborted:
			return "connection reset"
		case syscall.ETIMEDOUT, wsaeTimedOut:
			return "timeout"
		case syscall.ENETUNREACH, syscall.EHOSTUNREACH, syscall.ENETDOWN, wsaeNetUnreach, wsaeHostUnreach, wsaeNetDown:
			return "network unreachable"
		}
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "connection closed"
	}
	return "network error"
}
