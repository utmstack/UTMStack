package feed

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
	"github.com/utmstack/UTMStack/shared/logger"
)

const (
	maxManifestBytes = 2 << 20
	// The official containers are about 85, 22 and 0.3 MiB; these limits
	// leave headroom for their growth while bounding every download.
	maxArtifactBytes   = 256 << 20
	maxGenerationBytes = 512 << 20
	maxManifestFiles   = 128
	// minUpdateRate is the slowest sustained mirror throughput (about 2 Mbit/s)
	// at which a maximum-size generation still fits in updateTimeout. The
	// budget, not a per-request limit, bounds a transfer that keeps progressing.
	minUpdateRate = 256 << 10
	updateTimeout = maxGenerationBytes/minUpdateRate*time.Second + 5*time.Minute
	// defaultStallTimeout bounds waiting for response headers, then for the
	// first body bytes, then each gap between received body bytes, so a dead
	// mirror fails well before updateTimeout.
	defaultStallTimeout = time.Minute
)

var errTransferStalled = errors.New("signature resource transfer stalled")

// GenerationStore commits compiled signatures and anti-rollback state together.
// A rejected generation must leave both the active rules and state untouched.
type GenerationStore interface {
	ActivateGeneration(map[string][]byte, uint64, string) (string, error)
	FeedState() (uint64, string)
	Version() string
}

// preparedGenerationStore consumes decoded native indicators without parsing
// their durable compatibility representation again. Other stores keep the
// original GenerationStore interface and receive identical source bytes.
type preparedGenerationStore interface {
	ActivatePreparedGeneration(map[string][]byte, map[string]*signatures.PreparedHashList, uint64, string) (string, error)
}

// ManifestEnvelope authenticates the exact decoded payload bytes, not a
// reserialized JSON object. Signing keys are provisioned independently of feeds.
type ManifestEnvelope struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

type Manifest struct {
	Version   string         `json:"version"`
	Sequence  uint64         `json:"sequence"`
	ExpiresAt time.Time      `json:"expires_at"`
	Files     []ManifestFile `json:"files"`
}

type ManifestFile struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Format string `json:"format,omitempty"`
}

type Feed struct {
	cfg           config.EDRConfig
	cache         *cache.Cache
	engine        GenerationStore
	client        *http.Client
	now           func() time.Time
	updateMu      sync.Mutex
	mu            sync.Mutex
	lastSuccess   time.Time
	lastError     string
	spool         *event.Spool
	staleNotified bool
	stallTimeout  time.Duration
}

// New retains the source-compatible constructor while requiring an engine for
// actual updates. Missing credentials or an engine are explicit update errors.
// The client has no whole-request Timeout: in Go it also covers reading the
// body, which would abort large artifacts on slow links while bytes still
// arrive. read bounds stalls and updateTimeout bounds the whole update.
func New(cfg config.EDRConfig, c *cache.Cache, engines ...GenerationStore) *Feed {
	f := &Feed{cfg: cfg, cache: c, now: time.Now, stallTimeout: defaultStallTimeout, client: &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("signature redirects are disabled") },
	}}
	if len(engines) == 1 {
		f.engine = engines[0]
	}
	return f
}

func NewNative(cfg config.EDRConfig, c *cache.Cache, e GenerationStore) *Feed { return New(cfg, c, e) }

var engineNameRedactor = strings.NewReplacer(
	"database.clamav.net", "the signature database",
	"ClamAV", "the signature engine", "clamav", "the signature engine",
	"clamd", "the signature engine", "freshclam", "the signature updater",
)

// NativeMirrorURL intentionally does not inherit legacy HTTP or TLS-skip
// exceptions. A pinned manifest key and authenticated HTTPS are both required.
func NativeMirrorURL(cfg config.EDRConfig) (string, error) {
	raw := strings.TrimRight(cfg.SigMirror, "/")
	if raw == "cdn" {
		return "", errors.New("legacy public database format is not supported by native updates")
	}
	if raw == "" {
		if cfg.Server == "" {
			return "", errors.New("signature mirror is not configured")
		}
		raw = "https://" + cfg.Server + ":" + config.MirrorPort + config.MirrorBasePath + "/native"
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" {
		return "", errors.New("signature mirror must be an HTTPS URL without credentials, query or fragment")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func decodeStrict(data []byte, out any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("trailing JSON data")
	}
	return nil
}

// read cancels the request when the headers, or the next body bytes, do not
// arrive within the stall limit. Progress re-arms the limit, so a slow transfer
// is bounded by the caller's update budget rather than by its duration.
func (f *Feed) read(parent context.Context, address string, max int64) ([]byte, error) {
	ctx, cancel := context.WithCancelCause(parent)
	defer cancel(nil)
	stall := f.stallTimeout
	if stall <= 0 {
		stall = defaultStallTimeout
	}
	watchdog := time.AfterFunc(stall, func() { cancel(errTransferStalled) })
	defer watchdog.Stop()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, errors.New("invalid signature resource URL")
	}
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("User-Agent", "UTMStack-EDR-Native/1")
	resp, err := f.client.Do(req)
	if err != nil {
		if cause := context.Cause(ctx); cause != nil {
			return nil, cause
		}
		return nil, fmt.Errorf("signature resource request failed: %s", requestFailureCause(err))
	}
	defer resp.Body.Close()
	// The headers arrived; the wait for the first body bytes is a new gap.
	watchdog.Reset(stall)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signature resource HTTP status %d", resp.StatusCode)
	}
	if resp.ContentLength > max {
		return nil, errors.New("signature resource exceeds size limit")
	}
	b, err := readBounded(progressReader{r: resp.Body, progress: func() { watchdog.Reset(stall) }}, max)
	// A cancelled transport can finish its body with EOF, making ReadAll
	// return nil. Even checksum-correct bytes must not authorize activation
	// after the request's stall limit or the caller's cancellation.
	if cause := context.Cause(ctx); cause != nil {
		return nil, cause
	}
	return b, err
}

// progressReader reports every read that returned data.
type progressReader struct {
	r        io.Reader
	progress func()
}

func (p progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.progress()
	}
	return n, err
}

func readBounded(r io.Reader, max int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, errors.New("signature resource exceeds size limit")
	}
	return b, nil
}

func artifactURL(base string, item ManifestFile) (string, error) {
	if err := signatures.ValidateFileName(item.Name); err != nil {
		return "", err
	}
	if item.Name == "" || len(item.Name) > 128 || path.Base(item.Name) != item.Name || item.Name == "." || item.Name == ".." || strings.ContainsAny(item.Name, "\\:\x00\r\n") {
		return "", errors.New("invalid signature artifact name")
	}
	for _, r := range item.Name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.') {
			return "", errors.New("invalid signature artifact name")
		}
	}
	u, err := url.Parse(item.URL)
	if err != nil || u.IsAbs() || u.Host != "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || u.Path == "" || strings.HasPrefix(u.Path, "/") || strings.ContainsAny(u.Path, "\\\x00") {
		return "", errors.New("artifact URL must be a relative path within the signature mirror")
	}
	for _, p := range strings.Split(u.Path, "/") {
		if p == ".." || p == "." || p == "" {
			return "", errors.New("invalid artifact URL path")
		}
	}
	root, _ := url.Parse(base + "/")
	return root.ResolveReference(u).String(), nil
}

func (f *Feed) verifyManifest(b []byte, key ed25519.PublicKey) (Manifest, string, error) {
	var envelope ManifestEnvelope
	var m Manifest
	if err := decodeStrict(b, &envelope); err != nil {
		return m, "", errors.New("invalid signature manifest envelope")
	}
	payload, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		return m, "", errors.New("invalid signature manifest payload encoding")
	}
	sig, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil || !ed25519.Verify(key, payload, sig) {
		return m, "", errors.New("signature manifest authentication failed")
	}
	if err = decodeStrict(payload, &m); err != nil {
		return m, "", errors.New("invalid signature manifest payload")
	}
	if m.Sequence == 0 || strings.TrimSpace(m.Version) == "" || len(m.Version) > 128 || len(m.Files) == 0 || len(m.Files) > maxManifestFiles {
		return m, "", errors.New("invalid signature manifest metadata")
	}
	if !f.now().Before(m.ExpiresAt) {
		return m, "", errors.New("signature manifest has expired")
	}
	digest := sha256.Sum256(payload)
	hash := hex.EncodeToString(digest[:])
	sequence, previous := f.engine.FeedState()
	if m.Sequence < sequence || m.Sequence == sequence && hash != previous {
		return m, "", errors.New("signature manifest rollback or sequence reuse rejected")
	}
	return m, hash, nil
}

func (f *Feed) updateOnce() (string, error) { return f.UpdateOnce(context.Background()) }

// UpdateOnce downloads and validates an entire signed generation before the
// engine compiles and atomically publishes it. No unsigned fallback is used.
func (f *Feed) UpdateOnce(parent context.Context) (string, error) {
	f.updateMu.Lock()
	defer f.updateMu.Unlock()
	version, err := f.updateWithContext(parent)
	f.mu.Lock()
	f.lastError = ""
	if err != nil {
		f.lastError = engineNameRedactor.Replace(err.Error())
	}
	f.mu.Unlock()
	return version, err
}

func (f *Feed) updateWithContext(parent context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(parent, updateTimeout)
	defer cancel()
	if f.engine == nil {
		return "", errors.New("native signature engine is not configured")
	}
	key, err := base64.StdEncoding.DecodeString(f.cfg.SignaturePublicKey)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return "", errors.New("a pinned Ed25519 signature public key is required")
	}
	base, err := NativeMirrorURL(f.cfg)
	if err != nil {
		return "", err
	}
	b, err := f.read(ctx, base+"/manifest.json", maxManifestBytes)
	if err != nil {
		return "", err
	}
	m, manifestHash, err := f.verifyManifest(b, ed25519.PublicKey(key))
	if err != nil {
		return "", err
	}
	// The mirror still serves the active generation. Its exact signed payload
	// was fully verified and compiled when activated, so downloading,
	// compiling and durably rewriting it again would change nothing.
	if sequence, previous := f.engine.FeedState(); m.Sequence == sequence && manifestHash == previous {
		if version := f.engine.Version(); version != "" {
			return f.recordSuccess(version)
		}
	}
	files := make(map[string][]byte, len(m.Files))
	prepared := make(map[string]*signatures.PreparedHashList)
	native, usePrepared := f.engine.(preparedGenerationStore)
	names := make(map[string]bool, len(m.Files))
	var total int
	for _, item := range m.Files {
		nameKey := strings.ToLower(item.Name)
		if names[nameKey] {
			return "", errors.New("duplicate signature artifact name")
		}
		names[nameKey] = true
		address, err := artifactURL(base, item)
		if err != nil {
			return "", err
		}
		want, err := hex.DecodeString(item.SHA256)
		if err != nil || len(want) != sha256.Size {
			return "", errors.New("invalid signature artifact checksum")
		}
		data, err := f.read(ctx, address, maxArtifactBytes)
		if err != nil {
			return "", err
		}
		total += len(data)
		if total > maxGenerationBytes {
			return "", errors.New("signature generation exceeds size limit")
		}
		got := sha256.Sum256(data)
		if !bytes.Equal(want, got[:]) {
			return "", errors.New("signature artifact checksum mismatch")
		}
		converted, nativeHashes, err := normalizeArtifact(item, data, usePrepared)
		if err != nil {
			return "", fmt.Errorf("invalid signature artifact: %w", err)
		}
		if len(converted) > len(data) {
			total += len(converted) - len(data)
		}
		if total > maxGenerationBytes {
			return "", errors.New("signature generation exceeds size limit")
		}
		files[item.Name] = converted
		if nativeHashes != nil {
			prepared[item.Name] = nativeHashes
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if !f.now().Before(m.ExpiresAt) {
		return "", errors.New("signature manifest expired during download")
	}
	var version string
	if usePrepared && len(prepared) > 0 {
		version, err = native.ActivatePreparedGeneration(files, prepared, m.Sequence, manifestHash)
	} else {
		version, err = f.engine.ActivateGeneration(files, m.Sequence, manifestHash)
	}
	if err != nil {
		return "", fmt.Errorf("signature generation rejected: %w", err)
	}
	return f.recordSuccess(version)
}

// recordSuccess invalidates clean verdicts from other generations, so a failed
// invalidation is retried by the next cycle even when nothing changed, and then
// records update health.
func (f *Feed) recordSuccess(version string) (string, error) {
	if f.cache != nil {
		if _, err := f.cache.MarkStaleBySigDB(version); err != nil {
			return version, err
		}
	}
	f.mu.Lock()
	f.lastSuccess = f.now()
	f.staleNotified = false
	f.mu.Unlock()
	return version, nil
}

func nextDelay(err error, interval time.Duration) time.Duration {
	const retry = 15 * time.Minute
	if err != nil && retry < interval {
		return retry
	}
	return interval
}

func (f *Feed) Run(ctx context.Context) {
	interval := time.Duration(f.cfg.SigUpdateHours) * time.Hour
	if interval <= 0 {
		interval = 4 * time.Hour
	}
	for {
		_, err := f.UpdateOnce(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			logger.Error("UTMStack EDR: signature update failed: %s", engineNameRedactor.Replace(err.Error()))
		}
		f.noteHealth(interval)
		timer := time.NewTimer(nextDelay(err, interval))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
